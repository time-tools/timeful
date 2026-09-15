---
id: CAND-213
title: Permit Development MongoDB Access Without Authentication
verdict: excluded
related_requirements: []
confidence: confirmed
---

# CAND-213: Permit Development MongoDB Access Without Authentication

## Source

> Dev MongoDB shall allow unauthenticated access.

## Candidate behavior

No durable requirement behavior remains; the source describes unauthenticated access to a development MongoDB deployment that the system retired.

## Applicability

Actor: developer.
Location: development MongoDB deployment, removed with the MongoDB retirement.
Event kind: not applicable.
Interaction mode: local development configuration.
Viewport: not applicable.
State: retired store.
Exclusions: every current deployment, which uses PostgreSQL.

## Classification

implementation detail

## Existing Requirements and Confidence

None.
Confidence: confirmed.

## Disposition

Deprecated on 2026-09-11 with the MongoDB retirement; the development MongoDB service and its unauthenticated configuration were removed, and PostgreSQL is the only supported store.
Not reformulated because development and test PostgreSQL access requires configured credentials, so the unauthenticated-access behavior has no PostgreSQL analog.

## Open Questions

None.
