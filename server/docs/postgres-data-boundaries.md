# PostgreSQL Data Boundaries

## Scope

PostgreSQL is the only store for accounts, profiles, events of every kind, responses, attendees, signup data, availability groups, folders, folder membership, calendar integrations, OTP challenges, and historical daily user logs.
This document is the durable backend contract for store ownership, account identity, calendar identities and keys, OTP handling, the provider-credential encryption boundary, and reporting reads.
The observable API behavior of PostgreSQL-owned events is governed by the [PostgreSQL Event API Contract](postgres-event-api-contract.md).
The durable decisions are recorded as [ADR-018](../../docs/design/architecture/adr/ADR-018.md), [ADR-020](../../docs/design/architecture/adr/ADR-020.md), and [ADR-021](../../docs/design/architecture/adr/ADR-021.md).

## Authoritative Store Ownership

Every record has exactly one authoritative store, and PostgreSQL is that store for every record kind.
A record is authoritative in the store that accepts its reads and writes, and no other store may be consulted for that record or re-authorize it.
Every record is authored in exactly one store, and no record kind has a second read path.

| Record kind                                                                                                                   | Authoritative store      | Notes                                                                                       |
| ----------------------------------------------------------------------------------------------------------------------------- | ------------------------ | ------------------------------------------------------------------------------------------- |
| [Calendar Connection](../../docs/terminology/glossary.md#calendar-connection), provider credentials, and calendar preferences | PostgreSQL               | The whole integration aggregate lives together per account.                                 |
| Sub-calendar and its enabled state                                                                                            | PostgreSQL               | Child of its [Calendar Connection](../../docs/terminology/glossary.md#calendar-connection). |
| OTP challenge                                                                                                                 | PostgreSQL               | Ephemeral; one active challenge per email.                                                  |
| Historical daily user log and its account membership                                                                          | PostgreSQL               | Membership references an authoritative account.                                             |
| Event-creator analytics                                                                                                       | PostgreSQL event storage | Reads `events.creator_posthog_id`.                                                          |
| Active-user and signed-up-user reporting                                                                                      | PostgreSQL               | Reads `daily_user_logs`, `daily_user_log_members`, and `accounts`.                          |
| Core account, event, response, attendee, signup, group, and folder data                                                       | PostgreSQL               | Governed by the event API contract and the PostgreSQL schema.                               |

## Account Identity

`platform_identities.id` is the account's sole identifier and is a native UUIDv7.
Every account reference is a native `uuid` column referencing `platform_identities(id)`, and the sign-in session carries the canonical lowercase hyphenated UUID string.
Account deletion records that uuid in `account_deletion_tombstones` without a foreign key, so the tombstone survives the platform identity's deletion.
The tombstone is an audit record rather than a runtime gate, because deletion removes the `platform_identities` row that account resolution keys on and a later sign-in mints a fresh UUIDv7 identity, so a deleted identity can never be adopted again.
The all-zero UUID is the wire representation of an absent account identity, such as a guest response's `userId` or an unowned event's `ownerId`, and it is never a stored platform identity.

## Calendar Identities And Keys

Each [Calendar Connection](../../docs/terminology/glossary.md#calendar-connection) receives a UUIDv7 `calendar_accounts.id` and is owned by exactly one `platform_identities` row.
Its runtime key is the `email_calendarType` map key: email-like identifiers are trimmed, normalized, and lowercased, ICS identifiers are trimmed without email normalization, and the key ends with `_<calendarType>`.
Connection lookup recomputes that canonical key, and every writer stores it, so a connection always resolves through its canonical key.
Sub-calendars receive UUIDv7 identities scoped to their **Calendar Connection**, and each sub-calendar keeps its provider identifier as its runtime key.
The connection, its encrypted credentials, its sub-calendars, and the calendar preferences are read and written only in PostgreSQL.

## OTP Challenges

Each OTP challenge receives a UUIDv7 identity and is keyed by email.
A challenge grants no account authority: a successful verification resolves the authoritative PostgreSQL account.
At most one challenge is active per email, and sending a new challenge replaces any existing challenge for that email.
The one-time code is stored as a salted one-way hash, never in plaintext, and verification compares hashes.
The ten-minute expiry and the five-attempt lockout apply.

## Daily Logs

Each daily user log is keyed by its account-local date, and each membership references an authoritative account.
Membership is idempotent per account per day and preserves first-seen order.
The log and its memberships are read and written only in PostgreSQL.

## Credential Encryption Boundary

Provider credentials are encrypted at rest with authenticated AES-256-GCM and an envelope that carries a format version.

- Algorithm: AES-256 in Galois/Counter Mode with a 128-bit authentication tag.
- Key source: the `ENCRYPTION_KEY` environment variable, read as exactly 32 raw bytes; startup validation refuses a missing or wrong-length key rather than deriving one, and the key is never logged.
- Nonce: a fresh 96-bit random nonce per encryption, never reused for a key.
- Envelope: the versioned string `v1:` followed by the base64 encoding of `nonce || ciphertext || tag`, so a future key or algorithm change is distinguishable without ambiguity.
- Encrypted fields: the OAuth2 access token, the OAuth2 refresh token, the Apple app password, and the ICS feed URL.
- Plaintext fields: access-token expiry, granted OAuth2 scope, email, picture, enabled state, sub-calendar name and enabled state, primary calendar preference, token origin, and calendar options; none of these is a provider secret.
- OTP one-time codes are not provider credentials and are not stored with this envelope; they are stored as salted one-way hashes so the code is not readable at rest.

Decryption failures are errors, never empty credentials.
A failed decrypt of an access token fails the token refresh for that connection and is reported; it never submits an empty credential.
A failed decrypt of a refresh token, Apple password, or ICS feed URL surfaces as an error to the caller and does not partially load a connection.

Credentials are never returned through the API.
No calendar route may expose a token, password, or feed URL in a response.

## Reporting Reads

Event-creator analytics aggregate `events.creator_posthog_id`, counting each event exactly once.
Active-user reporting reads `daily_user_logs` and `daily_user_log_members`, and signed-up-user reporting reads `accounts`.
Reporting reads only PostgreSQL records.
