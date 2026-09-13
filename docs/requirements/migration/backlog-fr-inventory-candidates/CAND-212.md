---
id: CAND-212
title: Secure Production And Staging MongoDB Access
verdict: excluded
related_requirements: []
confidence: confirmed
---

# CAND-212: Secure Production And Staging MongoDB Access

## Source

> Production and staging MongoDB shall disallow unauthenticated access.
>
> - Root user + limited app user
> - Authentication required
> - Credentials stored in corresponding .env files

## Candidate behavior

No durable requirement behavior remains; the source describes authentication for a MongoDB deployment that the system retired.

## Applicability

Actor: operator.
Location: production and staging MongoDB deployments, removed with the MongoDB retirement.
Event kind: not applicable.
Interaction mode: deployment configuration.
Viewport: not applicable.
State: retired store.
Exclusions: every current deployment, which uses PostgreSQL.

## Classification

implementation detail

## Existing Requirements and Confidence

None.
Confidence: confirmed.

## Disposition

Deprecated on 2026-09-11 with the MongoDB retirement; the MongoDB runtime, Compose services, environment variables, and credentials were removed, so no MongoDB access requirement remains.
Reformulated as [QR-014](../../quality/qr/QR-014.md) for staging and production PostgreSQL access.

## Open Questions

None.
