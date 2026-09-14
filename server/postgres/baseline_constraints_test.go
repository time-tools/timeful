package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// expectConstraintViolation requires fn to fail with the named CHECK
// constraint, so a dropped or renamed baseline invariant is diagnosable
// instead of passing as any other error.
func expectConstraintViolation(t *testing.T, ctx context.Context, tx pgx.Tx, constraint string, fn func() error) {
	t.Helper()
	err := expectSavepointError(t, ctx, tx, fn)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23514" || postgresError.ConstraintName != constraint {
		t.Fatalf("expected %s check violation, got %v", constraint, err)
	}
}

// TestBaselineEventPayloadAndShortIDConstraints proves the baseline event
// invariants with no runtime coverage of their own: short_id must be the
// canonical eight-character storage identifier and payload must be a JSON
// object rather than any other JSON shape.
func TestBaselineEventPayloadAndShortIDConstraints(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)

	if _, err := tx.Exec(ctx, `INSERT INTO events (short_id, name, type, payload)
VALUES ($1, 'Baseline', 'signup', '{"name":"Baseline"}')`, signupTestShortID(t)); err != nil {
		t.Fatalf("accept canonical short id and object payload: %v", err)
	}
	for _, shortID := range []string{"", "AAAAAAA", "AAAAAAAAA", "aaaaaaaa", "IIIIIIII", "AAAAAAAO"} {
		shortID := shortID
		expectConstraintViolation(t, ctx, tx, "events_short_id_format", func() error {
			_, err := tx.Exec(ctx, `INSERT INTO events (short_id, name, type) VALUES ($1, 'Baseline', 'signup')`, shortID)
			return err
		})
	}
	for _, payload := range []string{"[]", `"text"`, "null", "7"} {
		payload := payload
		expectConstraintViolation(t, ctx, tx, "events_payload_object", func() error {
			_, err := tx.Exec(ctx, `INSERT INTO events (short_id, name, type, payload)
VALUES ($1, 'Baseline', 'signup', $2::jsonb)`, signupTestShortID(t), payload)
			return err
		})
	}
}

// TestBaselineResponsePayloadObjectConstraint proves response payloads are
// objects at the storage boundary, matching the event payload invariant.
func TestBaselineResponsePayloadObjectConstraint(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)

	if _, err := tx.Exec(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind, payload)
VALUES ($1, $2, 'guest', '{"name":"Ada"}')`, eventID, visitorID); err != nil {
		t.Fatalf("accept object payload: %v", err)
	}
	expectConstraintViolation(t, ctx, tx, "event_responses_payload_object", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO event_responses (event_id, event_visitor_identity_id, respondent_kind, payload)
VALUES ($1, $2, 'guest', '[]'::jsonb)`, eventID, visitorID)
		return err
	})
}

// TestBaselineAccessTransferIdentityConstraint proves a transfer names exactly
// one source: a credential or a platform identity, never both and never
// neither. It also proves the retained source hash and state guards.
func TestBaselineAccessTransferIdentityConstraint(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	var credentialID string
	if err := tx.QueryRow(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash)
VALUES ($1, decode(repeat('ab',32),'hex')) RETURNING id`, visitorID).Scan(&credentialID); err != nil {
		t.Fatal(err)
	}
	var platformIdentityID string
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&platformIdentityID); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte("baseline transfer"))

	expectConstraintViolation(t, ctx, tx, "access_transfers_source_xor_platform_identity", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO access_transfers (event_id, source_hash) VALUES ($1, $2)`, eventID, hash[:])
		return err
	})
	expectConstraintViolation(t, ctx, tx, "access_transfers_source_xor_platform_identity", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO access_transfers (event_id, source_hash, source_credential_id, platform_identity_id)
VALUES ($1, $2, $3, $4)`, eventID, hash[:], credentialID, platformIdentityID)
		return err
	})
	for _, source := range []struct {
		name             string
		credentialID     any
		platformIdentity any
	}{
		{name: "credential source", credentialID: credentialID, platformIdentity: nil},
		{name: "platform identity source", credentialID: nil, platformIdentity: platformIdentityID},
	} {
		source := source
		t.Run("accepts "+source.name, func(t *testing.T) {
			if _, err := tx.Exec(ctx, `INSERT INTO access_transfers (event_id, source_hash, source_credential_id, platform_identity_id)
VALUES ($1, $2, $3, $4)`, eventID, hash[:], source.credentialID, source.platformIdentity); err != nil {
				t.Fatalf("insert %s: %v", source.name, err)
			}
		})
	}
	expectConstraintViolation(t, ctx, tx, "access_transfers_source_hash_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO access_transfers (event_id, source_hash, platform_identity_id) VALUES ($1, 'short', $2)`, eventID, platformIdentityID)
		return err
	})
	expectConstraintViolation(t, ctx, tx, "access_transfers_state_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO access_transfers (event_id, source_hash, platform_identity_id, state) VALUES ($1, $2, $3, 'expired')`, eventID, hash[:], platformIdentityID)
		return err
	})
}

