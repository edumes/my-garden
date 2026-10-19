# Phase 5 — Marketplace

## Objective

Add trading and delivery systems on top of the proven economy without compromising ownership or balances.

## Dependencies

- Phase 2 economic invariants are enforced.
- Phase 3 defines harvestable assets and rewards.
- Phase 4 provides stable store and inventory interactions.

## Scope

### 1. Product rules

- Define what can be sold: a plant, harvest output, or inventory item.
- Define listing expiration, cancellation, fees, and ownership while listed.
- Define delivery acceptance, proof of completion, deadlines, penalties, and refunds.
- Keep prices and balances as integer minor units.

### 2. Transaction lifecycle

- Reserve the asset and required funds before accepting an operation.
- Make completion, cancellation, and expiration explicit state transitions.
- Store an immutable economic history separate from mutable current state.
- Require idempotency for purchase and delivery commands.

### 3. Experience

- Show only actions valid for the current user and state.
- Explain fees and final amounts before confirmation.
- Expose owned listings, purchases, deliveries, and outcomes.
- Provide filters and pagination only after the core flow is correct.
- Fix wallet creation from the UI: the frontend sends `encrypted_private_key`, but `POST /wallet` requires `password`, so the request always fails. Accept, complete, and buy actions also send placeholder signatures.

### 4. Blockchain decision gate

Do not expand blockchain behavior by default. Continue only if a written requirement explains why an append-only database ledger cannot satisfy the product need. Otherwise, remove unused consensus and cryptographic complexity while preserving auditable economic history.

## Validation

- Concurrent buyers cannot acquire the same asset.
- Failures never leave ownership and payment inconsistent.
- Repeated commands return the original result without applying effects twice.
- Browser tests cover listing, purchase, cancellation, delivery, and insufficient funds.

## Risks

- The current plant-transfer model silently chooses a buyer garden.
- Automatically created wallets currently mint starting balances.
- Commission rounding for small integer prices needs an explicit rule.

## Exit criteria

- Marketplace operations preserve currency and ownership invariants under concurrency.
- Players can understand the lifecycle and recover from expired or cancelled operations.

