# Phase 1 — Engineering Baseline

## Objective

Create a trustworthy development baseline before changing domain behavior or redesigning the interface.

## Scope

### 1. Reproducible local environment

- Add a documented way to start PostgreSQL, Redis, the API, and the frontend.
- Separate development defaults from production requirements.
- Make the frontend API URL configurable by environment.
- Ensure startup reports actionable failures without exposing secrets.

### 2. Repository hygiene

- Stop tracking compiled binaries such as `bin/server`, `main.exe`, and `tmp/main.exe`.
- Align `.gitignore`, Make targets, and documented commands.
- Remove documentation for scripts and features that do not exist.
- Keep generated Swagger artifacts reproducible.

### 3. Dependency health

- Review direct dependency updates before running automated fixes.
- Resolve production-exposed high-severity advisories first.
- Update build-only dependencies separately to make regressions easier to locate.
- Record any advisory intentionally deferred and why it is not reachable.

### 4. Green validation

- Fix frontend lint errors without weakening rules globally.
- Add missing frontend type contracts instead of replacing `any` with unsafe casts.
- Keep Go formatting, tests, and vet checks green.
- Add CI for clean install, build, lint, `go vet`, and tests.

### 5. Test foundations

- Define separate unit, PostgreSQL integration, and browser smoke suites.
- Add integration-test database isolation and deterministic fixtures.
- Cover authentication and one garden happy path as the initial contract.

## Validation

- A new clone can start all local services from the documented commands.
- `go test ./...`, `go vet ./...`, `npm run lint`, and `npm run build` pass.
- CI executes the same commands from a clean checkout.
- No compiled executable is tracked.
- Runtime API configuration contains no hard-coded localhost dependency.

## Status

Validation executed on 2026-10-19, before review:

- [x] A clean copy of the repository starts PostgreSQL and Redis with `docker compose up -d --wait`, and the API and frontend with the documented commands.
- [x] `go test ./...`, `go vet ./...`, `npm run lint` (0 errors), and `npm run build` (with type checking) pass; `make check` runs them all.
- [ ] CI runs the same commands from a clean checkout: the workflow exists and `make check` passes from a clean local copy, but it has not run on GitHub yet.
- [x] No compiled executable is tracked.
- [x] The frontend calls the API through a relative URL or `VITE_API_URL`.
- [x] Unit, PostgreSQL integration, and browser smoke suites exist; integration tests cover authentication and a garden happy path, and the smoke test covers sign-up through planting.

## Deferred advisories

Recorded on 2026-10-19 after the dependency upgrades. `npm audit --omit=dev` and `govulncheck ./...` report no reachable vulnerabilities; the remaining findings are:

| Source | Severity | Why it is not reachable | Resolution |
|---|---|---|---|
| `tailwindcss@3.4.19` → `braces`, `micromatch`, `chokidar`, `fast-glob` | High | Tailwind runs only during build and dev. These packages expand the `content` globs from `tailwind.config.js` and watch local files; no user input reaches them, and none ship to the browser. | Tailwind 4 upgrade in [Phase 4](./04-experience-redesign.md) |
| `tailwindcss@3.4.19` → `postcss-nested`, `postcss-selector-parser` | Moderate | Same build-only path; they parse the project's own CSS. | Same |
| `golang.org/x/crypto/openpgp` (GO-2026-5932) | Unmaintained package | Not imported by the project; no upstream fix exists. | None while it stays unimported |

Held back on purpose:

- `react-router` 8 requires React 19, which moves to Phase 4 with Tailwind 4.
- TypeScript 7 is not supported by `typescript-eslint` yet; the project uses TypeScript 6.0.

## Risks

- Dependency upgrades may expose previously hidden API incompatibilities.
- Existing local databases may contain schema drift created by `AutoMigrate`.
- Cleaning lint findings must not silently change application behavior.

## Exit criteria

- The repository has one documented validation path that passes locally and in CI.
- The next phases can add tests without inventing new infrastructure each time.

