import fs from "fs";
import path from "path";
import os from "os";

const SERVICE_LOG_BASE = path.join(os.homedir(), ".crona", "logs", "kernel");

function date(): string {
  return new Date().toISOString().slice(0, 10); // YYYY-MM-DD
}

function logDir(): string {
  const dir = path.join(SERVICE_LOG_BASE, date());
  try { fs.mkdirSync(dir, { recursive: true, mode: 0o700 }); } catch { /* ignore */ }
  return dir;
}

function formatEntry(level: string, msg: string, err?: unknown): string {
  const ts = new Date().toISOString();
  let entry = `[${ts}] [${level}] ${msg}`;
  if (err instanceof Error) {
    entry += `\n  Error: ${err.message}`;
    if (err.stack) entry += `\n  Stack: ${err.stack.split("\n").slice(1).join("\n  ")}`;
  } else if (err !== undefined) {
    entry += `\n  Detail: ${String(err)}`;
  }
  return entry + "\n";
}

export function logError(msg: string, err?: unknown): void {
  const entry = formatEntry("ERROR", msg, err);
  const dir = logDir();
  try { fs.appendFileSync(path.join(dir, "info.log"), entry); } catch { /* never throw from logger */ }
  try { fs.appendFileSync(path.join(dir, "error.log"), entry); } catch { /* never throw from logger */ }
}

export function logInfo(msg: string): void {
  const dir = logDir();
  try { fs.appendFileSync(path.join(dir, "info.log"), formatEntry("INFO", msg)); } catch { /* never throw from logger */ }
}
