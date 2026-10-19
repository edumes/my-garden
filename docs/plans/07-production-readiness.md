# Phase 7 — Production Readiness

## Objective

Prepare the complete product for real users with controlled releases, recovery, and observable behavior.

## Dependencies

- Product scope for the first public release is fixed.
- Earlier phase checks are green and representative of production behavior.

## Scope

### 1. Database lifecycle

- Replace startup `AutoMigrate` with reviewed, versioned migrations.
- Test upgrades from the last supported schema and forward-compatible rollback behavior.
- Define backup, restore, retention, and recovery objectives.
- Add indexes from measured query patterns.

### 2. Runtime safety

- Separate liveness from dependency-aware readiness.
- Configure connection pools, timeouts, request limits, and graceful shutdown.
- Share rate-limit counters across instances; today each instance keeps its own in memory.
- Prevent multiple game-engine workers from applying the same scheduled work.
- Run production with secure database transport and managed secrets.

### 3. Observability

- Add structured logs with request and operation identifiers.
- Add metrics for latency, errors, database pools, game ticks, queues, and WebSockets.
- Add tracing only where logs and metrics cannot explain cross-service latency.
- Redact credentials, tokens, private keys, and sensitive request content.

### 4. Release process

- Build immutable frontend and backend artifacts.
- Run migration compatibility checks before deployment.
- Add smoke tests against the deployed commit.
- Define forward-only recovery and rollback constraints.

### 5. Account and privacy controls

- Implement password change, account recovery, session revocation, and deletion before advertising them.
- Define audit-log access, retention, and deletion exceptions.
- Document privacy expectations for shared gardens and public marketplace data.

### 6. Operational verification

- Load-test the core loop and marketplace contention points.
- Exercise database loss, Redis loss, restart, and degraded dependency scenarios.
- Verify backup restoration in an isolated environment.
- Record external infrastructure evidence separately from local test evidence.

## Validation

- A clean environment can be deployed from documented automation.
- The deployed commit is observable and identifiable through readiness and smoke checks.
- Restore and failure drills meet the declared recovery objectives.
- Security-sensitive configuration fails closed.

## Risks

- Local success does not prove provider capacity, backup restoration, or network policy.
- Rollback may be unsafe after destructive schema changes.
- Operational controls added before product scope stabilizes can become wasted complexity.

## Exit criteria

- The release process is repeatable, observable, and recoverable.
- Remaining launch actions identify their owner and required external evidence.

