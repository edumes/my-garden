# Phase 3 — Core Gameplay Loop

## Objective

Deliver a complete and rewarding garden loop before expanding secondary systems.

## Dependencies

- Phase 2 has secured inventory, balances, and garden access.

## Decisions

Made by Eduardo on 2026-10-19:

- `growth_time` is real minutes. `GAME_SPEED` multiplies growth to speed up development and tests without changing the rules.
- Harvesting removes the plant and frees its plot; replanting needs a new seed. The seed catalog is rebalanced so every harvest pays back its seed with a profit.
- The plant's grower earns the harvest's coins and experience, whoever harvests: the player who planted it, or who bought it on the marketplace.
- Ripe plants never wither.
- Planting, harvesting, marketplace sales, and completed deliveries grant experience. Going from level n to n+1 costs 100×n experience.
- Levels unlock seed types and more gardens.
- The `/game/status` and `/game/leaderboard` placeholders are removed; leaderboards belong to Phase 6.
- Redis is removed: its only use was a weather cache that was written and never read. It returns when Phase 6 or Phase 7 needs shared state across instances.

Values chosen during implementation, open to tuning after playtesting:

- Planting grants 2 XP; harvesting grants the plant's growth time in minutes; each sale or completed delivery grants 10 XP.
- A player owns at most (level + 1) / 2 gardens: one at levels 1 and 2, two at levels 3 and 4, and so on.
- Seeds unlock in six tiers by price, from level 1 to level 6. Each harvest returns the seed price plus growth minutes × a rate from 0.20 coins per minute at level 1 to 0.45 at level 6.
- Marketplace purchases need a free plot in one of the buyer's gardens, so the garden limit has no loophole.

## Player loop

1. Create a garden.
2. Buy seeds.
3. Plant a seed in an available plot.
4. Observe growth and weather effects.
5. Harvest a mature plant.
6. Receive currency and experience.
7. Unlock meaningful progression.

## Scope

### 1. Domain rules

- Define growth stages, tick behavior, weather modifiers, harvest results, and withering policy.
- Grant experience for gameplay actions: today nothing increases `experience`, so levels never change.
- Make time calculations deterministic and testable.
- Decide what happens after harvest: reset, regrow, or remove the plant.
- Decide who earns rewards in shared gardens: today the harvester receives the coins, not the owner or the planter.
- Define level thresholds and actual unlocks.

### 2. Reliable engine

- Prevent multiple application instances from processing the same plant tick twice.
- Make ticks idempotent and safe after restarts.
- Replace the tick's whole-row `Save` of each plant with a conditional update of growth and stage. Today a plant removed, harvested, or sold between the tick's read and write comes back with stale values, and a deleted plant is recreated.
- Fix weather serialization in Redis or remove caching if it provides no demonstrated value.
- Add readiness information for database, Redis, and the game engine.

### 3. API contract

- Replace placeholder game-status responses with useful state or remove the route.
- Return typed, stable responses for garden, plant, inventory, and progression operations.
- Use consistent error codes for invalid state, missing resources, and denied actions.

### 4. Gameplay feedback

- Show the next meaningful action for every plant.
- Display real growth time and the effect of current weather.
- Refresh currency, experience, inventory, and plant state after successful actions.
- Provide useful empty, loading, offline, and failure states.

## Validation

- Service and integration tests cover every transition in the player loop.
- Time-controlled tests cover growth boundaries and weather modifiers.
- Browser tests complete the loop from a new account through first harvest.
- Restarting the API does not duplicate rewards or growth.

## Risks

- Existing seed data may not produce a balanced economy.
- A five-second development tick can hide production-time race conditions.
- Progression design may change once real playtesting begins.

## Exit criteria

- A new player can complete the full loop without database intervention.
- Every reward and inventory change has one server-authoritative cause.


## Completion evidence (2026-10-24)

Phase 3 is implemented and locally validated.

- Garden plots expose their stage, next action, base growth time and total harvest payout. Planting, harvesting and removal use visible, keyboard-accessible controls and current server permissions.
- Garden refresh failures preserve the last state, show a retry action and suspend plot commands until refreshed. Store and seed inventory failures are distinguished from empty results.
- Current weather refreshes periodically and displays its growth multiplier. User profile summaries follow refreshed balances and progression instead of keeping a stale copy.
- The browser smoke test completes registration, login, garden creation, purchase, planting, automatic growth and harvest. It verifies a profitable reward, XP, a free plot, consumed inventory, persistence after reload and refresh recovery. Harvest also runs at a 390px viewport; screenshots cover desktop and mobile.
- `go test -tags integration ./...` and `go vet -tags integration ./...` passed against an isolated local PostgreSQL 16 instance. Existing suites cover growth boundaries, concurrent ticks, reward ownership, repeated harvests and progression unlocks.
- A fresh-engine integration check repeats the last tick against persisted plants and valid weather to verify restart idempotency.
- `npm ci`, `npm run lint`, `npm run build` and `npm run test:e2e` passed. Browser growth is accelerated only in the test-started API.

Remaining Phase 4 work: shared UI ref warnings, eight existing Fast Refresh lint warnings, route splitting (the main bundle remains above 500 kB), and the broader experience redesign. Local validation does not replace production deployment or failure drills from Phase 7.
