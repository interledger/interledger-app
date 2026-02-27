# Official Migration Plan: Remix v2 → React Router v7

Sources:
- https://reactrouter.com/upgrading/remix
- https://v2.remix.run/docs/start/future-flags/
- https://reactrouter.com/explanation/type-safety
- https://reactrouter.com/how-to/route-module-type-safety

---

## Prerequisites

- Node 20+
- React 18 + React DOM 18
- Must be on Remix v2 with the **Vite plugin** already adopted (classic compiler migration must be done first)

---

## Phase 1 — Adopt Future Flags (on Remix v2)

Enable all six `v3_*` flags **one at a time** in the Remix Vite plugin config. Verify build + smoke test after each.

```ts
// vite.config.ts
remix({
  future: {
    v3_fetcherPersist: true,
    v3_relativeSplatPath: true,
    v3_throwAbortReason: true,
    v3_singleFetch: true,
    v3_lazyRouteDiscovery: true,
    v3_routeConfig: true,
  },
})
```

### 1.1 `v3_fetcherPersist`

**What changes:** Fetchers now persist based on idle state rather than unmounting. Fetchers may live longer than before.

**Code changes:**
- Review `useFetchers()` usage — components consuming it may render longer than expected
- No API change required; behavioral only

---

### 1.2 `v3_relativeSplatPath`

**What changes:** Relative path resolution inside multi-segment splat routes (e.g. `dashboard/*`) now works correctly.

**Code changes:**
- If you have relative `<Link to="...">` inside a splat layout, add extra `..` segments
  - e.g. `<Link to="team">` → `<Link to="../team">`
- Consider splitting splat routes into a layout + child route structure

---

### 1.3 `v3_throwAbortReason`

**What changes:** On server abort, throws `request.signal.reason` instead of a generic `Error` string.

**Code changes:**
- Only relevant if you have custom `handleError` logic that matches on the old abort error message string
- Typically: no changes needed

---

### 1.4 `v3_singleFetch`

**What changes:** Data requests during client-side navigation are batched into a single fetch (same as document requests). `json()` and `defer()` are deprecated.

**Required changes:**

#### Remove `json()` wrappers — return raw objects
```ts
// Before
import { json } from "@remix-run/node";
export async function loader() {
  return json(await fetchData());
}

// After
export async function loader() {
  return await fetchData();
}
```

#### For custom headers/status — use `data()` instead of `json()`
```ts
import { data } from "@remix-run/node";

export async function loader() {
  return data(await fetchData(), {
    headers: { "Cache-Control": "public, max-age=604800" },
  });
}
```

#### Remove `defer()` — return promises directly
```ts
// Before
import { defer } from "@remix-run/node";
export async function loader() {
  return defer({ slow: slowPromise() });
}

// After
export async function loader() {
  return { slow: slowPromise() };
}
```

#### Update `entry.server.tsx` — replace `ABORT_DELAY` with `streamTimeout`
```ts
// Add this export (replaces ABORT_DELAY constant)
export const streamTimeout = 5000;

// In the abort logic, use streamTimeout + 1000 for the setTimeout
// Remove abortDelay prop from <RemixServer>
```

> **Note:** `SerializeFrom<T>` is deprecated. Replace with `ReturnType<typeof useLoaderData<T>>`.

---

### 1.5 `v3_lazyRouteDiscovery`

**What changes:** Remix no longer sends the full route manifest on initial load — routes are discovered as users navigate.

**Code changes:**
- Generally none required
- Optional: use `<Link discover="render" | "none">` to control eager/lazy discovery per-link

---

### 1.6 `v3_routeConfig`

**What changes:** File-based routing is replaced by a `routes.ts` config file.

**Install packages:**
```bash
pnpm add @remix-run/route-config @remix-run/fs-routes
# If keeping existing routes option:
pnpm add @remix-run/routes-option-adapter
```

**Create `app/routes.ts`:**
```ts
import { flatRoutes } from "@remix-run/fs-routes";
import type { RouteConfig } from "@remix-run/route-config";

export default flatRoutes() satisfies RouteConfig;
```

Or if using existing `routes()` option:
```ts
import { remixRoutesOptionAdapter } from "@remix-run/routes-option-adapter";
import type { RouteConfig } from "@remix-run/route-config";

export default remixRoutesOptionAdapter((defineRoutes) => {
  return defineRoutes((route) => {
    // existing route definitions
  });
}) satisfies RouteConfig;
```

**Remove from `vite.config.ts`:** `ignoredRouteFiles`, `routes()` option — they now live in `routes.ts`.

---

## Phase 2 — Package Swap to React Router v7

After all flags are stable, run the automated codemod first:

```bash
npx codemod remix/2/react-router/upgrade
```

This handles steps 2–8 below automatically. Verify each one manually after.

---

### 2.1 Update `package.json` dependencies

| Remove | Add |
|--------|-----|
| `@remix-run/react` | `react-router` |
| `@remix-run/node` | `@react-router/node` |
| `@remix-run/dev` | `@react-router/dev` |
| `@remix-run/serve` | `@react-router/serve` |
| `@remix-run/fs-routes` | `@react-router/fs-routes` |
| `@remix-run/route-config` | `@react-router/dev` (re-exports it) |
| `@remix-run/routes-option-adapter` | `@react-router/remix-routes-option-adapter` |
| `@sentry/remix` | `@sentry/react-router` |
| `routes-gen`, `@routes-gen/remix` | *(remove entirely — use RR7 typegen)* |

