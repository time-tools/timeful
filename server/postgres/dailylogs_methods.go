package postgres

import (
	"context"
	"time"
)

// Handwritten sibling beside the generated dailylogs.go.
//
// dailyLogDate, recordDailyUserLogMembershipQuery, listActiveUserDaysQuery,
// and the free-function bodies of the four methods are transpiled from
// dailylogs.gala. DailyUserLog and DailyUserLogMember are declared in
// dailylogs_types.go. The methods stay here as one-line wrappers because a
// method declared on Repository in a .gala file makes the GALA-declared set
// Repository's complete method set, which hides the eighteen methods
// repository_methods.go declares.

// RecordDailyUserLogMembership records one account's sign-in for its
// account-local day. It is idempotent per account per day and appends new
// accounts after existing members so first-seen order is preserved. The log and
// its new membership are written in one transaction.
func (r *Repository) RecordDailyUserLogMembership(ctx context.Context, platformIdentityID string, timezoneOffset int) error {
	return recordDailyUserLogMembership(r, ctx, platformIdentityID, timezoneOffset)
}

// recordDailyUserLogMembershipAt is the deterministic core of
// RecordDailyUserLogMembership; tests call it with a fixed instant to exercise
// timezone bucketing without depending on the wall clock.
func (r *Repository) recordDailyUserLogMembershipAt(ctx context.Context, platformIdentityID string, timezoneOffset int, now time.Time) error {
	return recordDailyUserLogMembershipAt(r, ctx, platformIdentityID, timezoneOffset, now)
}

// ListActiveUserDays returns active-user reporting days from startDate through
// the UTC calendar date of now, newest first.
func (r *Repository) ListActiveUserDays(ctx context.Context, startDate, now time.Time) ([]DailyUserLog, error) {
	return listActiveUserDays(r, ctx, startDate, now)
}

// CountAccounts returns the number of signed-up accounts for reporting.
func (r *Repository) CountAccounts(ctx context.Context) (int64, error) {
	return countAccounts(r, ctx)
}
