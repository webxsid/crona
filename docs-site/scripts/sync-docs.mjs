import { cp, mkdir, readFile, readdir, rename, rm, unlink, writeFile } from "node:fs/promises";
import { dirname, join, relative, resolve, sep } from "node:path";
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
const sourceMarkdown = new Map();
const markdownFiles = [];
async function collectMarkdown(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) await collectMarkdown(path);
    else if (entry.name.endsWith(".md")) markdownFiles.push(path);
  }
}
const sourceMarkdownFiles = [];
async function collectSourceMarkdown(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) await collectSourceMarkdown(path);
    else if (entry.name.endsWith(".md")) sourceMarkdownFiles.push(path);
  }
}
await collectSourceMarkdown(source);
for (const path of sourceMarkdownFiles) {
  sourceMarkdown.set(relative(source, path).split(sep).join("/"), frontmatter(await readFile(path, "utf8")));
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
  const sourcePath = path === index ? "README.md" : relative(destination, path).split(sep).join("/");
  const rewritten = content
    .replace(/\]\(\.\.\/\.\.\/(shared\/[^)#]+)(#[^)]+)?\)/g, (_match, file, fragment = "") =>
      `](${sourceBase}${file}${fragment})`,
    )
    .replace(/\]\((?!https?:|\/|#|mailto:)([^)#]+\.md)(#[^)]+)?\)/g, (_match, target, fragment = "") => {
      const targetPath = relative(source, resolve(source, dirname(sourcePath), target)).split(sep).join("/");
      const metadata = sourceMarkdown.get(targetPath);
      if (!metadata) return _match;
      if (metadata.get("hosted") === "true") {
        const route = targetPath === "README.md"
          ? "/"
          : targetPath.endsWith("/index.md")
            ? `/${targetPath.slice(0, -"index.md".length)}`
            : `/${targetPath.slice(0, -3)}/`;
        return `](${route}${fragment})`;
      }
      return `](${sourceBase}${targetPath}${fragment})`;
    });
  if (rewritten !== original) await writeFile(path, rewritten);
}

console.log(`Synced local docs from ${source}`);
