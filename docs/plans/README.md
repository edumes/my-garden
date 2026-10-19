# My Garden Development Plan

This directory turns the project audit into incremental, reviewable phases. Work through the phases in order unless a later plan explicitly has no unmet dependency.

## Working agreement

- Keep each change small enough to review independently.
- Do not begin a phase while its listed dependencies are incomplete.
- Add or update automated checks with every non-trivial behavior change.
- Treat authorization, money, inventory, and shared access as server-authoritative.
- Do not claim a feature is complete until its acceptance criteria have been executed.
- Keep changes uncommitted until Eduardo reviews them, unless he asks otherwise.

## Roadmap

| Phase | Plan | Outcome | Depends on |
|---|---|---|---|
| 1 | [Engineering baseline](./01-engineering-baseline.md) | Reproducible local and CI validation | None |
| 2 | [Security and integrity](./02-security-and-integrity.md) | Isolated data and atomic state changes | Phase 1 |
| 3 | [Core gameplay loop](./03-core-gameplay-loop.md) | A complete, tested garden progression loop | Phases 1-2 |
| 4 | [Experience redesign](./04-experience-redesign.md) | A coherent, accessible product experience | Phase 3 contracts |
| 5 | [Marketplace](./05-marketplace.md) | Safe trading and delivery flows | Phases 2-4 |
| 6 | [Social and real-time](./06-social-and-realtime.md) | Reliable collaboration and live updates | Phases 2-4 |
| 7 | [Production readiness](./07-production-readiness.md) | Operable deployment for real users | Phases 1-6 |

## Current baseline

Recorded on 2026-10-19 from local execution:

- `go test -timeout 90s ./...`: passes, with tests only in the marketplace package.
- `go vet ./...`: passes.
- `npm ci`: passes.
- `npm run build`: passes; the main JavaScript bundle is about 875 kB before gzip.
- `npm run lint`: fails with 62 errors and 16 warnings.
- `npm audit`: reports 37 vulnerabilities: 28 high, 7 moderate, and 2 low.
- Browser smoke: registration, login, garden creation, and opening the store pass.
- The working branch is local `dev`, created from `main`; the older `origin/dev` was not changed.

## Confirmed high-priority risks

- Authenticated users can read system-wide audit data.
- Store and marketplace state changes are not reliably atomic.
- Shared-garden permissions are defined but inconsistently enforced.
- Share-link usage limits can be exceeded by concurrent requests.
- Production configuration accepts unsafe defaults.
- Documentation advertises routes, scripts, and features that are not implemented.

## Definition of done for a phase

A phase is complete only when:

1. Every required task is complete or explicitly removed from scope.
2. Its automated checks pass from a clean install.
3. Its acceptance flows have been exercised against real local dependencies.
4. Documentation describes the behavior that actually exists.
5. Remaining risks and external actions are recorded.

