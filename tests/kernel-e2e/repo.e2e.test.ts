import { describe, it, beforeAll, afterAll, expect } from "vitest";
import { startTestKernel, stopTestKernel, type IKernelTestHandle } from "./helpers/kernel";
import { api } from "./helpers/http";
import type { Repo } from "@crona/core"

describe("@repo @e2e", () => {
  let kernel: IKernelTestHandle;

  beforeAll(async () => {
    kernel = await startTestKernel();
  });

  afterAll(async () => {
    await stopTestKernel(kernel);
  });

  it("creates a repo", async () => {
    const repo = await api<Repo>(
      kernel.baseUrl,
      "/commands/repo",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: "Office" }),
      }
    )

    expect(repo).toHaveProperty("id");
    expect(repo.name).toBe("Office");

  });

  it("lists repos", async () => {
    const repos = await api<Repo[]>(
      kernel.baseUrl,
      "/repos",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(Array.isArray(repos)).toBe(true);
    expect(repos.length).toBeGreaterThan(0);

    expect(repos[0]).toHaveProperty("id");
    expect(repos[0]).toHaveProperty("name");

    expect(repos.find(r => r.name === "Office")).toBeDefined();
  });

  it("updates a repo", async () => {
    // create
    const repo = await api<Repo>(
      kernel.baseUrl,
      "/commands/repo",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: "Old Name" }),
      }
    )

    // update
    const updatedRepo = await api<Repo>(
      kernel.baseUrl,
      `/commands/repo/${repo.id}`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: "New Name" }),
      }
    )

    expect(updatedRepo.id).toBe(repo.id);
    expect(updatedRepo.name).toBe("New Name");
  });

  it("deletes a repo", async () => {

    const repo = await api<Repo>(
      kernel.baseUrl,
      "/commands/repo",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: "To Be Deleted" }),
      }
    )

    const deleteResponse = await api<{ ok: boolean }>(
      kernel.baseUrl,
      `/commands/repo/${repo.id}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    )

    expect(deleteResponse.ok).toBe(true);

    // Verify deletion
    const repos = await api<Repo[]>(
      kernel.baseUrl,
      "/repos",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(repos.find(r => r.id === repo.id)).toBeUndefined();

  });
});
