# Phase 6 — Social and Real-Time

## Objective

Enable collaboration and timely feedback without making transient connections authoritative for durable game state.

## Dependencies

- Phase 2 permission rules are complete.
- Phase 3 actions are idempotent and server-authoritative.
- Phase 4 provides accessible shared-garden interactions.

## Scope

### 1. Collaboration

- Complete owner controls for links, members, permissions, expiration, and removal.
- Show collaborators and their effective capabilities clearly.
- Record who performed shared-garden mutations.
- Notify affected users of access changes.

### 2. Real-time transport

- Define events for plant, garden, weather, marketplace, and access changes.
- Authenticate connections and authorize every subscription.
- Treat WebSocket messages as notifications to refetch authoritative state, not as the source of truth.
- Support reconnect, missed-event recovery, and duplicate delivery.

### 3. Notifications

- Start with an in-app notification inbox.
- Define read state, retention, and user preferences.
- Add browser or email delivery only after a concrete need and provider are selected.

### 4. Competitive features

- Define leaderboard metric, period, tie-breaking, and privacy rules.
- Derive rankings from durable facts rather than client-submitted totals.
- Add achievements only when their trigger can be verified server-side.

## Validation

- Users cannot subscribe to gardens they cannot view.
- Permission revocation closes or invalidates active subscriptions promptly.
- Reconnect produces the same final state as a continuous connection.
- Duplicate events do not duplicate UI state or rewards.

## Risks

- Multi-instance deployment requires shared fan-out and connection ownership.
- Presence is ephemeral and must not replace durable membership.
- Notification volume can become noisy without aggregation rules.

## Exit criteria

- Collaboration remains correct with WebSockets disabled or temporarily unavailable.
- Real-time behavior improves freshness without weakening authorization or consistency.

