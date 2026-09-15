-- +goose Up
-- The legacy rows that forced the NOT VALID guard are gone: TASK-0224.01 found
-- no live legacy rows, and a database that predates the baseline is recreated
-- from it instead of upgraded. Validate the constraint added NOT VALID by
-- 20260913000002_events_name_length.sql so PostgreSQL also trusts it
-- for plan-time checks, not only for new inserts and updates.
ALTER TABLE events
    VALIDATE CONSTRAINT events_name_length;

-- +goose Down
-- PostgreSQL has no unvalidate operation, so restore the state left by
-- 20260913000002_events_name_length.sql by replacing the constraint
-- with a NOT VALID one of the same definition.
ALTER TABLE events
    DROP CONSTRAINT events_name_length;

ALTER TABLE events
    ADD CONSTRAINT events_name_length
    CHECK (name <> '' AND char_length(name) <= 100) NOT VALID;
