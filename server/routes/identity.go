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
	pgstore "timeful/server/postgres"
)

type visitor struct {
	identity           *pgstore.EventVisitorIdentity
	platformIdentityID string
	authorized         bool
	granted            bool
	owner              bool
	grantedVisitorID   string
}

// resolveSessionPlatformIdentity resolves the platform identity UUID a session
// carries. An empty, retired, or deleted value reports no identity, so a stale
// session never adopts an account and never reaches a uuid column as an invalid literal.
func resolveSessionPlatformIdentity(ctx context.Context, repo *pgstore.Repository, platformIdentityID string) (*pgstore.PlatformIdentity, error) {
	if platformIdentityID == "" {
		return nil, nil
	}
	platform, err := repo.GetPlatformIdentity(ctx, platformIdentityID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return platform, nil
}

// @Summary Associate browser Event Visitor Identities with the authenticated account
// @Description Source EVCC proof associates response identity; independent Event Owner Edit Token proof associates or moves event ownership without moving responses. Granted EVCC association awaits the transfer confirmation flow.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body object{identities=[]object{eventId=string,eventVisitorId=string}} true "Browser-local public identities; matching HttpOnly credentials are required"
// @Success 200
// @Failure 401
// @Router /auth/visitor-identities [post]
func associateVisitorIdentities(c *gin.Context) {
	platformIdentityID, ok := sessions.Default(c).Get("userId").(string)
	if !ok || platformIdentityID == "" {
		c.Status(http.StatusUnauthorized)
		return
	}
	var input struct {
		Identities []struct {
			EventID        string `json:"eventId"`
			EventVisitorID string `json:"eventVisitorId"`
		} `json:"identities"`
	}
	if err := c.BindJSON(&input); err != nil {
		return
	}
	if len(input.Identities) > 200 {
		c.Status(http.StatusBadRequest)
		return
	}
	repo := defaultRepository(c)
	if repo == nil {
		return
	}
	err := repo.WithTransaction(c.Request.Context(), func(ctx context.Context, tx *pgstore.Repository) error {
		for _, item := range input.Identities {
			event, err := tx.GetEventByShortID(ctx, item.EventID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			event, err = tx.LockEvent(ctx, event.ID)
			if err != nil {
				return err
			}
			if event.IsDeleted {
				continue
			}
			if _, err := authorizeOwner(c, tx, event); err != nil {
				return err
			}
			visitor, err := tx.GetEventVisitorIdentity(ctx, event.ID, item.EventVisitorID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			credential, err := provenCredential(c, tx, visitor, event.ShortID)
			if err != nil {
				return err
			}
			if credential == nil || credential.Kind != pgstore.CredentialKindBase || visitor.PlatformIdentityID != nil {
				continue
			}
			platform, err := resolveSessionPlatformIdentity(ctx, tx, platformIdentityID)
			if err != nil {
				return err
			}
			if platform == nil {
				continue
			}
			if err := tx.AssociateEventVisitorIdentity(ctx, visitor.ID, platform.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		mutationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func credentialCookieName(eventID string) string { return "timeful_evcc_" + eventID }

func issueCredential(ctx context.Context, repo *pgstore.Repository, visitorID string) (string, error) {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", err
	}
	value := base64.RawURLEncoding.EncodeToString(secret[:])
	hash := sha256.Sum256([]byte(value))
	credential := &pgstore.EventVisitorCredential{EventVisitorIdentityID: visitorID, CredentialHash: hash[:]}
	if err := repo.CreateEventVisitorCredential(ctx, credential); err != nil {
		return "", err
	}
	return credential.ID + "." + value, nil
}

func setCredentialCookie(c *gin.Context, eventID, publicID, credential string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: credentialCookieName(eventID), Value: publicID + "." + credential,
		Path: "/api", MaxAge: 34560000, HttpOnly: true, SameSite: http.SameSiteLaxMode,
		Secure: c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https",
	})
}

func provenCredential(c *gin.Context, repo *pgstore.Repository, visitor *pgstore.EventVisitorIdentity, eventID string) (*pgstore.EventVisitorCredential, error) {
	return provenCredentialCookie(c, repo, visitor, credentialCookieName(eventID))
}

func provenCredentialCookie(c *gin.Context, repo *pgstore.Repository, visitor *pgstore.EventVisitorIdentity, name string) (*pgstore.EventVisitorCredential, error) {
	cookie, err := c.Cookie(name)
	if err != nil {
		return nil, nil
	}
	parts := strings.Split(cookie, ".")
	if len(parts) != 3 || parts[0] != visitor.PublicID {
		return nil, nil
	}
	credential, err := repo.GetEventVisitorCredential(c.Request.Context(), visitor.ID, parts[1])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(parts[2]))
	if subtle.ConstantTimeCompare(hash[:], credential.CredentialHash) != 1 || credential.RevokedAt != nil {
		return nil, nil
	}
	return credential, nil
}

// The public identifier selects a visitor but never proves control of it.
func resolveVisitor(c *gin.Context, repo *pgstore.Repository, event *pgstore.Event) (*visitor, error) {
	owner, err := resolveOwner(c, repo, event)
	if err != nil {
		return nil, err
	}
	platformIdentityID, _ := sessions.Default(c).Get("userId").(string)
	// Normalize the session value once: only a live canonical platform identity
	// stays visible to callers, so a stale or retired session behaves as
	// anonymous and never reaches a uuid column as an invalid literal.
	if platformIdentityID != "" {
		platform, err := resolveSessionPlatformIdentity(c.Request.Context(), repo, platformIdentityID)
		if err != nil {
			return nil, err
		}
		if platform == nil {
			platformIdentityID = ""
		} else {
			platformIdentityID = platform.ID
		}
	}
	publicID := c.Query("eventVisitorId")
	if publicID == "" {
		if cookie, err := c.Cookie(credentialCookieName(event.ShortID)); err == nil {
			publicID = strings.Split(cookie, ".")[0]
		}
	}
	result := &visitor{platformIdentityID: platformIdentityID, owner: owner}
	grantVisitor, _, err := provenGrant(c, repo, event)
	if err != nil {
		return nil, err
	}
	if grantVisitor != nil {
		result.grantedVisitorID = grantVisitor.ID
	}
	if publicID != "" {
		visitor, err := repo.GetEventVisitorIdentity(c.Request.Context(), event.ID, publicID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			credential, err := provenCredential(c, repo, visitor, event.ShortID)
			if err != nil {
				return nil, err
			}
			valid := credential != nil
			result.granted = valid && credential.Kind == pgstore.CredentialKindGranted
			account := false
			if platformIdentityID != "" {
				account, err = repo.VisitorBelongsToAccount(c.Request.Context(), visitor.ID, platformIdentityID)
				if err != nil {
					return nil, err
				}
			}
			result.identity = visitor
			if valid || account {
				result.authorized = true
				if account && !valid {
					credential, err := issueCredential(c.Request.Context(), repo, visitor.ID)
					if err != nil {
						return nil, err
					}
					setCredentialCookie(c, event.ShortID, visitor.PublicID, credential)
				}
			}
		}
	}
	if result.identity == nil {
		var credential string
		err := repo.WithTransaction(c.Request.Context(), func(ctx context.Context, tx *pgstore.Repository) error {
			visitor, err := tx.CreateEventVisitorIdentity(ctx, event.ID)
			if err != nil {
				return err
			}
			result.identity = visitor
			result.authorized = true
			credential, err = issueCredential(ctx, tx, visitor.ID)
			return err
		})
		if err != nil {
			return nil, err
		}
		setCredentialCookie(c, event.ShortID, result.identity.PublicID, credential)
	}
	if platformIdentityID != "" && result.authorized && !result.granted && result.identity.PlatformIdentityID == nil {
		if err := repo.AssociateEventVisitorIdentity(c.Request.Context(), result.identity.ID, platformIdentityID); err != nil {
			return nil, err
		}
		result.identity.PlatformIdentityID = &platformIdentityID
	}
	return result, nil
}

func (v *visitor) controls(ctx context.Context, repo *pgstore.Repository, visitorID string) (bool, error) {
	if v.grantedVisitorID == visitorID || (v.authorized && v.identity.ID == visitorID) {
		return true, nil
	}
	if v.platformIdentityID == "" {
		return false, nil
	}
	return repo.VisitorBelongsToAccount(ctx, visitorID, v.platformIdentityID)
}

// controlsBatch answers controls for many visitor identities with at most one
// ownership read, preserving the single-control precedence: a granted visitor,
// the acting authorized visitor, then the signed-in account's associated
// visitors. A visitor whose identity is not determined by the first two rules
// and has no platform identity is unauthorized.
func (v *visitor) controlsBatch(ctx context.Context, repo *pgstore.Repository, visitorIDs []string) (map[string]bool, error) {
	authorized := make(map[string]bool, len(visitorIDs))
	lookup := make([]string, 0, len(visitorIDs))
	for _, visitorID := range visitorIDs {
		if _, known := authorized[visitorID]; known {
			continue
		}
		if v.grantedVisitorID == visitorID || (v.authorized && v.identity.ID == visitorID) {
			authorized[visitorID] = true
			continue
		}
		authorized[visitorID] = false
		if v.platformIdentityID != "" {
			lookup = append(lookup, visitorID)
		}
	}
	if len(lookup) == 0 {
		return authorized, nil
	}
	owned, err := repo.EventVisitorIdentitiesBelongingToAccount(ctx, v.platformIdentityID, lookup)
	if err != nil {
		return nil, err
	}
	for _, visitorID := range lookup {
		authorized[visitorID] = owned[visitorID]
	}
	return authorized, nil
}
