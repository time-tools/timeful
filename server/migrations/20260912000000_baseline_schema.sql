-- +goose Up
-- Baseline schema for a fresh PostgreSQL 18 database. This migration replaces
-- the retired MongoDB transition chain and the account identity cutover chain:
-- it creates the complete consolidated schema the current server requires in
-- one step. The retired migration tooling tables (migration_ledger,
-- migration_quarantine) are intentionally not created, and the retired account
-- identifier columns do not exist. Future schema changes are added as new
-- incremental goose migrations.
--
-- Physical column order follows the cutover chain's final schema so a schema
-- comparison against a database migrated by that chain is exact.

-- The platform identity is the account's sole identifier: its uuid is the
-- account reference everywhere, and no external account identifier is stored.
CREATE TABLE platform_identities (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Events: public identifiers remain storage-opaque. public_id was retired in
-- favor of the canonical short identifier, and ownership association is
-- separate from the creator's response identity. The owner is a platform
-- identity uuid.
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    short_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    num_responses INTEGER NOT NULL DEFAULT 0 CHECK (num_responses >= 0),
    schedule_version INTEGER NOT NULL DEFAULT 1,
    creator_posthog_id TEXT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    owner_event_visitor_identity_id UUID,
    owner_edit_token_hash BYTEA CHECK (octet_length(owner_edit_token_hash) = 32),
    owner_platform_identity_id UUID REFERENCES platform_identities(id),
    CONSTRAINT events_short_id_format CHECK (
        short_id ~ '^[0-9A-HJKMNPQRSTVWXYZ]{8}$'
    ),
    CONSTRAINT events_type CHECK (type IN ('specific_dates', 'dow', 'signup', 'group')),
    CONSTRAINT events_payload_object CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX events_active_creator_posthog_id_idx
    ON events (creator_posthog_id, created_at DESC)
    WHERE NOT is_deleted;

CREATE INDEX events_owner_platform_idx
    ON events (owner_platform_identity_id);

CREATE TABLE event_visitor_identities (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    platform_identity_id UUID REFERENCES platform_identities(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (event_id, id)
);

CREATE INDEX event_visitor_platform_idx
    ON event_visitor_identities (platform_identity_id, event_id);

CREATE TABLE event_visitor_credentials (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    event_visitor_identity_id UUID NOT NULL REFERENCES event_visitor_identities(id) ON DELETE CASCADE,
    credential_hash BYTEA NOT NULL CHECK (octet_length(credential_hash) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    revoked_at TIMESTAMPTZ,
    kind TEXT NOT NULL DEFAULT 'base' CHECK (kind IN ('base', 'granted')),
    grants_owner BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT owner_grant_requires_granted CHECK (NOT grants_owner OR kind = 'granted')
);

CREATE INDEX event_visitor_credentials_visitor_idx
    ON event_visitor_credentials (event_visitor_identity_id);

-- Ownership association is a separate relation from the creator's response
-- identity, so it is added after the visitor identity table exists.
ALTER TABLE events
    ADD CONSTRAINT events_owner_visitor_fk
    FOREIGN KEY (id, owner_event_visitor_identity_id) REFERENCES event_visitor_identities(event_id, id);

-- Responses are owned by an Event Visitor Identity, and the account reference
-- is the platform identity uuid. Respondent identity columns and the retained
-- credential column stay for runtime compatibility.
CREATE TABLE event_responses (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    respondent_kind TEXT NULL,
    guest_id TEXT NULL,
    canonical_guest_name TEXT NULL,
    guest_edit_policy TEXT NULL,
    guest_ownership_mode TEXT NULL,
    -- Retained temporarily so tokenless open mutations can return credentials.
    guest_edit_token TEXT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    public_id UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    event_visitor_identity_id UUID NOT NULL,
    platform_identity_id UUID REFERENCES platform_identities(id),
    CONSTRAINT event_responses_kind CHECK (
        respondent_kind IN ('account', 'guest')
    ),
    CONSTRAINT event_responses_payload_object CHECK (jsonb_typeof(payload) = 'object'),
    CONSTRAINT event_responses_visitor_event_fk
        FOREIGN KEY (event_id, event_visitor_identity_id)
        REFERENCES event_visitor_identities(event_id, id)
);

CREATE INDEX event_responses_event_id_idx
    ON event_responses (event_id);

CREATE INDEX event_responses_visitor_idx
    ON event_responses (event_visitor_identity_id);

CREATE INDEX event_responses_platform_identity_idx
    ON event_responses (platform_identity_id);

-- The source account reference is the platform identity uuid while the
-- credential-based path keeps source_credential_id.
CREATE TABLE access_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    source_hash BYTEA NOT NULL CHECK (octet_length(source_hash) = 32),
    source_credential_id UUID REFERENCES event_visitor_credentials(id),
    grants_owner BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (clock_timestamp() + interval '5 minutes'),
    state TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'approved', 'redeemed', 'cancelled')),
    approved_request_id UUID,
    grant_id UUID REFERENCES event_visitor_credentials(id),
    platform_identity_id UUID REFERENCES platform_identities(id),
    CONSTRAINT access_transfers_source_xor_platform_identity
        CHECK ((source_credential_id IS NULL) <> (platform_identity_id IS NULL))
);

