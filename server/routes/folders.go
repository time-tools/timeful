package routes

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"timeful/server/middleware"
	pgstore "timeful/server/postgres"
)

var folderIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validFolderID(id string) bool { return folderIDPattern.MatchString(id) }

// FolderResponse mirrors the documented folder wire shape while carrying
// canonical public event identifiers in eventIds.
type FolderResponse struct {
	Id        string   `json:"_id"`
	UserId    string   `json:"userId"`
	Name      string   `json:"name,omitempty"`
	Color     *string  `json:"color,omitempty"`
	IsDeleted *bool    `json:"isDeleted,omitempty"`
	EventIds  []string `json:"eventIds"`
}

func InitFolders(router *gin.RouterGroup) {
	folderRouter := router.Group("/user/folders")
	folderRouter.Use(middleware.AuthRequired())

	folderRouter.GET("", GetAllFolders)
	folderRouter.POST("", CreateFolder)
	folderRouter.GET("/:folderId", GetFolder)
	folderRouter.PATCH("/:folderId", UpdateFolder)
	folderRouter.DELETE("/:folderId", DeleteFolder)
}

func folderPlatformIdentityID(c *gin.Context) (string, bool) {
	userIdString, ok := sessions.Default(c).Get("userId").(string)
	if !ok || userIdString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return "", false
	}
	return userIdString, true
}

// canonicalFolderEventIDs renders each member as the canonical public event
// identifier the frontend uses to match an event. Members without a canonical
// identifier are skipped.
func canonicalFolderEventIDs(members []pgstore.FolderMember) []string {
	ids := make([]string, 0, len(members))
	for _, member := range members {
		if member.EventShortID != nil && *member.EventShortID != "" {
			ids = append(ids, *member.EventShortID)
		}
	}
	return ids
}

func folderResponseFrom(folder pgstore.Folder) FolderResponse {
	return FolderResponse{
		Id:        folder.ID,
		UserId:    folder.PlatformIdentityID,
		Name:      folder.Name,
		Color:     folder.Color,
		IsDeleted: folder.IsDeleted,
		EventIds:  canonicalFolderEventIDs(folder.Members),
	}
}

// @Summary Get all folders
// @Tags folders
// @Produce json
// @Success 200 {array} FolderResponse "A list of all folders for the user"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Failed to get folders"
// @Router /user/folders [get]
func GetAllFolders(c *gin.Context) {
	platformIdentityID, ok := folderPlatformIdentityID(c)
	if !ok {
		return
	}
	repository := defaultRepository(c)
	if repository == nil {
		return
	}

	folders, err := repository.ListFolders(c.Request.Context(), platformIdentityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get folders"})
		return
	}

	result := make([]FolderResponse, 0, len(folders))
	for _, folder := range folders {
		result = append(result, folderResponseFrom(folder))
	}
	c.JSON(http.StatusOK, result)
}

// @Summary Get a folder by its ID and its contents
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} FolderResponse "The folder object with events"
// @Failure 400 {object} map[string]string "Invalid user ID or folder ID"
// @Failure 404 {object} map[string]string "Folder not found"
// @Failure 500 {object} map[string]string "Failed to get events in folder"
// @Router /user/folders/{folderId} [get]
func GetFolder(c *gin.Context) {
	platformIdentityID, ok := folderPlatformIdentityID(c)
	if !ok {
		return
	}
	repository := defaultRepository(c)
	if repository == nil {
		return
	}

	folder, err := repository.GetFolderByID(c.Request.Context(), c.Param("folderId"), platformIdentityID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events in folder"})
		return
	}

	c.JSON(http.StatusOK, folderResponseFrom(*folder))
}

type CreateFolderResponse struct {
	Id string `json:"id"`
}

// @Summary Create a new folder
// @Tags folders
// @Accept json
// @Produce json
// @Param payload body object{name=string,color=string} true "Folder name and optional color"
// @Success 201 {object} CreateFolderResponse "The ID of the created folder"
// @Failure 400 {object} map[string]string "Invalid user ID or request body"
// @Failure 500 {object} map[string]string "Failed to create folder"
// @Router /user/folders [post]
func CreateFolder(c *gin.Context) {
	var body struct {
		Name  string  `json:"name" binding:"required"`
		Color *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platformIdentityID, ok := folderPlatformIdentityID(c)
	if !ok {
		return
	}
	repository := defaultRepository(c)
	if repository == nil {
		return
	}

	folder := pgstore.Folder{
		PlatformIdentityID: platformIdentityID,
		Name:               body.Name,
		Color:              body.Color,
	}
	if err := repository.CreateFolder(c.Request.Context(), &folder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, CreateFolderResponse{Id: folder.ID})
}

// @Summary Update a folder's name or color
// @Tags folders
// @Accept json
// @Produce json
// @Param folderId path string true "Folder ID"
// @Param payload body object{name=string,color=string} true "New folder name and/or color"
// @Success 200
// @Failure 400 {object} map[string]string "Invalid user ID or folder ID"
// @Failure 404 {object} map[string]string "Folder not found"
// @Failure 500 {object} map[string]string "Failed to update folder"
// @Router /user/folders/{folderId} [patch]
func UpdateFolder(c *gin.Context) {
	var body struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platformIdentityID, ok := folderPlatformIdentityID(c)
	if !ok {
		return
	}
	repository := defaultRepository(c)
	if repository == nil {
		return
	}

	err := repository.UpdateFolder(c.Request.Context(), c.Param("folderId"), platformIdentityID, body.Name, body.Color)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.Status(http.StatusOK)
}

// @Summary Delete a folder
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200
// @Failure 400 {object} map[string]string "Invalid user ID or folder ID"
// @Failure 404 {object} map[string]string "Folder not found"
// @Failure 500 {object} map[string]string "Failed to delete folder"
// @Router /user/folders/{folderId} [delete]
func DeleteFolder(c *gin.Context) {
	platformIdentityID, ok := folderPlatformIdentityID(c)
	if !ok {
		return
	}
	repository := defaultRepository(c)
	if repository == nil {
		return
	}

	err := repository.DeleteFolder(c.Request.Context(), c.Param("folderId"), platformIdentityID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	c.Status(http.StatusOK)
}
