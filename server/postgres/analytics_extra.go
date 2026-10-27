package postgres

import (
	"context"
	"time"
)

// CountDistinctMonthlyActiveEventCreatorsByDay returns one count per day-end
// instant in the same ascending order as the supplied spine. A day without any
// qualifying event yields zero, so the result always has one count per input
// day, and the caller issues one query for the whole reporting range.
func (r *Repository) CountDistinctMonthlyActiveEventCreatorsByDay(ctx context.Context, dayEnds []time.Time) ([]int64, error) {
	return countDistinctMonthlyActiveEventCreatorsByDay(r, ctx, dayEnds)
}

// CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay returns one
// count per day-end instant in the same ascending order as the supplied spine.
// Each day counts the distinct non-empty creators with at least x events
// created in that day's half-open [day-30d, day) window, and a day without any
// qualifying creator yields zero.
func (r *Repository) CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(ctx context.Context, dayEnds []time.Time, x int) ([]int64, error) {
	return countDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(r, ctx, dayEnds, x)
}
