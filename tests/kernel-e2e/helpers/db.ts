import os from "os";
import path from "path";
import fs from "fs/promises";

export async function createTestDbPath() {
  const dir = await fs.mkdtemp(
    path.join(os.tmpdir(), "crona-test-")
  );

  const dbPath = path.join(dir, "crona.db");

  console.log("dbPath:", dbPath);

  return {
    dbPath,
    cleanup: async () => {
      await fs.rm(dir, { recursive: true, force: true });
    }
  };
}