```bash
pnpm install
```

---

### 2.2 Update `package.json` scripts

| Script | Before | After |
|--------|--------|-------|
| `dev` | `remix vite:dev` | `react-router dev` |
| `build` | `remix vite:build` | `react-router build` |
| `start` | `remix-serve build/server/index.js` | `react-router-serve build/server/index.js` |
| `typecheck` | `tsc` | `react-router typegen && tsc` |

---

### 2.3 Update `vite.config.ts`

```diff
-import { vitePlugin as remix } from "@remix-run/dev";
+import { reactRouter } from "@react-router/dev/vite";

 export default defineConfig({
   plugins: [
-    remix({
-      future: { /* all v3 flags */ },
-    }),
+    reactRouter(),
     tsconfigPaths(),
   ],
 });
```

All future flags and non-default config move to `react-router.config.ts` (next step).

---

### 2.4 Create `react-router.config.ts`

```ts
import type { Config } from "@react-router/dev/config";

export default {
  ssr: true,
  // appDirectory: "app",  // only if non-default
} satisfies Config;
```

---

### 2.5 Update `app/routes.ts` imports

```diff
-import type { RouteConfig } from "@remix-run/route-config";
-import { flatRoutes } from "@remix-run/fs-routes";
+import type { RouteConfig } from "@react-router/dev/routes";
+import { flatRoutes } from "@react-router/fs-routes";

 export default flatRoutes() satisfies RouteConfig;
```

---

### 2.6 Update `entry.server.tsx`

```diff
-import { RemixServer } from "@remix-run/react";
+import { ServerRouter } from "react-router";

-<RemixServer context={remixContext} url={request.url} />
+<ServerRouter context={remixContext} url={request.url} />
```

Also update Sentry instrumentation to `@sentry/react-router` API.

Fix remaining `@remix-run/*` imports (anything not caught by codemod):
```bash
grep -r "@remix-run" app/
```

---

### 2.7 Update `entry.client.tsx`

```diff
-import { RemixBrowser } from "@remix-run/react";
+import { HydratedRouter } from "react-router/dom";

 hydrateRoot(
   document,
   <StrictMode>
-    <RemixBrowser />
+    <HydratedRouter />
   </StrictMode>,
 );
```

---

### 2.8 Update `remix.env.d.ts` (or `env.d.ts`)

```diff
-/// <reference types="@remix-run/dev" />
+/// <reference types="@react-router/dev" />
```

---

## Phase 3 — Type Safety Setup

### 3.1 Update `.gitignore`

```
.react-router/
```

### 3.2 Update `tsconfig.json`

```json
{
  "include": [
    "...",
    ".react-router/types/**/*"
  ],
  "compilerOptions": {
    "types": ["@react-router/node", "vite/client"],
    "rootDirs": [".", "./.react-router/types"]
  }
}
```

### 3.3 Generate types

```bash
npx react-router typegen
```

Types are auto-generated when running `react-router dev` (via the Vite plugin). In CI, run `react-router typegen && tsc`.

### 3.4 Migrate route types (optional but recommended)

Replace generic `LoaderFunctionArgs` / `ActionFunctionArgs` with per-route generated types:

```ts
// Before
import { LoaderFunctionArgs } from "react-router";
export async function loader({ params }: LoaderFunctionArgs) {}

// After
import type { Route } from "./+types/my-route";
export async function loader({ params }: Route.LoaderArgs) {}
// params is now fully typed to the route's URL params
```

Available generated types per route:
- `Route.LoaderArgs` / `Route.ActionArgs`
- `Route.ClientLoaderArgs` / `Route.ClientActionArgs`
- `Route.ComponentProps`
- `Route.ErrorBoundaryProps`
- `Route.HydrateFallbackProps`

### 3.5 Type `AppLoadContext` (if using custom server context)

```ts
// app/env.ts (or a .d.ts file)
declare module "react-router" {
  interface AppLoadContext {
    // your context shape
  }
}
export {};
```

---

## Phase 4 — Verification

```bash
# No @remix-run/* imports should remain
grep -r "@remix-run" app/

# Full build
pnpm build

# TypeScript
pnpm typecheck

# Dev smoke test
pnpm dev
```

Smoke test checklist:
- [ ] Home page loads
- [ ] Client-side navigation works
- [ ] Forms submit correctly
- [ ] Sentry initializes without console errors
- [ ] Storybook builds: `pnpm storybook build`

---

## Key Deprecations Summary

| API | Status | Replacement |
|-----|--------|-------------|
| `json()` | Deprecated | Return raw object; use `data()` for headers/status |
| `defer()` | Deprecated | Return promises directly |
| `SerializeFrom<T>` | Deprecated | `ReturnType<typeof useLoaderData<T>>` |
| `LoaderFunctionArgs` (generic) | Still works, but | Prefer `Route.LoaderArgs` from typegen |
| `ActionFunctionArgs` (generic) | Still works, but | Prefer `Route.ActionArgs` from typegen |
| `@remix-run/eslint-config` | Deprecated | Migrate to standard ESLint configs |
| `routes-gen` / `@routes-gen/remix` | Remove | React Router v7 typegen replaces this |
| Multipart form utilities (`unstable_*`) | Deprecated | `@mjackson/form-data-parser` + `@mjackson/file-storage` |
