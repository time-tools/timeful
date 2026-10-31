package routes

import (
	"crypto/sha256"

	pgstore "timeful/server/postgres"
)

// visitor is the resolved caller identity for one event. The struct stays
// handwritten because routes/event_routes.go constructs and reads it as a plain
// Go value.
type visitor struct {
	identity           *pgstore.EventVisitorIdentity
	platformIdentityID string
	authorized         bool
	granted            bool
	owner              bool
	grantedVisitorID   string
}

func newVisitor(platformIdentityID string, owner bool) *visitor {
	return &visitor{platformIdentityID: platformIdentityID, owner: owner}
}

// hashBytes returns the SHA-256 digest of value. The helper keeps the .gala
// file off sha256.Sum256's fixed-size [32]byte result, which has no slice
// spelling in GALA.
func hashBytes(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

// visitorIdentityInput is the browser-local identity association request body.
// The struct stays handwritten because its JSON tags define the wire shape.
type visitorIdentityInput struct {
	Identities []visitorIdentityItem `json:"identities"`
}

type visitorIdentityItem struct {
	EventID        string `json:"eventId"`
	EventVisitorID string `json:"eventVisitorId"`
}

func newVisitorIdentityInput() visitorIdentityInput {
	return visitorIdentityInput{}
}