CREATE INDEX access_transfers_event_idx ON access_transfers (event_id);

-- Redeemed transfers are retained as revocation anchors, so pruning skips them.
CREATE INDEX access_transfers_prune_idx
    ON access_transfers (expires_at)
    WHERE state IN ('pending', 'approved', 'cancelled');

CREATE TABLE access_transfer_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES access_transfers(id) ON DELETE CASCADE,
    target_hash BYTEA NOT NULL CHECK (octet_length(target_hash) = 32),
    code TEXT NOT NULL,
    UNIQUE (transfer_id, code)
);

ALTER TABLE access_transfers
    ADD CONSTRAINT access_transfers_approved_request_fkey
    FOREIGN KEY (approved_request_id) REFERENCES access_transfer_requests(id);

-- Accounts are the sole authority for account identity and profile. Email is
-- deliberately not unique: distinct accounts stay distinct even when emails
-- compare equal.
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    platform_identity_id UUID NOT NULL UNIQUE REFERENCES platform_identities(id),
    email TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    picture TEXT NOT NULL DEFAULT '',
    has_custom_name BOOLEAN NULL,
    timezone_offset INTEGER NOT NULL DEFAULT 0,
    num_events_created INTEGER NOT NULL DEFAULT 0 CHECK (num_events_created >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX accounts_email_lower_idx ON accounts (lower(email));

-- Account deletion is recorded as a tombstone so a concurrent or repeated
-- account backfill cannot recreate a deleted account. The tombstone carries the
-- platform identity uuid with no foreign key so it survives identity deletion.
CREATE TABLE account_deletion_tombstones (
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    platform_identity_id UUID PRIMARY KEY
);

-- Folders are account-scoped organization records owned by a platform identity.
CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL DEFAULT '',
    color TEXT NULL,
    -- NULL preserves an omitted isDeleted, distinct from an explicit false.
    is_deleted BOOLEAN NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    platform_identity_id UUID NOT NULL REFERENCES platform_identities(id)
);

CREATE INDEX folders_platform_identity_id_idx ON folders (platform_identity_id);

-- A membership holds exactly one event reference and belongs to one
-- account.
CREATE TABLE folder_events (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    platform_identity_id UUID NOT NULL REFERENCES platform_identities(id)
);

CREATE INDEX folder_events_folder_id_idx ON folder_events (folder_id);

-- One folder per account per event.
CREATE UNIQUE INDEX folder_events_event_unique_idx
    ON folder_events (platform_identity_id, event_id)
    WHERE event_id IS NOT NULL;

