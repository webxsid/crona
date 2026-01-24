import { describe, it, beforeAll, afterAll, expect } from "vitest";
import {
  startTestKernel,
  stopTestKernel,
  type IKernelTestHandle
} from "./helpers/kernel";
import { api } from "./helpers/http";
import type { Repo, Stream } from "@crona/core";

describe("@stream @e2e", () => {
  let kernel: IKernelTestHandle;
  let repo: Repo;

  beforeAll(async () => {
    kernel = await startTestKernel();

    // Create parent repo
    repo = await api<Repo>(
      kernel.baseUrl,
      "/commands/repo",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: "Office1" }),
      }
    );
  });

  afterAll(async () => {
    await stopTestKernel(kernel);
  });

  it("creates a stream under a repo", async () => {
    const stream = await api<Stream>(
      kernel.baseUrl,
      "/stream",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: "feature-auth",
          repoId: repo.id,
        }),
      }
    );

    expect(stream).toHaveProperty("id");
    expect(stream.name).toBe("feature-auth");
    expect(stream.repoId).toBe(repo.id);
  });

  it("lists streams for a repo", async () => {
    const streams = await api<Stream[]>(
      kernel.baseUrl,
      `/streams?repoId=${repo.id}`,
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(Array.isArray(streams)).toBe(true);
    expect(streams.length).toBeGreaterThan(0);

    expect(
      streams.find(s => s.name === "feature-auth")
    ).toBeDefined();
  });

  it("updates a stream", async () => {
    const stream = await api<Stream>(
      kernel.baseUrl,
      "/stream",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: "old-stream",
          repoId: repo.id,
        }),
      }
    );

    const updated = await api<Stream>(
      kernel.baseUrl,
      `/stream/${stream.id}`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: "new-stream",
        }),
      }
    );

    expect(updated.id).toBe(stream.id);
    expect(updated.name).toBe("new-stream");
  });

  it("deletes a stream", async () => {
    const stream = await api<Stream>(
      kernel.baseUrl,
      "/stream",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: "to-delete",
          repoId: repo.id,
        }),
      }
    );

    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      `/stream/${stream.id}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(res.ok).toBe(true);

    const streams = await api<Stream[]>(
      kernel.baseUrl,
      `/streams?repoId=${repo.id}`,
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(
      streams.find(s => s.id === stream.id)
    ).toBeUndefined();
  });
});
