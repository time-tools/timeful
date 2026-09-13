package postgres

import (
	"context"
	"errors"
	"time"
)

// DailyUserLog is a PostgreSQL-owned historical daily user log. LogDate is the
// account-local month/day/year at UTC midnight, and Members holds one row per
// account that signed in that day in first-seen order.
type DailyUserLog struct {
	ID      string
	LogDate time.Time
	Members []DailyUserLogMember
}

// DailyUserLogMember is one account's membership in a daily log. The profile
// fields are rebuilt from the authoritative accounts table at read time and are
// never stored on the log.
type DailyUserLogMember struct {
	PlatformIdentityID string
	FirstName          string
	LastName           string
	Email              string
	Position           int
}

// dailyLogDate returns the start of the account-local month/day/year as UTC
// midnight. It preserves the legacy bucketing: shift the server instant by the
// account's timezone offset and take that calendar date, so a sign-in late at
// night and one the next morning remain distinct in the account's own timezone.
func dailyLogDate(now time.Time, timezoneOffset int) time.Time {
	adjusted := now.Add(time.Duration(timezoneOffset) * time.Minute).UTC()
	return time.Date(adjusted.Year(), adjusted.Month(), adjusted.Day(), 0, 0, 0, 0, time.UTC)
}

// recordDailyUserLogMembershipQuery appends one account to its day's log after
// the existing members. The supporting-index forced-plan test runs this
// statement directly so the MAX access path cannot drift from production.
const recordDailyUserLogMembershipQuery = `INSERT INTO daily_user_log_members (daily_user_log_id, platform_identity_id, first_seen_position)
VALUES ($1, $2, COALESCE((SELECT MAX(first_seen_position) + 1 FROM daily_user_log_members WHERE daily_user_log_id = $1), 0))
ON CONFLICT (daily_user_log_id, platform_identity_id) DO NOTHING`

// listActiveUserDaysQuery returns active-user reporting days from startDate
// through the UTC calendar date of now, newest first. The day spine is
// generated in SQL so days without a log appear with an empty member list, and
// any stored log on or after startDate is retained even when it falls outside
// the spine. The supporting-index forced-plan test runs this statement directly.
const listActiveUserDaysQuery = `WITH reporting_days AS (
    SELECT generate_series($1::date, $2::date, interval '1 day')::date AS log_date
    UNION
    SELECT log_date FROM daily_user_logs WHERE log_date >= $1::date
)
SELECT l.id, d.log_date, m.platform_identity_id, COALESCE(a.first_name, ''), COALESCE(a.last_name, ''), COALESCE(a.email, ''), m.first_seen_position
FROM reporting_days d
LEFT JOIN daily_user_logs l ON l.log_date = d.log_date
LEFT JOIN daily_user_log_members m ON m.daily_user_log_id = l.id
LEFT JOIN accounts a ON a.platform_identity_id = m.platform_identity_id
ORDER BY d.log_date DESC, m.first_seen_position, m.id`

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