-- A signup block is an ordered, capacity-limited slot on a signup form event.
-- capacity is NULL when unlimited; zero capacity rejects every signup.
CREATE TABLE event_signup_blocks (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    capacity INTEGER NULL CHECK (capacity IS NULL OR capacity >= 0),
    start_date TIMESTAMPTZ NULL,
    end_date TIMESTAMPTZ NULL,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX event_signup_blocks_event_id_idx
    ON event_signup_blocks (event_id, position, created_at, id);

-- A signup response is owned by an Event Visitor Identity and its account
-- reference is the platform identity uuid. block_ids holds the claimed
-- event_signup_blocks identities as text because a PostgreSQL array cannot
-- carry a foreign key; membership is validated in the repository.
CREATE TABLE event_signup_responses (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    event_visitor_identity_id UUID NOT NULL,
    respondent_kind TEXT NOT NULL CHECK (respondent_kind IN ('account', 'guest')),
    canonical_guest_name TEXT NULL,
    name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    block_ids TEXT[] NOT NULL DEFAULT '{}'::text[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    platform_identity_id UUID REFERENCES platform_identities(id),
    CONSTRAINT event_signup_responses_visitor_event_fk
        FOREIGN KEY (event_id, event_visitor_identity_id)
        REFERENCES event_visitor_identities(event_id, id) ON DELETE CASCADE,
    CONSTRAINT event_signup_responses_identity CHECK (
        (respondent_kind = 'account' AND platform_identity_id IS NOT NULL)
        OR
        (respondent_kind = 'guest' AND platform_identity_id IS NULL AND canonical_guest_name IS NOT NULL AND canonical_guest_name <> '')
    )
);

CREATE INDEX event_signup_responses_event_id_idx
    ON event_signup_responses (event_id, created_at, id);

CREATE UNIQUE INDEX event_signup_responses_account_unique_idx
    ON event_signup_responses (event_id, platform_identity_id)
    WHERE respondent_kind = 'account';

CREATE UNIQUE INDEX event_signup_responses_guest_name_unique_idx
    ON event_signup_responses (event_id, canonical_guest_name)
    WHERE respondent_kind = 'guest';

-- An attendee is one group invitation keyed by the invited email address. The
-- platform identity reference is released (set NULL) when that account is
-- deleted so the email-keyed membership survives. declined is NULL when
-- omitted.
CREATE TABLE event_attendees (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    email TEXT NOT NULL CHECK (email <> ''),
    declined BOOLEAN NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    platform_identity_id UUID REFERENCES platform_identities(id)
);

CREATE UNIQUE INDEX event_attendees_event_email_unique_idx
    ON event_attendees (event_id, email);

CREATE INDEX event_attendees_platform_identity_id_idx
    ON event_attendees (platform_identity_id);

-- Each calendar connection is owned by exactly one platform identity and keeps
-- its legacy email_CALENDARTYPE map key as calendar_key.
CREATE TABLE calendar_accounts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    platform_identity_id UUID NOT NULL REFERENCES platform_identities(id) ON DELETE CASCADE,
    calendar_key TEXT NOT NULL CHECK (calendar_key <> ''),
    calendar_type TEXT NOT NULL CHECK (calendar_type IN ('google', 'outlook', 'apple', 'ics')),
    email TEXT NOT NULL DEFAULT '',
    picture TEXT NOT NULL DEFAULT '',
    -- NULL preserves an omitted enabled, distinct from an explicit false.
    enabled BOOLEAN NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (platform_identity_id, calendar_key)
);

CREATE INDEX calendar_accounts_platform_identity_id_idx
    ON calendar_accounts (platform_identity_id);

-- A sub-calendar receives a fresh identity scoped to its connection and keeps
-- the provider calendar id as its runtime key.
CREATE TABLE calendar_sub_calendars (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    calendar_account_id UUID NOT NULL REFERENCES calendar_accounts(id) ON DELETE CASCADE,
    sub_calendar_id TEXT NOT NULL CHECK (sub_calendar_id <> ''),
    name TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (calendar_account_id, sub_calendar_id)
);

-- Exactly one credential row per connection. Provider secrets are encrypted at
-- rest with the versioned AES-256-GCM envelope; a NULL ciphertext means the
-- credential was absent, never an empty secret.
CREATE TABLE calendar_account_credentials (
    calendar_account_id UUID PRIMARY KEY REFERENCES calendar_accounts(id) ON DELETE CASCADE,
    oauth_access_token_ciphertext TEXT NULL,
    oauth_refresh_token_ciphertext TEXT NULL,
    oauth_access_token_expires_at TIMESTAMPTZ NULL,
    oauth_scope TEXT NULL,
    apple_password_ciphertext TEXT NULL,
    ics_feed_url_ciphertext TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Exactly one preference row per platform identity. primary_account_key NULL
-- preserves the first-account fallback; token_origin NULL means undefined.
CREATE TABLE calendar_preferences (
    platform_identity_id UUID PRIMARY KEY REFERENCES platform_identities(id) ON DELETE CASCADE,
    primary_account_key TEXT NULL CHECK (primary_account_key IS NULL OR primary_account_key <> ''),
    token_origin TEXT NULL CHECK (token_origin IS NULL OR token_origin IN ('ios', 'android', 'web')),
    calendar_options JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Each OTP challenge is keyed by email, so at most one challenge is active per
-- email and a new send replaces the existing one. The one-time code is stored
-- only as a salted one-way hash.
CREATE TABLE otp_challenges (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email TEXT NOT NULL UNIQUE CHECK (email <> ''),
    code_hash TEXT NOT NULL CHECK (code_hash <> ''),
    expires_at TIMESTAMPTZ NOT NULL,
    -- Preserves the five-attempt lockout, counted before each compare.
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Expired challenges are swept explicitly; the index keeps the sweep and the
-- verification predicate bounded.
CREATE INDEX otp_challenges_expires_at_idx ON otp_challenges (expires_at);

-- Historical daily user logs: exactly one row per account-local date. The
-- denormalized user array is never stored; profiles are rebuilt at read time.
CREATE TABLE daily_user_logs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    log_date DATE NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- One membership per account per log. Account identity is the platform
-- identity uuid, and first_seen_position preserves insertion order so
-- reporting lists accounts in the order they first signed in.
CREATE TABLE daily_user_log_members (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    daily_user_log_id UUID NOT NULL REFERENCES daily_user_logs(id) ON DELETE CASCADE,
    first_seen_position INTEGER NOT NULL CHECK (first_seen_position >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    platform_identity_id UUID NOT NULL REFERENCES platform_identities(id),
    CONSTRAINT daily_user_log_members_log_platform_identity_key
        UNIQUE (daily_user_log_id, platform_identity_id)
);

CREATE INDEX daily_user_log_members_platform_identity_id_idx
    ON daily_user_log_members (platform_identity_id);

-- +goose Down
-- A baseline cannot be reversed: dropping it would destroy every record and
-- leave no migration history to restore. Refuse instead of pretending to roll
-- back.
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'The baseline schema migration cannot be reversed';
END $$;
-- +goose StatementEnd
