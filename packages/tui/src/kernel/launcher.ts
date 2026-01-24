import fs from "fs";
import path from "path";
import os from "os";
import { spawn } from "child_process";
import fetch from "node-fetch";

const KERNEL_INFO_PATH = path.join(
  os.homedir(),
  ".crona",
  "kernel.json"
);

export async function ensureKernelRunning(): Promise<{
  baseUrl: string;
  token: string;
}> {
  if (fs.existsSync(KERNEL_INFO_PATH)) {
    const info = JSON.parse(
      fs.readFileSync(KERNEL_INFO_PATH, "utf-8")
    );

    if (await isKernelHealthy(info)) {
      return {
        baseUrl: `http://127.0.0.1:${info.port}`,
        token: info.token,
      };
    }
  }

  spawnKernel();
  console.log("Spawning Crona kernel…");

  // wait for kernel.json to appear
  for (let i = 0; i < 20; i++) {
    await sleep(250);
    if (fs.existsSync(KERNEL_INFO_PATH)) {
      const info = JSON.parse(
        fs.readFileSync(KERNEL_INFO_PATH, "utf-8")
      );

      return {
        baseUrl: `http://127.0.0.1:${info.port}`,
        token: info.token,
      };
    }
  }

  throw new Error("Kernel failed to start");
}

async function isKernelHealthy(info: {
  pid: number;
  port: number;
}): Promise<boolean> {
  try {
    process.kill(info.pid, 0); // PID exists?

    const res = await fetch(
      `http://127.0.0.1:${info.port}/health`
    );
    console.log(`Health check status: ${res.status}`);
    return res.ok;
  } catch {
    return false;
  }
}

function spawnKernel() {
  const child = spawn(
    "crona-kernel",
    [],
    {
      detached: true,
      stdio: "ignore",
    }
  );

  child.unref();
}

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}
