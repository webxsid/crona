import fs from "fs/promises";
import path from "path";
import os from "os";

export interface KernelInfo {
  port: number;
  token: string;
  pid: number;
  startedAt: string;
  scratchDir: string;
}

const CRONA_DIR = path.join(os.homedir(), ".crona");
const KERNEL_INFO_FILE = path.join(CRONA_DIR, "kernel.json");
const CRONA_SCRATCH_DIR = path.join(CRONA_DIR, "scratch");

/**
 * Ensure ~/.crona exists with safe permissions
 */
async function ensureDir() {
  await fs.mkdir(CRONA_DIR, { recursive: true, mode: 0o700 });
  await fs.mkdir(CRONA_SCRATCH_DIR, { recursive: true, mode: 0o700 });
}

/**
 * Write kernel connection info
 * Overwrites existing file atomically
 */
export async function writeKernelInfo(info: Omit<KernelInfo, "scratchDir">): Promise<void> {
  await ensureDir();

  const tempFile = `${KERNEL_INFO_FILE}.tmp`;

  await fs.writeFile(
    tempFile,
    JSON.stringify({
      ...info,
      scratchDir: CRONA_SCRATCH_DIR,
    }, null, 2),
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
    console.error("Failed to clear kernel info:", err);
  }
}
