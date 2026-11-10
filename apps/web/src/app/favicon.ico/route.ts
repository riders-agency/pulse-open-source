import fs from "node:fs";
import path from "node:path";

function paletteColor(name: string) {
  const tokenPath = path.join(process.cwd(), "src/styles/tokens/primitives.css");
  const tokens = fs.readFileSync(tokenPath, "utf8");
  const match = tokens.match(new RegExp(`--${name}:\\s*(#[0-9a-fA-F]{6})`));
  return match?.[1] ?? "currentColor";
}

export function GET() {
  const brand = paletteColor("palette-brand");
  const surface = paletteColor("palette-surface");
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
    <rect width="64" height="64" rx="14" fill="${brand}"/>
    <path d="M18 42V18h16c7 0 12 4 12 10 0 4-2 8-6 9l8 5H36l-7-5h-1v5H18Zm10-14h6c2 0 4-1 4-3s-2-3-4-3h-6v6Z" fill="${surface}"/>
  </svg>`;

  return new Response(svg, {
    headers: {
      "Content-Type": "image/svg+xml; charset=utf-8",
      "Cache-Control": "public, max-age=86400",
    },
  });
}
