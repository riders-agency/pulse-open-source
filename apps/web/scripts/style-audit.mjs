import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const srcDir = path.join(root, "src");
const rel = (file) => path.relative(root, file).replaceAll(path.sep, "/");
const failures = [];

function walk(dir, predicate = () => true) {
  const out = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === "node_modules" || entry.name === ".next" || entry.name === "generated" || entry.name === "__generated__") continue;
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) out.push(...walk(full, predicate));
    else if (predicate(full)) out.push(full);
  }
  return out;
}

function lineCount(file) {
  return fs.readFileSync(file, "utf8").split(/\r?\n/).length - 1;
}

function fail(message) {
  failures.push(message);
}

const auditedFiles = walk(srcDir, (file) => /\.(css|tsx|ts)$/.test(file));
const largest = auditedFiles
  .map((file) => ({ file, lines: lineCount(file) }))
  .sort((a, b) => b.lines - a.lines)
  .slice(0, 30);

console.log("Style audit: top 30 css/tsx/ts files");
for (const item of largest) {
  console.log(`${String(item.lines).padStart(5)} ${rel(item.file)}`);
}

const appDir = path.join(srcDir, "app");
for (const file of walk(appDir, (candidate) => /\.(tsx|ts)$/.test(candidate))) {
  const relative = rel(file);
  const lines = lineCount(file);
  const text = fs.readFileSync(file, "utf8");
  if (relative.endsWith("/page.tsx") && lines > 80) {
    fail(`app_page_size_budget ${relative} lines=${lines} limit=80`);
  }
  if (/^\s*["']use client["'];/m.test(text)) {
    fail(`app_client_boundary ${relative}`);
  }
  if (/\buse(?:State|Effect|Memo|Callback|Ref|Reducer|Transition|Optimistic)\s*\(/.test(text)) {
    fail(`app_react_hook ${relative}`);
  }
  if (/\busePulseStore\s*\(/.test(text)) {
    fail(`app_store_access ${relative}`);
  }
}

for (const file of walk(srcDir, (candidate) => /\.(tsx|ts)$/.test(candidate))) {
  const relative = rel(file);
  const lines = lineCount(file);
  const limit = 180;
  if (lines > limit) fail(`source_size_budget ${relative} lines=${lines} limit=${limit}`);
}

for (const file of walk(srcDir, (candidate) => candidate.endsWith(".css"))) {
  const relative = rel(file);
  const lines = lineCount(file);
  let limit = 180;
  if (relative === "src/styles/index.css") limit = 80;
  if (relative === "src/app/globals.css") limit = 20;
  if (relative.startsWith("src/styles/ui/")) limit = 180;
  if (relative.startsWith("src/styles/tokens/")) limit = 180;
  if (lines > limit) fail(`css_size_budget ${relative} lines=${lines} limit=${limit}`);
}

const primitiveTokens = path.join(srcDir, "styles/tokens/primitives.css");
if (fs.existsSync(primitiveTokens)) {
  const text = fs.readFileSync(primitiveTokens, "utf8");
  const primitiveColorDeclarations = [...text.matchAll(/--palette-[a-z-]+:\s*#[0-9A-Fa-f]{6}\b/g)];
  const allRawColors = [...text.matchAll(/#[0-9A-Fa-f]{3,8}\b|\b(?:rgb|rgba|hsl|hsla|oklch|lab|lch)\s*\(/gi)];
  if (primitiveColorDeclarations.length !== 5 || allRawColors.length !== 5) {
    fail(`primitive_palette_size src/styles/tokens/primitives.css colors=${allRawColors.length} palette_declarations=${primitiveColorDeclarations.length} expected=5`);
  }
}

const rawColorPattern = /#[0-9A-Fa-f]{3,8}\b|\b(?:rgb|rgba|hsl|hsla|oklch|lab|lch)\s*\(/gi;
const namedColorPattern = /(?<![-\w])(?:black|white)(?![-\w])/gi;
for (const file of auditedFiles) {
  const relative = rel(file);
  if (relative === "src/styles/tokens/primitives.css") continue;
  const text = fs.readFileSync(file, "utf8");
  for (const match of text.matchAll(rawColorPattern)) {
    const line = text.slice(0, match.index).split(/\r?\n/).length;
    fail(`raw_color_outside_primitives ${relative}:${line} value=${match[0]}`);
  }
  for (const match of text.matchAll(namedColorPattern)) {
    const line = text.slice(0, match.index).split(/\r?\n/).length;
    fail(`named_color_outside_primitives ${relative}:${line} value=${match[0]}`);
  }
}

const colorMixLiteralPattern = /color-mix\([^)]*(#[0-9A-Fa-f]{3,8}\b|\b(?:rgb|rgba|hsl|hsla|oklch|lab|lch)\s*\(|(?<![-\w])(?:white|black)(?![-\w]))/i;
for (const file of walk(path.join(srcDir, "styles"), (candidate) => candidate.endsWith(".css"))) {
  const relative = rel(file);
  if (relative === "src/styles/tokens/primitives.css") continue;
  fs.readFileSync(file, "utf8")
    .split(/\r?\n/)
    .forEach((line, index) => {
      if (colorMixLiteralPattern.test(line)) {
        fail(`color_mix_literal ${relative}:${index + 1}`);
      }
    });
}

const deprecatedColorTokens = /--(?:color-[\w-]+|action-blue|live-accent|chat-accent|surface-mint|surface-cool|surface-cyan|surface-sky|surface-bluegray|surface-ink(?:-deep)?|neutral-bluegray|status-info-sky(?:-text)?|status-skipped-strong)\b/;
for (const file of auditedFiles) {
  const relative = rel(file);
  const text = fs.readFileSync(file, "utf8");
  text.split(/\r?\n/).forEach((line, index) => {
    if (deprecatedColorTokens.test(line)) fail(`deprecated_color_token ${relative}:${index + 1}`);
  });
}

const importantAllowlist = [
  ["src/styles/features/workflow/runtime.css", "color: var(--foreground) !important"],
  ["src/styles/features/workflow/runtime.css", "font-size: 13px !important"],
  ["src/styles/features/workflow/runtime.css", "line-height: 1.45 !important"],
  ["src/styles/features/workflow/runtime.css", "font-size: 15px !important"],
  ["src/styles/features/responsive/mobile-responsive.css", "width: 100% !important"],
];
for (const file of walk(srcDir, (candidate) => /\.(css|tsx|ts)$/.test(candidate))) {
  const relative = rel(file);
  fs.readFileSync(file, "utf8")
    .split(/\r?\n/)
    .forEach((line, index) => {
      if (!line.includes("!important")) return;
      const allowed = importantAllowlist.some(([allowedFile, fragment]) => relative === allowedFile && line.includes(fragment));
      if (!allowed) fail(`important_not_allowed ${relative}:${index + 1}`);
    });
}

function selectorDepth(selector) {
  return selector
    .replace(/:not\([^)]*\)/g, "")
    .split(/\s+|>|\+|~/)
    .map((part) => part.trim())
    .filter((part) => part && part !== "*" && part !== ",").length;
}

for (const file of walk(path.join(srcDir, "styles"), (candidate) => candidate.endsWith(".css"))) {
  const relative = rel(file);
  const text = fs.readFileSync(file, "utf8").replace(/\/\*[\s\S]*?\*\//g, "");
  for (const match of text.matchAll(/([^{}]+)\{/g)) {
    const prelude = match[1].trim();
    if (!prelude || prelude.startsWith("@") || /^(from|to|\d+%)$/.test(prelude)) continue;
    const selectors = prelude.split(",").map((selector) => selector.trim()).filter(Boolean);
    for (const selector of selectors) {
      const depth = selectorDepth(selector);
      if (depth > 3) {
        const line = text.slice(0, match.index).split(/\r?\n/).length;
        fail(`selector_depth ${relative}:${line} depth=${depth} selector=${selector}`);
      }
    }
  }
}

const globals = path.join(srcDir, "app/globals.css");
if (fs.existsSync(globals)) {
  const text = fs.readFileSync(globals, "utf8");
  const featureSelectorPattern = /\.(work|chat|payment|live|ai-employee)-/;
  if (featureSelectorPattern.test(text)) fail("globals_contains_feature_selector src/app/globals.css");
}

if (failures.length > 0) {
  console.error("Style audit failed:");
  for (const failure of failures) console.error(`- ${failure}`);
  process.exit(1);
}

console.log("Style audit passed");
