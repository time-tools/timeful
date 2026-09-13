package errs

import (
	"encoding/json"
	"fmt"
)

const EventOwnerCredentialRequired = "event-owner-credential-required"
const EventArchived = "event-archived"

// Errors enum
// TODO: make these an actual type (i.e. Errors.NotSignedIn)
const (
	NotSignedIn           string = "not-signed-in"
	UserDoesNotExist      string = "user-does-not-exist"
	AccountEmailMismatch  string = "account-email-mismatch"
	EventNotFound         string = "event-not-found"
	EventNameTooLong      string = "event-name-too-long"
	UserNotEventOwner     string = "user-not-event-owner"
	AttendeeEmailNotFound string = "attendee-email-not-found"
	EventNotGroup         string = "event-not-group"
	InvalidCredentials    string = "invalid-credentials"
	OtpExpired            string = "otp-expired"
	OtpInvalidCode        string = "otp-invalid-code"
	OtpTooManyAttempts    string = "otp-too-many-attempts"
	InvalidIdToken        string = "invalid-id-token"
)

type GoogleAPIError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Status  string      `json:"status"`
	Details interface{} `json:"details"`
	Errors  interface{} `json:"errors"`
}

func (e *GoogleAPIError) Error() string {
	s, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintln("GoogleAPIError: <error parsing json>")
	}

	return fmt.Sprintln("GoogleAPIError: ", string(s))
}
