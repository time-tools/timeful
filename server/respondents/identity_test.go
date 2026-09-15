package respondents

import (
	"strings"
	"testing"

	"timeful/server/models"
)

func TestNormalizeGuestNameCanonicalizesWhitespaceFormattingAndUnicode(t *testing.T) {
	name := NormalizeGuestName("  A\u200bda\u0301 \n")

	if name != "Adá" {
		t.Fatalf("expected canonical guest name, got %q", name)
	}
}

func TestValidateGuestNameRejectsBlankAndFormattingOnlyValues(t *testing.T) {
	if result := ValidateGuestName("   "); result.Code != GuestNameRequired {
		t.Fatalf("expected required validation error, got %q", result.Code)
	}

	if result := ValidateGuestName("\u200b\u200b"); result.Code != GuestNameInvalidFormatting {
		t.Fatalf("expected invalid formatting validation error, got %q", result.Code)
	}
}

func TestValidateGuestNameRejectsIdentifierLikeAndOverlengthValues(t *testing.T) {
	if result := ValidateGuestName(models.NewUUID().String()); result.Code != GuestNameAccountIDLike {
		t.Fatalf("expected identifier-like validation error, got %q", result.Code)
	}

	tooLong := strings.Repeat("a", MaxGuestNameLength+1)
	if result := ValidateGuestName(tooLong); result.Code != GuestNameTooLong {
		t.Fatalf("expected too-long validation error, got %q", result.Code)
	}

	tooLongCombining := strings.Repeat("e\u0301", MaxGuestNameLength+1)
	if result := ValidateGuestName(tooLongCombining); result.Code != GuestNameTooLong {
		t.Fatalf("expected combining-character name to be too long by rune count, got %q", result.Code)
	}
}

func TestSanitizeUserDisplayNameUsesLiveProfileData(t *testing.T) {
	firstName, lastName, displayName := SanitizeUserDisplayName(&models.User{
		FirstName: "  Ada\u200b ",
		LastName:  " Lovelace\n",
	})

	if firstName != "Ada" {
		t.Fatalf("expected sanitized first name, got %q", firstName)
	}
	if lastName != "Lovelace" {
		t.Fatalf("expected sanitized last name, got %q", lastName)
	}
	if displayName != "Ada Lovelace" {
		t.Fatalf("expected sanitized display name, got %q", displayName)
	}
}