// TestBaselineOtpChallengeConstraints proves an OTP challenge cannot carry an
// empty email or code hash and cannot take a negative attempt count, and that
// attempts default to zero.
func TestBaselineOtpChallengeConstraints(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	expiresAt := time.Now().Add(time.Minute)

	if _, err := tx.Exec(ctx, `INSERT INTO otp_challenges (email, code_hash, expires_at) VALUES ('ada@example.com', 'salt:hash', $1)`, expiresAt); err != nil {
		t.Fatalf("accept a challenge row: %v", err)
	}
	expectConstraintViolation(t, ctx, tx, "otp_challenges_email_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO otp_challenges (email, code_hash, expires_at) VALUES ('', 'salt:hash', $1)`, expiresAt)
		return err
	})
	expectConstraintViolation(t, ctx, tx, "otp_challenges_code_hash_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO otp_challenges (email, code_hash, expires_at) VALUES ('grace@example.com', '', $1)`, expiresAt)
		return err
	})
	expectConstraintViolation(t, ctx, tx, "otp_challenges_attempts_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO otp_challenges (email, code_hash, expires_at, attempts) VALUES ('grace@example.com', 'salt:hash', $1, -1)`, expiresAt)
		return err
	})
	var attempts int
	if err := tx.QueryRow(ctx, `SELECT attempts FROM otp_challenges WHERE email = 'ada@example.com'`).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 0 {
		t.Fatalf("default attempts = %d, want 0", attempts)
	}
}

// TestBaselineCalendarPreferenceConstraints proves calendar preferences keep
// the first-account fallback and undefined token origin representable while
// rejecting an empty key and an unknown origin.
func TestBaselineCalendarPreferenceConstraints(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	var firstID, secondID, thirdID string
	for _, id := range []*string{&firstID, &secondID, &thirdID} {
		if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(id); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := tx.Exec(ctx, `INSERT INTO calendar_preferences (platform_identity_id) VALUES ($1)`, firstID); err != nil {
		t.Fatalf("accept undefined preferences: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO calendar_preferences (platform_identity_id, primary_account_key, token_origin)
VALUES ($1, 'ada@example.com_google', 'ios')`, secondID); err != nil {
		t.Fatalf("accept defined preferences: %v", err)
	}
	expectConstraintViolation(t, ctx, tx, "calendar_preferences_primary_account_key_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_preferences (platform_identity_id, primary_account_key) VALUES ($1, '')`, thirdID)
		return err
	})
	expectConstraintViolation(t, ctx, tx, "calendar_preferences_token_origin_check", func() error {
		_, err := tx.Exec(ctx, `INSERT INTO calendar_preferences (platform_identity_id, token_origin) VALUES ($1, 'desktop')`, thirdID)
		return err
	})
}

