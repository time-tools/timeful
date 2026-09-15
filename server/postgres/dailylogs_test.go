package postgres

import (
	"testing"
	"time"
)

// TestDailyLogDateUsesTimezoneAdjustedCalendarDate proves the log date is the
// account-local month/day/year at UTC midnight, so a sign-in late at night and
// one the next morning in another timezone do not collapse into the same day.
func TestDailyLogDateUsesTimezoneAdjustedCalendarDate(t *testing.T) {
	now := time.Date(2026, 9, 10, 2, 30, 0, 0, time.UTC)
	cases := []struct {
		name   string
		offset int
		want   string
	}{
		{"server timezone keeps the date", 0, "2026-09-10"},
		{"west of UTC crosses to the previous day", -420, "2026-09-09"},
		{"east of UTC stays on the same day", 420, "2026-09-10"},
		{"a full day forward advances the date", 1440, "2026-09-11"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dailyLogDate(now, tc.offset)
			if got.Format("2006-01-02") != tc.want {
				t.Fatalf("dailyLogDate(%s, %d) = %s, want %s", now, tc.offset, got.Format("2006-01-02"), tc.want)
			}
			if !got.Equal(time.Date(got.Year(), got.Month(), got.Day(), 0, 0, 0, 0, time.UTC)) {
				t.Fatalf("dailyLogDate must be UTC midnight, got %s", got)
			}
		})
	}
}

// TestRecordDailyUserLogMembershipIsIdempotentAndOrdered proves repeated
// same-day sign-ins create exactly one membership and that new accounts append
// after existing members so first-seen order is preserved.
func TestRecordDailyUserLogMembershipIsIdempotentAndOrdered(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	first, err := repo.CreateAccount(ctx, Account{Email: "first@example.com", FirstName: "First", LastName: "One"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.CreateAccount(ctx, Account{Email: "second@example.com", FirstName: "Second", LastName: "Two"})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		if err := repo.recordDailyUserLogMembershipAt(ctx, first.PlatformIdentityID, 0, now); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, second.PlatformIdentityID, 0, now); err != nil {
		t.Fatal(err)
	}

	logs, err := repo.ListActiveUserDays(ctx, dailyLogDate(now, 0), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("same-day sign-ins created %d logs, want 1", len(logs))
	}
	if got := logs[0].LogDate.Format("2006-01-02"); got != "2026-09-10" {
		t.Fatalf("log date = %s, want 2026-09-10", got)
	}
	if len(logs[0].Members) != 2 {
		t.Fatalf("members = %d, want 2 (same-day idempotency)", len(logs[0].Members))
	}
	if logs[0].Members[0].PlatformIdentityID != first.PlatformIdentityID || logs[0].Members[1].PlatformIdentityID != second.PlatformIdentityID {
		t.Fatalf("member order = %q then %q, want first-seen order", logs[0].Members[0].PlatformIdentityID, logs[0].Members[1].PlatformIdentityID)
	}
	if logs[0].Members[0].Position != 0 || logs[0].Members[1].Position != 1 {
		t.Fatalf("positions = %d, %d, want 0, 1", logs[0].Members[0].Position, logs[0].Members[1].Position)
	}
	if logs[0].Members[0].FirstName != "First" || logs[0].Members[0].Email != "first@example.com" {
		t.Fatalf("profile not rebuilt from accounts: %#v", logs[0].Members[0])
	}
}

// TestRecordDailyUserLogMembershipBucketsByAccountTimezone proves two accounts
// signing in at the same instant are logged under different local dates when
// their timezone offsets differ.
func TestRecordDailyUserLogMembershipBucketsByAccountTimezone(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	instant := time.Date(2026, 9, 10, 2, 30, 0, 0, time.UTC)

	serverTZ, err := repo.CreateAccount(ctx, Account{Email: "server@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	westTZ, err := repo.CreateAccount(ctx, Account{Email: "west@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, serverTZ.PlatformIdentityID, 0, instant); err != nil {
		t.Fatal(err)
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, westTZ.PlatformIdentityID, -420, instant); err != nil {
		t.Fatal(err)
	}

	logs, err := repo.ListActiveUserDays(ctx,
		time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("logs = %d, want 2 timezone-bucketed days", len(logs))
	}
	if got := logs[0].LogDate.Format("2006-01-02"); got != "2026-09-10" {
		t.Fatalf("newest log = %s, want 2026-09-10", got)
	}
	if len(logs[0].Members) != 1 || logs[0].Members[0].PlatformIdentityID != serverTZ.PlatformIdentityID {
		t.Fatalf("server-timezone log members = %#v", logs[0].Members)
	}
	if got := logs[1].LogDate.Format("2006-01-02"); got != "2026-09-09" {
		t.Fatalf("west-timezone log = %s, want 2026-09-09", got)
	}
	if len(logs[1].Members) != 1 || logs[1].Members[0].PlatformIdentityID != westTZ.PlatformIdentityID {
		t.Fatalf("west-timezone log members = %#v", logs[1].Members)
	}
}

// TestListActiveUserDaysPadsEmptyDays proves the reporting read returns every
// day in the range, newest first, with empty days preserved as zero-member days.
func TestListActiveUserDaysPadsEmptyDays(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "active@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	seeded := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	if err := repo.recordDailyUserLogMembershipAt(ctx, account.PlatformIdentityID, 0, seeded); err != nil {
		t.Fatal(err)
	}

	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	days, err := repo.ListActiveUserDays(ctx, start, now)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"2026-09-05", "2026-09-04", "2026-09-03", "2026-09-02", "2026-09-01"}
	if len(days) != len(want) {
		t.Fatalf("padded days = %d, want %d", len(days), len(want))
	}
	for i, day := range days {
		if got := day.LogDate.Format("2006-01-02"); got != want[i] {
			t.Fatalf("day %d = %s, want %s", i, got, want[i])
		}
		if want[i] == "2026-09-03" {
			if len(day.Members) != 1 || day.Members[0].PlatformIdentityID != account.PlatformIdentityID {
				t.Fatalf("seeded day members = %#v", day.Members)
			}
			continue
		}
		if len(day.Members) != 0 {
			t.Fatalf("empty day %s has %d members", want[i], len(day.Members))
		}
		if day.Members == nil {
			t.Fatalf("empty day %s has a nil member slice", want[i])
		}
	}
}

