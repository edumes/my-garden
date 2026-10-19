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
- Show players the delivery requests they accepted: `GET /marketplace/delivery-requests` returns open requests only, so the complete action is unreachable from the interface.

### 4. Blockchain decision gate

Phase 2 removed wallets, keys, the ledger, and the `/wallet` and `/blockchain` endpoints; user coins are the only currency. Reintroduce a ledger only as an append-only database table, and only if a written requirement needs more than the economic history above.

## Validation

- Concurrent buyers cannot acquire the same asset.
- Failures never leave ownership and payment inconsistent.
- Repeated commands return the original result without applying effects twice.
- Browser tests cover listing, purchase, cancellation, delivery, and insufficient funds.

## Risks

- The current plant-transfer model silently chooses a buyer garden.
- A seller can still harvest a listed plant, which resets it to a seed that the buyer then receives.
- Concurrent trades in opposite directions between the same two players can deadlock; PostgreSQL aborts one transaction, which the API reports as a server error.
- Commission rounding for small integer prices needs an explicit rule.

## Exit criteria

- Marketplace operations preserve currency and ownership invariants under concurrency.
- Players can understand the lifecycle and recover from expired or cancelled operations.

