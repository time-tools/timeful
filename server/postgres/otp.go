package postgres

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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

// normalizeOtpEmail applies the same case folding the route applies so the
// unique email key always describes the same challenge.
func normalizeOtpEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// hashOtpCode returns the versionless salted one-way hash stored in
// otp_challenges.code_hash as base64(salt):base64(HMAC-SHA256(salt, code)). The
// salt is per challenge, and the code is never stored in readable form.
func hashOtpCode(code string) (string, error) {
	salt := make([]byte, otpHashSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return encodeOtpHash(salt, otpDigest(salt, code)), nil
}

func otpDigest(salt []byte, code string) []byte {
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(code))
	return mac.Sum(nil)
}

func encodeOtpHash(salt, digest []byte) string {
	return base64.StdEncoding.EncodeToString(salt) + ":" + base64.StdEncoding.EncodeToString(digest)
}

// otpCodeMatches recomputes the hash from the stored salt and compares in
// constant time. A malformed stored hash is a non-match, never an error.
func otpCodeMatches(stored, code string) bool {
	saltPart, digestPart, ok := strings.Cut(stored, ":")
	if !ok {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(saltPart)
	if err != nil {
		return false
	}
	expected, err := base64.StdEncoding.DecodeString(digestPart)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(otpDigest(salt, code), expected) == 1
}

// CreateOtpChallenge stores the only active challenge for email, replacing any
// existing challenge and resetting the attempt counter. The one-time code given
// by the caller is hashed before storage and is never persisted in plaintext.
func (r *Repository) CreateOtpChallenge(ctx context.Context, email, code string, expiresAt time.Time) error {
	email = normalizeOtpEmail(email)
	if email == "" {
		return errors.New("otp challenge email is required")
	}
	if code == "" {
		return errors.New("otp challenge code is required")
	}
	codeHash, err := hashOtpCode(code)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `INSERT INTO otp_challenges (email, code_hash, expires_at, attempts)
VALUES ($1, $2, $3, 0)
ON CONFLICT (email) DO UPDATE
SET code_hash = EXCLUDED.code_hash,
    expires_at = EXCLUDED.expires_at,
    attempts = 0,
    updated_at = clock_timestamp()`, email, codeHash, expiresAt)
	return err
}

// VerifyOtpChallenge consumes one attempt for the unexpired challenge owned by
// email. The attempt increment, expiry check, and code match happen under a
// single row-locked UPDATE so concurrent verifications cannot exceed the
// attempt limit. The challenge is deleted on success and on lockout.
func (r *Repository) VerifyOtpChallenge(ctx context.Context, email, code string) error {
	email = normalizeOtpEmail(email)
	var id, codeHash string
	var attempts int
	err := r.db.QueryRow(ctx, `UPDATE otp_challenges
SET attempts = attempts + 1, updated_at = clock_timestamp()
WHERE email = $1 AND expires_at > clock_timestamp()
RETURNING id, code_hash, attempts`, email).Scan(&id, &codeHash, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOtpExpired
	}
	if err != nil {
		return err
	}
	// attempts is the post-increment value, so the limit is enforced when the
	// sixth attempt is made.
	if attempts > otpMaxAttempts {
		if _, err := r.db.Exec(ctx, `DELETE FROM otp_challenges WHERE id = $1`, id); err != nil {
			return err
		}
		return ErrOtpTooManyAttempts
	}
	if !otpCodeMatches(codeHash, code) {
		return ErrOtpInvalidCode
	}
	_, err = r.db.Exec(ctx, `DELETE FROM otp_challenges WHERE id = $1`, id)
	return err
}

// DeleteExpiredOtpChallenges removes every challenge whose expiry has passed.
func (r *Repository) DeleteExpiredOtpChallenges(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM otp_challenges WHERE expires_at <= clock_timestamp()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// DeleteOtpChallenge removes the challenge owned by email. Deleting a missing
// challenge is a no-op.
func (r *Repository) DeleteOtpChallenge(ctx context.Context, email string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM otp_challenges WHERE email = $1`, normalizeOtpEmail(email))
	return err
}
