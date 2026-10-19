# Phase 3 — Core Gameplay Loop

## Objective

Deliver a complete and rewarding garden loop before expanding secondary systems.

## Dependencies

- Phase 2 has secured inventory, balances, and garden access.

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
- Make the garden page's "Gardener Stats" agree with the header: a new player currently sees level 0 and 0 coins there, but level 1 and 100 coins in the header.
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

