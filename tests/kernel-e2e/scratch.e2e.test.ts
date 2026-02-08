import { describe, it, beforeAll, afterAll, expect } from "vitest";
import {
  startTestKernel,
  stopTestKernel,
  type IKernelTestHandle
} from "./helpers/kernel";
import { api } from "./helpers/http";

type ScratchpadMeta = {
  path: string;
  name: string;
  pinned?: boolean;
  lastOpenedAt?: string;
};

describe("@scratch @e2e", () => {
  let kernel: IKernelTestHandle;

  beforeAll(async () => {
    kernel = await startTestKernel();
  });

  afterAll(async () => {
    await stopTestKernel(kernel);
  });

  it("exposes scratch directory via kernel info", async () => {
    const info = await api<{ scratchDir: string }>(
      kernel.baseUrl,
      "/kernel/info",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(info).toHaveProperty("scratchDir");
    expect(info.scratchDir).toContain(".crona");
  });

  it("registers a scratchpad", async () => {
    const scratch: ScratchpadMeta = {
      path: "notes/today.md",
      name: "Today Notes",
    };

    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      "/scratchpads/register",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(scratch),
      }
    );

    expect(res.ok).toBe(true);
  });

  it("lists scratchpads", async () => {
    const list = await api<ScratchpadMeta[]>(
      kernel.baseUrl,
      "/scratchpads",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(Array.isArray(list)).toBe(true);
    expect(list.length).toBeGreaterThan(0);

    const entry = list.find(s => s.path === "notes/today.md");
    expect(entry).toBeDefined();
    expect(entry?.name).toBe("Today Notes");
  });

  it("pins and unpins a scratchpad", async () => {
    // pin
    const pinRes = await api<{ ok: boolean }>(
      kernel.baseUrl,
      "/scratchpads/pin?path=notes/today.md",
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ pinned: true }),
      }
    );

    expect(pinRes.ok).toBe(true);

    let list = await api<ScratchpadMeta[]>(
      kernel.baseUrl,
      "/scratchpads",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(list.find(s => s.path === "notes/today.md")?.pinned).toBe(true);

    // unpin
    await api(
      kernel.baseUrl,
      "/scratchpads/pin?path=notes/today.md",
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ pinned: false }),
      }
    );

    list = await api<ScratchpadMeta[]>(
      kernel.baseUrl,
      "/scratchpads",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(list.find(s => s.path === "notes/today.md")?.pinned).toBe(false);
  });

  /* -------------------------------------------------------------------------- */
  /*                            VARIABLE PATH TESTS                             */
  /* -------------------------------------------------------------------------- */


  it("expands date variable in scratchpad path", async () => {
    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      "/scratchpads/register",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          path: "daily/[[date]].md",
          name: "Daily Notes",
        }),
      }
    );

    expect(res.ok).toBe(true);

    const list = await api<ScratchpadMeta[]>(
      kernel.baseUrl,
      "/scratchpads",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    const today = new Date().toISOString().split("T")[0];
    const entry = list.find(s => s.path.includes(`daily/${today}`));

    expect(entry).toBeDefined();
    expect(entry?.path).toContain(today);
  });

  it("supports multiple variables in path", async () => {
    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      "/scratchpads/register",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          path: "sessions/[[date]]/[[time]]-notes.md",
          name: "Session Notes",
        }),
      }
    );

    expect(res.ok).toBe(true);

    const list = await api<ScratchpadMeta[]>(
      kernel.baseUrl,
      "/scratchpads",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    const match = list.find(s =>
      /sessions\/\d{4}-\d{2}-\d{2}\/\d{2}-\d{2}-\d{2}-notes\.md/.test(s.path)
    );

    expect(match).toBeDefined();
  });

  it("rejects unsupported variables in path", async () => {
    const res = await fetch(
      `${kernel.baseUrl}/scratchpads/register`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          path: "notes/[[invalid]].md",
          name: "Invalid Notes",
        }),
      }
    );

    const body = await res.json();

    expect(res.status).toBe(500);
    expect(body.message).toContain("Invalid variable in path");
  });

});

