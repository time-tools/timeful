-- +goose Up
-- Supporting indexes for existing query shapes, plus retirement of two indexes
-- those shapes cannot use.

-- postgres_event_responses: ListResponses reads one event's responses ordered by
-- (created_at, id). The (event_id, created_at, id) index also serves the event
-- foreign-key cascade, so it replaces the narrower (event_id) index dropped
-- below.
CREATE INDEX postgres_event_responses_event_created_idx
    ON postgres_event_responses (event_id, created_at, id);

DROP INDEX postgres_event_responses_event_id_idx;

-- event_attendees: ListAttendees reads one event's members ordered by
-- (created_at, id), and the dashboard membership EXISTS compares lower(email)
-- within an event. The unique (event_id, email) index is case-sensitive and
-- cannot serve the lower(email) probe.
CREATE INDEX event_attendees_event_created_idx
    ON event_attendees (event_id, created_at, id);

CREATE INDEX event_attendees_event_email_lower_idx
    ON event_attendees (event_id, lower(email));

-- postgres_events: the analytics day-spine windows range over created_at and
-- read only non-empty creator_posthog_id while deliberately including
-- soft-deleted events, so the retired postgres_events_active_creator_posthog_id_idx
-- ((creator_posthog_id, created_at DESC) WHERE NOT is_deleted) could not serve
-- them. It is replaced by an explicit partial index: created_at leads for the
-- range, creator_posthog_id covers the distinct aggregation, and the predicate
-- matches the analytics filter exactly.
DROP INDEX postgres_events_active_creator_posthog_id_idx;

CREATE INDEX postgres_events_creator_created_at_idx
    ON postgres_events (created_at, creator_posthog_id)
    WHERE creator_posthog_id IS NOT NULL AND creator_posthog_id <> '';

-- daily_user_log_members: ListActiveUserDays orders one log's members by
-- (first_seen_position, id), and the append path reads MAX(first_seen_position)
-- for the log. The index leads with daily_user_log_id for both.
CREATE INDEX daily_user_log_members_log_position_idx
    ON daily_user_log_members (daily_user_log_id, first_seen_position, id);

-- No event_signup_responses (event_visitor_identity_id) index is added here.
-- Account deletion scans event_signup_responses with `platform_identity_id = $1
-- OR event_visitor_identity_id = ANY($2)`. The only index on either column is
-- event_signup_responses_account_unique_idx (event_id, platform_identity_id)
-- WHERE respondent_kind = 'account', which leads with event_id and so serves no
-- lookup by platform_identity_id alone; one new index cannot serve the OR, and
-- deletion is a cold path. Removing the ::text cast on
-- event_visitor_identity_id is type consistency only.

-- +goose Down
DROP INDEX daily_user_log_members_log_position_idx;

DROP INDEX postgres_events_creator_created_at_idx;

CREATE INDEX postgres_events_active_creator_posthog_id_idx
    ON postgres_events (creator_posthog_id, created_at DESC)
    WHERE NOT is_deleted;

DROP INDEX event_attendees_event_email_lower_idx;

DROP INDEX event_attendees_event_created_idx;

CREATE INDEX postgres_event_responses_event_id_idx
    ON postgres_event_responses (event_id);

DROP INDEX postgres_event_responses_event_created_idx;
