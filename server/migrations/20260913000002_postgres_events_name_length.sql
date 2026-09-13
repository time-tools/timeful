-- +goose Up
-- Enforce the FR-119 event-name cap at the storage boundary. char_length
-- counts Unicode code points, matching utf8.RuneCountInString and
-- models.MaxEventNameLength. The constraint is added NOT VALID so legacy rows
-- that predate the API cap (over 100 code points or empty) do not fail the
-- migration while every new insert and update is checked. After legacy data is
-- audited and cleaned, run:
--   ALTER TABLE postgres_events VALIDATE CONSTRAINT postgres_events_name_length;
ALTER TABLE postgres_events
    ADD CONSTRAINT postgres_events_name_length
    CHECK (name <> '' AND char_length(name) <= 100) NOT VALID;

-- +goose Down
ALTER TABLE postgres_events
    DROP CONSTRAINT postgres_events_name_length;
