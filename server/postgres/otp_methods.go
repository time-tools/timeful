package postgres

import (
	"context"
	"errors"
	"time"
)

// This file keeps the OTP constants, the three API error values, and the four
// *Repository wrappers beside the generated otp.go. The hash helpers and the
// method bodies live in otp.gala; the wrappers stay handwritten because a
// method declared on Repository in a .gala file makes the GALA-declared set
// Repository's complete method set, which hides the eighteen methods
// repository_methods.go declares.

const (
	// otpHashSaltLength is the number of random bytes mixed into each OTP hash so
	// equal codes stored at different times never share a digest.
	otpHashSaltLength = 16
	// otpMaxAttempts is the number of wrong attempts allowed before the sixth
	// verification attempt is locked out.
	otpMaxAttempts = 5
)

// OTP challenge errors are reported to the route layer, which maps them to the
// existing API error values and status codes.
var (
	// ErrOtpExpired reports that no unexpired challenge exists for the email.
	ErrOtpExpired = errors.New("otp challenge is missing or expired")
	// ErrOtpInvalidCode reports that the submitted code does not match the hash.
	ErrOtpInvalidCode = errors.New("otp challenge code does not match")
	// ErrOtpTooManyAttempts reports that the challenge exceeded its attempt
	// limit; the challenge is deleted before this is returned.
	ErrOtpTooManyAttempts = errors.New("otp challenge exceeded the attempt limit")
)

// CreateOtpChallenge stores the only active challenge for email, replacing any
// existing challenge and resetting the attempt counter. The one-time code given
// by the caller is hashed before storage and is never persisted in plaintext.
func (r *Repository) CreateOtpChallenge(ctx context.Context, email, code string, expiresAt time.Time) error {
	return createOtpChallenge(r, ctx, email, code, expiresAt)
}

// VerifyOtpChallenge consumes one attempt for the unexpired challenge owned by
// email. The attempt increment, expiry check, and code match happen under a
// single row-locked UPDATE so concurrent verifications cannot exceed the
// attempt limit. The challenge is deleted on success and on lockout.
func (r *Repository) VerifyOtpChallenge(ctx context.Context, email, code string) error {
	return verifyOtpChallenge(r, ctx, email, code)
}

// DeleteExpiredOtpChallenges removes every challenge whose expiry has passed.
func (r *Repository) DeleteExpiredOtpChallenges(ctx context.Context) (int64, error) {
	return deleteExpiredOtpChallenges(r, ctx)
}

// DeleteOtpChallenge removes the challenge owned by email. Deleting a missing
// challenge is a no-op.
func (r *Repository) DeleteOtpChallenge(ctx context.Context, email string) error {
	return deleteOtpChallenge(r, ctx, email)
}
