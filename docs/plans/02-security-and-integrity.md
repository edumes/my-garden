# Phase 2 — Security and Integrity

## Objective

Make authorization, shared access, inventory, and currency safe under failure and concurrency.

## Dependencies

- Phase 1 validation and integration-test foundations are complete.

## Scope

### 1. Authorization model

- Decide whether audit data is administrator-only or user-scoped.
- Add an explicit role or capability model instead of treating authentication as authorization.
- Centralize garden access checks for view, plant, harvest, manage, and owner-only actions.
- Enforce share expiration on every protected operation.
- Test cross-user access with default-deny expectations.

### 2. Shared access

- Define whether permissions are independent capabilities or an ordered hierarchy.
- Make access-link redemption and `used_count` updates one atomic operation.
- Add database constraints preventing duplicate shares for the same garden and user.
- Use full-entropy tokens and avoid returning active tokens outside owner-authorized responses.

### 3. Store and inventory

- Repair the transaction boundary so every write uses the transaction handle.
- Lock or conditionally update balances and inventory to prevent overspending.
- Define whether repeated seed purchases aggregate inventory or create immutable lots.
- Add database constraints for non-negative balances and quantities where appropriate.

### 4. Marketplace and wallet

- Define one authoritative currency model; resolve the overlap between user coins and wallet balance.
- Wrap purchase, delivery completion, plant transfer, transaction history, and reputation changes atomically.
- Propagate every persistence error.
- Make commands idempotent and reject duplicate purchases or completions.
- Reassess private-key storage and remove blockchain mechanics unless a concrete product requirement justifies them.

### 5. Authentication and configuration

- Reject default JWT secrets outside development.
- Validate JWT issuer and expected signing algorithm explicitly.
- Define logout and token revocation semantics honestly.
- Respect the configured database TLS mode.
- Implement rate limiting at authentication and mutation boundaries.

## Validation

- Integration tests prove users cannot read or mutate another user's private data.
- Concurrent purchase, planting, sharing, and marketplace tests preserve all invariants.
- Forced failures in the middle of a transaction leave no partial state.
- Expired and exhausted share links fail closed.
- Production configuration refuses unsafe defaults.

## Risks

- Existing rows may violate new uniqueness or non-negative constraints.
- Changing currency authority may require a one-time data migration.
- Permission semantics affect both the API and existing frontend controls.

## Exit criteria

- Authorization is centralized and covered by a permission matrix.
- Every multi-record economic operation is atomic and tested under concurrency.

