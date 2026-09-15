-- +goose Up
-- Enforce the FR-119 event-name cap at the storage boundary. char_length
-- counts Unicode code points, matching utf8.RuneCountInString and
-- models.MaxEventNameLength. The constraint is added NOT VALID so legacy rows
-- that predate the API cap (over 100 code points or empty) do not fail the
-- migration while every new insert and update is checked.
-- 20260914000000_validate_events_name_length.sql validates the
-- constraint once no legacy rows remain.
ALTER TABLE events
    ADD CONSTRAINT events_name_length
    CHECK (name <> '' AND char_length(name) <= 100) NOT VALID;

-- +goose Down
ALTER TABLE events
    DROP CONSTRAINT events_name_length;