// TestBaselineAdditionalCheckConstraints covers the remaining baseline CHECK
// invariants that no other test exercises directly. The FR-119
// events_name_length constraint is owned by name_constraint_test.go.
func TestBaselineAdditionalCheckConstraints(t *testing.T) {
	ctx, _, tx := newMigrationTestRepository(t)
	eventID := seedSignupEvent(t, ctx, tx, signupTestShortID(t))
	visitorID := seedSignupVisitor(t, ctx, tx, eventID)
	var platformIdentityID string
	if err := tx.QueryRow(ctx, `INSERT INTO platform_identities DEFAULT VALUES RETURNING id`).Scan(&platformIdentityID); err != nil {
		t.Fatal(err)
	}
	var logID string
	if err := tx.QueryRow(ctx, `INSERT INTO daily_user_logs (log_date) VALUES (CURRENT_DATE) RETURNING id`).Scan(&logID); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte("baseline additional"))

	cases := []struct {
		name       string
		constraint string
		run        func() error
	}{
		{
			name:       "event response counter rejects negative",
			constraint: "events_num_responses_check",
			run: func() error {
				_, err := tx.Exec(ctx, `INSERT INTO events (short_id, name, type, num_responses) VALUES ($1, 'Baseline', 'signup', -1)`, signupTestShortID(t))
				return err
			},
		},
		{
			name:       "event owner edit token hash requires 32 bytes",
			constraint: "events_owner_edit_token_hash_check",
			run: func() error {
				_, err := tx.Exec(ctx, `INSERT INTO events (short_id, name, type, owner_edit_token_hash) VALUES ($1, 'Baseline', 'signup', 'short')`, signupTestShortID(t))
				return err
			},
		},
		{
			name:       "credential hash requires 32 bytes",
			constraint: "event_visitor_credentials_credential_hash_check",
			run: func() error {
				_, err := tx.Exec(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash) VALUES ($1, 'short')`, visitorID)
				return err
			},
		},
		{
			name:       "credential kind is base or granted",
			constraint: "event_visitor_credentials_kind_check",
			run: func() error {
				_, err := tx.Exec(ctx, `INSERT INTO event_visitor_credentials (event_visitor_identity_id, credential_hash, kind) VALUES ($1, $2, 'owner')`, visitorID, hash[:])
				return err
			},
		},
		{
			name:       "transfer request target hash requires 32 bytes",
			constraint: "access_transfer_requests_target_hash_check",
			run: func() error {
				var transferID string
				if err := tx.QueryRow(ctx, `INSERT INTO access_transfers (event_id, source_hash, platform_identity_id) VALUES ($1, $2, $3) RETURNING id`, eventID, hash[:], platformIdentityID).Scan(&transferID); err != nil {
					return err
				}
				_, err := tx.Exec(ctx, `INSERT INTO access_transfer_requests (transfer_id, target_hash, code) VALUES ($1, 'short', 'CODE1234')`, transferID)
				return err
			},
		},
		{
			name:       "account event counter rejects negative",
			constraint: "accounts_num_events_created_check",
			run: func() error {
				_, err := tx.Exec(ctx, `INSERT INTO accounts (platform_identity_id, num_events_created) VALUES ($1, -1)`, platformIdentityID)
				return err
			},
		},
		{
			name:       "daily log member position rejects negative",
			constraint: "daily_user_log_members_first_seen_position_check",
			run: func() error {
				_, err := tx.Exec(ctx, `INSERT INTO daily_user_log_members (daily_user_log_id, platform_identity_id, first_seen_position) VALUES ($1, $2, -1)`, logID, platformIdentityID)
				return err
			},
		},
		{
			name:       "sub calendar id rejects empty",
			constraint: "calendar_sub_calendars_sub_calendar_id_check",
			run: func() error {
				var accountID string
				if err := tx.QueryRow(ctx, `INSERT INTO calendar_accounts (platform_identity_id, calendar_key, calendar_type) VALUES ($1, 'a_google', 'google') RETURNING id`, platformIdentityID).Scan(&accountID); err != nil {
					return err
				}
				_, err := tx.Exec(ctx, `INSERT INTO calendar_sub_calendars (calendar_account_id, sub_calendar_id) VALUES ($1, '')`, accountID)
				return err
			},
		},
	}
	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			expectConstraintViolation(t, ctx, tx, testCase.constraint, testCase.run)
		})
	}
}
