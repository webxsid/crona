import { bootstrapKernel, clearKernelInfo, readKernelInfo } from "@crona/kernel";
import { createTestDbPath } from "./db";

export interface IKernelTestHandle {
  baseUrl: string;
  token: string;
  pid: number;
  cleanup: () => Promise<void>;
}

export async function startTestKernel(): Promise<IKernelTestHandle> {
  const { dbPath, cleanup } = await createTestDbPath()

  console.log(`Starting kernel with test db at ${dbPath}`, bootstrapKernel);
  await bootstrapKernel({ dbPath });

  const info = await readKernelInfo();
  if (!info) {
    throw new Error("Kernel did not write kernel info");
  }

  return {
    baseUrl: `http://127.0.0.1:${info.port}`,
    token: info.token,
    pid: info.pid,
    cleanup,
  };
}

export async function stopTestKernel(kernel: IKernelTestHandle): Promise<void> {
  try {
    clearKernelInfo()
  } catch { }
  try {
    await kernel.cleanup();
  } catch { }
}
