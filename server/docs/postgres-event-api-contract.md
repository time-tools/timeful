# PostgreSQL Event API Contract

## Scope

PostgreSQL is the only store for every [Event Kind](../../docs/terminology/glossary.md#event-kind): [Timed Events](../../docs/terminology/glossary.md#timed-event), [Dates-Only Events](../../docs/terminology/glossary.md#dates-only-event), day-of-week events, availability groups, and signup forms, whether an [Anonymous Event Visitor](../../docs/terminology/glossary.md#anonymous-event-visitor) or an [Authenticated Event Visitor](../../docs/terminology/glossary.md#authenticated-event-visitor) creates them.
This document is the API-behavior contract for those records: public identifiers, payload round-trip, authority credentials, response mutation, access transfers, and transaction rules.
[Calendar Connections](../../docs/terminology/glossary.md#calendar-connection), provider tokens, OTP challenges, historical daily user logs, and reporting reads are governed by the [PostgreSQL Data Boundaries](postgres-data-boundaries.md) contract.
`postgres_events` and `postgres_event_responses` are persistence tables, not HTTP DTOs, and route handlers project their records into API responses.
The rules below govern the observable API behavior of every stored event record.

## Identifier Exposure

`postgres_events.id` and `postgres_event_responses.id` are internal UUIDv7 identities.
Event routes address events by the canonical public event identifier, an eight-character Crockford Base32 `short_id`.
The `/api/events/{eventId}/ids` read returns that identifier in both the `shortId` and `longId` fields.
Each [Event Response](../../docs/terminology/glossary.md#event-response) carries an opaque `public_id` that keys response maps and names the mutation target.
The PostgreSQL UUID primary keys remain internal.

## Stored Relations

Event columns hold the public identifier, soft-delete state, name, type, response count, schedule version, creator PostHog ID, ownership association, owner-token hash, and timestamps.
Response columns hold the event relation, an [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) owner, the opaque public identifier, a NOT NULL respondent kind, account identity, and timestamps.
[Platform Visitor Identities](../../docs/terminology/glossary.md#platform-visitor-identity) and **Event Visitor Identities** relate through internal UUID columns, and no external account identifier is stored or referenced.
Signup response `canonical_guest_name` is produced by `respondents.NormalizeGuestName` in Go, and PostgreSQL must not reimplement guest-name normalization; generic response guest names live in the JSONB payload.
The generic-response columns `guest_id`, `canonical_guest_name`, `guest_edit_policy`, `guest_ownership_mode`, and `guest_edit_token` are retained only for the prior server release's rollback window, and the current server neither reads nor writes them.
PostgreSQL permits multiple responses per **Event Visitor Identity**.
Each response's event relation and its owner **Event Visitor Identity** must identify the same event through a composite database constraint.

## JSONB Payloads

`postgres_events.payload` holds all remaining event state, including:

- Description and nullable/default settings.
- Dates, [Active Slots](../../docs/terminology/glossary.md#active-slots), [Event Timezone](../../docs/terminology/glossary.md#event-timezone), slot generation, and timed recurrence.
- The days-only schedule columns `duration`, `dates`, `timeIncrement`, `hasSpecificTimes`, `times`, and `startOnMonday`.
- The selected-schedule snapshot, remindees, attendee-compatible fields, and unsupported-feature fields accepted by current request decoding.

Payload encoding round-trips every persisted field, including the days-only schedule columns.
`models.Event.MarshalJSON` is the persistence encoding and must never drop a stored field.
`models.Event.MarshalAPIJSON` is the API projection: it omits `duration`, `dates`, `timeIncrement`, `hasSpecificTimes`, `times`, and `startOnMonday` for events whose `daysOnly` flag is not true, while [Dates-Only Events](../../docs/terminology/glossary.md#dates-only-event) keep them.

`postgres_event_responses.payload` holds display name, email, availability, if-needed availability, manual availability, and calendar-related fields.

JSON arrays preserve their input behavior: date and recurrence arrays keep input order and duplicates, route validation normalizes active slots, and response availability keeps first-seen order after deduplication.
Instants are normalized to millisecond precision before writing JSONB and before API output.

## Request Semantics

The repository distinguishes absent fields, JSON null, empty arrays/maps, and zero scalar values.
An omitted description preserves the existing value, and an explicit empty description persists.
A timed edit with an explicit empty `activeSlots` retains the existing [Active Slots](../../docs/terminology/glossary.md#active-slots) to match the field-omission behavior of partial payloads.
Saving, replacing, or clearing the [Event Occurrence Span](../../docs/terminology/glossary.md#event-occurrence-span) is owner-only, and archived events reject occurrence-span mutations.

## Event Visitor Identity And Credentials

An [Event Visitor Control Credential (EVCC)](../../docs/terminology/glossary.md#event-visitor-control-credential-evcc) authorizes management of every [Event Response](../../docs/terminology/glossary.md#event-response) owned by its [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) in that event.
The public `eventVisitorId` is an identifier, not proof.
A [Granted Event Visitor Control Credential (Granted EVCC)](../../docs/terminology/glossary.md#granted-event-visitor-control-credential-granted-evcc) is a distinct, source-revocable delegated credential.
Neither credential value is exposed to application JavaScript.

Event creation establishes an event-scoped [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) and returns its public `eventVisitorId` with the created event.
Creation issues the private EVCC as an HttpOnly, SameSite=Lax cookie scoped to `/api`, and the credential value never reaches application JavaScript.
The browser retains `eventVisitorId` per event in localStorage so the identity survives reloads and sign-out, while authority always comes from the EVCC cookie or an authenticated [Platform Visitor Identity](../../docs/terminology/glossary.md#platform-visitor-identity) session.

Signing in associates known browser [Event Visitor Identities](../../docs/terminology/glossary.md#event-visitor-identity) with the private [Platform Visitor Identity](../../docs/terminology/glossary.md#platform-visitor-identity) through `POST /auth/visitor-identities`.
Association never grants authority by itself, and response authorization still requires the source EVCC or the associated **Platform Visitor Identity** session.

Event creation records the creator's [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) separately from the event's ownership association.

## Response Mutation

PostgreSQL response maps use opaque response IDs.
The API defines explicit creation, selected-response, update, and deletion contracts that remain valid when one [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) owns multiple responses.
When [Blind Availability Mode](../../docs/terminology/glossary.md#blind-availability-mode) is enabled, a read exposes all responses only with [Event Owner](../../docs/terminology/glossary.md#event-owner) authority, and other visitors see only responses they are authorized to manage, with other-response counts omitted.

Response mutation uses an explicit-selection contract: `createResponse: true` creates a new response for the calling [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity), and every edit, deletion, and rename must carry the target `responseId`; otherwise the route rejects with `select-response-or-explicitly-create`.
Each response-map entry carries `canEdit` so the client can offer exactly the edits the server will honor.
The browser plugin `set-slots` wire contract is unchanged; the frontend maps the plugin's named response onto the explicit-selection contract (existing named response, else selected response, else create) before calling the API, and the respondents-list delete submits the same `responseId` contract.

## Event Owner Authority

PostgreSQL creation issues a distinct [Event Owner Edit Token](../../docs/terminology/glossary.md#event-owner-edit-token) in an HttpOnly, SameSite=Lax cookie scoped to `/api`, with Secure enabled for HTTPS requests.
Only its SHA-256 hash is stored, and the credential value never reaches application JavaScript.
The token authorizes [Event Settings](../../docs/terminology/glossary.md#event-settings) edits, [Event Occurrence Span](../../docs/terminology/glossary.md#event-occurrence-span) save, replace, and clear, archive/unarchive, and deletion, but it does not authorize [Event Response](../../docs/terminology/glossary.md#event-response) edits.
Base [EVCCs](../../docs/terminology/glossary.md#event-visitor-control-credential-evcc) never authorize these owner actions, including the creator's credential.

Ownership has its own [Platform Visitor Identity](../../docs/terminology/glossary.md#platform-visitor-identity) association, separate from the creator's [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) and [Event Responses](../../docs/terminology/glossary.md#event-response).
An associated account can manage the event without the original cookie.
Proving the [Event Owner Edit Token](../../docs/terminology/glossary.md#event-owner-edit-token) while signed in associates ownership with that account, replacing any previous ownership association without transferring response ownership.
Ownership takeover and protected mutations serialize under the event row lock.

Event reads expose server-proven `canEditSettings` and `canManageEvent` capabilities for frontend controls.
Archived events remain readable and allow authorized unarchive or deletion, but reject settings, response, rename, and occurrence-span mutations.
Deleted events and their responses stop resolving through event routes.

The credential schema and validator distinguish an owner-issued [Granted EVCC](../../docs/terminology/glossary.md#granted-event-visitor-control-credential-granted-evcc) through explicit credential-kind and owner-grant metadata, and reject revoked grants.

An event with no recoverable [Event Owner Edit Token](../../docs/terminology/glossary.md#event-owner-edit-token) and no ownership association cannot have its settings, archive state, or deletion managed.

## Access Transfers

[Access Transfers](../../docs/terminology/glossary.md#access-transfer) are source-confirmed browser-to-browser processes that delegate event authority or establish a platform session on another browser.
A source creates a transfer link and approves the target browser's matching code within five minutes of creating the link.
Opening the link alone grants no access, and each browser opening it receives an independent code.
The source can cancel a pending or approved-but-unredeemed link or create a new link after expiry.

A signed-in source creates a normal session for the same [Platform Visitor Identity](../../docs/terminology/glossary.md#platform-visitor-identity) on the target.
Redemption replaces only the target session's identity and preserves its unrelated session keys.
An anonymous source must prove its [EVCC](../../docs/terminology/glossary.md#event-visitor-control-credential-evcc), and an anonymous [Event Owner](../../docs/terminology/glossary.md#event-owner) must also prove the [Event Owner Edit Token](../../docs/terminology/glossary.md#event-owner-edit-token).
The target receives a distinct [Granted EVCC](../../docs/terminology/glossary.md#granted-event-visitor-control-credential-granted-evcc), preserving the source role and ownership while retaining the target's own [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity).
Owner grants permit settings edits, visibility of all responses, archive/unarchive, and deletion.
Ordinary grants permit only the source's response authority, including the same privacy restrictions in [Blind Availability Mode](../../docs/terminology/glossary.md#blind-availability-mode).

The source retains revocation handles across reloads and can revoke issued anonymous grants.
Grants have no fixed server-side expiry; clearing target browser data or source revocation removes that browser's delegated authority.
Expired pending, approved, and cancelled transfers are pruned automatically together with their requests, while redeemed transfers are retained as revocation anchors.
Normal signed-in sessions use the ordinary session lifecycle and do not offer grant revocation.
When the target signs in, the app asks before associating the source [Event Visitor Identity](../../docs/terminology/glossary.md#event-visitor-identity) with its [Platform Visitor Identity](../../docs/terminology/glossary.md#platform-visitor-identity), including sign-in from outside the event page.
`Not now` leaves the grant usable without associating the source; `Confirm association` enables durable response recovery without associating event ownership.
Explicitly accepted account recovery is independent of later grant revocation.

All paths below are relative to `/api`.

| Request                                                 | Contract                                                                                                                                                                                                                                                       |
| ------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `POST /events/{eventId}/transfers`                      | Requires source authority and returns `id` and `expiresAt`; stores only hashed source proof.                                                                                                                                                                   |
| `POST /events/{eventId}/transfers/{transferId}/open`    | Empty JSON object creates or retrieves this browser's independent `requestId` and `code`, without issuing event authority; after approval but before redemption it re-serves the approved request and code to the browser holding that request's target proof. |
| `POST /events/{eventId}/transfers/{transferId}/status`  | Source proof returns `state`, `requests`, and `revocable`; target requests cannot inspect source status.                                                                                                                                                       |
| `POST /events/{eventId}/transfers/{transferId}/approve` | Source proof plus exact `{requestId, code}` selects one target and consumes the pending state.                                                                                                                                                                 |
| `POST /events/{eventId}/transfers/{transferId}/redeem`  | Only the approved target proof can redeem once before the original deadline.                                                                                                                                                                                   |
| `POST /events/{eventId}/transfers/{transferId}/cancel`  | Source proof cancels a pending or approved-but-unredeemed transfer; cancelled transfers reject approval and redemption.                                                                                                                                        |
| `POST /events/{eventId}/transfers/{transferId}/revoke`  | Source proof revokes the issued grant without a transfer deadline.                                                                                                                                                                                             |
| `POST /events/{eventId}/grant-association`              | `{confirm: false}` inspects consent requirements; only explicit `{confirm: true}` with an active grant and authenticated session associates the source response identity.                                                                                      |

Source proofs, target proofs, and anonymous grants use separate HttpOnly, SameSite=Lax cookies scoped to `/api`, with Secure enabled over HTTPS.
Raw source credentials and owner tokens are never copied to the target or exposed to JavaScript.
Lifecycle mutations serialize under event and transfer row locks, preserving single redemption under concurrent requests.
Session encoding occurs before committing redemption, and failed transactions discard session cookie headers.
Expired, cancelled, mismatched, unapproved, reused, cross-event, unauthorized, and revoked credentials or transfers are rejected.

## Transactions

Response create, update, and delete lock the event row and update the response row plus `num_responses` in one transaction.
Guest rename and signup capacity enforcement are also transactional.
Unique-index conflicts map to the existing duplicate-name route error.
Event edit and selected schedule replace/clear write the event aggregate atomically.
Transactions prevent duplicate response races and response-count drift.
