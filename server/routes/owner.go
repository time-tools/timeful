package routes

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"timeful/server/errs"
	pgstore "timeful/server/postgres"
)

func ownerCookieName(eventID string) string { return "timeful_owner_" + eventID }

func issueOwnerToken(ctx context.Context, repo *pgstore.Repository, event *pgstore.Event) (string, error) {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(secret[:])
	hash := sha256.Sum256([]byte(token))
	if err := repo.SetEventOwnerToken(ctx, event.ID, hash[:]); err != nil {
		return "", err
	}
	return token, nil
}

func setOwnerCookie(c *gin.Context, eventID, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: ownerCookieName(eventID), Value: token, Path: "/api",
		MaxAge: 34560000, HttpOnly: true, SameSite: http.SameSiteLaxMode,
		Secure: c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https",
	})
}

// authorizeOwner runs under the event row lock, serializing takeover
// with protected mutations. Token proof never changes response ownership.
func authorizeOwner(c *gin.Context, repo *pgstore.Repository, event *pgstore.Event) (bool, error) {
	if event.IsDeleted {
		return false, pgx.ErrNoRows
	}
	ctx := c.Request.Context()
	platformIdentityID, _ := sessions.Default(c).Get("userId").(string)
	token, err := c.Cookie(ownerCookieName(event.ShortID))
	hash := sha256.Sum256([]byte(token))
	if err == nil && token != "" && subtle.ConstantTimeCompare(hash[:], event.OwnerEditTokenHash) == 1 {
		if platformIdentityID != "" {
			platform, err := resolveSessionPlatformIdentity(ctx, repo, platformIdentityID)
			if err != nil {
				return false, err
			}
			if platform != nil && (event.OwnerPlatformIdentityID == nil || *event.OwnerPlatformIdentityID != platform.ID) {
				if err := repo.AssociateEventOwner(ctx, event.ID, platform.ID); err != nil {
					return false, err
				}
				event.OwnerPlatformIdentityID = &platform.ID
			}
		}
		return true, nil
	}
	if platformIdentityID != "" {
		owned, err := repo.EventOwnerBelongsToAccount(ctx, event.ID, platformIdentityID)
		if err != nil || owned {
			return owned, err
		}
	}
	grantVisitor, grant, err := provenGrant(c, repo, event)
	if err != nil {
		return false, err
	}
	if grant != nil && grant.GrantsOwner && event.OwnerEventVisitorIdentityID != nil && grantVisitor.ID == *event.OwnerEventVisitorIdentityID {
		return true, nil
	}
	// Retain validation of grants stored in the foundation cookie slot.
	// A base EVCC, even the creator's, never grants Event Owner powers.
	cookie, err := c.Cookie(credentialCookieName(event.ShortID))
	if err != nil {
		return false, nil
	}
	parts := strings.Split(cookie, ".")
	if len(parts) != 3 {
		return false, nil
	}
	visitor, err := repo.GetEventVisitorIdentity(ctx, event.ID, parts[0])
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	credential, err := provenCredential(c, repo, visitor, event.ShortID)
	if err != nil {
		return false, err
	}
	return credential != nil && credential.Kind == pgstore.CredentialKindGranted && credential.GrantsOwner &&
		event.OwnerEventVisitorIdentityID != nil && visitor.ID == *event.OwnerEventVisitorIdentityID, nil
}

func resolveOwner(c *gin.Context, repo *pgstore.Repository, event *pgstore.Event) (bool, error) {
	var authorized bool
	err := repo.WithTransaction(c.Request.Context(), func(ctx context.Context, tx *pgstore.Repository) error {
		locked, err := tx.LockEvent(ctx, event.ID)
		if err != nil {
			return err
		}
		authorized, err = authorizeOwner(c, tx, locked)
		if err == nil {
			*event = *locked
		}
		return err
	})
	return authorized, err
}

func writableEvent(event *pgstore.Event) error {
	if event.IsDeleted {
		return pgx.ErrNoRows
	}
	if event.IsArchived {
		return guestForbidden{errs.EventArchived}
	}
	return nil
}

func ownerMutation(c *gin.Context, allowArchived bool, mutate func(context.Context, *pgstore.Repository, *pgstore.Event) error) bool {
	repo := defaultRepository(c)
	if repo == nil {
		return false
	}
	event := loadEvent(c, repo)
	if event == nil {
		return false
	}
	err := repo.WithTransaction(c.Request.Context(), func(ctx context.Context, tx *pgstore.Repository) error {
		locked, err := tx.LockEvent(ctx, event.ID)
		if err != nil {
			return err
		}
		authorized, err := authorizeOwner(c, tx, locked)
		if err != nil {
			return err
		}
		if !authorized {
			return guestForbidden{errs.EventOwnerCredentialRequired}
		}
		if !allowArchived {
			if err := writableEvent(locked); err != nil {
				return err
			}
		}
		return mutate(ctx, tx, locked)
	})
	if err != nil {
		mutationError(c, err)
		return false
	}
	c.Status(http.StatusOK)
	return true
}

// @Summary Archive an event
// @Description Requires the same owner credentials as settings edits; archive makes the event read-only and unarchive restores mutations.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path string true "Event ID"
// @Param payload body object{archive=bool} true "Archive status"
// @Success 200
// @Failure 403 {object} responses.Error "Owner authority required or event archived"
// @Failure 404 {object} responses.Error "Event not found"
// @Router /events/{eventId}/archive [post]
func archiveEvent(c *gin.Context) {
	var input struct {
		Archive *bool `json:"archive" binding:"required"`
	}
	if err := c.BindJSON(&input); err != nil {
		return
	}
	ownerMutation(c, true, func(ctx context.Context, tx *pgstore.Repository, event *pgstore.Event) error {
		return tx.SetEventArchived(ctx, event.ID, *input.Archive)
	})
}

// @Summary Deletes an event based on its id
// @Description Requires the same owner credentials as settings edits; deleted events and responses stop resolving.
// @Tags events
// @Produce json
// @Param eventId path string true "Event ID"
// @Success 200
// @Failure 403 {object} responses.Error "Owner authority required or event archived"
// @Failure 404 {object} responses.Error "Event not found"
// @Router /events/{eventId} [delete]
func deleteEvent(c *gin.Context) {
	ownerMutation(c, true, func(ctx context.Context, tx *pgstore.Repository, event *pgstore.Event) error {
		return tx.SetEventDeleted(ctx, event.ID, true)
	})
}
