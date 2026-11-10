# Web Design System

Pulse frontend styles are organized as a layered design-system contract, not as
route-local global CSS.

## Layers

`src/styles/index.css` declares the cascade order:

```css
@layer tokens, base, ui, layout, features, vendor;
```

The import order in `index.css` is the public style composition contract. Route
files and feature components should not import global CSS directly.

## Token Hierarchy

- `tokens/primitives.css` contains the only raw color values in the frontend:
  `--palette-surface`, `--palette-ink`, `--palette-brand`,
  `--palette-success`, and `--palette-danger`. Do not add color ramps here.
- `tokens/semantic.css` maps primitives to app aliases such as `--background`,
  `--foreground`, `--card`, `--border`, `--primary`, `--ring`, shared
  surfaces, and shadows. Additional shades must be derived with variables and
  `color-mix()`.
- `tokens/themes.dark.css` overrides semantic tokens under
  `[data-theme="dark"]`. It must not define component or feature selectors.
- `tokens/status.css` owns success, info, warning, danger, skipped, and shared
  status-adjacent aliases used by feature modules. `info` and `warning` derive
  from brand unless a future ADR introduces more brand colors.
- `tokens/density.css` owns compact/default density variables.

Feature files may define local custom properties only in their namespace, for
example `--work-*`, `--chat-*`, `--workflow-*`, `--payment-*`, and
`--ai-widget-*`.

## Component Contract

Thin wrappers in `src/components/ui` map directly to `styles/ui` classes:

- `Button` -> `.ui-button`, `.ui-button--*`;
- `Input`, `Textarea`, `Field` -> `.ui-input`, `.ui-textarea`, `.ui-field`;
- `Card` -> `.ui-card`, `.ui-card-*`;
- `Table` -> `.ui-table-*`;
- `Dialog`, `Sheet`, `Tabs`, `Badge`, `Switch` -> matching `.ui-*` classes.

UI primitive CSS must not depend on route-specific feature classes.

## Feature Contract

Feature CSS lives under `src/styles/features/<feature>/`. Existing class names
are preserved to avoid route behavior and JSX churn. Large root surfaces may add
`data-feature` attributes for future scoping, but selectors should continue to
work without changing DOM semantics.

Current feature roots include `work`, `workflow`, `payments`, `live-assist`,
`ai-employee`, and `status`.

## File Budgets

- `src/styles/index.css`: max 80 lines.
- `src/styles/tokens/*.css`: max 180 lines.
- `src/styles/ui/*.css`: max 180 lines.
- all other CSS modules: max 260 lines.
- `src/app/globals.css`: thin entrypoint only.

## Forbidden Patterns

- raw hex colors outside `src/styles/tokens/primitives.css`;
- raw `rgb()`, `hsl()`, `oklch()`, or named colors outside
  `tokens/primitives.css`;
- legacy hue aliases such as `--action-blue`, `--live-accent`,
  `--chat-accent`, `--surface-cyan`, or `--color-*`;
- `color-mix()` from raw colors or color keywords instead of variables;
- `!important` outside the explicit audit allowlist;
- selector nesting depth greater than 3;
- feature selectors such as `.work-*`, `.chat-*`, `.payment-*`, `.live-*`, or
  `.ai-employee-*` in `src/app/globals.css`;
- route-local global CSS imports.

## Migration Policy

Do not migrate to Tailwind v4 or shadcn in this architecture pass. If that
becomes necessary, write a separate ADR covering migration scope, compatibility,
visual regression checks, and rollback.

New visual work must first choose the right layer, add tokens before raw values,
and pass `pnpm style:audit`.
