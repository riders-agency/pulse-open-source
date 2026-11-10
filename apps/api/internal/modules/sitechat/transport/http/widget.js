(function () {
  function isOperatorPreviewWindow() {
    try {
      return window.name === "riders-pulse-webvisor-preview" || new URLSearchParams(window.location.search || "").get("rp_operator_preview") === "1";
    } catch (_) {
      return window.name === "riders-pulse-webvisor-preview";
    }
  }
  function installOperatorPreviewController() {
    if (window.__RP_OPERATOR_PREVIEW_CONTROLLER_INSTALLED__) return;
    window.__RP_OPERATOR_PREVIEW_CONTROLLER_INSTALLED__ = true;
    window.__RP_OPERATOR_PREVIEW_CONTROLLER_READY__ = true;
    window.__RP_MIRROR_CONTROLLER_READY__ = true;
    window.__RP_MIRROR_APPLY_COUNT__ = 0;
    window.__RP_MIRROR_LAST_STATE__ = null;
    var applySoonTimers = [];

    function post(payload) {
      try {
        payload = payload || {};
        payload.source = "riders-pulse-webvisor-preview";
        parent.postMessage(payload, "*");
      } catch (_) {}
    }
    function metrics() {
      var html = document.documentElement || {};
      var body = document.body || {};
      var scrolling = document.scrollingElement || html || body;
      return {
        scrollX: Math.round(window.scrollX || scrolling.scrollLeft || 0),
        scrollY: Math.round(window.scrollY || scrolling.scrollTop || 0),
        viewportWidth: Math.round(window.innerWidth || html.clientWidth || body.clientWidth || 0),
        viewportHeight: Math.round(window.innerHeight || html.clientHeight || body.clientHeight || 0),
        documentWidth: Math.round(Math.max(html.scrollWidth || 0, body.scrollWidth || 0, window.innerWidth || 0)),
        documentHeight: Math.round(Math.max(html.scrollHeight || 0, body.scrollHeight || 0, window.innerHeight || 0))
      };
    }
    function nodeByPath(path) {
      if (!path) return null;
      if (path === "html") return document.documentElement;
      var parts = String(path).split(".");
      var node = document;
      for (var i = 0; i < parts.length; i += 1) {
        if (parts[i] === "html") {
          node = document.documentElement;
          continue;
        }
        var index = Number(parts[i]);
        if (!Number.isFinite(index) || !node || !node.childNodes) return null;
        node = node.childNodes[index];
      }
      return node && node.nodeType === 1 ? node : null;
    }
    function apply(data) {
      data = data || {};
      var sx = Number(data.scrollX || 0);
      var sy = Number(data.scrollY || 0);
      if (!Number.isFinite(sx)) sx = 0;
      if (!Number.isFinite(sy)) sy = 0;
      window.__RP_MIRROR_APPLY_COUNT__ += 1;
      window.__RP_MIRROR_LAST_STATE__ = data;
      try {
        document.documentElement.dataset.rpAppliedScrollX = String(Math.round(sx));
        document.documentElement.dataset.rpAppliedScrollY = String(Math.round(sy));
      } catch (_) {}
      try {
        if (document.body) {
          document.body.dataset.rpAppliedScrollX = String(Math.round(sx));
          document.body.dataset.rpAppliedScrollY = String(Math.round(sy));
        }
      } catch (_) {}
      try {
        window.scrollTo({left: sx, top: sy, behavior: "auto"});
      } catch (_) {
        try { window.scrollTo(sx, sy); } catch (__) {}
      }
      var items = Array.isArray(data.scrollContainers) ? data.scrollContainers : [];
      items.forEach(function (item) {
        if (!item || !item.path) return;
        try {
          var node = document.querySelector('[data-rp-path="' + String(item.path).replace(/"/g, '\\"') + '"]') || nodeByPath(item.path);
          if (node) {
            node.scrollLeft = Number(item.left || 0);
            node.scrollTop = Number(item.top || 0);
          }
        } catch (_) {}
      });
      post(Object.assign({type: "rp:webvisor:mirror-applied", requestedScrollY: Math.round(sy), applyCount: window.__RP_MIRROR_APPLY_COUNT__}, metrics()));
    }
    function applySoon() {
      var data = window.__RP_MIRROR_LAST_STATE__;
      if (!data) return;
      applySoonTimers.forEach(function (timer) { clearTimeout(timer); });
      applySoonTimers = [];
      [120, 320, 800, 1600, 2600].forEach(function (delay) {
        applySoonTimers.push(setTimeout(function () { apply(data); }, delay));
      });
    }
    setInterval(function () {
      var data = window.__RP_MIRROR_LAST_STATE__;
      if (!data) return;
      var wantedY = Number(data.scrollY || 0);
      var current = metrics().scrollY;
      if (Math.abs(current - wantedY) > 12) apply(data);
    }, 750);
    addEventListener("message", function (event) {
      var data = event && event.data;
      if (!data || data.type !== "rp:webvisor:apply-state") return;
      apply(data);
      applySoon();
    });
    addEventListener("load", applySoon);
    addEventListener("resize", applySoon);
    document.addEventListener("DOMContentLoaded", applySoon);
    post(Object.assign({type: "rp:webvisor:mirror-ready"}, metrics()));
  }
  if (isOperatorPreviewWindow()) {
    installOperatorPreviewController();
    return;
  }
  var script = document.currentScript || {};
  var publicKey = (script.dataset && script.dataset.siteChatKey) || "";
  var guardKey = publicKey || "__default__";
  window.__ridersLiveAssistLoadedKeys = window.__ridersLiveAssistLoadedKeys || {};
  if (window.__ridersLiveAssistLoadedKeys[guardKey]) return;
  window.__ridersLiveAssistLoadedKeys[guardKey] = true;
  window.__ridersLiveAssistLoaded = true;

  var origin = new URL(script.src || window.location.href).origin;
  var storePrefix = "rp_live_assist_" + publicKey + "_";
  var state = {
    cfg: {
      title: "Чат на сайте",
      subtitle: "",
      accent_color: "#00B2FF",
      position: "right",
      settings_json: {}
    },
    open: false,
    connected: false,
    reconnecting: false,
    monitoring: false,
    call: null,
    muted: false,
    media: null,
    peer: null,
    remoteAudio: null,
    pendingIce: [],
    voice: null,
    voiceConfig: null,
    audioRelayRecorder: null,
    audioRelaySeq: 0,
    relayMediaSource: null,
    relaySourceBuffer: null,
    relayQueue: [],
    relayObjectUrl: "",
    relayFallbackQueue: [],
    relayFallbackPlaying: false,
    relayPlaybackActive: false,
    anonymousId: get("anonymous_id") || makeID("anon"),
    pageSessionId: getSession("page_session_id") || makeID("page"),
    visitorId: get("visitor_id"),
    visitSessionId: getSession("visit_session_id"),
    conversationId: getSession("conversation_id"),
    visitorName: get("name"),
    visitorContact: get("contact"),
    privacyAccepted: get("privacy_ok") === "1",
    messages: [],
    startedAt: Date.now(),
    lastOperatorMessageId: "",
    pendingAttachments: [],
    uploading: 0,
    ws: null,
    wsQueue: [],
    reconnectTimer: 0,
    reconnectAttempts: 0,
    pollTimer: 0,
    heartbeatTimer: 0,
    eventBuffer: [],
    conversationPromise: null,
    sequence: Number(get("sequence") || "0"),
    root: null,
    shadow: null,
    els: {},
    lastCursorAt: 0,
    lastScrollAt: 0,
    mutationTimer: 0,
    mutationCount: 0,
    mirrorTimer: 0,
    mirrorReason: "",
    audioCtx: null,
    ringTimer: 0,
  };
  set("anonymous_id", state.anonymousId);
  try {
    localStorage.removeItem(storePrefix + "visit_session_id");
    localStorage.removeItem(storePrefix + "conversation_id");
  } catch (_) {}
  setSession("page_session_id", state.pageSessionId);

  function makeID(prefix) {
    if (crypto && crypto.randomUUID) return prefix + "_" + crypto.randomUUID();
    var bytes = new Uint8Array(16);
    if (crypto && crypto.getRandomValues) crypto.getRandomValues(bytes);
    return prefix + "_" + Array.prototype.map.call(bytes, function (b) { return b.toString(16).padStart(2, "0"); }).join("");
  }
  function get(key) { try { return localStorage.getItem(storePrefix + key) || ""; } catch (_) { return ""; } }
  function set(key, value) { try { localStorage.setItem(storePrefix + key, value || ""); } catch (_) {} }
  function getSession(key) { try { return sessionStorage.getItem(storePrefix + key) || ""; } catch (_) { return ""; } }
  function setSession(key, value) { try { sessionStorage.setItem(storePrefix + key, value || ""); } catch (_) {} }
  function esc(value) { return String(value == null ? "" : value).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"); }
  function setting(path, fallback) {
    var cur = state.cfg.settings_json || {};
    path.split(".").forEach(function (part) { cur = cur && cur[part]; });
    return cur == null ? fallback : cur;
  }
  function labelSetting(path, fallback) {
    var value = String(setting(path, fallback) || "").trim();
    if (!value) value = fallback;
    if (path === "launcher.call_text" && value === "Позвонить") value = "Звонок";
    if (path === "active_invite.call_text" && value === "Позвонить") value = "Звонок";
    if (path === "launcher.primary_text" && value === "Написать") value = "Консультация онлайн";
    return value;
  }
  function accentText(bg) {
    var m = String(bg || "").match(/^#?([0-9a-f]{6})$/i);
    if (!m) return "#fff";
    var v = m[1], r = parseInt(v.slice(0, 2), 16), g = parseInt(v.slice(2, 4), 16), b = parseInt(v.slice(4, 6), 16);
    return ((r * 299 + g * 587 + b * 114) / 1000) > 165 ? "#0f172a" : "#fff";
  }
  function api(path, opts) {
    opts = opts || {};
    return fetch(origin + path, {
      method: opts.method || "GET",
      headers: {"Content-Type": "application/json", "Accept": "application/json"},
      body: opts.body ? JSON.stringify(opts.body) : undefined,
      keepalive: !!opts.keepalive
    }).then(function (res) {
      if (!res.ok) throw new Error("request failed");
      if (res.status === 204) return {};
      return res.json().catch(function () { return {}; });
    });
  }
  function absoluteURL(path) {
    if (!path) return "";
    try { return new URL(path, origin).toString(); } catch (_) { return path; }
  }
  function upload(path, form) {
    return fetch(origin + path, {
      method: "POST",
      headers: {"Accept": "application/json"},
      body: form
    }).then(function (res) {
      if (!res.ok) throw new Error("upload failed");
      return res.json().catch(function () { return {}; });
    });
  }
  function beacon(path, body) {
    var data = JSON.stringify(body || {});
    if (navigator.sendBeacon) {
      try {
        if (navigator.sendBeacon(origin + path, new Blob([data], {type: "application/json"}))) return Promise.resolve({});
      } catch (_) {}
    }
    return api(path, {method: "POST", body: body, keepalive: true}).catch(function () { return {}; });
  }
  function wsURL(path, params) {
    var url = new URL(origin + path);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    Object.keys(params || {}).forEach(function (key) {
      if (params[key]) url.searchParams.set(key, params[key]);
    });
    return url.toString();
  }
  function metrikaClientID() { try { return localStorage.getItem("_ym_uid") || ""; } catch (_) { return ""; } }
  function viewportMetrics() {
    var visual = window.visualViewport;
    var width = Math.round((visual && visual.width) || window.innerWidth || 0);
    var height = Math.round((visual && visual.height) || window.innerHeight || 0);
    var doc = documentMetrics();
    return Object.assign({
      viewport_width: width,
      viewport_height: height,
      width: width,
      height: height,
      layout_width: window.innerWidth || width,
      layout_height: window.innerHeight || height,
      viewport_offset_x: Math.round((visual && visual.offsetLeft) || 0),
      viewport_offset_y: Math.round((visual && visual.offsetTop) || 0),
      viewport_scale: Number((visual && visual.scale) || 1),
      dpr: window.devicePixelRatio || 1
    }, doc);
  }
  function documentMetrics() {
    var root = document.documentElement || {};
    var body = document.body || {};
    return {
      document_width: Math.round(Math.max(
        root.scrollWidth || 0,
        body.scrollWidth || 0,
        root.offsetWidth || 0,
        body.offsetWidth || 0,
        root.clientWidth || 0,
        window.innerWidth || 0
      )),
      document_height: Math.round(Math.max(
        root.scrollHeight || 0,
        body.scrollHeight || 0,
        root.offsetHeight || 0,
        body.offsetHeight || 0,
        root.clientHeight || 0,
        window.innerHeight || 0
      ))
    };
  }
  function isWidgetNode(node) {
    if (!node) return false;
    var root = node.getRootNode && node.getRootNode();
    if (root === state.shadow) return true;
    for (var cur = node; cur; cur = cur.parentNode) {
      if (cur === state.root) return true;
    }
    return false;
  }
  function nodePath(node) {
    if (!node) return "";
    if (node.nodeType === 9) node = document.documentElement;
    if (node === document.documentElement) return "html";
    var parts = [];
    for (var cur = node; cur && cur.parentNode; cur = cur.parentNode) {
      if (cur === document.documentElement) {
        parts.unshift("html");
        break;
      }
      var parent = cur.parentNode;
      var index = Array.prototype.indexOf.call(parent.childNodes, cur);
      if (index < 0) break;
      parts.unshift(String(index));
    }
    return parts.join(".");
  }
  function absoluteAttr(value) {
    value = String(value || "").trim();
    if (!value || /^(javascript|data:text\/html)/i.test(value)) return "";
    try { return new URL(value, location.href).toString(); } catch (_) { return value; }
  }
  function absoluteSrcset(value) {
    return String(value || "").split(",").map(function (part) {
      var trimmed = part.trim();
      if (!trimmed) return "";
      var match = trimmed.match(/^(\S+)(.*)$/);
      if (!match) return trimmed;
      var url = absoluteAttr(match[1]);
      return url ? url + (match[2] || "") : "";
    }).filter(Boolean).join(", ");
  }
  function selectorLabel(el) {
    if (!el || el.nodeType !== 1) return "";
    var tag = el.tagName.toLowerCase();
    var id = el.id ? "#" + el.id : "";
    var cls = String(el.className || "").split(/\s+/).filter(Boolean).slice(0, 3).map(function (part) { return "." + part; }).join("");
    return tag + id + cls;
  }
  function isBlockedMirrorElement(el, block) {
    if (!el || el.nodeType !== 1 || isWidgetNode(el)) return true;
    var tag = el.tagName;
    if (/^(TEMPLATE)$/i.test(tag)) return true;
    return block.some(function (sel) { try { return el.matches && el.matches(sel); } catch (_) { return false; } });
  }
  function isMaskedMirrorElement(el, mask) {
    if (!el || el.nodeType !== 1) return false;
    var tag = el.tagName;
    var type = el.getAttribute("type") || "";
    var autocomplete = el.getAttribute("autocomplete") || "";
    var sensitive = /^(INPUT|TEXTAREA|SELECT)$/i.test(tag) || el.isContentEditable || /password|email|tel|search|number/i.test(type) || /cc-|card|password|one-time-code/i.test(autocomplete);
    return sensitive || mask.some(function (sel) { try { return el.matches && el.matches(sel); } catch (_) { return false; } });
  }
  function sanitizeMirrorClone(original, clone, path, maskedParent, block, mask) {
    if (!original || !clone) return;
    if (original.nodeType === 3) {
      if (maskedParent) clone.textContent = "••••";
      return;
    }
    if (original.nodeType !== 1 || clone.nodeType !== 1) return;
    if (isBlockedMirrorElement(original, block)) {
      if (clone.parentNode) clone.parentNode.removeChild(clone);
      return;
    }
    var tag = original.tagName.toUpperCase();
    if (/^(IFRAME|OBJECT|EMBED|CANVAS|VIDEO|AUDIO)$/i.test(tag)) {
      var placeholder = document.createElement("div");
      placeholder.setAttribute("data-rp-path", path);
      placeholder.setAttribute("class", "rp-mirror-placeholder");
      placeholder.textContent = tag.toLowerCase() + " скрыт в live-показе";
      if (clone.parentNode) clone.parentNode.replaceChild(placeholder, clone);
      return;
    }
    clone.setAttribute("data-rp-path", path);
    Array.prototype.slice.call(clone.attributes || []).forEach(function (attr) {
      var name = attr.name.toLowerCase();
      if (name.indexOf("on") === 0 || name === "srcdoc" || name === "autofocus") clone.removeAttribute(attr.name);
      if ((name === "href" || name === "src" || name === "poster" || name === "action") && clone.hasAttribute(attr.name)) {
        var next = absoluteAttr(clone.getAttribute(attr.name));
        if (next) clone.setAttribute(attr.name, next); else clone.removeAttribute(attr.name);
      }
      if (name === "srcset" && clone.hasAttribute(attr.name)) {
        var srcset = absoluteSrcset(clone.getAttribute(attr.name));
        if (srcset) clone.setAttribute(attr.name, srcset); else clone.removeAttribute(attr.name);
      }
    });
    var masked = maskedParent || isMaskedMirrorElement(original, mask);
    if (masked) clone.setAttribute("data-rp-masked", "true");
    if (/^(INPUT|TEXTAREA|SELECT)$/i.test(tag)) {
      clone.removeAttribute("name");
      clone.setAttribute("value", masked ? "••••" : "");
      if (tag === "TEXTAREA") clone.textContent = masked ? "••••" : "";
      if (tag === "SELECT") {
        Array.prototype.slice.call(clone.options || []).forEach(function (opt) { opt.textContent = masked ? "••••" : opt.textContent; });
      }
    }
    var originalChildren = Array.prototype.slice.call(original.childNodes || []);
    var cloneChildren = Array.prototype.slice.call(clone.childNodes || []);
    for (var i = 0; i < originalChildren.length && i < cloneChildren.length; i += 1) {
      sanitizeMirrorClone(originalChildren[i], cloneChildren[i], path + "." + i, masked, block, mask);
    }
  }
  function collectScrollableState(limit) {
    limit = limit || 80;
    var seen = [];
    var items = [];
    function add(el) {
      if (!el || el.nodeType !== 1 || isWidgetNode(el) || seen.indexOf(el) >= 0) return;
      seen.push(el);
      var top = Math.round(el === document.scrollingElement ? (window.scrollY || el.scrollTop || 0) : (el.scrollTop || 0));
      var left = Math.round(el === document.scrollingElement ? (window.scrollX || el.scrollLeft || 0) : (el.scrollLeft || 0));
      var scrollW = Math.round(el.scrollWidth || 0);
      var scrollH = Math.round(el.scrollHeight || 0);
      var clientW = Math.round(el.clientWidth || window.innerWidth || 0);
      var clientH = Math.round(el.clientHeight || window.innerHeight || 0);
      if (top || left || scrollH > clientH + 2 || scrollW > clientW + 2 || el === document.scrollingElement) {
        items.push({
          path: nodePath(el),
          selector: selectorLabel(el),
          top: top,
          left: left,
          width: clientW,
          height: clientH,
          scroll_width: scrollW,
          scroll_height: scrollH
        });
      }
    }
    add(document.scrollingElement || document.documentElement);
    Array.prototype.slice.call(document.querySelectorAll("*")).some(function (el) {
      add(el);
      return items.length >= limit;
    });
    return items.slice(0, limit);
  }
  function collectViewportContext(limit) {
    limit = limit || 8;
    var monitoring = setting("monitoring", {});
    var block = monitoring.block_selectors || ["[data-rp-block]", ".rp-block"];
    var mask = monitoring.mask_selectors || ["[data-rp-mask]", ".rp-mask"];
    var vw = Math.max(window.innerWidth || 0, 1);
    var vh = Math.max(window.innerHeight || 0, 1);
    var viewportArea = vw * vh;
    var seen = [];
    var items = [];
    function blocked(el) {
      if (!el || el.nodeType !== 1 || isWidgetNode(el)) return true;
      return block.some(function (sel) { try { return el.matches && el.matches(sel); } catch (_) { return false; } });
    }
    function masked(el) {
      if (!el || el.nodeType !== 1) return false;
      var tag = el.tagName;
      var type = el.getAttribute("type") || "";
      var autocomplete = el.getAttribute("autocomplete") || "";
      var sensitive = /^(INPUT|TEXTAREA|SELECT)$/i.test(tag) || el.isContentEditable || /password|email|tel|search|number/i.test(type) || /cc-|card|password|one-time-code/i.test(autocomplete);
      return sensitive || mask.some(function (sel) { try { return el.matches && el.matches(sel); } catch (_) { return false; } });
    }
    function cleanText(value, max) {
      value = String(value || "").replace(/\s+/g, " ").trim();
      return value.length > max ? value.slice(0, max - 1) + "…" : value;
    }
    function textFor(el, max) {
      if (!el || masked(el)) return "";
      var text = "";
      try { text = el.innerText || el.textContent || ""; } catch (_) { text = el.textContent || ""; }
      return cleanText(text, max || 260);
    }
    function headingFor(el) {
      if (!el) return "";
      var heading = null;
      try { heading = el.matches("h1,h2,h3,[role='heading']") ? el : el.querySelector("h1,h2,h3,[role='heading']"); } catch (_) {}
      return heading ? cleanText(heading.textContent || "", 120) : "";
    }
    function add(el, reason) {
      if (!el || seen.indexOf(el) >= 0 || blocked(el)) return;
      seen.push(el);
      var rect;
      try { rect = el.getBoundingClientRect(); } catch (_) { return; }
      if (!rect || rect.width < 8 || rect.height < 8) return;
      var left = Math.max(rect.left, 0);
      var top = Math.max(rect.top, 0);
      var right = Math.min(rect.right, vw);
      var bottom = Math.min(rect.bottom, vh);
      var visibleW = Math.max(0, right - left);
      var visibleH = Math.max(0, bottom - top);
      var visibleArea = visibleW * visibleH;
      if (visibleArea < 96) return;
      var text = textFor(el, 360);
      var heading = headingFor(el);
      if (!text && !heading) return;
      var tag = el.tagName.toLowerCase();
      var role = el.getAttribute("role") || "";
      var style = {};
      try { style = getComputedStyle(el); } catch (_) {}
      var position = style.position || "";
      var isOverlay = /^(DIALOG)$/i.test(el.tagName) || /dialog|menu/i.test(role) || el.getAttribute("aria-modal") === "true";
      var isChrome = !isOverlay && (/^(HEADER|NAV|FOOTER)$/i.test(el.tagName) || /navigation/i.test(role) || ((position === "fixed" || position === "sticky") && rect.height < vh * 0.35));
      var score = (visibleArea / viewportArea) + (heading ? 0.15 : 0) + (/^(MAIN|ARTICLE|SECTION|DIALOG)$/i.test(el.tagName) ? 0.16 : 0) + (/dialog|main/i.test(role) ? 0.12 : 0) + (reason === "viewport_sample" ? 0.08 : 0);
      if (isOverlay) score += 0.5;
      if (isChrome) score -= 0.45;
      items.push({
        path: nodePath(el),
        selector: selectorLabel(el),
        tag: tag,
        role: role,
        position: position,
        is_page_chrome: isChrome,
        reason: reason || "",
        heading: heading,
        text: text,
        visible_ratio: Number(Math.min(1, visibleArea / viewportArea).toFixed(3)),
        rect: {
          x: Math.round(rect.left),
          y: Math.round(rect.top),
          width: Math.round(rect.width),
          height: Math.round(rect.height)
        },
        score: Number(score.toFixed(4))
      });
    }
    function nearestBlock(el) {
      for (var cur = el; cur && cur.nodeType === 1 && cur !== document.body; cur = cur.parentElement) {
        try {
          if (cur.matches("main,section,article,aside,nav,header,footer,dialog,[role='dialog'],[role='main'],[role='navigation'],[aria-modal='true'],details[open],[aria-expanded='true'],[data-section],.modal,.popup,.drawer,.dropdown,.menu,.hero,.section,.card")) return cur;
        } catch (_) {}
      }
      return el;
    }
    Array.prototype.slice.call(document.querySelectorAll("main,section,article,aside,nav,header,footer,dialog,[role='dialog'],[role='main'],[role='navigation'],[aria-modal='true'],details[open],[aria-expanded='true'],[data-section],.modal,.popup,.drawer,.dropdown,.menu,.hero,.section,.card,h1,h2,h3"), 0, 220).forEach(function (el) {
      add(el, "candidate");
    });
    [[0.5,0.18],[0.5,0.5],[0.5,0.82],[0.22,0.5],[0.78,0.5]].forEach(function (point) {
      var el = null;
      try { el = document.elementFromPoint(Math.round(vw * point[0]), Math.round(vh * point[1])); } catch (_) {}
      if (el) add(nearestBlock(el), "viewport_sample");
    });
    items.sort(function (a, b) { return b.score - a.score; });
    var blocks = items.slice(0, limit).map(function (item) {
      delete item.score;
      return item;
    });
    return {
      version: 1,
      captured_at: new Date().toISOString(),
      url: location.href,
      title: document.title || "",
      viewport: {
        width: vw,
        height: vh,
        scroll_x: Math.round(window.scrollX || 0),
        scroll_y: Math.round(window.scrollY || 0)
      },
      primary_block: blocks[0] || null,
      blocks: blocks,
      visible_text: cleanText(blocks.map(function (item) { return item.heading || item.text; }).filter(Boolean).join(" · "), 1200)
    };
  }
  function mirrorControllerSource() {
    return [
      "(function(){",
      "if(window.__RP_MIRROR_CONTROLLER_INSTALLED__)return;",
      "window.__RP_MIRROR_CONTROLLER_INSTALLED__=true;",
      "window.__RP_MIRROR_CONTROLLER_READY__=true;",
      "window.__RP_MIRROR_APPLY_COUNT__=0;",
      "window.__RP_MIRROR_LAST_STATE__=null;",
      "var applySoonTimers=[];",
      "function apply(data){data=data||{};window.__RP_MIRROR_APPLY_COUNT__+=1;window.__RP_MIRROR_LAST_STATE__=data;",
      "var sx=Number(data.scrollX||0),sy=Number(data.scrollY||0),vw=Math.max(Number(data.viewportWidth||innerWidth||1),1),vh=Math.max(Number(data.viewportHeight||innerHeight||1),1),dw=Math.max(Number(data.documentWidth||vw),vw),dh=Math.max(Number(data.documentHeight||vh),vh);",
      "var html=document.documentElement,body=document.body;",
      "if(html){html.style.overflow='auto';html.style.width=vw+'px';html.style.minWidth=dw+'px';html.style.minHeight=dh+'px';html.style.scrollbarWidth='none';html.dataset.rpAppliedScrollX=String(Math.round(sx));html.dataset.rpAppliedScrollY=String(Math.round(sy));}",
      "if(body){body.style.overflow='visible';body.style.width=vw+'px';body.style.minWidth=dw+'px';body.style.minHeight=dh+'px';body.style.scrollbarWidth='none';body.dataset.rpAppliedScrollX=String(Math.round(sx));body.dataset.rpAppliedScrollY=String(Math.round(sy));}",
      "try{scrollTo({left:sx,top:sy,behavior:'auto'});}catch(_){try{scrollTo(sx,sy);}catch(__){}}",
      "var items=Array.isArray(data.scrollContainers)?data.scrollContainers:[];items.forEach(function(item){if(!item||!item.path)return;try{var node=document.querySelector('[data-rp-path=\"'+String(item.path).replace(/\"/g,'\\\\\"')+'\"]');if(node){node.scrollLeft=Number(item.left||0);node.scrollTop=Number(item.top||0);}}catch(_){}});",
      "try{parent.postMessage({type:'rp:webvisor:mirror-applied',scrollY:Math.round(scrollY||0),requestedScrollY:Math.round(sy),applyCount:window.__RP_MIRROR_APPLY_COUNT__},'*');}catch(_){}",
      "}",
      "function applySoon(){var data=window.__RP_MIRROR_LAST_STATE__;if(!data)return;applySoonTimers.forEach(function(timer){clearTimeout(timer);});applySoonTimers=[];[120,320,800,1600,2600].forEach(function(delay){applySoonTimers.push(setTimeout(function(){apply(data);},delay));});}",
      "setInterval(function(){var data=window.__RP_MIRROR_LAST_STATE__;if(!data)return;var wantedY=Number(data.scrollY||0),current=Math.round(scrollY||document.scrollingElement&&document.scrollingElement.scrollTop||0);if(Math.abs(current-wantedY)>12)apply(data);},750);",
      "addEventListener('message',function(event){var data=event&&event.data;if(!data||data.type!=='rp:webvisor:apply-state')return;apply(data);applySoon();});",
      "addEventListener('load',applySoon);",
      "document.addEventListener('DOMContentLoaded',applySoon);",
      "try{parent.postMessage({type:'rp:webvisor:mirror-ready'},'*');}catch(_){}",
      "})();"
    ].join("");
  }
  function mirrorSnapshot(reason) {
    var monitoring = setting("monitoring", {});
    var block = monitoring.block_selectors || ["[data-rp-block]", ".rp-block"];
    var mask = monitoring.mask_selectors || ["[data-rp-mask]", ".rp-mask"];
    var maxBytes = Math.max(80000, Math.min(Number(monitoring.max_dom_bytes || 650000), 1200000));
    var root = document.documentElement;
    var clone = root && root.cloneNode(true);
    if (!clone) return null;
    sanitizeMirrorClone(root, clone, "html", false, block, mask);
    var head = clone.querySelector("head") || clone.insertBefore(document.createElement("head"), clone.firstChild);
    var base = document.createElement("base");
    base.setAttribute("href", location.href);
    head.insertBefore(base, head.firstChild);
    var controller = document.createElement("script");
    controller.setAttribute("data-rp-mirror-controller", "");
    controller.textContent = mirrorControllerSource();
    head.insertBefore(controller, base.nextSibling);
    var style = document.createElement("style");
    style.setAttribute("data-rp-mirror-style", "");
    style.textContent = "*,*::before,*::after{animation:none!important;transition:none!important;caret-color:transparent!important}html,body{margin:0!important;min-height:100%;scrollbar-width:none!important}::-webkit-scrollbar{display:none!important}.rp-mirror-placeholder{display:grid;place-items:center;min-height:72px;border:1px dashed #b7c2d0;border-radius:8px;background:#f6f8fb;color:#64748b;font:13px system-ui,sans-serif}";
    head.appendChild(style);
    var html = "<!doctype html>" + clone.outerHTML;
    var truncated = false;
    if (html.length > maxBytes) {
      truncated = true;
      html = '<!doctype html><html><head><base href="' + esc(location.href) + '"><style>body{margin:0;font:14px system-ui,sans-serif;background:#f8fafc;color:#334155}.rp-too-large{padding:24px}</style></head><body data-rp-path="html.1"><div class="rp-too-large"><strong>DOM слишком большой для live-показа</strong><p>Курсор, клики, скролл и события продолжают передаваться.</p></div></body></html>';
    }
    return {
      version: 1,
      reason: reason || "snapshot",
      base_url: location.href,
      captured_at: new Date().toISOString(),
      html: html,
      dom_size: html.length,
      truncated: truncated,
      scroll_containers: collectScrollableState(80)
    };
  }
  function scheduleMirrorIncremental(reason) {
    if (!state.monitoring) return;
    state.mirrorReason = reason || state.mirrorReason || "mutation";
    if (state.mirrorTimer) return;
    state.mirrorTimer = setTimeout(function () {
      var mirror = mirrorSnapshot(state.mirrorReason || "mutation");
      var payload = scrollPayload();
      payload.mutation_count = state.mutationCount;
      payload.title = document.title;
      payload.url = location.href;
      payload.mirror = mirror;
      state.mirrorTimer = 0;
      state.mirrorReason = "";
      pushEvent("monitoring.incremental", payload, true);
      state.mutationCount = 0;
    }, 450);
  }
  function presenceBody() {
    var metrics = viewportMetrics();
    return {
      public_key: publicKey,
      anonymous_id: state.anonymousId,
      visit_session_id: state.visitSessionId,
      url: location.href,
      title: document.title || "",
      referrer: document.referrer || "",
      viewport_width: metrics.viewport_width,
      viewport_height: metrics.viewport_height,
      scroll_x: Math.round(window.scrollX || 0),
      scroll_y: Math.round(window.scrollY || 0),
      duration_seconds: elapsedSeconds(),
      viewport_context: collectViewportContext(6),
      browser: navigator.userAgentData && navigator.userAgentData.brands ? navigator.userAgentData.brands.map(function (b) { return b.brand; }).join(", ") : "",
      os: navigator.platform || "",
      device_type: /Mobi|Android/i.test(navigator.userAgent) ? "mobile" : "desktop",
      screen_width: screen.width || 0,
      screen_height: screen.height || 0,
      language: navigator.language || "",
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "",
      metrika_client_id: metrikaClientID(),
      user_agent: navigator.userAgent || "",
      metadata: {conversation_id: state.conversationId, page_session_id: state.pageSessionId}
    };
  }
  function loadConfig() {
    return api("/api/v1/public/site-chat/widgets/" + encodeURIComponent(publicKey) + "/config").then(function (cfg) {
      state.cfg = Object.assign(state.cfg, cfg || {});
      state.voiceConfig = cfg && cfg.voice ? cfg.voice : state.voiceConfig;
      if (!state.cfg.accent_color) state.cfg.accent_color = setting("theme.accent_color", "#00B2FF");
      if (!state.cfg.position) state.cfg.position = setting("theme.position", "right");
      if (/ответ/i.test(state.cfg.subtitle || "")) state.cfg.subtitle = "";
    }).catch(function () {});
  }
  function presence() {
    return api("/api/v1/public/site-chat/presence", {method: "POST", body: presenceBody()}).then(function (data) {
      state.visitorId = data.visitor_id || state.visitorId;
      state.visitSessionId = data.visit_session_id || state.visitSessionId;
      state.conversationId = data.conversation_id || state.conversationId;
      set("visitor_id", state.visitorId);
      setSession("visit_session_id", state.visitSessionId);
      setSession("conversation_id", state.conversationId);
      if (data.voice) state.voiceConfig = data.voice;
      syncMonitoringStatus(data);
      setStatus(state.connected ? "На связи" : "Подключаемся");
      return data;
    }).catch(function () {
      setStatus("Связь восстанавливается");
    });
  }
  function syncMonitoringStatus(data) {
    var status = data && data.monitoring_status;
    if (status === "active" && !state.monitoring) startMonitoringCapture();
    if (status === "stopped" && state.monitoring) stopMonitoringCapture();
  }
  function connectWS() {
    if (!state.visitorId || !state.visitSessionId || !state.conversationId || state.ws) return;
    var ws = new WebSocket(wsURL("/api/v1/public/site-chat/ws", {
      public_key: publicKey,
      visitor_id: state.visitorId,
      visit_session_id: state.visitSessionId,
      conversation_id: state.conversationId,
      connection_id: makeID("conn")
    }));
    state.ws = ws;
    ws.onopen = function () {
      state.connected = true;
      state.reconnecting = false;
      state.reconnectAttempts = 0;
      if (state.reconnectTimer) clearTimeout(state.reconnectTimer);
      state.reconnectTimer = 0;
      setStatus("На связи");
      stopPolling();
      flushWSQueue();
      heartbeat();
      state.heartbeatTimer = setInterval(heartbeat, 10000);
      sendWS("viewport", viewportPayload(), true);
      sendWS("page.navigate", pagePayload(), true);
    };
    ws.onclose = function () {
      state.connected = false;
      state.ws = null;
      clearInterval(state.heartbeatTimer);
      setStatus("Офлайн, включен резервный режим");
      startPolling();
      scheduleReconnect();
    };
    ws.onerror = function () {
      state.reconnecting = true;
      setStatus("Связь восстанавливается");
    };
    ws.onmessage = function (event) {
      var msg;
      try { msg = JSON.parse(event.data); } catch (_) { return; }
      handleWS(msg);
    };
  }
  function sendWS(type, payload, low) {
    payload = payload || {};
    var payloadCall = payload.call_session && payload.call_session.call_session_id;
    var callID = state.call && state.call.call_session_id || payload.call_session_id || payloadCall || "";
    var msg = {
      id: makeID("msg"),
      type: type,
      sent_at: new Date().toISOString(),
      visitor_id: state.visitorId,
      visit_session_id: state.visitSessionId,
      conversation_id: state.conversationId,
      call_session_id: callID,
      payload: payload
    };
    if (state.ws && state.ws.readyState === WebSocket.OPEN) {
      if (low && state.ws.bufferedAmount > 128000) return;
      state.ws.send(JSON.stringify(msg));
      return;
    }
    if (isRealtimeSignal(type)) {
      state.wsQueue.push(msg);
      state.wsQueue = state.wsQueue.slice(-120);
      connectWS();
      return;
    }
    if (!low) startPolling();
  }
  function isRealtimeSignal(type) {
    return /^webrtc\./.test(type) || /^call\./.test(type) || /^consent\./.test(type);
  }
  function flushWSQueue() {
    if (!state.ws || state.ws.readyState !== WebSocket.OPEN || !state.wsQueue.length) return;
    var queue = state.wsQueue.splice(0);
    queue.forEach(function (msg) {
      try { state.ws.send(JSON.stringify(msg)); } catch (_) {}
    });
  }
  function scheduleReconnect() {
    if (state.reconnectTimer) return;
    state.reconnectAttempts += 1;
    var delay = Math.min(30000, 2500 * Math.pow(1.6, Math.max(0, state.reconnectAttempts - 1)));
    state.reconnectTimer = setTimeout(function () {
      state.reconnectTimer = 0;
      connectWS();
    }, delay);
  }
  function elapsedSeconds() {
    return Math.max(0, Math.round((Date.now() - state.startedAt) / 1000));
  }
  function heartbeat() { sendWS("visitor.heartbeat", {url: location.href, title: document.title, duration_seconds: elapsedSeconds()}, true); }
  function handleWS(msg) {
    if (msg.type === "operator.invite") showInvite(msg.payload || {});
    if (msg.type === "chat.message") {
      loadMessages();
      if ((msg.payload || {}).author_kind === "operator" && !state.open) {
        showMessageToast((msg.payload || {}).body || "В чате новый ответ.");
      }
    }
    if (msg.type === "monitoring.start") startMonitoringCapture();
    if (msg.type === "monitoring.stop") stopMonitoringCapture();
    if (msg.type === "call.offer") incomingCall(msg.payload && msg.payload.call_session ? msg.payload.call_session : null);
    if (msg.type === "call.accept") handleCallAccepted(msg.payload && msg.payload.call_session);
    if (msg.type === "call.reject" || msg.type === "call.end") endCallUI(msg.type);
    if (msg.type === "call.audio_chunk") playAudioRelayChunk(msg.payload || {});
    if (msg.type === "webrtc.offer" || msg.type === "webrtc.answer" || msg.type === "webrtc.ice") handleSignal(msg);
  }
  function render() {
    ["riders-site-chat-root-", "rp-live-assist-root-"].forEach(function (prefix) {
      var old = document.getElementById(prefix + publicKey);
      if (old) old.remove();
    });
    Array.prototype.slice.call(document.querySelectorAll(".rp-site-chat")).forEach(function (node) {
      if (node.getAttribute("data-site-chat-key") === publicKey) node.remove();
    });
    state.root = document.createElement("div");
    state.root.id = "riders-site-chat-root-" + publicKey;
    state.root.className = "rp-site-chat";
    state.root.setAttribute("data-site-chat-key", publicKey);
    state.shadow = state.root.attachShadow({mode: "open"});
    state.shadow.innerHTML = "<style>" + css() + "</style>" + markup();
    document.body.appendChild(state.root);
    bind();
    paintMessages();
  }
  function markup() {
    var primary = labelSetting("launcher.primary_text", "Консультация онлайн");
    var callText = labelSetting("launcher.call_text", "Звонок");
    var subtitle = state.cfg.subtitle ? '<span>' + esc(state.cfg.subtitle) + '</span>' : "";
    return '' +
      '<section class="rp-wrap rp-' + esc(state.cfg.position === "left" ? "left" : "right") + '">' +
      '  <div class="rp-toast" hidden data-toast></div>' +
      '  <div class="rp-invite" hidden data-invite></div>' +
      '  <div class="rp-monitor" hidden data-monitor><strong>Показ экрана включён</strong><span>Можно остановить в любой момент.</span><button data-stop-monitor>Остановить</button></div>' +
      '  <section class="rp-panel" role="complementary" aria-label="Чат с командой" hidden data-panel>' +
      '    <header class="rp-head"><div class="rp-brand"><span class="rp-avatar" aria-hidden="true"></span><div><strong>' + esc(state.cfg.title) + '</strong>' + subtitle + '</div></div><div class="rp-head-actions"><button class="rp-icon rp-close" aria-label="Закрыть" data-close>×</button></div></header>' +
      '    <div class="rp-conversation" aria-live="polite"><div><span class="rp-presence-dot" aria-hidden="true"></span><span class="rp-status" data-status>Подключаемся</span></div><div class="rp-call-controls"><button class="rp-call-mini" data-call-open>' + esc(callText) + '</button></div></div>' +
      '    <div class="rp-call" hidden data-call></div>' +
      '    <details class="rp-profile"><summary>Контакты для ответа <small>необязательно</small></summary><label><span>Имя</span><input data-name autocomplete="name" placeholder="Как к вам обращаться"></label><label><span>Телефон</span><input data-contact autocomplete="tel" inputmode="tel" placeholder="+7"></label></details>' +
      '    <div class="rp-agreements" data-agreements><label><input type="checkbox" data-privacy> <span>Согласен на обработку персональных данных и связь по обращению</span></label>' + linksHTML() + '</div>' +
      '    <div class="rp-consent" hidden data-consent></div>' +
      '    <audio class="rp-remote-audio" data-remote-audio autoplay playsinline></audio>' +
      '    <div class="rp-messages" role="log" aria-label="Переписка" aria-live="polite" data-messages></div>' +
      '    <div class="rp-pending" hidden data-pending></div>' +
      '    <div class="rp-send-error" role="alert" hidden data-send-error></div><footer class="rp-compose"><textarea aria-label="Сообщение" data-draft rows="2" placeholder="Напишите сообщение"></textarea><input type="file" multiple hidden data-file><button class="rp-attach" data-attach title="Прикрепить файл" aria-label="Прикрепить файл">＋</button><button class="rp-send" data-send>Отправить</button></footer>' +
      '  </section>' +
      '  <div class="rp-launcher" data-launcher><button class="rp-main" data-open><span class="rp-main-mark" aria-hidden="true"></span><span><strong>' + esc(primary) + '</strong></span></button></div>' +
      '</section>';
  }
  function bind() {
    var q = function (sel) { return state.shadow.querySelector(sel); };
    state.els.panel = q("[data-panel]");
    state.els.launcher = q("[data-launcher]");
    state.els.status = q("[data-status]");
    state.els.messages = q("[data-messages]");
    state.els.name = q("[data-name]");
    state.els.contact = q("[data-contact]");
    state.els.privacy = q("[data-privacy]");
    state.els.agreements = q("[data-agreements]");
    state.els.draft = q("[data-draft]");
    state.els.sendError = q("[data-send-error]");
    state.els.invite = q("[data-invite]");
    state.els.toast = q("[data-toast]");
    state.els.call = q("[data-call]");
    state.els.callOpen = q("[data-call-open]");
    state.els.consent = q("[data-consent]");
    state.els.monitor = q("[data-monitor]");
    state.els.file = q("[data-file]");
    state.els.pending = q("[data-pending]");
    state.els.remoteAudio = q("[data-remote-audio]");
    state.els.name.value = state.visitorName;
    state.els.contact.value = state.visitorContact;
    state.els.privacy.checked = state.privacyAccepted;
    q("[data-open]").addEventListener("click", function () { unlockAudio(); openPanel(true); });
    state.els.callOpen.addEventListener("click", requestCall);
    q("[data-close]").addEventListener("click", function () { openPanel(false); });
    q("[data-send]").addEventListener("click", sendMessage);
    q("[data-attach]").addEventListener("click", function () { state.els.file.click(); });
    q("[data-stop-monitor]").addEventListener("click", function () { sendWS("monitoring.stop", {}, false); stopMonitoringCapture(); });
    state.els.name.addEventListener("input", function () { state.visitorName = state.els.name.value; set("name", state.visitorName); });
    state.els.contact.addEventListener("input", function () { state.visitorContact = state.els.contact.value; set("contact", state.visitorContact); });
    state.els.privacy.addEventListener("change", function () { state.privacyAccepted = !!state.els.privacy.checked; set("privacy_ok", state.privacyAccepted ? "1" : ""); });
    state.els.file.addEventListener("change", function () { addFiles(state.els.file.files); state.els.file.value = ""; });
    state.els.draft.addEventListener("keydown", function (e) {
      if (e.key !== "Enter" || e.isComposing) return;
      if (e.ctrlKey || e.metaKey || e.shiftKey) {
        e.preventDefault();
        insertDraftNewline();
        return;
      }
      e.preventDefault();
      sendMessage();
    });
    state.els.draft.addEventListener("paste", function (e) {
      var files = e.clipboardData && e.clipboardData.files;
      if (files && files.length) addFiles(files);
    });
    state.shadow.addEventListener("keydown", function (e) { if (e.key === "Escape") openPanel(false); });
    state.shadow.addEventListener("click", unlockAudio, {once: true});
  }
  function openPanel(open) {
    state.open = !!open;
    state.els.panel.hidden = !state.open;
    state.els.launcher.hidden = state.open;
    if (!state.open) state.shadow.querySelector("[data-open]").focus();
    if (state.open) {
      hideToast();
      ensureConversation().then(loadMessages).catch(function () { setStatus("Не удалось подключиться. Повторите попытку."); });
      setTimeout(function () { state.els.draft && state.els.draft.focus(); }, 40);
    }
    sendWS(state.open ? "chat.typing" : "chat.closed", {open: state.open}, true);
    pushEvent("widget_state", {open: state.open, active_tab: "chat", url: location.href, title: document.title}, false);
    if (state.monitoring) scheduleMirrorIncremental(state.open ? "widget-open" : "widget-close");
    if (!state.connected) flushEvents(false);
  }
  function setStatus(text) { if (state.els.status) state.els.status.textContent = text; }
  function ensureConversation() {
    if (state.conversationId) return Promise.resolve(state.conversationId);
    if (state.conversationPromise) return state.conversationPromise;
    state.conversationPromise = presence().then(function () {
      state.conversationPromise = null;
      return state.conversationId;
    }).catch(function (err) {
      state.conversationPromise = null;
      throw err;
    });
    return state.conversationPromise;
  }
  function loadMessages() {
    if (!state.conversationId || !state.visitorId) return;
    api("/api/v1/public/site-chat/conversations/" + encodeURIComponent(state.conversationId) + "/messages?public_key=" + encodeURIComponent(publicKey) + "&visitor_id=" + encodeURIComponent(state.visitorId))
      .then(function (rows) {
        var next = Array.isArray(rows) ? rows.filter(function (message) { return !isCallRecordingMessage(message); }).slice(-220) : [];
        var latestOperator = null;
        for (var i = next.length - 1; i >= 0; i -= 1) {
          if (next[i].author_kind === "operator") {
            latestOperator = next[i];
            break;
          }
        }
        if (latestOperator && latestOperator.message_id !== state.lastOperatorMessageId) {
          var isFresh = Date.parse(latestOperator.created_at || "") >= state.startedAt - 1000;
          if (!state.open && isFresh) showMessageToast(latestOperator.body || "В чате новый ответ.");
          state.lastOperatorMessageId = latestOperator.message_id;
        }
        state.messages = next;
        paintMessages();
      })
      .catch(function () {});
  }
  function isCallRecordingMessage(message) {
    var body = String(message && message.body || "").trim().toLowerCase();
    if (body !== "запись звонка" && body !== "запись разговора") return false;
    var attachments = Array.isArray(message.attachments) ? message.attachments : [];
    return attachments.some(function (file) { return file && (file.call_session_id || file.kind === "audio"); });
  }
  function paintMessages() {
    if (!state.els.messages) return;
    if (!state.messages.length) {
      state.els.messages.innerHTML = '<div class="rp-empty"><strong>Начните диалог</strong></div>';
      return;
    }
    var follow = state.followMessages || state.els.messages.scrollHeight - state.els.messages.scrollTop - state.els.messages.clientHeight < 80;
    var scrollTop = state.els.messages.scrollTop;
    state.followMessages = false;
    var lastDay = "";
    state.els.messages.innerHTML = state.messages.map(function (m) {
      var day = dayLabel(m.created_at);
      var sep = day !== lastDay ? '<div class="rp-day">' + esc(day) + '</div>' : "";
      lastDay = day;
      var system = m.author_kind === "system";
      return sep + '<article class="rp-msg ' + (system ? "system" : (m.author_kind === "visitor" ? "mine" : "their")) + '"><span>' + esc(system ? "Событие" : (m.author_kind === "operator" ? "Команда" : (m.author_name || "Вы"))) + ' · ' + timeLabel(m.created_at) + '</span>' + (m.body ? '<p>' + esc(m.body || "") + '</p>' : "") + attachmentsHTML(m.attachments || []) + '</article>';
    }).join("");
    state.els.messages.scrollTop = follow ? state.els.messages.scrollHeight : scrollTop;
  }
  function attachmentsHTML(items) {
    if (!items.length) return "";
    return '<div class="rp-attachments">' + items.map(function (file) {
      var url = absoluteURL(file.public_url);
      var name = esc(file.file_name || "file");
      if (file.kind === "image") return '<a class="rp-att rp-att-image" href="' + esc(url) + '" target="_blank" rel="noreferrer"><img src="' + esc(url) + '" alt="' + name + '"><span>' + name + '</span></a>';
      if (file.kind === "audio") return '<div class="rp-att rp-att-audio"><span>' + name + '</span><audio controls src="' + esc(url) + '"></audio></div>';
      return '<a class="rp-att" href="' + esc(url) + '" target="_blank" rel="noreferrer"><strong>' + name + '</strong><span>' + esc(formatBytes(file.size_bytes || 0)) + '</span></a>';
    }).join("") + '</div>';
  }
  function dayLabel(raw) {
    var d = raw ? new Date(raw) : new Date();
    if (Number.isNaN(d.getTime())) return "";
    return String(d.getDate()).padStart(2, "0") + "." + String(d.getMonth() + 1).padStart(2, "0") + "." + d.getFullYear();
  }
  function timeLabel(raw) { var d = raw ? new Date(raw) : new Date(); return d.toLocaleTimeString("ru-RU", {hour: "2-digit", minute: "2-digit"}); }
  function sendMessage() {
    if (state.sendingMessage) return;
    var body = (state.els.draft.value || "").trim();
    var attachments = state.pendingAttachments.slice();
    var attachmentIDs = attachments.filter(function (item) { return item.attachment && !item.error; }).map(function (item) { return item.attachment.attachment_id; });
    if (!body && !attachmentIDs.length) return;
    if (state.uploading) { setStatus("Файл ещё загружается."); return; }
    if (!requirePrivacy("Перед отправкой примите согласие на обработку данных.")) return;
    state.sendingMessage = true;
    state.els.sendError.hidden = true;
    state.els.draft.disabled = true;
    var send = state.shadow.querySelector("[data-send]");
    send.disabled = true; send.textContent = "Отправляем…";
    ensureConversation().then(function () {
      // HTTP acknowledges persistence; the websocket remains the realtime receive channel.
      return api("/api/v1/public/site-chat/messages", {method: "POST", body: messagePayload(body, attachmentIDs)});
    }).then(function () {
      state.els.draft.value = "";
      attachments.forEach(function (item) { if (item.localUrl) URL.revokeObjectURL(item.localUrl); });
      state.pendingAttachments = state.pendingAttachments.filter(function (item) { return attachments.indexOf(item) === -1; });
      paintPending(); state.followMessages = true;
      return loadMessages();
    }).catch(function () {
      state.els.sendError.textContent = "Не удалось отправить сообщение. Текст сохранён — попробуйте ещё раз.";
      state.els.sendError.hidden = false;
    }).finally(function () {
      state.sendingMessage = false; state.els.draft.disabled = false;
      send.disabled = false; send.textContent = "Отправить";
      state.els.draft.focus();
    });
  }
  function messagePayload(body, attachmentIDs) {
    return {
      public_key: publicKey,
      conversation_id: state.conversationId,
      visitor_id: state.visitorId,
      visit_session_id: state.visitSessionId,
      author_name: state.visitorName || "Посетитель",
      visitor_contact: state.visitorContact,
      body: body,
      attachment_ids: attachmentIDs || []
    };
  }
  function showMessageToast(text) {
    state.els.toast.hidden = false;
    state.els.toast.innerHTML = '<strong>Новый ответ в чате</strong><p>' + esc(trimLimitText(text, 120)) + '</p><div><button data-reply>Ответить</button><button data-close-toast aria-label="Закрыть">×</button></div>';
    state.els.toast.querySelector("[data-reply]").onclick = function () { hideToast(); openPanel(true); };
    state.els.toast.querySelector("[data-close-toast]").onclick = hideToast;
  }
  function hideToast() {
    if (!state.els.toast) return;
    state.els.toast.hidden = true;
    state.els.toast.innerHTML = "";
  }
  function addFiles(files) {
    if (!files || !files.length) return;
    if (!requirePrivacy("Перед отправкой файла примите согласие под телефоном.")) return;
    ensureConversation().then(function () {
      Array.prototype.slice.call(files, 0, 8).forEach(uploadFile);
    }).catch(function () { setStatus("Не удалось подготовить файл"); });
  }
  function uploadFile(file) {
    if (!file || file.size > 25 * 1024 * 1024) {
      setStatus("Файл должен быть меньше 25 МБ.");
      return;
    }
    var item = {
      id: makeID("file"),
      file: file,
      localUrl: file.type && file.type.indexOf("image/") === 0 ? URL.createObjectURL(file) : "",
      uploading: true,
      error: ""
    };
    state.pendingAttachments.push(item);
    state.uploading += 1;
    paintPending();
    var form = new FormData();
    form.append("file", file);
    form.append("public_key", publicKey);
    form.append("visitor_id", state.visitorId || "");
    form.append("visit_session_id", state.visitSessionId || "");
    upload("/api/v1/public/site-chat/conversations/" + encodeURIComponent(state.conversationId) + "/attachments", form).then(function (attachment) {
      item.attachment = attachment;
      item.uploading = false;
    }).catch(function () {
      item.error = "Не загрузился";
      item.uploading = false;
    }).finally(function () {
      state.uploading = Math.max(0, state.uploading - 1);
      paintPending();
    });
  }
  function paintPending() {
    if (!state.els.pending) return;
    state.els.pending.hidden = state.pendingAttachments.length === 0;
    state.els.pending.innerHTML = state.pendingAttachments.map(function (item) {
      var name = esc(item.file && item.file.name || "file");
      var thumb = item.localUrl ? '<img src="' + esc(item.localUrl) + '" alt="">' : '<span class="rp-file-ico">' + esc(fileExt(name)) + '</span>';
      return '<button class="rp-pending-file ' + (item.error ? "is-error" : "") + '" data-remove="' + esc(item.id) + '">' + thumb + '<span>' + name + '<small>' + esc(item.error || (item.uploading ? "Загрузка" : formatBytes(item.file.size))) + '</small></span></button>';
    }).join("");
    Array.prototype.forEach.call(state.els.pending.querySelectorAll("[data-remove]"), function (node) {
      node.onclick = function () {
        var id = node.getAttribute("data-remove");
        state.pendingAttachments = state.pendingAttachments.filter(function (item) {
          if (item.id !== id) return true;
          if (item.localUrl) URL.revokeObjectURL(item.localUrl);
          return false;
        });
        paintPending();
      };
    });
  }
  function insertDraftNewline() {
    var el = state.els.draft;
    var start = el.selectionStart || 0;
    var end = el.selectionEnd || 0;
    el.value = el.value.slice(0, start) + "\n" + el.value.slice(end);
    el.selectionStart = el.selectionEnd = start + 1;
  }
  function showInvite(payload) {
    var cooldownKey = "invite_cooldown";
    var force = !!payload.force;
    var cooldown = Number(setting("active_invite.cooldown_minutes", 10)) * 60000;
    var last = Number(get(cooldownKey) || "0");
    if (!force && cooldown && Date.now() - last < cooldown) return;
    set(cooldownKey, String(Date.now()));
    state.els.invite.hidden = false;
    var inviteTitle = setting("active_invite.title", "Команда на связи");
    if (inviteTitle === "Оператор на связи") inviteTitle = "Команда на связи";
    state.els.invite.innerHTML = '<strong>' + esc(inviteTitle) + '</strong><p>' + esc(payload.message || setting("active_invite.body", "Помочь на этой странице?")) + '</p><div><button data-accept>' + esc(setting("active_invite.accept_text", "Ответить")) + '</button><button data-call>' + esc(labelSetting("active_invite.call_text", "Звонок")) + '</button><button data-decline>' + esc(setting("active_invite.decline_text", "Не сейчас")) + '</button></div>';
    sendWS("invite.seen", {public_key: publicKey}, false);
    state.els.invite.querySelector("[data-accept]").onclick = function () { unlockAudio(); state.els.invite.hidden = true; openPanel(true); sendWS("invite.accepted", {public_key: publicKey}, false); };
    state.els.invite.querySelector("[data-call]").onclick = function () { unlockAudio(); state.els.invite.hidden = true; sendWS("invite.accepted", {public_key: publicKey}, false); requestCall(); };
    state.els.invite.querySelector("[data-decline]").onclick = function () { state.els.invite.hidden = true; sendWS("invite.declined", {public_key: publicKey}, false); };
  }
  function requestCall() {
    unlockAudio();
    if (!requirePrivacy("Перед звонком примите согласие под телефоном.")) return;
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      setStatus("Браузер не поддерживает звонок с сайта.");
      return;
    }
    showCall("Запрашиваем доступ к микрофону.", "permission");
    state.muted = false;
    navigator.mediaDevices.getUserMedia(callMediaConstraints(false)).then(function (stream) {
      if (!hasLiveAudio(stream)) throw new Error("microphone unavailable");
      stream.getAudioTracks().forEach(function (track) { track.enabled = true; });
      state.media = stream;
      return ensureConversation();
    }).then(function () {
      return api("/api/v1/public/site-chat/calls/request", {method: "POST", body: {
        public_key: publicKey,
        visitor_id: state.visitorId,
        visit_session_id: state.visitSessionId,
        conversation_id: state.conversationId,
        media_kind: "audio",
        privacy_consent_accepted: true,
        recording_enabled: true
      }});
    }).then(function (call) {
      state.call = call;
      showCall("Запрос на звонок отправлен.", "ringing");
      sendWS("call.request", {call_session: call}, false);
    }).catch(function () {
      if (state.media) state.media.getTracks().forEach(function (t) { t.stop(); });
      state.media = null;
      showCall("Браузер не дал доступ к микрофону.", "failed");
    });
  }
  function incomingCall(call) {
    if (!call) return;
    state.call = call;
    playRing();
    state.els.toast.hidden = false;
    state.els.toast.innerHTML = '<strong>Звонок в чате</strong><p>Команда готова подключиться голосом.</p><div><button data-accept-call>Принять</button><button data-reject-call>Отклонить</button></div>';
    state.els.toast.querySelector("[data-accept-call]").onclick = function () {
      stopRing();
      hideToast();
      openPanel(true);
      showConsent(call);
    };
    state.els.toast.querySelector("[data-reject-call]").onclick = function () {
      stopRing();
      hideToast();
      rejectCurrentCall();
    };
  }
  function showConsent(call) {
    var rec = !!call.recording_enabled;
    state.els.consent.hidden = false;
    state.els.consent.innerHTML = '<strong>Звонок в чате</strong><p>' + esc(setting("call_consent_text", "Перед звонком подтвердите доступ к микрофону и обработку данных.")) + '</p><label><input type="checkbox" data-call-privacy' + (state.privacyAccepted ? " checked" : "") + '> Согласен с обработкой данных для звонка</label>' + (rec ? '<label><input type="checkbox" data-call-recording> Согласен на запись разговора</label>' : '') + '<div class="rp-consent-actions"><button data-accept>Принять</button><button data-cancel>Отклонить</button></div>';
    state.els.consent.querySelector("[data-cancel]").onclick = rejectCurrentCall;
    state.els.consent.querySelector("[data-accept]").onclick = function () { acceptCurrentCall(false); };
  }
  function linksHTML() {
    var out = "";
    var privacy = setting("privacy_policy_url", "");
    var terms = setting("call_terms_url", "");
    if (privacy) out += '<a target="_blank" rel="noreferrer" href="' + esc(privacy) + '">Политика конфиденциальности</a>';
    if (terms) out += '<a target="_blank" rel="noreferrer" href="' + esc(terms) + '">Условия звонка</a>';
    return out ? '<div class="rp-links">' + out + '</div>' : "";
  }
  function requirePrivacy(text) {
    if (state.els.privacy && state.els.privacy.checked) return true;
    setStatus(text || "Примите согласие на обработку данных.");
    if (state.els.agreements) {
      state.els.agreements.classList.add("rp-attention");
      setTimeout(function () {
        state.els.agreements && state.els.agreements.classList.remove("rp-attention");
      }, 900);
    }
    return false;
  }
  function acceptCurrentCall(video) {
    var privacy = state.els.consent.querySelector("[data-call-privacy]");
    var recording = state.els.consent.querySelector("[data-call-recording]");
    if (!privacy || !privacy.checked || (recording && !recording.checked)) {
      showCall("Нужно согласие перед запросом доступа к микрофону.", "failed");
      return;
    }
    state.privacyAccepted = true;
    state.els.privacy.checked = true;
    set("privacy_ok", "1");
    state.muted = false;
    navigator.mediaDevices.getUserMedia(callMediaConstraints(video)).then(function (stream) {
      if (!hasLiveAudio(stream)) throw new Error("microphone unavailable");
      stream.getAudioTracks().forEach(function (track) { track.enabled = true; });
      state.media = stream;
      return api("/api/v1/public/site-chat/calls/" + encodeURIComponent(state.call.call_session_id) + "/accept", {method: "POST", body: {
        public_key: publicKey,
        visitor_id: state.visitorId,
        visit_session_id: state.visitSessionId,
        conversation_id: state.conversationId,
        privacy_consent_accepted: true,
        recording_consent_accepted: !!(recording && recording.checked)
      }});
    }).then(function (call) {
      state.call = call;
      state.els.consent.hidden = true;
      showCall("Ждем подключение команды.", "connecting");
      sendWS("call.accept", {call_session: call}, false);
      startAudioRelay();
      startPeer();
    }).catch(function () {
      showCall("Браузер не дал доступ к микрофону.", "failed");
      rejectCurrentCall("media_permission_denied");
    });
  }
  function rejectCurrentCall(reason) {
    if (!state.call) return;
    api("/api/v1/public/site-chat/calls/" + encodeURIComponent(state.call.call_session_id) + "/reject", {method: "POST", body: {
      public_key: publicKey,
      visitor_id: state.visitorId,
      visit_session_id: state.visitSessionId,
      conversation_id: state.conversationId,
      end_reason: reason || "visitor_rejected"
    }}).catch(function () {});
    state.els.consent.hidden = true;
    stopRing();
    sendWS("call.reject", {reason: reason || "visitor_rejected"}, false);
    showCall("Звонок отклонен.", "rejected");
  }
  function endCurrentCall(reason) {
    var call = state.call;
    var endReason = reason || "visitor_ended";
    if (call && call.call_session_id) {
      api("/api/v1/public/site-chat/calls/" + encodeURIComponent(call.call_session_id) + "/end", {method: "POST", body: {
        public_key: publicKey,
        visitor_id: state.visitorId,
        visit_session_id: state.visitSessionId,
        conversation_id: state.conversationId,
        end_reason: endReason
      }}).then(function (next) {
        state.call = next || state.call;
        sendWS("call.end", {call_session: state.call, reason: endReason}, false);
      }).catch(function () {
        sendWS("call.end", {reason: endReason}, false);
      });
    } else {
      sendWS("call.end", {reason: endReason}, false);
    }
    endCallUI("call.end");
  }
  function handleCallAccepted(call) {
    if (call) state.call = call;
    showCall("Ждем подключение команды.", "connecting");
    startAudioRelay();
    startPeer();
  }
  function callStatusLabel(status) {
    if (status === "permission") return "Нужен микрофон";
    if (status === "ringing") return "Ждем подключения";
    if (status === "connecting") return "Подключаем";
    if (status === "active") return "Разговор идет";
    if (status === "rejected") return "Отклонен";
    if (status === "failed") return "Не подключился";
    if (status === "ended") return "Завершен";
    return "Звонок";
  }
  function paintCallControls(text, status) {
    var live = ["permission", "ringing", "connecting", "active"].indexOf(status) >= 0;
    if (state.els.callOpen) {
      state.els.callOpen.hidden = live;
      state.els.callOpen.textContent = labelSetting("launcher.call_text", "Звонок");
    }
    if (text) setStatus(text);
  }
  function bindCallButtons() {
    Array.prototype.forEach.call(state.els.call.querySelectorAll("[data-end]"), function (node) {
      node.onclick = function () { endCurrentCall("visitor_ended"); };
    });
    Array.prototype.forEach.call(state.els.call.querySelectorAll("[data-play-audio]"), function (node) {
      node.onclick = playRemoteAudio;
    });
  }
  function callMediaConstraints(video) {
    return {
      audio: {
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
        channelCount: 1
      },
      video: !!video
    };
  }
  function rtcConfiguration() {
    var cfg = state.voiceConfig || {};
    var iceServers = Array.isArray(cfg.ice_servers) && cfg.ice_servers.length ? cfg.ice_servers : [{urls: "stun:stun.l.google.com:19302"}];
    var policy = cfg.ice_transport_policy === "relay" ? "relay" : "all";
    return {
      iceServers: iceServers,
      iceTransportPolicy: policy,
      bundlePolicy: "balanced",
      rtcpMuxPolicy: "require"
    };
  }
  function hasLiveAudio(stream) {
    return !!(stream && stream.getAudioTracks && stream.getAudioTracks().some(function (track) { return track.readyState === "live"; }));
  }
  function playRemoteAudio() {
    var audio = state.els.remoteAudio;
    if (!audio) return;
    audio.muted = false;
    audio.volume = 1;
    audio.controls = false;
    audio.play().catch(function () {
      setStatus("Нажмите «Включить звук» в звонке.");
    });
  }
  function showCall(text, status) {
    state.els.call.hidden = false;
    paintCallControls(text, status || "");
    var live = ["permission", "ringing", "connecting", "active"].indexOf(status || "") >= 0;
    var actions = live ? '<div class="rp-call-actions"><button data-play-audio>Включить звук</button><button data-end>Отбой</button></div>' : "";
    state.els.call.innerHTML = '<div class="rp-call-copy"><span>Разговор</span><strong>' + esc(text) + '</strong><small>' + esc(callStatusLabel(status || "")) + '</small></div>' + actions;
    bindCallButtons();
  }
  function endCallUI(reason) {
    stopAudioRelay();
    stopRing();
    if (state.media) state.media.getTracks().forEach(function (t) { t.stop(); });
    if (state.peer) state.peer.close();
    if (state.els.remoteAudio) state.els.remoteAudio.srcObject = null;
    state.media = null;
    state.peer = null;
    state.pendingIce = [];
    state.voice = null;
    state.remoteAudio = null;
    state.muted = false;
    showCall(reason === "call.reject" ? "Звонок отклонен." : "Разговор завершен.", "ended");
  }
  function unlockAudio() {
    if (!window.AudioContext && !window.webkitAudioContext) return;
    if (!state.audioCtx) {
      var Ctx = window.AudioContext || window.webkitAudioContext;
      try { state.audioCtx = new Ctx(); } catch (_) { return; }
    }
    if (state.audioCtx.state === "suspended") state.audioCtx.resume().catch(function () {});
  }
  function playRing() {
    unlockAudio();
    if (state.ringTimer) return;
    var playTone = function (frequency, delay, duration, level) {
      if (!state.audioCtx) return;
      try {
        var osc = state.audioCtx.createOscillator();
        var gain = state.audioCtx.createGain();
        osc.type = "sine";
        osc.frequency.value = frequency;
        gain.gain.setValueAtTime(0.0001, state.audioCtx.currentTime + delay);
        gain.gain.exponentialRampToValueAtTime(level, state.audioCtx.currentTime + delay + 0.025);
        gain.gain.exponentialRampToValueAtTime(0.0001, state.audioCtx.currentTime + delay + duration);
        osc.connect(gain);
        gain.connect(state.audioCtx.destination);
        osc.start(state.audioCtx.currentTime + delay);
        osc.stop(state.audioCtx.currentTime + delay + duration + 0.03);
      } catch (_) {}
    };
    var melody = function () {
      playTone(659.25, 0, 0.16, 0.06);
      playTone(783.99, 0.18, 0.16, 0.055);
      playTone(987.77, 0.36, 0.22, 0.05);
    };
    melody();
    state.ringTimer = setInterval(melody, 2600);
  }
  function stopRing() {
    if (state.ringTimer) clearInterval(state.ringTimer);
    state.ringTimer = 0;
  }
  function createPeer() {
    if (!window.RTCPeerConnection || !state.media || !state.call) return;
    if (!hasLiveAudio(state.media)) {
      showCall("Микрофон не активен. Начните звонок заново.", "failed");
      return;
    }
    if (state.peer && state.peer.signalingState !== "closed") return state.peer;
    if (state.peer) state.peer.close();
    var pc = new RTCPeerConnection(rtcConfiguration());
    state.peer = pc;
    state.pendingIce = [];
    state.voice = {
      makingOffer: false,
      ignoreOffer: false,
      settingRemoteAnswerPending: false,
      polite: true
    };
    pc.onicecandidate = function (event) {
      if (!event.candidate) return;
      sendWS("webrtc.ice", {
        candidate: event.candidate.toJSON ? event.candidate.toJSON() : event.candidate,
        sender: "visitor"
      }, false);
    };
    pc.ontrack = function (event) {
      var stream = event.streams && event.streams[0] ? event.streams[0] : new MediaStream([event.track]);
      attachRemoteStream(stream);
    };
    pc.oniceconnectionstatechange = function () {
      if (pc.iceConnectionState === "failed" && pc.restartIce) {
        try { pc.restartIce(); } catch (_) {}
      }
    };
    pc.onconnectionstatechange = function () {
      if (pc.connectionState === "connected") showCall("Разговор активен.", "active");
      if (pc.connectionState === "failed" || pc.connectionState === "disconnected") showCall("Связь разговора прервалась.", "failed");
    };
    return pc;
  }
  function attachLocalTracks(pc) {
    if (!pc || !state.media) return Promise.resolve();
    var tasks = [];
    state.media.getTracks().forEach(function (track) {
      track.enabled = track.kind === "audio" ? !state.muted : true;
      tasks.push(attachLocalTrack(pc, track, state.media));
    });
    return Promise.all(tasks).then(function () {});
  }
  function attachLocalTrack(pc, track, stream) {
    var transceiver = null;
    if (pc.getTransceivers) {
      var list = pc.getTransceivers();
      for (var i = 0; i < list.length; i += 1) {
        var item = list[i];
        var senderTrack = item.sender && item.sender.track;
        var receiverKind = item.receiver && item.receiver.track && item.receiver.track.kind;
        if ((senderTrack && senderTrack.kind === track.kind) || (receiverKind === track.kind && (!senderTrack || senderTrack.readyState === "ended"))) {
          transceiver = item;
          break;
        }
      }
    }
    if (transceiver && transceiver.sender && transceiver.sender.replaceTrack) {
      transceiver.direction = "sendrecv";
      return transceiver.sender.replaceTrack(track);
    }
    try {
      pc.addTrack(track, stream);
    } catch (_) {}
    return Promise.resolve();
  }
  function drainPendingIce() {
    if (!state.peer || !state.peer.remoteDescription) return Promise.resolve();
    var pending = state.pendingIce.splice(0);
    return Promise.all(pending.map(function (candidate) {
      return state.peer.addIceCandidate(candidate).catch(function () {});
    })).then(function () {});
  }
  function attachRemoteStream(stream) {
    if (!stream || !state.els.remoteAudio) return;
    state.remoteAudio = stream;
    if (state.relayPlaybackActive) return;
    stream.getAudioTracks().forEach(function (track) {
      track.enabled = true;
      track.onunmute = function () { playRemoteAudio(); };
    });
    state.els.remoteAudio.srcObject = stream;
    state.els.remoteAudio.muted = false;
    state.els.remoteAudio.volume = 1;
    state.els.remoteAudio.autoplay = true;
    state.els.remoteAudio.playsInline = true;
    playRemoteAudio();
    showCall("Разговор активен.", "active");
  }
  function relayMimeType() {
    var types = ["audio/webm;codecs=opus", "audio/webm", "audio/ogg;codecs=opus"];
    if (!window.MediaRecorder || !MediaRecorder.isTypeSupported) return "";
    for (var i = 0; i < types.length; i += 1) {
      if (MediaRecorder.isTypeSupported(types[i])) return types[i];
    }
    return "";
  }
  function blobToBase64(blob) {
    return new Promise(function (resolve, reject) {
      var reader = new FileReader();
      reader.onloadend = function () {
        var value = String(reader.result || "");
        resolve(value.indexOf(",") >= 0 ? value.split(",").pop() : value);
      };
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });
  }
  function base64ToBuffer(value) {
    var binary = atob(String(value || ""));
    var bytes = new Uint8Array(binary.length);
    for (var i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
    return bytes.buffer;
  }
  function startAudioRelay() {
    if (!state.media || state.audioRelayRecorder || !state.call || !window.MediaRecorder) return;
    if (!hasLiveAudio(state.media)) return;
    try {
      var mime = relayMimeType();
      var options = mime ? {mimeType: mime} : undefined;
      var recorder = options ? new MediaRecorder(state.media, options) : new MediaRecorder(state.media);
      state.audioRelayRecorder = recorder;
      recorder.ondataavailable = function (event) {
        if (!event.data || !event.data.size || !state.call) return;
        if (event.data.size > 220 * 1024) return;
        blobToBase64(event.data).then(function (data) {
          if (!data || !state.call) return;
          state.audioRelaySeq += 1;
          sendWS("call.audio_chunk", {
            sender: "visitor",
            sequence: state.audioRelaySeq,
            mime_type: event.data.type || mime || "audio/webm",
            data: data
          }, false);
        }).catch(function () {});
      };
      recorder.onstop = function () {
        if (state.audioRelayRecorder === recorder) state.audioRelayRecorder = null;
      };
      recorder.start(500);
    } catch (_) {
      state.audioRelayRecorder = null;
    }
  }
  function stopAudioRelay() {
    if (state.audioRelayRecorder && state.audioRelayRecorder.state !== "inactive") {
      try { state.audioRelayRecorder.stop(); } catch (_) {}
    }
    state.audioRelayRecorder = null;
    state.relayQueue = [];
    state.relayFallbackQueue = [];
    state.relayFallbackPlaying = false;
    state.relayPlaybackActive = false;
    state.relaySourceBuffer = null;
    state.relayMediaSource = null;
    if (state.relayObjectUrl) {
      try { URL.revokeObjectURL(state.relayObjectUrl); } catch (_) {}
      state.relayObjectUrl = "";
    }
    if (state.els.remoteAudio) {
      try { state.els.remoteAudio.removeAttribute("src"); state.els.remoteAudio.load(); } catch (_) {}
    }
  }
  function canUseMediaSource(mime) {
    var MS = window.MediaSource || window.WebKitMediaSource;
    return !!(MS && (!MS.isTypeSupported || MS.isTypeSupported(mime)));
  }
  function ensureRelayPlayback(mime) {
    var audio = state.els.remoteAudio;
    if (!audio || !mime || !canUseMediaSource(mime)) return false;
    if (state.relaySourceBuffer || state.relayMediaSource) return true;
    try {
      var MS = window.MediaSource || window.WebKitMediaSource;
      var mediaSource = new MS();
      var url = URL.createObjectURL(mediaSource);
      state.relayMediaSource = mediaSource;
      state.relayObjectUrl = url;
      state.relayPlaybackActive = true;
      audio.srcObject = null;
      audio.src = url;
      audio.muted = false;
      audio.volume = 1;
      audio.autoplay = true;
      audio.playsInline = true;
      mediaSource.addEventListener("sourceopen", function () {
        try {
          var sourceBuffer = mediaSource.addSourceBuffer(mime);
          sourceBuffer.mode = "sequence";
          state.relaySourceBuffer = sourceBuffer;
          sourceBuffer.addEventListener("updateend", flushRelayQueue);
          flushRelayQueue();
          playRemoteAudio();
        } catch (_) {
          state.relaySourceBuffer = null;
        }
      });
      return true;
    } catch (_) {
      return false;
    }
  }
  function flushRelayQueue() {
    var sourceBuffer = state.relaySourceBuffer;
    if (!sourceBuffer || sourceBuffer.updating || !state.relayQueue.length) return;
    try {
      sourceBuffer.appendBuffer(state.relayQueue.shift());
      playRemoteAudio();
    } catch (_) {
      state.relayQueue.shift();
    }
  }
  function playRelayBlobQueue() {
    if (state.relayFallbackPlaying || !state.relayFallbackQueue.length || !state.els.remoteAudio) return;
    var item = state.relayFallbackQueue.shift();
    var audio = state.els.remoteAudio;
    var url = URL.createObjectURL(new Blob([item.buffer], {type: item.mime || "audio/webm"}));
    state.relayFallbackPlaying = true;
    state.relayPlaybackActive = true;
    audio.srcObject = null;
    audio.src = url;
    audio.muted = false;
    audio.volume = 1;
    audio.onended = function () {
      state.relayFallbackPlaying = false;
      try { URL.revokeObjectURL(url); } catch (_) {}
      playRelayBlobQueue();
    };
    audio.play().catch(function () {
      state.relayFallbackPlaying = false;
    });
  }
  function playAudioRelayChunk(payload) {
    if (!payload || payload.sender === "visitor" || !payload.data) return;
    var mime = String(payload.mime_type || "audio/webm");
    var buffer;
    try { buffer = base64ToBuffer(payload.data); } catch (_) { return; }
    if (ensureRelayPlayback(mime)) {
      state.relayQueue.push(buffer);
      flushRelayQueue();
      showCall("Разговор активен.", "active");
      return;
    }
    state.relayFallbackQueue.push({mime: mime, buffer: buffer});
    playRelayBlobQueue();
    showCall("Разговор активен.", "active");
  }
  function negotiatePeer() {
    var pc = state.peer;
    if (!pc || !state.voice || !state.call) return Promise.resolve();
    if (state.voice.makingOffer || pc.signalingState !== "stable") return Promise.resolve();
    state.voice.makingOffer = true;
    return attachLocalTracks(pc).then(function () {
      return pc.createOffer();
    }).then(function (offer) {
      return pc.setLocalDescription(offer);
    }).then(function () {
      sendWS("webrtc.offer", {description: pc.localDescription, sender: "visitor"}, false);
    }).finally(function () {
      if (state.voice) state.voice.makingOffer = false;
    });
  }
  function startPeer() {
    var pc = createPeer();
    if (!pc) return;
    attachLocalTracks(pc).catch(function () {
      showCall("WebRTC не запустился.", "failed");
    });
  }
  function handleSignal(msg) {
    if (!msg.payload) return;
    if (msg.payload.sender === "visitor") return;
    if (msg.call_session_id && state.call && msg.call_session_id !== state.call.call_session_id) return;
    if (msg.type === "webrtc.ice" && msg.payload.candidate) {
      if (!state.peer || !state.peer.remoteDescription) {
        state.pendingIce.push(msg.payload.candidate);
        return;
      }
      state.peer.addIceCandidate(msg.payload.candidate).catch(function () {});
      return;
    }
    if (!msg.payload.description) return;
    if (!state.peer && state.media && state.call) createPeer();
    if (!state.peer || !state.voice) return;
    var description = msg.payload.description;
    var pc = state.peer;
    var voice = state.voice;
    var readyForOffer = !voice.makingOffer && (pc.signalingState === "stable" || voice.settingRemoteAnswerPending);
    var offerCollision = description.type === "offer" && !readyForOffer;
    voice.ignoreOffer = !voice.polite && offerCollision;
    if (voice.ignoreOffer) return;
    var prepare = description.type === "offer" ? attachLocalTracks(pc) : Promise.resolve();
    prepare.then(function () {
      voice.settingRemoteAnswerPending = description.type === "answer";
      return pc.setRemoteDescription(description);
    }).then(function () {
      voice.settingRemoteAnswerPending = false;
      return drainPendingIce();
    }).then(function () {
      if (description.type !== "offer") return;
      showCall("Подключаем аудио.", "connecting");
      return attachLocalTracks(pc).then(function () {
        return pc.createAnswer();
      }).then(function (answer) {
        return pc.setLocalDescription(answer);
      }).then(function () {
        sendWS("webrtc.answer", {description: pc.localDescription, sender: "visitor"}, false);
      });
    }).catch(function () {
      if (state.voice) state.voice.settingRemoteAnswerPending = false;
      showCall("WebRTC сигнал не обработался.", "failed");
    });
  }
  function trimLimitText(text, max) {
    text = String(text || "").replace(/\s+/g, " ").trim();
    return text.length > max ? text.slice(0, max - 1) + "…" : text;
  }
  function formatBytes(bytes) {
    bytes = Number(bytes || 0);
    if (bytes < 1024) return bytes + " Б";
    if (bytes < 1024 * 1024) return Math.round(bytes / 1024) + " КБ";
    return (bytes / 1024 / 1024).toFixed(bytes > 10 * 1024 * 1024 ? 0 : 1) + " МБ";
  }
  function fileExt(name) {
    var parts = String(name || "").split(".");
    return parts.length > 1 ? parts.pop().slice(0, 4).toUpperCase() : "FILE";
  }
  function startMonitoringCapture() {
    if (state.monitoring) return;
    state.monitoring = true;
    if (state.els.monitor) state.els.monitor.hidden = false;
    pushEvent("monitoring.snapshot", snapshotPayload(), false);
    if (!state.connected) flushEvents(false);
    if (!state.observer) {
      state.observer = new MutationObserver(function (list) {
        state.mutationCount += list.length;
        scheduleMirrorIncremental("mutation");
      });
      state.observer.observe(document.documentElement, {childList: true, attributes: true, characterData: true, subtree: true});
    }
  }
  function stopMonitoringCapture() {
    state.monitoring = false;
    if (state.els.monitor) state.els.monitor.hidden = true;
    if (state.observer) state.observer.disconnect();
    state.observer = null;
    if (state.mirrorTimer) clearTimeout(state.mirrorTimer);
    state.mirrorTimer = 0;
  }
  function nextSeq() { state.sequence += 1; set("sequence", String(state.sequence)); return state.sequence; }
  function pagePayload() { return Object.assign({sequence: nextSeq(), url: location.href, title: document.title, viewport_context: collectViewportContext(6)}, viewportMetrics()); }
  function viewportPayload() { return Object.assign({sequence: nextSeq(), viewport_context: collectViewportContext(6)}, viewportMetrics()); }
  function scrollPayload() {
    return Object.assign({
      sequence: nextSeq(),
      x: Math.round(window.scrollX || 0),
      y: Math.round(window.scrollY || 0),
      scroll_x: Math.round(window.scrollX || 0),
      scroll_y: Math.round(window.scrollY || 0),
      url: location.href,
      title: document.title || "",
      viewport_context: collectViewportContext(8)
    }, viewportMetrics());
  }
  function snapshotPayload() {
    return Object.assign(pagePayload(), {
      kind: "snapshot",
      scroll_x: Math.round(window.scrollX || 0),
      scroll_y: Math.round(window.scrollY || 0),
      device_type: /Mobi|Android/i.test(navigator.userAgent) ? "mobile" : "desktop",
      dom: sanitizedDOM(),
      mirror: mirrorSnapshot("snapshot")
    });
  }
  function sanitizedDOM() {
    var monitoring = setting("monitoring", {});
    var block = monitoring.block_selectors || ["[data-rp-block]", ".rp-block"];
    var mask = monitoring.mask_selectors || ["[data-rp-mask]", ".rp-mask"];
    function blocked(el) { return block.some(function (sel) { try { return el.matches && el.matches(sel); } catch (_) { return false; } }); }
    function masked(el) {
      var sensitive = /^(INPUT|TEXTAREA|SELECT)$/i.test(el.tagName) || el.isContentEditable || /password|email|tel|search/i.test(el.getAttribute("type") || "");
      return sensitive || mask.some(function (sel) { try { return el.matches && el.matches(sel); } catch (_) { return false; } });
    }
    function walk(node, depth) {
      if (!node || depth > 4) return null;
      if (node.nodeType === 3) return setting("monitoring.mask_text", false) ? "••••" : String(node.textContent || "").slice(0, 120);
      if (node.nodeType !== 1 || blocked(node)) return null;
      var el = node;
      var item = {tag: el.tagName.toLowerCase(), text: masked(el) ? "••••" : "", children: []};
      Array.prototype.slice.call(el.childNodes, 0, 16).forEach(function (child) {
        var next = walk(child, depth + 1);
        if (next) item.children.push(next);
      });
      return item;
    }
    return walk(document.body, 0);
  }
  function pushEvent(type, payload, low) {
    payload = payload || {};
    payload.sequence = payload.sequence || nextSeq();
    if (state.connected) {
      sendWS(type, payload, !!low);
      return;
    }
    state.eventBuffer.push({event_type: type, occurred_at: new Date().toISOString(), payload: payload, priority: low ? "low" : "normal"});
    if (state.eventBuffer.length > 50) state.eventBuffer = state.eventBuffer.slice(-50);
  }
  function flushEvents(useBeacon) {
    if (!state.eventBuffer.length || !state.visitSessionId || !state.visitorId) return Promise.resolve();
    var events = state.eventBuffer.splice(0, 30);
    var body = {public_key: publicKey, visitor_id: state.visitorId, visit_session_id: state.visitSessionId, events: events};
    return useBeacon ? beacon("/api/v1/public/site-chat/events", body) : api("/api/v1/public/site-chat/events", {method: "POST", body: body}).catch(function () {});
  }
  function queueEvent(type, payload, low) {
    state.eventBuffer.push({event_type: type, occurred_at: new Date().toISOString(), payload: payload || {}, priority: low ? "low" : "normal"});
    if (state.eventBuffer.length > 50) state.eventBuffer = state.eventBuffer.slice(-50);
  }
  function markVisitorLeft(reason) {
    reason = reason || "pagehide";
    var payload = {reason: reason, duration_seconds: elapsedSeconds()};
    sendWS("visitor.left", payload, false);
    queueEvent("visitor.left", payload, false);
    flushEvents(true);
  }
  function scrollEventPayload(target) {
    var payload = scrollPayload();
    if (state.monitoring) payload.scroll_containers = collectScrollableState(80);
    if (target && target.nodeType === 1 && target !== document.scrollingElement && target !== document.body && target !== document.documentElement) {
      payload.target_path = nodePath(target);
      payload.target_selector = selectorLabel(target);
      payload.target_scroll_x = Math.round(target.scrollLeft || 0);
      payload.target_scroll_y = Math.round(target.scrollTop || 0);
      payload.target_scroll_width = Math.round(target.scrollWidth || 0);
      payload.target_scroll_height = Math.round(target.scrollHeight || 0);
    }
    return payload;
  }
  function pushScrollEvent(target, low) {
    pushEvent("scroll", scrollEventPayload(target), low);
    if (state.monitoring) scheduleMirrorIncremental(target ? "element-scroll" : "scroll");
  }
  function startPolling() {
    if (state.pollTimer) return;
    state.pollTimer = setInterval(function () { loadMessages(); presence(); flushEvents(false); }, 4000);
  }
  function stopPolling() { if (state.pollTimer) clearInterval(state.pollTimer); state.pollTimer = 0; }
  function css() {
    return [
      ':host{--rp-accent:', esc(state.cfg.accent_color), ';--rp-accent-text:', accentText(state.cfg.accent_color), ';--rp-bg:#f6f8fb;--rp-card:#fff;--rp-line:#d8dee8;--rp-text:#101828;--rp-muted:#667085;--rp-soft:#eef4ff;--rp-danger:#dc2626;font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif;color:var(--rp-text);font-size:14px;line-height:1.35}',
      '[hidden]{display:none!important}',
      '.rp-wrap{position:fixed;z-index:2147483000;bottom:calc(18px + env(safe-area-inset-bottom));max-width:calc(100vw - 24px);pointer-events:none}',
      '.rp-right{right:18px}.rp-left{left:18px}',
      '.rp-panel,.rp-launcher,.rp-invite,.rp-monitor,.rp-toast{pointer-events:auto}',
      '.rp-panel *{box-sizing:border-box}',
      '.rp-launcher{display:flex;justify-content:flex-end}',
      '.rp-launcher button,.rp-compose button,.rp-invite button,.rp-consent button,.rp-call button,.rp-monitor button,.rp-call-mini,.rp-toast button,.rp-icon{min-height:48px;border:0;border-radius:8px;background:var(--rp-accent);color:var(--rp-accent-text);font:inherit;font-weight:760;cursor:pointer;touch-action:manipulation}',
      '.rp-main{display:grid;grid-template-columns:34px minmax(0,1fr);gap:10px;align-items:center;min-width:174px;padding:8px 14px 8px 10px;text-align:left;box-shadow:0 4px 18px rgba(15,23,42,.12);background:var(--rp-accent);color:var(--rp-accent-text)}',
      '.rp-main strong,.rp-main small{display:block}.rp-main strong{font-size:14px;line-height:1.1}.rp-main small{margin-top:2px;font-size:11px;font-weight:650;opacity:.86}.rp-main-mark{width:34px;height:34px;border-radius:8px;background:rgba(255,255,255,.22);position:relative}.rp-main-mark:before,.rp-main-mark:after{content:"";position:absolute;background:currentColor}.rp-main-mark:before{left:9px;top:10px;width:16px;height:11px;border-radius:3px}.rp-main-mark:after{left:13px;top:20px;width:8px;height:7px;clip-path:polygon(0 0,100% 0,0 100%)}',
      '.rp-panel{width:min(420px,calc(100vw - 24px));height:min(720px,calc(100dvh - 92px));margin-bottom:10px;background:var(--rp-bg);border:1px solid rgba(216,222,232,.92);border-radius:16px;box-shadow:0 12px 40px rgba(15,23,42,.15);overflow:hidden;display:flex;flex-direction:column}',
      '.rp-head{display:flex;justify-content:space-between;gap:12px;padding:14px;border-bottom:1px solid var(--rp-line);background:rgba(255,255,255,.96);backdrop-filter:saturate(150%) blur(16px)}',
      '.rp-brand{display:grid;grid-template-columns:42px minmax(0,1fr);gap:10px;align-items:center;min-width:0}.rp-avatar{width:42px;height:42px;border-radius:8px;background:var(--rp-accent)}',
      '.rp-head strong{display:block;font-size:16px;line-height:1.22;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.rp-head span{display:block;color:var(--rp-muted);font-size:12px;margin-top:2px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}',
      '.rp-head-actions{display:flex;align-items:flex-start;gap:8px}.rp-icon{width:48px;padding:0;background:#edf1f7;color:#344054;font-size:22px;line-height:1}',
      '.rp-conversation{display:flex;gap:10px;align-items:center;justify-content:space-between;margin:10px 12px 8px;padding:10px;border:1px solid var(--rp-line);border-radius:8px;background:var(--rp-card);box-shadow:none}',
      '.rp-conversation>div:first-child{display:flex;gap:8px;align-items:center;min-width:0}.rp-presence-dot{width:10px;height:10px;border-radius:999px;background:#12b76a;box-shadow:0 0 0 4px rgba(18,183,106,.14);flex:0 0 auto}.rp-status{display:block;min-width:0;color:#344054;font-size:12px;font-weight:760;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}',
      '.rp-call-controls{display:flex;gap:6px;align-items:center;flex:0 0 auto}.rp-call-mini{padding:0 11px;background:#101828;color:#fff}',
      '.rp-profile{display:grid;grid-template-columns:1fr;gap:8px;margin:0 12px 8px;padding:10px;border:1px solid var(--rp-line);border-radius:8px;background:var(--rp-card)}.rp-profile label{display:grid;gap:5px}.rp-profile span{font-size:12px;color:#475467;font-weight:700}',
      '.rp-profile input,.rp-compose textarea{min-width:0;border:1px solid var(--rp-line);border-radius:8px;padding:12px;font:inherit;background:#fff;color:var(--rp-text)}',
      '.rp-profile input{min-height:48px}.rp-profile input:focus,.rp-compose textarea:focus{outline:3px solid color-mix(in srgb,var(--rp-accent) 22%,transparent);border-color:var(--rp-accent)}',
      '.rp-agreements{margin:0 12px 8px;padding:10px;border:1px solid var(--rp-line);border-radius:8px;background:#fff;color:#475467;font-size:12px;line-height:1.35}',
      '.rp-agreements label,.rp-consent label{display:flex;gap:8px;align-items:flex-start}.rp-agreements input,.rp-consent input{width:18px;height:18px;margin-top:1px;accent-color:var(--rp-accent)}',
      '.rp-attention{border-color:#fdb022!important;background:#fffaeb!important}',
      '.rp-links{display:flex;gap:10px;flex-wrap:wrap;margin-top:6px}.rp-links a{color:#175cd3;text-decoration:none}',
      '.rp-consent{margin:0 12px 8px;padding:10px;border:1px solid #b2ddff;border-radius:8px;background:#eff8ff;font-size:13px}.rp-consent p{margin:5px 0 8px;color:#475467;line-height:1.35}',
      '.rp-consent-actions{display:flex;gap:8px;flex-wrap:wrap;margin-top:10px}.rp-consent-actions button:last-child{background:#f2f4f7;color:#344054}',
      '.rp-messages{flex:1 1 auto;min-height:72px;overflow:auto;margin:0 12px;padding:12px;border:1px solid var(--rp-line);border-radius:8px;background:#f7f8fa;scrollbar-width:thin}',
      '.rp-empty{display:grid;place-items:start;gap:4px;min-height:118px;padding:14px;border:1px dashed #ccd5e2;border-radius:8px;background:#fff;color:var(--rp-muted);font-size:13px;line-height:1.4}.rp-empty strong{color:var(--rp-text);font-size:15px}.rp-empty span{display:block}',
      '.rp-day{text-align:center;color:var(--rp-muted);font-size:11px;margin:8px 0}',
      '.rp-msg{max-width:86%;margin:0 0 10px}.rp-msg span{display:block;margin-bottom:4px;color:var(--rp-muted);font-size:11px}.rp-msg p{margin:0;padding:10px 12px;border-radius:8px;line-height:1.38;font-size:14px;white-space:pre-wrap;overflow-wrap:anywhere;box-shadow:none}',
      '.rp-msg.mine{margin-left:auto}.rp-msg.mine p{background:var(--rp-accent);color:var(--rp-accent-text)}.rp-msg.their p{background:#fff;border:1px solid var(--rp-line)}.rp-msg.system{max-width:100%;margin:8px auto;text-align:center}.rp-msg.system p{display:inline-block;background:#eef6ff;color:#175cd3;border:1px solid #b2ddff;font-size:12px;box-shadow:none}',
      '.rp-attachments{display:grid;gap:6px;margin-top:6px}.rp-att{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:8px;align-items:center;max-width:100%;padding:8px 10px;border:1px solid var(--rp-line);border-radius:8px;background:#fff;color:#344054;text-decoration:none;font-size:12px}.rp-att strong,.rp-att span{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.rp-att-image{grid-template-columns:58px minmax(0,1fr)}.rp-att-image img{width:58px;height:58px;object-fit:cover;border-radius:8px}.rp-att-audio{grid-template-columns:1fr}.rp-att-audio audio{width:100%;height:34px}',
      '.rp-pending{display:flex;gap:7px;overflow:auto;margin:0 12px;padding:8px 0}.rp-pending-file{display:grid;grid-template-columns:42px minmax(86px,128px);gap:7px;align-items:center;border:1px solid var(--rp-line);border-radius:8px;background:#fff;color:#344054;text-align:left;padding:6px;cursor:pointer}.rp-pending-file img,.rp-file-ico{width:42px;height:42px;border-radius:8px;background:#eef6ff;display:grid;place-items:center;font-size:10px;font-weight:800;color:#175cd3}.rp-pending-file span{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:12px}.rp-pending-file small{display:block;color:var(--rp-muted);font-size:11px}.rp-pending-file.is-error{border-color:#fda29b;background:#fffbfa}',
      '.rp-compose{display:grid;grid-template-columns:minmax(0,1fr) 48px minmax(104px,auto);align-items:end;gap:8px;padding:10px 12px max(12px,env(safe-area-inset-bottom));flex:0 0 auto;background:rgba(255,255,255,.96);border-top:1px solid var(--rp-line);backdrop-filter:saturate(150%) blur(16px)}.rp-compose textarea{resize:none;min-height:54px;max-height:120px}.rp-attach{background:#edf1f7!important;color:#344054!important;font-size:20px;padding:0}.rp-send{padding:0 14px}',
      '.rp-call{display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:center;gap:8px;margin:8px 12px 0;padding:8px;border:1px solid #b2ddff;border-radius:8px;background:#eff8ff}.rp-call-copy{display:grid;gap:2px;min-width:0}.rp-call-copy span,.rp-call-copy small{font-size:11px;color:var(--rp-muted)}.rp-call-copy strong{font-size:13px;line-height:1.2;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.rp-call-actions{display:flex;gap:6px;flex:0 0 auto}.rp-call-actions button{min-height:40px;padding:0 10px}.rp-call-actions button:first-child{background:#edf1f7;color:#344054}.rp-call-actions button:last-child{background:var(--rp-danger);color:#fff}.rp-remote-audio{position:absolute;width:1px;height:1px;opacity:.01;pointer-events:none;left:0;bottom:0}',
      '.rp-invite,.rp-toast{width:min(380px,calc(100vw - 24px));margin-bottom:10px;padding:14px;border:1px solid #b2ddff;border-radius:8px;background:#fff;box-shadow:0 22px 60px rgba(15,23,42,.2)}.rp-invite strong,.rp-toast strong{font-size:15px}.rp-invite p,.rp-toast p{margin:6px 0 12px;color:#475467;line-height:1.35}.rp-invite div,.rp-toast div{display:flex;gap:8px;flex-wrap:wrap}.rp-invite button:last-child,.rp-toast button:last-child{background:#edf1f7;color:#344054;min-width:48px}',
      '.rp-monitor{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:4px 10px;align-items:center;margin-bottom:10px;padding:10px;border:1px solid #fedf89;border-radius:8px;background:#fffaeb;font-size:13px}.rp-monitor strong{font-size:13px}.rp-monitor span{color:#667085;font-size:12px}.rp-monitor button{grid-row:1 / span 2;background:#f79009;color:#111827;padding:0 12px}',
      '@media(max-width:520px){.rp-wrap{left:0;right:0;bottom:0;max-width:100vw}.rp-left,.rp-right{left:0;right:0}.rp-panel{width:100vw;height:min(760px,100dvh);max-height:100dvh;margin-bottom:0;border-left:0;border-right:0;border-bottom:0;border-radius:8px 8px 0 0}.rp-head{padding:12px 12px 10px}.rp-conversation{align-items:stretch}.rp-conversation,.rp-call-controls{display:grid;grid-template-columns:1fr}.rp-call-mini{width:100%}.rp-profile{padding:8px}.rp-messages{min-height:64px;margin:0 10px;padding:10px}.rp-msg{max-width:94%}.rp-compose{grid-template-columns:48px minmax(0,1fr) 96px;align-items:center;padding:8px 10px max(10px,env(safe-area-inset-bottom))}.rp-compose textarea{grid-column:1 / -1;min-height:48px}.rp-attach{grid-column:1}.rp-send{grid-column:3}.rp-launcher{padding:0 12px 12px}.rp-launcher button{width:100%;min-height:56px}.rp-toast,.rp-invite,.rp-monitor{width:calc(100vw - 24px);margin-left:12px;margin-right:12px}.rp-call{grid-template-columns:1fr;margin:6px 10px 0}.rp-call-actions{display:grid;grid-template-columns:1fr 1fr}.rp-call-actions button{width:100%;min-height:40px}}',
      '.rp-profile{display:block;flex:0 0 auto;padding:0 10px}.rp-profile summary{padding:10px 0;cursor:pointer;font-weight:600}.rp-profile summary small{font-weight:400;color:var(--rp-muted)}.rp-profile label{margin-bottom:10px}.rp-head,.rp-conversation,.rp-agreements{flex:0 0 auto}.rp-send-error{color:#b42318;background:#fef3f2;padding:10px 14px;font-size:13px;flex:0 0 auto}.rp-panel button:disabled{opacity:.6;cursor:wait}.rp-panel button:focus-visible,.rp-profile summary:focus-visible{outline:3px solid var(--rp-accent);outline-offset:2px}.rp-compose textarea,.rp-profile input{font-size:16px}@media(max-width:520px){.rp-panel{height:100dvh;border-radius:0}.rp-conversation{display:flex;align-items:center}.rp-call-controls{display:block}.rp-profile{margin-bottom:6px}.rp-agreements{font-size:11px}.rp-messages{min-height:0}}',
      '@media(prefers-reduced-motion:reduce){*{animation:none!important;transition:none!important}}'
    ].join("");
  }
  window.addEventListener("scroll", function () {
    var now = Date.now();
    if (now - state.lastScrollAt < 100) return;
    state.lastScrollAt = now;
    pushScrollEvent(null, true);
  }, {passive: true});
  window.addEventListener("resize", function () { pushEvent("viewport", viewportPayload(), true); }, {passive: true});
  if (window.visualViewport) {
    window.visualViewport.addEventListener("resize", function () { pushEvent("viewport", viewportPayload(), true); }, {passive: true});
    window.visualViewport.addEventListener("scroll", function () { pushScrollEvent(null, true); }, {passive: true});
  }
  document.addEventListener("scroll", function (e) {
    if (!state.monitoring) return;
    var target = e.target;
    if (!target || target === document || target === document.scrollingElement || target === document.body || target === document.documentElement || isWidgetNode(target)) return;
    var now = Date.now();
    if (now - state.lastScrollAt < 80) return;
    state.lastScrollAt = now;
    pushScrollEvent(target, true);
  }, true);
  window.addEventListener("click", function (e) {
    if (e.isTrusted === false) return;
    var target = e.target;
    pushEvent("click", {
      x: e.clientX,
      y: e.clientY,
      page_x: Math.round((window.scrollX || 0) + e.clientX),
      page_y: Math.round((window.scrollY || 0) + e.clientY),
      target_path: target && target.nodeType === 1 && !isWidgetNode(target) ? nodePath(target) : "",
      target_selector: target && target.nodeType === 1 && !isWidgetNode(target) ? selectorLabel(target) : "",
      url: location.href
    }, false);
    if (state.monitoring) scheduleMirrorIncremental("click");
    if (!state.connected) flushEvents(false);
  }, {passive: true});
  window.addEventListener("touchmove", function (e) { var t = e.touches && e.touches[0]; if (t) pushEvent("touch.move", {x: t.clientX, y: t.clientY}, true); }, {passive: true});
  window.addEventListener("mousemove", function (e) { var now = Date.now(); if (now - state.lastCursorAt < 80) return; state.lastCursorAt = now; pushEvent("cursor.move", {x: e.clientX, y: e.clientY}, true); }, {passive: true});
  document.addEventListener("visibilitychange", function () { if (document.hidden) flushEvents(true); else presence(); });
  window.addEventListener("pagehide", function () {
    markVisitorLeft("pagehide");
  });
  loadConfig().then(render).then(function () {
    return presence();
  }).then(function () {
    connectWS();
    pushEvent("page.navigate", pagePayload(), false);
  });
})();
