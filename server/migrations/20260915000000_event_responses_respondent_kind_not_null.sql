-- +goose Up
-- respondent_kind on event_responses was nullable for rows that
-- predated the consolidated baseline, and readers coerced a NULL kind to an
-- empty string. Every baseline-created database starts without such rows and
-- every writer sets 'account' or 'guest', so the column becomes explicitly NOT
-- NULL and the coercion is retired.
--
-- The legacy guest columns on this table (guest_id, canonical_guest_name,
-- guest_edit_policy, guest_ownership_mode, guest_edit_token) are intentionally
-- not dropped here: the prior PostgreSQL-aware server release still names them
-- in its queries, so they stay for the documented rollback window.
ALTER TABLE event_responses
    ALTER COLUMN respondent_kind SET NOT NULL;

-- +goose Down
ALTER TABLE event_responses
    ALTER COLUMN respondent_kind DROP NOT NULL;
