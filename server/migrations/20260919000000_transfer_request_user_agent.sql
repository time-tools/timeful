-- +goose Up
-- Record the User-Agent each target browser presents when it opens an access
-- transfer link. The source browser uses it to describe the target that
-- received granted access and the target that is showing a matching code.
-- The client treats an empty or unparseable value as an unknown browser, so
-- legacy request rows and header-less clients keep the default.
ALTER TABLE access_transfer_requests
    ADD COLUMN user_agent TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE access_transfer_requests
    DROP COLUMN user_agent;
