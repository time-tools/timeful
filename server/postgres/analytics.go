package postgres

import (
	"context"
	"time"
)

// monthlyActiveCreatorLookback is the window length used by the event-creator
// reporting queries. A creator is active when it appears on an event created in
// the 30 days before the reporting instant.
const monthlyActiveCreatorLookback = 30

// monthlyActiveCreatorDaySpine is the shared day spine for the creator
// aggregations. unnest with ordinality keeps the caller's ascending day order,
// so each reporting day yields exactly one row even when it has no events. Each
// event contributes to a day only when it was created in that day's half-open
// [day-30d, day) window with a non-empty creator_posthog_id; created_at has
// one-second precision, so the upper bound excludes the reporting second
// itself. The queries deliberately do not filter is_deleted.
const monthlyActiveCreatorDaySpine = `FROM unnest($1::timestamptz[]) WITH ORDINALITY AS day(day_end, position)
LEFT JOIN events e
  ON e.created_at >= day.day_end - $2::int * interval '1 day'
 AND e.created_at < day.day_end
 AND e.creator_posthog_id IS NOT NULL
 AND e.creator_posthog_id <> ''`

// countDistinctMonthlyActiveEventCreatorsByDayQuery returns one distinct-creator
// count per day end. The supporting-index forced-plan test runs this statement
// directly.
const countDistinctMonthlyActiveEventCreatorsByDayQuery = `SELECT count(DISTINCT e.creator_posthog_id)` + monthlyActiveCreatorDaySpine + `
GROUP BY day.position
ORDER BY day.position`

// countDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDayQuery returns
// one count per day end of distinct creators with at least x events in that
// day's window. The supporting-index forced-plan test runs this statement
// directly.
const countDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDayQuery = `SELECT count(grouped.creator_posthog_id)
FROM unnest($1::timestamptz[]) WITH ORDINALITY AS day(day_end, position)
LEFT JOIN LATERAL (
    SELECT e.creator_posthog_id
    FROM events e
    WHERE e.created_at >= day.day_end - $2::int * interval '1 day'
      AND e.created_at < day.day_end
      AND e.creator_posthog_id IS NOT NULL
      AND e.creator_posthog_id <> ''
    GROUP BY e.creator_posthog_id
    HAVING count(*) >= $3
) AS grouped ON TRUE
GROUP BY day.position
ORDER BY day.position`

// CountDistinctMonthlyActiveEventCreatorsByDay returns one count per day-end
// instant in the same ascending order as the supplied spine. A day without any
// qualifying event yields zero, so the result always has one count per input
// day, and the caller issues one query for the whole reporting range.
func (r *Repository) CountDistinctMonthlyActiveEventCreatorsByDay(ctx context.Context, dayEnds []time.Time) ([]int64, error) {
	rows, err := r.db.Query(ctx, countDistinctMonthlyActiveEventCreatorsByDayQuery, dayEnds, monthlyActiveCreatorLookback)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make([]int64, 0, len(dayEnds))
	for rows.Next() {
		var count int64
		if err := rows.Scan(&count); err != nil {
			return nil, err
		}
		counts = append(counts, count)
	}
	return counts, rows.Err()
}

// CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay returns one
// count per day-end instant in the same ascending order as the supplied spine.
// Each day counts the distinct non-empty creators with at least x events
// created in that day's half-open [day-30d, day) window, and a day without any
// qualifying creator yields zero.
func (r *Repository) CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(ctx context.Context, dayEnds []time.Time, x int) ([]int64, error) {
	rows, err := r.db.Query(ctx, countDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDayQuery, dayEnds, monthlyActiveCreatorLookback, x)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make([]int64, 0, len(dayEnds))
	for rows.Next() {
		var count int64
		if err := rows.Scan(&count); err != nil {
			return nil, err
		}
		counts = append(counts, count)
	}
	return counts, rows.Err()
}
