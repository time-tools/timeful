-- +goose Up
-- folder_events.event_id is NOT NULL, so the partial predicate
-- `WHERE event_id IS NOT NULL` on folder_events_event_unique_idx never
-- excluded a row. The index is recreated without the vestigial predicate.
DROP INDEX folder_events_event_unique_idx;

CREATE UNIQUE INDEX folder_events_event_unique_idx
    ON folder_events (platform_identity_id, event_id);

-- +goose Down
DROP INDEX folder_events_event_unique_idx;

CREATE UNIQUE INDEX folder_events_event_unique_idx
    ON folder_events (platform_identity_id, event_id)
    WHERE event_id IS NOT NULL;
