package respondents

const MaxGuestNameLength = 100

type GuestNameValidationCode string

const (
	GuestNameValid             GuestNameValidationCode = ""
	GuestNameRequired          GuestNameValidationCode = "required"
	GuestNameInvalidFormatting GuestNameValidationCode = "invalid_formatting"
	GuestNameAccountIDLike     GuestNameValidationCode = "account_id_like"
	GuestNameTooLong           GuestNameValidationCode = "too_long"
)

type GuestNameValidationResult struct {
	Name string
	Code GuestNameValidationCode
}
