package routes

import (
	"encoding/json"
	"strings"
	"testing"

	"timeful/server/models"
	"timeful/server/respondents"
)

func TestGuestResponseJSONOmitsRawEditTokenAndIncludesOwnershipMetadata(t *testing.T) {
	response := &models.Response{
		Name:               "Token Ada",
		GuestId:            "guest_opaque_id",
		GuestEditToken:     "secret-token",
		GuestEditPolicy:    "protected",
		GuestOwnershipMode: guestOwnershipModeToken,
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	jsonText := string(payload)

	if strings.Contains(jsonText, "secret-token") || strings.Contains(jsonText, "guestEditToken") {
		t.Fatalf("expected raw guest edit token to stay private, got %s", jsonText)
	}
	if !strings.Contains(jsonText, `"guestOwnershipMode":"token"`) {
		t.Fatalf("expected ownership mode in payload, got %s", jsonText)
	}
	if !strings.Contains(jsonText, `"guestEditPolicy":"protected"`) {
		t.Fatalf("expected guest edit policy in payload, got %s", jsonText)
	}
}

func TestHasValidGuestNameRejectsBlankNames(t *testing.T) {
	if hasValidGuestName("") {
		t.Fatal("expected empty guest name to be invalid")
	}
	if hasValidGuestName("   ") {
		t.Fatal("expected whitespace-only guest name to be invalid")
	}
	if hasValidGuestName(models.NewUUID().String()) {
		t.Fatal("expected identifier-like guest name to be invalid")
	}
	if !hasValidGuestName(" A\u200bda ") {
		t.Fatal("expected non-empty guest name to be valid")
	}
}

func TestCanonicalGuestNameStripsFormattingAndNormalizesUnicode(t *testing.T) {
	if canonicalName := canonicalGuestName(" A\u200bda\u0301 "); canonicalName != "Adá" {
		t.Fatalf("expected canonical guest name, got %q", canonicalName)
	}
}

func TestGuestNameValidationErrorMessageUsesSpecificMessages(t *testing.T) {
	if message := guestNameValidationErrorMessage(respondents.GuestNameRequired); message != "Guest name is required" {
		t.Fatalf("unexpected required guest name message %q", message)
	}
	if message := guestNameValidationErrorMessage(respondents.GuestNameInvalidFormatting); message == "" {
		t.Fatal("expected invalid-formatting message")
	}
	if message := guestNameValidationErrorMessage(respondents.GuestNameAccountIDLike); message == "" {
		t.Fatal("expected account-id-like message")
	}
	if message := guestNameValidationErrorMessage(respondents.GuestNameTooLong); message == "" {
		t.Fatal("expected too-long message")
	}
}

func TestShouldExposeGuestSignUpResponsePayload(t *testing.T) {
	if !shouldExposeGuestSignUpResponsePayload("", &models.SignUpResponse{Name: "Ada"}) {
		t.Fatal("expected named guest sign-up payload row to be exposed")
	}
	if shouldExposeGuestSignUpResponsePayload("", &models.SignUpResponse{Name: ""}) {
		t.Fatal("expected blank-named guest sign-up payload row to be hidden")
	}
	if shouldExposeGuestSignUpResponsePayload("", &models.SignUpResponse{Name: "   "}) {
		t.Fatal("expected whitespace-only guest sign-up payload row to be hidden")
	}
	if !shouldExposeGuestSignUpResponsePayload("", &models.SignUpResponse{
		UserId: models.NewUUID(),
	}) {
		t.Fatal("expected signed-in sign-up payload row to remain exposed")
	}
}
