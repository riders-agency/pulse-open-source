package http

import _ "embed"

// embeddedSiteChatWidgetJS is rebuilt by air when this Go wrapper changes after widget edits.
// 2026-06-08: rebuild after operator-preview no-op and per-tab session storage fixes.
// 2026-06-08: rebuild after real iframe preview controller for webvisor.
// 2026-06-08: rebuild after iframe scroll watchdog for fullscreen webvisor.
// 2026-06-08: rebuild after inertia-scroll anti-jitter coalescing.
//
//go:embed widget.js
var embeddedSiteChatWidgetJS string