// TestListActiveUserDaysKeepsLogsBeyondTheSpine proves a stored log after the
// report's now instant still appears, newest first, alongside the padded spine,
// and that both padded and seeded days keep non-nil member slices.
func TestListActiveUserDaysKeepsLogsBeyondTheSpine(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "future@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, account.PlatformIdentityID, 0, time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	days, err := repo.ListActiveUserDays(ctx,
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"2026-09-20", "2026-09-05", "2026-09-04", "2026-09-03", "2026-09-02", "2026-09-01"}
	if len(days) != len(want) {
		t.Fatalf("days = %d, want %d (%v)", len(days), len(want), dailyLogDates(days))
	}
	for i, day := range days {
		if got := day.LogDate.Format("2006-01-02"); got != want[i] {
			t.Fatalf("day %d = %s, want %s", i, got, want[i])
		}
		if day.Members == nil {
			t.Fatalf("day %s has a nil member slice", want[i])
		}
	}
	if len(days[0].Members) != 1 || days[0].Members[0].PlatformIdentityID != account.PlatformIdentityID {
		t.Fatalf("log beyond the spine members = %#v", days[0].Members)
	}
}

// TestListActiveUserDaysNormalizesBoundsToUTC proves the report bounds use the
// UTC calendar date of each instant, so a local-late start date does not pull in
// the previous UTC day's log.
func TestListActiveUserDaysNormalizesBoundsToUTC(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "utc@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, account.PlatformIdentityID, 0, time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	// 23:00 UTC-7 is 06:00 UTC on the next calendar day, so the 09-10 log is
	// outside the UTC-normalized start bound.
	start := time.Date(2026, 9, 10, 23, 0, 0, 0, time.FixedZone("UserOffset", -7*60*60))
	days, err := repo.ListActiveUserDays(ctx, start, time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 || days[0].LogDate.Format("2006-01-02") != "2026-09-11" {
		t.Fatalf("UTC-normalized days = %v, want just 2026-09-11", dailyLogDates(days))
	}
}

func dailyLogDates(days []DailyUserLog) []string {
	dates := make([]string, 0, len(days))
	for _, day := range days {
		dates = append(dates, day.LogDate.Format("2006-01-02"))
	}
	return dates
}

// TestDeleteAccountRemovesDailyLogMembershipAndEmptiedLogs proves the
// PostgreSQL deletion path removes the account's memberships, deletes a log the
// removal emptied, and preserves a log shared with another account.
func TestDeleteAccountRemovesDailyLogMembershipAndEmptiedLogs(t *testing.T) {
	ctx, repo, _ := newAccountsTestRepository(t)
	account, err := repo.CreateAccount(ctx, Account{Email: "delete@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := repo.CreateAccount(ctx, Account{Email: "other@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	shared := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	solo := shared.AddDate(0, 0, -1)
	if err := repo.recordDailyUserLogMembershipAt(ctx, account.PlatformIdentityID, 0, shared); err != nil {
		t.Fatal(err)
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, other.PlatformIdentityID, 0, shared); err != nil {
		t.Fatal(err)
	}
	if err := repo.recordDailyUserLogMembershipAt(ctx, account.PlatformIdentityID, 0, solo); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteAccountByPlatformIdentityID(ctx, account.PlatformIdentityID); err != nil {
		t.Fatal(err)
	}

	logs, err := repo.ListActiveUserDays(ctx, shared, shared)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs after deletion = %d, want only the shared log", len(logs))
	}
	if got := logs[0].LogDate.Format("2006-01-02"); got != "2026-09-10" {
		t.Fatalf("surviving log = %s, want the shared 2026-09-10 log", got)
	}
	if len(logs[0].Members) != 1 || logs[0].Members[0].PlatformIdentityID != other.PlatformIdentityID {
		t.Fatalf("shared log members after deletion = %#v", logs[0].Members)
	}
}
