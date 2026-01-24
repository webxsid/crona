import fs from "fs/promises";
import path from "path";
import os from "os";

export interface KernelInfo {
  port: number;
  token: string;
  pid: number;
  startedAt: string;
}

const CRONA_DIR = path.join(os.homedir(), ".crona");
const KERNEL_INFO_FILE = path.join(CRONA_DIR, "kernel.json");

/**
 * Ensure ~/.crona exists with safe permissions
 */
async function ensureDir() {
  await fs.mkdir(CRONA_DIR, { recursive: true, mode: 0o700 });
}

/**
 * Write kernel connection info
 * Overwrites existing file atomically
 */
export async function writeKernelInfo(info: KernelInfo): Promise<void> {
  await ensureDir();

  const tempFile = `${KERNEL_INFO_FILE}.tmp`;

  await fs.writeFile(
    tempFile,
    JSON.stringify(info, null, 2),
    { mode: 0o600 }
  );

  await fs.rename(tempFile, KERNEL_INFO_FILE);
}

/**
 * Read kernel info if present
 */
export async function readKernelInfo(): Promise<KernelInfo | null> {
  try {
    const raw = await fs.readFile(KERNEL_INFO_FILE, "utf8");
    return JSON.parse(raw) as KernelInfo;
  } catch {
    return null;
  }
}

/**
 * Remove kernel info (on shutdown / crash recovery)
 */
export async function clearKernelInfo(): Promise<void> {
  try {
    await fs.unlink(KERNEL_INFO_FILE);
  } catch (err) {
    // ignore
    console.error("Failed to clear kernel info:", err);
  }
}
