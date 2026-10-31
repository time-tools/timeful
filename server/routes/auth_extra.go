package routes

import (
	"timeful/server/models"
)

// The request-body structs stay handwritten because their tags define the
// request wire shape; auth.gala constructs each one through its constructor
// below instead of naming the local type.

// signInPayload carries the OAuth sign-in body.
type signInPayload struct {
	Code           string              `json:"code" binding:"required"`
	Scope          string              `json:"scope" binding:"required"`
	CalendarType   models.CalendarType `json:"calendarType" binding:"required"`
	TimezoneOffset *int                `json:"timezoneOffset" binding:"required"`
}

func newSignInPayload() signInPayload { return signInPayload{} }

// signInMobilePayload carries the mobile token sign-in body.
type signInMobilePayload struct {
	AccessToken    string                 `json:"accessToken" binding:"required"`
	Scope          string                 `json:"scope" binding:"required"`
	IdToken        string                 `json:"idToken" binding:"required"`
	ExpiresIn      int                    `json:"expiresIn" binding:"required"`
	RefreshToken   string                 `json:"refreshToken" binding:"required"`
	TokenOrigin    models.TokenOriginType `json:"tokenOrigin" binding:"required"`
	CalendarType   models.CalendarType    `json:"calendarType" binding:"required"`
	TimezoneOffset int                    `json:"timezoneOffset" binding:"required"`
}

func newSignInMobilePayload() signInMobilePayload { return signInMobilePayload{} }

// otpEmailPayload carries the email-only OTP check and send bodies.
type otpEmailPayload struct {
	Email string `json:"email" binding:"required"`
}

func newOtpEmailPayload() otpEmailPayload { return otpEmailPayload{} }

// verifyOtpPayload carries the OTP verification body.
type verifyOtpPayload struct {
	Email          string `json:"email" binding:"required"`
	Code           string `json:"code" binding:"required"`
	TimezoneOffset *int   `json:"timezoneOffset" binding:"required"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
}

func newVerifyOtpPayload() verifyOtpPayload { return verifyOtpPayload{} }
