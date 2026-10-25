package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	pgstore "timeful/server/postgres"
)

// FolderResponse mirrors the documented folder wire shape while carrying
// canonical public event identifiers in eventIds.
//
// The response structs, the request-body structs, and the identity helper stay
// handwritten: their tags define the wire shape, the helper takes a comma-ok
// assertion, and folders.gala builds them through the constructors below.
type FolderResponse struct {
	Id        string   `json:"_id"`
	UserId    string   `json:"userId"`
	Name      string   `json:"name,omitempty"`
	Color     *string  `json:"color,omitempty"`
	IsDeleted *bool    `json:"isDeleted,omitempty"`
	EventIds  []string `json:"eventIds"`
}

type CreateFolderResponse struct {
	Id string `json:"id"`
}

// folderCreateBody carries the create-folder request body. Its tags define the
// request wire shape, and folders.gala constructs it through
// newFolderCreateBody instead of naming the local type.
type folderCreateBody struct {
	Name  string  `json:"name" binding:"required"`
	Color *string `json:"color"`
}

func newFolderCreateBody() *folderCreateBody { return &folderCreateBody{} }

// folderUpdateBody carries the update-folder request body.
type folderUpdateBody struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

func newFolderUpdateBody() *folderUpdateBody { return &folderUpdateBody{} }

func folderPlatformIdentityID(c *gin.Context) (string, bool) {
	platformIdentityID, ok := sessions.Default(c).Get("userId").(string)
	if !ok || platformIdentityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return "", false
	}
	return platformIdentityID, true
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

func folderResponsesFrom(folders []pgstore.Folder) []FolderResponse {
	result := make([]FolderResponse, 0, len(folders))
	for _, folder := range folders {
		result = append(result, folderResponseFrom(folder))
	}
	return result
}

func newCreateFolderResponse(id string) CreateFolderResponse {
	return CreateFolderResponse{Id: id}
}
