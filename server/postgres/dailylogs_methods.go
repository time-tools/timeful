package postgres

import (
	"context"
	"errors"
	"time"
)

// Handwritten sibling beside the generated dailylogs.go.
//
// The declarations it needs — DailyUserLog, DailyUserLogMember, dailyLogDate,
// recordDailyUserLogMembershipQuery, and listActiveUserDaysQuery — are
// transpiled from dailylogs.gala. The methods stay here because each one
// returns two values or calls a Repository method, and a Repository method
// declared in a handwritten file is not resolvable from a .gala source.

// RecordDailyUserLogMembership records one account's sign-in for its
// account-local day. It is idempotent per account per day and appends new
// accounts after existing members so first-seen order is preserved. The log and
// its new membership are written in one transaction.
func (r *Repository) RecordDailyUserLogMembership(ctx context.Context, platformIdentityID string, timezoneOffset int) error {
	return r.recordDailyUserLogMembershipAt(ctx, platformIdentityID, timezoneOffset, time.Now())
}

// recordDailyUserLogMembershipAt is the deterministic core of
// RecordDailyUserLogMembership; tests call it with a fixed instant to exercise
// timezone bucketing without depending on the wall clock.
func (r *Repository) recordDailyUserLogMembershipAt(ctx context.Context, platformIdentityID string, timezoneOffset int, now time.Time) error {
	if platformIdentityID == "" {
		return errors.New("daily log platform identity ID is required")
	}
	logDate := dailyLogDate(now, timezoneOffset)
	return r.withTransaction(ctx, func(ctx context.Context, tx *Repository) error {
		var logID string
		if err := tx.db.QueryRow(ctx, `INSERT INTO daily_user_logs (log_date) VALUES ($1)
ON CONFLICT (log_date) DO UPDATE SET updated_at = daily_user_logs.updated_at
RETURNING id`, logDate).Scan(&logID); err != nil {
			return err
		}
		_, err := tx.db.Exec(ctx, recordDailyUserLogMembershipQuery, logID, platformIdentityID)
		return err
	})
}

// ListActiveUserDays returns active-user reporting days from startDate through
// the UTC calendar date of now, newest first. The day spine is generated in SQL
// so days without a log appear with an empty member list, and any stored log on
// or after startDate is retained even when it falls outside the spine. Member
// profiles are rebuilt from authoritative accounts in first-seen order.
func (r *Repository) ListActiveUserDays(ctx context.Context, startDate, now time.Time) ([]DailyUserLog, error) {
	rows, err := r.db.Query(ctx, listActiveUserDaysQuery,
		startDate.UTC().Format("2006-01-02"), now.UTC().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []DailyUserLog{}
	var current *DailyUserLog
	for rows.Next() {
		var logID, platformIdentityID, firstName, lastName, email *string
		var logDate time.Time
		var position *int
		if err := rows.Scan(&logID, &logDate, &platformIdentityID, &firstName, &lastName, &email, &position); err != nil {
			return nil, err
		}
		if current == nil || !current.LogDate.Equal(logDate) {
			logs = append(logs, DailyUserLog{LogDate: logDate, Members: []DailyUserLogMember{}})
			current = &logs[len(logs)-1]
			if logID != nil {
				current.ID = *logID
			}
		}
		if platformIdentityID == nil {
			continue
		}
		member := DailyUserLogMember{PlatformIdentityID: *platformIdentityID}
		if firstName != nil {
			member.FirstName = *firstName
		}
		if lastName != nil {
			member.LastName = *lastName
		}
		if email != nil {
			member.Email = *email
		}
		if position != nil {
			member.Position = *position
		}
		current.Members = append(current.Members, member)
	}
	return logs, rows.Err()
}

// CountAccounts returns the number of signed-up accounts for reporting.
func (r *Repository) CountAccounts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM accounts`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
