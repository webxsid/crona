import { describe, it, beforeAll, afterAll, expect } from "vitest";
import {
  startTestKernel,
  stopTestKernel,
  type IKernelTestHandle,
} from "./helpers/kernel";
import { api } from "./helpers/http";
import type { Repo, Stream, Issue, Stash, ActiveContext } from "@crona/core";

describe("@stash @e2e", () => {
  let kernel: IKernelTestHandle;

  let repo: Repo;
  let stream: Stream;
  let issue: Issue;

  beforeAll(async () => {
    kernel = await startTestKernel();

    // ---------- repo ----------
    repo = await api<Repo>(kernel.baseUrl, "/commands/repo", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name: "Stash Repo" }),
    });

    // ---------- stream ----------
    stream = await api<Stream>(kernel.baseUrl, "/stream", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        repoId: repo.id,
        name: "main",
      }),
    });

    // ---------- issue ----------
    issue = await api<Issue>(kernel.baseUrl, "/issue", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        streamId: stream.id,
        title: "Stash Test Issue",
      }),
    });

    // ---------- context ----------
    await api(kernel.baseUrl, "/context", {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        repoId: repo.id,
        streamId: stream.id,
        issueId: issue.id,
      }),
    });
  });

  afterAll(async () => {
    await stopTestKernel(kernel);
  });

  it("creates a stash from active timer + context", async () => {
    // start timer
    await api(kernel.baseUrl, `/timer/start?issueId=${issue.id}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    await new Promise(r => setTimeout(r, 1000));

    // stash
    const stash = await api<Stash>(kernel.baseUrl, "/stash", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        stashNote: "Switching context temporarily",
      }),
    });

    expect(stash).toHaveProperty("id");
    expect(stash.issueId).toBe(issue.id);
    expect(stash.note).toContain("Switching context temporarily");

    // timer should now be idle
    const timer = await api<{ state: string }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(timer.state).toBe("idle");

    // context should be cleared
    const context = await api<ActiveContext>(kernel.baseUrl, "/context", {
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    expect(context.issueId).toBeUndefined();
  });

  it("lists stashes", async () => {
    const stashes = await api<Stash[]>(kernel.baseUrl, "/stash", {
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    expect(Array.isArray(stashes)).toBe(true);
    expect(stashes.length).toBeGreaterThan(0);
  });

  it("applies stash and restores context", async () => {
    const stashes = await api<Stash[]>(kernel.baseUrl, "/stash", {
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    const stashId = stashes[0]?.id;

    // apply stash
    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      `/stash/${stashId}/apply`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(res.ok).toBe(true);

    // context restored
    const context = await api<ActiveContext>(kernel.baseUrl, "/context", {
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    expect(context.issueId).toBe(issue.id);

    // timer remains idle (explicit start required)
    const timer = await api<{ state: string }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(timer.state).toBe("idle");
  });

  it("drops stash without applying", async () => {
    const stash = await api<Stash>(kernel.baseUrl, "/stash", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        stashNote: "Disposable stash",
      }),
    });

    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      `/stash/${stash.id}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(res.ok).toBe(true);

    const stashes = await api<Stash[]>(kernel.baseUrl, "/stash", {
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    expect(stashes.find(s => s.id === stash.id)).toBeUndefined();
  });
});
