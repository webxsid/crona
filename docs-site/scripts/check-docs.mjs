import { access, readFile, readdir } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
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
const hostedFiles = new Set();
const metadataByName = new Map();
for (const path of files) {
  const name = relative(root, path);
  const content = await readFile(path, "utf8");
  const frontmatter = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!frontmatter) continue;
  const fields = new Map();
  for (const line of frontmatter[1].split("\n")) {
    const match = line.match(/^([A-Za-z][\w-]*):\s*(.*)$/);
    if (match) fields.set(match[1], match[2].trim());
  }
  metadataByName.set(name, fields);
}
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
  if (fields.get("hosted") === "true") hostedFiles.add(name);
  if (fields.get("hosted") === "true") {
    for (const match of content.matchAll(/\]\(([^)]+)\)/g)) {
      const target = match[1].split("#")[0];
      if (target.endsWith(".md")) {
        const targetName = relative(root, join(dirname(path), target)).replaceAll("\\", "/");
        const metadata = metadataByName.get(targetName);
        if (metadata?.get("hosted") !== "true" && metadata) {
          errors.push(`${name}: public docs cannot link to excluded Markdown (${target})`);
        }
      }
    }
  }
}

for (const name of hostedFiles) {
  const content = await readFile(join(root, name), "utf8");
  for (const match of content.matchAll(/!\[[^\]]*\]\(([^)]+)\)/g)) {
    const target = match[1].split("#")[0];
    if (target.startsWith("../../../assets/")) {
      const asset = dirname(join(root, "../docs-site/src/content/docs", name));
      const resolved = join(asset, target);
      try { await access(resolved); } catch { errors.push(`${name}: missing asset ${target}`); }
    }
  }
}

if (errors.length) {
  console.error(errors.join("\n"));
  process.exit(1);
}

console.log(`Validated frontmatter in ${files.length} Markdown files.`);
