<!-- Generated from registry/standards/ui/shadcn-compact@1.1.0. Update the registry standard, then regenerate. -->

# shadcn Compact UI Profile

This UI profile applies to web interfaces that choose compact, high-contrast, reusable shadcn-based UI.

It is a separate standard because UI policy is not the same thing as a frontend runtime. React, Next.js, Vue, Svelte or another frontend can adopt or ignore this profile explicitly through the active stack.

## May 2026 Baseline

1. WCAG 2.2 AA is the accessibility target.
2. Tailwind v4-compatible token thinking is preferred when Tailwind is used.
3. shadcn primitives are consumed through a local `components/ui` layer.
4. Compact operational UI should be dense, readable and predictable.
5. Marketing-style hero/card layouts are not the default for apps and tools.

## shadcn-First Policy

1. `SHADCN_UI_FOR_INTERACTIVE_PRIMITIVES_PERCENT = 100`
2. `RAW_INTERACTIVE_HTML_TAGS_IN_FEATURE_CODE = 0`
3. `DIRECT_RADIX_IMPORTS_IN_FEATURE_CODE = 0`
4. `CUSTOM_UI_PRIMITIVE_WITHOUT_ADR = 0`

Interactive primitives in feature code use the local `components/ui` layer.

If a generated template does not yet install shadcn/Tailwind, this profile is a target policy, not a claim that those packages already exist.

## Unified List Standard

All repeated list/table screens use one list kit:

1. `LIST_KIT_COUNT = 1`
2. `LIST_BASE_COMPONENT = DataList`
3. `LIST_TOOLBAR_COMPONENT = DataListToolbar`
4. `LIST_FILTERS_COMPONENT = DataListFilters`
5. `LIST_EMPTY_COMPONENT = DataListEmpty`
6. `LIST_SKELETON_COMPONENT = DataListSkeleton`
7. `LIST_ROW_LAYOUT_VARIANTS = 1`
8. `DUPLICATE_LIST_IMPLEMENTATIONS_ALLOWED = 0`

## Compact Density

1. `MOBILE_FIRST_BREAKPOINT_ORDER = base -> sm -> md -> lg -> xl`
2. `MIN_TOUCH_TARGET_PX = 36` for dense mouse-first operational controls
3. `DEFAULT_TOUCH_TARGET_PX = 40`
4. `MAX_PRIMARY_LAYOUT_GAP_PX = 8`
5. `MAX_CARD_PADDING_PX = 8`
6. `MAX_SECTION_MARGIN_BOTTOM_PX = 12`
7. `DEFAULT_LINE_HEIGHT_RATIO = 1.25`
8. `DENSITY_MODE_COUNT = 1`

Touch-heavy or mobile-first screens should use larger targets and may override compact density.

## Visual Style

1. `TOTAL_PROJECT_COLORS_MAX = 5`
2. `ACCENT_COLOR_COUNT = 1`
3. `SHADOW_USAGE_COUNT = 0`
4. `BLACK_BORDER_USAGE_COUNT = 0`
5. `INLINE_STYLE_USAGE_COUNT = 0`
6. `RAW_HEX_IN_FEATURE_CODE_COUNT = 0`
7. `CONTRAST_REQUIRED_PERCENT = 100`
8. `NEGATIVE_LETTER_SPACING = 0`
9. `VIEWPORT_WIDTH_FONT_SCALING = 0`

Baseline colors:

1. `#FFFFFF` background
2. `#F3F5F7` surface
3. `#101418` text
4. `#0A84FF` accent
5. `#D92D20` danger

Avoid one-note palettes. A product interface should not be dominated by slight variations of a single hue.

## Accessibility

1. Visible focus states are mandatory.
2. Focus must not be obscured by sticky headers/toolbars.
3. Drag interactions require pointer alternatives.
4. Target size follows WCAG 2.2 intent while respecting compact desktop density.
5. Form errors are linked to fields.
6. Color is never the only signal.

## Motion

1. `LAYOUT_SHIFTING_ANIMATIONS_ALLOWED = 0`
2. `TRANSFORM_SHIFT_ANIMATIONS_ALLOWED = 0` except drag/reorder.
3. `TRANSITION_ALL_USAGE_COUNT = 0`
4. `DEFAULT_ANIMATION_DURATION_MS = 120`
5. `MAX_ANIMATION_DURATION_MS = 160`
6. Allowed transition properties: `color`, `background-color`, `border-color`, `opacity`.
7. `CLS_BUDGET = 0.1`

## Layout Stability

1. Fixed-format controls need stable dimensions.
2. Toolbar buttons must not resize on hover or loading.
3. Tables/lists must handle long labels without overlap.
4. Cards are for repeated items or tools, not page section wrappers.
5. Do not nest cards inside cards.

## Data-Dense Screens

1. Repeated operational lists use one list/table kit.
2. Filters, sorting and pagination are consistent.
3. Very large lists use pagination or virtualization.
4. Empty/error/loading states reuse shared components.
5. Destructive actions have confirmation and undo when appropriate.

## Enforcement Reality

This profile is documented by default. It becomes linted/CI-blocking only when the generated project includes an actual UI lint check that scans feature code.
