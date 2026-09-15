package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// newOtpTestRepository applies the schema migrations into a transaction-scoped
// set of temporary tables. Temp tables shadow the real schema so the isolated
// test never mutates test-stack records.
func newOtpTestRepository(t *testing.T) (context.Context, *Repository, pgx.Tx) {
	t.Helper()
	return newMigrationTestRepository(t)
}

// TestOtpCodeHashRoundTripRejectsTampering proves the stored challenge hash is
// salted, never contains the code, matches in constant time, rejects a wrong or
// malformed hash, and uses a fresh salt per challenge.
func TestOtpCodeHashRoundTripRejectsTampering(t *testing.T) {
	hash, err := hashOtpCode("123456")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(hash, "123456") || !strings.Contains(hash, ":") {
		t.Fatalf("hash is not salted or leaked the code: %q", hash)
	}
	if !otpCodeMatches(hash, "123456") {
		t.Fatal("correct code did not match")
	}
	if otpCodeMatches(hash, "123457") {
		t.Fatal("wrong code matched")
	}
	if otpCodeMatches("not-a-hash", "123456") {
		t.Fatal("malformed stored hash matched")
	}
	if otpCodeMatches("!!!:!!!", "123456") {
		t.Fatal("undecodable stored hash matched")
	}
	again, err := hashOtpCode("123456")
	if err != nil {
		t.Fatal(err)
	}
	if again == hash {
		t.Fatal("two challenges reused the same salt")
	}
}

// TestOtpChallengeCreateHashesAndReplaces proves send stores a normalized email
// with a hashed code, replaces any existing challenge, and resets attempts.
func TestOtpChallengeCreateHashesAndReplaces(t *testing.T) {
	ctx, repo, tx := newOtpTestRepository(t)

	if err := repo.CreateOtpChallenge(ctx, "Ada@Example.com ", "123456", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var email, codeHash string
	var attempts int
	if err := tx.QueryRow(ctx, `SELECT email, code_hash, attempts FROM otp_challenges`).Scan(&email, &codeHash, &attempts); err != nil {
		t.Fatal(err)
	}
	if email != "ada@example.com" {
		t.Fatalf("email not normalized: %q", email)
	}
	if strings.Contains(codeHash, "123456") || attempts != 0 {
		t.Fatalf("stored challenge is not a fresh hash: hash=%q attempts=%d", codeHash, attempts)
	}

	// One wrong attempt records a count.
	if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "000000"); !errors.Is(err, ErrOtpInvalidCode) {
		t.Fatalf("wrong code error = %v, want ErrOtpInvalidCode", err)
	}
	// A new send replaces the challenge and resets the attempt counter.
	if err := repo.CreateOtpChallenge(ctx, "ada@example.com", "654321", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*), max(attempts) FROM otp_challenges WHERE email = $1`, "ada@example.com").Scan(&count, &attempts); err != nil {
		t.Fatal(err)
	}
	if count != 1 || attempts != 0 {
		t.Fatalf("re-send did not replace and reset: count=%d attempts=%d", count, attempts)
	}
	// The replaced code no longer verifies.
	if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "123456"); !errors.Is(err, ErrOtpInvalidCode) {
		t.Fatalf("replaced code error = %v, want ErrOtpInvalidCode", err)
	}
}

// TestOtpChallengeVerifySuccessDeletesChallenge proves a matching code deletes
// the challenge so it cannot be replayed.
func TestOtpChallengeVerifySuccessDeletesChallenge(t *testing.T) {
	ctx, repo, tx := newOtpTestRepository(t)
	if err := repo.CreateOtpChallenge(ctx, "ada@example.com", "123456", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "123456"); err != nil {
		t.Fatalf("verify = %v, want success", err)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM otp_challenges WHERE email = $1`, "ada@example.com").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("successful verification left %d challenges", count)
	}
	if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "123456"); !errors.Is(err, ErrOtpExpired) {
		t.Fatalf("replay error = %v, want ErrOtpExpired", err)
	}
}

// TestOtpChallengeVerifyExpiry proves an expired or missing challenge is
// rejected without consuming an attempt.
func TestOtpChallengeVerifyExpiry(t *testing.T) {
	ctx, repo, tx := newOtpTestRepository(t)
	if err := repo.CreateOtpChallenge(ctx, "ada@example.com", "123456", time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "123456"); !errors.Is(err, ErrOtpExpired) {
		t.Fatalf("expired error = %v, want ErrOtpExpired", err)
	}
	var attempts int
	if err := tx.QueryRow(ctx, `SELECT attempts FROM otp_challenges WHERE email = $1`, "ada@example.com").Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 0 {
		t.Fatalf("expired verification incremented attempts to %d", attempts)
	}
	if err := repo.VerifyOtpChallenge(ctx, "missing@example.com", "123456"); !errors.Is(err, ErrOtpExpired) {
		t.Fatalf("missing challenge error = %v, want ErrOtpExpired", err)
	}
}

// TestOtpChallengeVerifyLockoutDeletesChallenge proves wrong attempts increment
// one at a time and the attempt past the limit is locked out and deleted even
// when the code is correct.
func TestOtpChallengeVerifyLockoutDeletesChallenge(t *testing.T) {
	ctx, repo, tx := newOtpTestRepository(t)
	if err := repo.CreateOtpChallenge(ctx, "ada@example.com", "123456", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt <= otpMaxAttempts; attempt++ {
		if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "000000"); !errors.Is(err, ErrOtpInvalidCode) {
			t.Fatalf("attempt %d error = %v, want ErrOtpInvalidCode", attempt, err)
		}
		var attempts int
		if err := tx.QueryRow(ctx, `SELECT attempts FROM otp_challenges WHERE email = $1`, "ada@example.com").Scan(&attempts); err != nil {
			t.Fatal(err)
		}
		if attempts != attempt {
			t.Fatalf("attempt %d stored attempts = %d", attempt, attempts)
		}
	}
	if err := repo.VerifyOtpChallenge(ctx, "ada@example.com", "123456"); !errors.Is(err, ErrOtpTooManyAttempts) {
		t.Fatalf("lockout error = %v, want ErrOtpTooManyAttempts", err)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM otp_challenges WHERE email = $1`, "ada@example.com").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("lockout left %d challenges", count)
	}
}

// TestOtpChallengeDeleteExpiredChallenges proves the explicit sweep removes only
// expired challenges.
func TestOtpChallengeDeleteExpiredChallenges(t *testing.T) {
	ctx, repo, tx := newOtpTestRepository(t)
	if err := repo.CreateOtpChallenge(ctx, "expired@example.com", "123456", time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateOtpChallenge(ctx, "active@example.com", "654321", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	removed, err := repo.DeleteExpiredOtpChallenges(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	var remaining string
	if err := tx.QueryRow(ctx, `SELECT email FROM otp_challenges`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != "active@example.com" {
		t.Fatalf("remaining challenge = %q, want the unexpired one", remaining)
	}
}
