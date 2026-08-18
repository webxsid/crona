import { readFile, readdir } from "node:fs/promises";
import { join, relative } from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../docs", import.meta.url));
const errors = [];

async function collect(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await collect(path)));
    else if (entry.name.endsWith(".md")) files.push(path);
  }
  return files;
}

const files = await collect(root);
for (const path of files) {
  const name = relative(root, path);
  const content = await readFile(path, "utf8");
  const frontmatter = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!frontmatter) {
    errors.push(`${name}: missing frontmatter`);
    continue;
  }
  const fields = new Map();
  for (const line of frontmatter[1].split("\n")) {
    const match = line.match(/^([A-Za-z][\w-]*):\s*(.*)$/);
    if (match) fields.set(match[1], match[2].trim());
  }
  if (!fields.has("title") || !fields.get("title")) errors.push(`${name}: missing title`);
  if (fields.get("hosted") !== "true" && fields.get("hosted") !== "false") {
    errors.push(`${name}: hosted must be true or false`);
  }
}

if (errors.length) {
  console.error(errors.join("\n"));
  process.exit(1);
}

console.log(`Validated frontmatter in ${files.length} Markdown files.`);
