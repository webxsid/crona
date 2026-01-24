import { describe, it, beforeAll, afterAll, expect } from "vitest";
import {
  startTestKernel,
  stopTestKernel,
  type IKernelTestHandle,
} from "./helpers/kernel";
import { api } from "./helpers/http";
import type {
  Repo,
  Stream,
  Issue,
  Stash,
  TimerStatePayload,
} from "@crona/core";

describe("@full @e2e", () => {
  let kernel: IKernelTestHandle;

  let repoA: Repo;
  let repoB: Repo;
  let streamA: Stream;
  let streamB: Stream;
  let issueA: Issue;
  let issueB: Issue;

  beforeAll(async () => {
    kernel = await startTestKernel();

    // ---------- repos ----------
    repoA = await api<Repo>(kernel.baseUrl, "/commands/repo", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({ name: "Repo A" }),
    });

    repoB = await api<Repo>(kernel.baseUrl, "/commands/repo", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({ name: "Repo B" }),
    });

    // ---------- streams ----------
    streamA = await api<Stream>(kernel.baseUrl, "/stream", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({ repoId: repoA.id, name: "main" }),
    });

    streamB = await api<Stream>(kernel.baseUrl, "/stream", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({ repoId: repoB.id, name: "dev" }),
    });

    // ---------- issues ----------
    issueA = await api<Issue>(kernel.baseUrl, "/issue", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({
        streamId: streamA.id,
        title: "Issue A",
        estimateMinutes: 30,
      }),
    });

    issueB = await api<Issue>(kernel.baseUrl, "/issue", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({
        streamId: streamB.id,
        title: "Issue B",
        estimateMinutes: 15,
      }),
    });
  });

  afterAll(async () => {
    await stopTestKernel(kernel);
  });

  it("runs session → pause → stash → restore → resume → end", async () => {
    // ---------- set context ----------
    await api(kernel.baseUrl, "/context", {
      method: "PUT",
      headers: auth(kernel),
      body: JSON.stringify({
        repoId: repoA.id,
        streamId: streamA.id,
        issueId: issueA.id,
      }),
    });

    // ---------- idle ----------
    let state = await timerState(kernel);
    expect(state.state).toBe("idle");

    // ---------- start ----------
    await api(kernel.baseUrl, `/timer/start`, {
      method: "POST",
      headers: auth(kernel, true),
    });

    state = await timerState(kernel);
    expect(state.state).toBe("running");
    if (state.state === "running") {
      expect(state.segmentType).toBe("work");
      expect(state.issueId).toBe(issueA.id);
    }
    await sleep(1000);

    // ---------- pause ----------
    await api(kernel.baseUrl, "/timer/pause", {
      method: "POST",
      headers: auth(kernel, true),
    });

    state = await timerState(kernel);
    expect(state.state).toBe("paused");

    // ---------- stash ----------
    const stash = await api<Stash>(kernel.baseUrl, "/stash", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({
        stashNote: "Switching to another task",
      }),
    });

    expect(stash.issueId).toBe(issueA.id);
    expect(stash.pausedSegmentType).toBeDefined();

    // timer should now be idle
    state = await timerState(kernel);
    expect(state.state).toBe("idle");

    // ---------- switch context ----------
    await api(kernel.baseUrl, "/context", {
      method: "PUT",
      headers: auth(kernel),
      body: JSON.stringify({
        repoId: repoB.id,
        streamId: streamB.id,
        issueId: issueB.id,
      }),
    });

    // ---------- start new session ----------
    await api(kernel.baseUrl, `/timer/start`, {
      method: "POST",
      headers: auth(kernel, true),
    });

    state = await timerState(kernel);
    expect(state.state).toBe("running");
    if (state.state === 'running')
      expect(state.issueId).toBe(issueB.id);

    // ---------- end ----------
    await api(kernel.baseUrl, "/timer/end", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({
        commitMessage: "Finished Issue B work",
      }),
    });

    state = await timerState(kernel);
    expect(state.state).toBe("idle");

    // ---------- restore stash ----------
    await api(kernel.baseUrl, `/stash/${stash.id}/apply`, {
      method: "POST",
      headers: auth(kernel, true),
    });

    // context restored
    const context = await api<{
      repoId: string;
      streamId: string;
      issueId: string;
    }>(kernel.baseUrl, "/context", {
      headers: auth(kernel, true),
    });

    expect(context.issueId).toBe(issueA.id);

    // ---------- resume ----------
    await api(kernel.baseUrl, `/timer/start`, {
      method: "POST",
      headers: auth(kernel, true),
    });

    state = await timerState(kernel);
    expect(state.state).toBe("running");
    if (state.state === 'running')
      expect(state.issueId).toBe(issueA.id);

    // ---------- final end ----------
    await api(kernel.baseUrl, "/timer/end", {
      method: "POST",
      headers: auth(kernel),
      body: JSON.stringify({
        commitMessage: "Resumed and completed Issue A",
      }),
    });

    state = await timerState(kernel);
    expect(state.state).toBe("idle");
  });
});

/* ---------------- helpers ---------------- */

function auth(kernel: IKernelTestHandle, noBody: boolean = false) {
  return {
    Authorization: `Bearer ${kernel.token}`,
    ...(!noBody ? {
      "Content-Type": "application/json",
    } : {})
  };
}

async function timerState(kernel: IKernelTestHandle) {
  return api<TimerStatePayload>(kernel.baseUrl, "/timer/state", {
    headers: auth(kernel, true),
  });
}

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}
