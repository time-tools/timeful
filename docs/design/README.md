# Design Records

This directory is the canonical record of durable Timeful design decisions.
Design records describe architectural choices that constrain implementation, while requirements specify outcomes.

## Layout

- `architecture/README.md` contains architecture-decision-record authoring
  guidance.
- `architecture/adr/` contains architecture decision records (ADRs).
- Future design-record categories belong in their own subdirectory and are
  indexed here.

## Related Artifacts

- Functional requirements (FRs) state required system behavior.
- Quality requirements (QRs) state required quality attributes or runtime
  constraints.
- Specifications (SPECs) describe a feature or domain in broader context and
  may link to multiple requirements.
- Architecture decision records (ADRs) record architectural decisions.
- ADRs may link to the FRs and QRs that they enable, constrain, or satisfy.
  FRs and QRs remain self-contained specifications and must not cite ADRs as
  normative dependencies.
  Requirement provenance is recorded separately from
  the requirement statement.
- Backlog tasks track work to investigate, implement, or verify requirements.

## Architecture decision records

Active records are normative and constrain implementation.
Inactive records are retained for decision history and must not be followed.
The inactive statuses are `deprecated`, `superseded`, and `rejected`.

### Active records

| ID                                     | Scope                             | Status   | Title                                                               |
| -------------------------------------- | --------------------------------- | -------- | ------------------------------------------------------------------- |
| [ADR-001](architecture/adr/ADR-001.md) | Frontend                          | accepted | Frontend Boundary Models and Canonical Internal Shapes              |
| [ADR-002](architecture/adr/ADR-002.md) | Frontend                          | accepted | Frontend Timezone Decoding and Fixed-Offset Boundaries              |
| [ADR-004](architecture/adr/ADR-004.md) | Frontend                          | accepted | Frontend Temporal Runtime Model                                     |
| [ADR-005](architecture/adr/ADR-005.md) | Frontend                          | accepted | Frontend Civil-Date and End-of-Day Model                            |
| [ADR-006](architecture/adr/ADR-006.md) | Frontend                          | accepted | Frontend Temporal Collection Semantics                              |
| [ADR-007](architecture/adr/ADR-007.md) | Frontend                          | accepted | Frontend Semantic Styling Tokens                                    |
| [ADR-008](architecture/adr/ADR-008.md) | Frontend                          | accepted | Frontend Event Ownership Semantics                                  |
| [ADR-009](architecture/adr/ADR-009.md) | Frontend                          | accepted | Frontend Guest Response Ownership Semantics                         |
| [ADR-010](architecture/adr/ADR-010.md) | Frontend, Backend                 | accepted | Source-Confirmed Access Transfers                                   |
| [ADR-018](architecture/adr/ADR-018.md) | Backend, Infrastructure           | accepted | PostgreSQL Is The Single Authoritative Store                        |
| [ADR-020](architecture/adr/ADR-020.md) | Backend, Infrastructure           | accepted | Encrypt Provider Credentials At Rest With Authenticated AES-256-GCM |
| [ADR-021](architecture/adr/ADR-021.md) | Frontend, Backend, Infrastructure | accepted | Use The Platform Identity UUID As The Account Identifier            |

### Inactive records

| ID                                     | Scope                   | Status     | Superseded by                          | Title                                                                                      |
| -------------------------------------- | ----------------------- | ---------- | -------------------------------------- | ------------------------------------------------------------------------------------------ |
| [ADR-003](architecture/adr/ADR-003.md) | Frontend                | deprecated | —                                      | Frontend Freemium Operational Gating                                                       |
| [ADR-011](architecture/adr/ADR-011.md) | Backend, Infrastructure | superseded | [ADR-016](architecture/adr/ADR-016.md) | PostgreSQL Owns Core Records While MongoDB Retains Integrations                            |
| [ADR-012](architecture/adr/ADR-012.md) | Backend, Infrastructure | superseded | [ADR-019](architecture/adr/ADR-019.md) | Resolve Legacy Accounts By external_user_id And Rewrite Migrated Relationships             |
| [ADR-013](architecture/adr/ADR-013.md) | Backend, Infrastructure | deprecated | —                                      | Cut Over By Resumable Backfill With A Short Write Freeze                                   |
| [ADR-014](architecture/adr/ADR-014.md) | Backend, Infrastructure | deprecated | —                                      | Quarantine Legacy Guest Credentials And Ambiguous Ownership                                |
| [ADR-015](architecture/adr/ADR-015.md) | Backend, Infrastructure | superseded | [ADR-020](architecture/adr/ADR-020.md) | Encrypt Provider Credentials At Rest With Authenticated AES-256-GCM                        |
| [ADR-016](architecture/adr/ADR-016.md) | Backend, Infrastructure | superseded | [ADR-018](architecture/adr/ADR-018.md) | PostgreSQL Owns Retained Calendar, OTP, And Daily-Log Data And Friend Requests Are Retired |
| [ADR-017](architecture/adr/ADR-017.md) | Backend, Infrastructure | deprecated | —                                      | Backfill Retained Data On The Migration Ledger And Remove MongoDB In Dependency Order      |
| [ADR-019](architecture/adr/ADR-019.md) | Backend, Infrastructure | superseded | [ADR-021](architecture/adr/ADR-021.md) | Resolve Accounts By External User Identifier                                               |
