import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const cssDir = path.join(root, ".next", "static", "css");
const failures = [];

if (!fs.existsSync(cssDir)) {
  failures.push(`missing_css_output ${cssDir}`);
} else {
  const cssFiles = fs.readdirSync(cssDir)
    .filter((name) => name.endsWith(".css"))
    .map((name) => path.join(cssDir, name));

  if (cssFiles.length === 0) failures.push("missing_css_files");

  for (const file of cssFiles) {
    const text = fs.readFileSync(file, "utf8");
    const relative = path.relative(root, file).replaceAll(path.sep, "/");
    if (/@media\s+layer\s*\(/i.test(text)) {
      failures.push(`invalid_cascade_layer_media ${relative}`);
    }
  }
}

if (failures.length > 0) {
  console.error("Build artifact verification failed:");
  for (const failure of failures) console.error(`- ${failure}`);
  process.exit(1);
}

console.log("Build artifact verification passed: CSS cascade layers are valid");
