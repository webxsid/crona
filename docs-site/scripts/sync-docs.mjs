import { cp, mkdir, readFile, readdir, rename, rm, unlink, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const source = join(root, "..", "docs");
const destination = join(root, "src/content/docs");

const frontmatter = (content) => {
  const match = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) return new Map();
  return new Map(match[1].split("\n").flatMap((line) => {
    const field = line.match(/^([A-Za-z][\w-]*):\s*(.*)$/);
    return field ? [[field[1], field[2].trim()]] : [];
  }));
};

await rm(destination, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });

const readme = join(destination, "README.md");
const index = join(destination, "index.md");
await rename(readme, index);

const sourceBase = "https://github.com/webxsid/crona/blob/main/";
const markdownFiles = [];
async function collectMarkdown(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) await collectMarkdown(path);
    else if (entry.name.endsWith(".md")) markdownFiles.push(path);
  }
}
await collectMarkdown(destination);

for (const path of markdownFiles) {
  const original = await readFile(path, "utf8");
  if (frontmatter(original).get("hosted") !== "true") {
    await unlink(path);
    continue;
  }
  let content = original;
  if (!content.startsWith("---\n")) {
    const heading = content.match(/^#\s+(.+)\s*\n/);
    if (heading) {
      content = `---\ntitle: ${JSON.stringify(heading[1].trim())}\n---\n\n${content.slice(heading[0].length)}`;
    }
  }
  const rewritten = content.replace(/\]\(\.\.\/\.\.\/(shared\/[^)#]+)(#[^)]+)?\)/g, (_match, file, fragment = "") =>
    `](${sourceBase}${file}${fragment})`,
  );
  if (rewritten !== original) await writeFile(path, rewritten);
}

console.log(`Synced local docs from ${source}`);
