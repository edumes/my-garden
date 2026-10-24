# Phase 4 — Experience Redesign

## Objective

Redesign the experience around the garden while preserving working behavior through incremental screen changes.

## Dependencies

- The Phase 3 API and gameplay contracts are stable enough to design against.

## Decisions (2026-10-24)

Eduardo chose a casual-game direction: bright colors, playful details and visible progression. Registration should immediately enter the game and guide the player through creating a garden and planting a seed.

## First increment completed (2026-10-24)

- Login and registration use a lime, lilac and pink palette with a decorative garden grid, responsive layouts and dark-mode support. Existing optional profile fields remain available.
- Registration uses the session already issued by the API and replaces the signup route with the dashboard. No authentication contract or gameplay rule changed.
- The first quest follows persisted player state: create the first garden, open it automatically, buy seeds, choose a plot and plant. The planting XP ends the quest; returning accounts see their actual level and remaining XP.
- Fields have associated labels and autocomplete, password visibility is keyboard accessible, and API failures appear inline. Failed garden loading offers retry instead of presenting a false empty state.
- The shared Button now forwards its ref under React 18. A real browser test exposed an off-screen account menu before this fix; logout and sign-in pass after it.
- `npm run build` and `npm run lint` passed (eight existing Fast Refresh warnings). Three Playwright tests passed: automatic session and auth failures, dark/mobile/keyboard entry, and the full onboarding-to-harvest loop. Desktop and mobile entry screenshots were inspected.

Phase 4 is still in progress. The broader garden/store redesign, navigation, route splitting, dependency upgrades and remaining screens are not completed by this increment.

## Garden increment completed (2026-10-24)

- Garden cards have a nine-plot preview, planted/harvestable counts, a clear open action, keyboard support and a separate labeled share action. Shared cards display all effective capabilities rather than only the first one.
- Garden detail uses the casual-game palette, a compact summary and a plot-first layout on mobile. Plant details expand on demand, while growth stages and permitted actions remain visible. Weather, progression, plant counts, metadata and reward ownership remain available.
- Created/updated dates follow the player's locale and timezone, with a fallback for unsupported saved preferences.
- The overview keeps progression and gardens first. Audit activity and generic tips were removed from the dashboard; the audit route remains available. Mobile navigation has an accessible name/state and no longer overlaps the brand.
- `npm run build` and `npm run lint` passed, with the existing bundle-size and eight Fast Refresh warnings. Four Playwright tests passed, covering authentication, the full garden loop, keyboard card navigation, viewer restrictions and live capability changes; the shared harvest still rewards the grower. Light/mobile and dark shared-garden captures were inspected.

Phase 4 remains in progress; the store increment below follows the garden work.

## Store and inventory increment completed (2026-10-24)

- The store follows the casual-game palette and exposes Shop seeds and My seeds views, search, owned quantities, purchase totals, remaining balance and level locks. Purchase controls have explicit accessible names and prevent overlapping submissions within the shop.
- Inventory comes from the server and refreshes after purchases. Empty results are distinct from request failures; catalog and inventory loading can fail independently, and retry recovers without repeating a purchase.
- `/store` now opens the same store interface used by quick-buy dialogs, fixing the existing sidebar destination. Closing the route returns to gardens; the inventory also links back to gardens for planting.
- Build and lint passed, with the existing bundle and eight Fast Refresh warnings. The existing four browser tests passed. The new store test exceeded its initial 30-second budget during a parallel cold run, then passed in isolation with a 90-second budget (11.6 seconds): empty bag, two-seed purchase, balance/count refresh, persistence, a failed inventory request and retry, level locks, insufficient funds, responsive layout and route close.
- Store desktop and inventory mobile screenshots were inspected. A separate local preview database is used for Eduardo's manual testing, with GAME_SPEED=60 and one-second ticks; this does not alter default gameplay settings.

Remaining Phase 4 work includes progression/achievements, marketplace, sharing, settings, navigation cleanup, route splitting and the deferred dependency upgrades.

## Experience principles

- The current garden and its next actions are the primary focus.
- Navigation reflects player goals rather than backend modules.
- Unavailable features are hidden or clearly labeled as planned.
- Mobile, keyboard, reduced-motion, and screen-reader use are first-class.

## Scope

### 1. Information architecture

- Define the minimum top-level destinations: garden, collection/inventory, market, community, and account.
- Remove the broken `/store` navigation path or implement one canonical store destination.
- Move audit tooling out of ordinary player navigation unless the user has the required role.
- Separate profile content from the garden dashboard.

### 2. Visual foundation

- Define color, typography, spacing, elevation, focus, and motion tokens.
- Consolidate duplicate UI primitives before redesigning screens.
- Use animation only to explain state change or improve continuity.
- Preserve theme support without mounting hidden application chrome on authentication routes.

### 3. Incremental screen order

1. Authentication and onboarding.
2. Garden overview and garden detail.
3. Store and inventory.
4. Progression and achievements.
5. Marketplace.
6. Sharing and community.
7. Settings.

### 4. Frontend architecture

- Replace API `any` values with domain response types.
- Add route-level loading and code splitting.
- Standardize query, mutation, error, and refresh behavior without adding a state library unless existing tools cannot solve it clearly.
- Fix invalid ref forwarding and hook dependency problems, and invalid DOM nesting such as the `<div>` inside the plant-seed dialog's `<p>` description.
- Upgrade Tailwind CSS 3 → 4 and React 18 → 19, deferred from Phase 1. Tailwind 4 clears the build-time advisories recorded in [Phase 1](./01-engineering-baseline.md#deferred-advisories); React 19 unblocks `react-router` 8 and removes the ref warnings from shadcn components that do not forward refs.
- Format dates using the selected locale and timezone.

## Validation

- Every screen works at mobile and desktop breakpoints.
- Keyboard navigation and visible focus work for primary flows and dialogs.
- Browser console is clean during smoke flows.
- No navigation item leads to an unimplemented route.
- The main bundle is split by meaningful route boundaries.

## Risks

- Redesigning before stabilizing contracts would cause repeated rework.
- Replacing every component at once would make regressions difficult to isolate.
- Decorative motion can harm usability if not reduced for user preferences.

## Exit criteria

- The complete core loop uses the new experience.
- Legacy screens are removed as each replacement becomes verified.

