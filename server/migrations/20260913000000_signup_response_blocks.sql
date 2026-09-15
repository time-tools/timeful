-- +goose Up
-- Signup block membership becomes a relational join table so foreign keys can
-- prove a claimed block exists. The retired block_ids array could not carry the
-- relation, which forced repository-side validation and per-block capacity
-- counts.
CREATE TABLE event_signup_response_blocks (
    response_id UUID NOT NULL REFERENCES event_signup_responses(id) ON DELETE CASCADE,
    block_id UUID NOT NULL REFERENCES event_signup_blocks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    CONSTRAINT event_signup_response_blocks_unique UNIQUE (response_id, block_id)
);

-- The (block_id, response_id) index serves grouped capacity counts by block and
-- keeps the block foreign-key cascade cheap. The (response_id, position,
-- block_id) index serves the ordered membership aggregate and keeps the
-- response foreign-key cascade cheap; the unique constraint backstops a
-- repeated claim.
CREATE INDEX event_signup_response_blocks_block_id_idx
    ON event_signup_response_blocks (block_id, response_id);

CREATE INDEX event_signup_response_blocks_response_position_idx
    ON event_signup_response_blocks (response_id, position, block_id);

-- Backfill in stored array order. DISTINCT ON keeps the first occurrence of a
-- repeated block, the join skips values that do not name a block of the same
-- event, and invalid or dangling text never reaches a uuid cast because only the
-- stored block key is cast to text.
INSERT INTO event_signup_response_blocks (response_id, block_id, position)
SELECT DISTINCT ON (response.id, block.id) response.id, block.id, claimed.ordinality
FROM event_signup_responses response
CROSS JOIN LATERAL unnest(response.block_ids) WITH ORDINALITY AS claimed(block_text, ordinality)
JOIN event_signup_blocks block
    ON block.id::text = claimed.block_text AND block.event_id = response.event_id
ORDER BY response.id, block.id, claimed.ordinality;

ALTER TABLE event_signup_responses DROP COLUMN block_ids;

-- +goose Down
-- Rebuild the retired array column from the join table so the down migration
-- preserves membership content and order.
ALTER TABLE event_signup_responses ADD COLUMN block_ids TEXT[] NOT NULL DEFAULT '{}'::text[];

UPDATE event_signup_responses response
SET block_ids = COALESCE((
    SELECT array_agg(membership.block_id::text ORDER BY membership.position)
    FROM event_signup_response_blocks membership
    WHERE membership.response_id = response.id
), '{}'::text[]);

DROP TABLE event_signup_response_blocks;
