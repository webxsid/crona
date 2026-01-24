import { describe, it, expect, beforeAll, afterAll } from "vitest";
import type { IKernelTestHandle } from "./helpers/kernel";
import { startTestKernel, stopTestKernel } from "./helpers/kernel";
import { api } from "./helpers/http";
import { listenEvents } from "./helpers/events";
import type { Repo, Stream, Issue, Session } from "@crona/core";
import { SessionNotesParser } from "@crona/core";

let kernel: IKernelTestHandle;

beforeAll(async () => {
  kernel = await startTestKernel();
});

afterAll(async () => {
  await stopTestKernel(kernel);
});

describe("@session @e2e", () => {
  let repo: Repo;
  let stream: Stream;
  let issue: Issue;

  beforeAll(async () => {
    // Repo
    repo = await api<Repo>(
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
    );

    // Stream
    stream = await api<Stream>(
      kernel.baseUrl,
      "/stream",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: "backend",
          repoId: repo.id,
        }),
      }
    );

    // Issue
    issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          title: "Implement timer",
          repoId: repo.id,
          streamId: stream.id,
        }),
      }
    );

    // Active context (git-like)
    await api(
      kernel.baseUrl,
      "/context",
      {
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
      }
    );
  });

  it("start → pause → resume → end (timer-driven)", async () => {
    const { events, close } = listenEvents(kernel.baseUrl, kernel.token);

    // 1. Idle
    const idle = await api<{ state: "idle" }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );
    expect(idle.state).toBe("idle");

    // 2. Start timer (creates session implicitly)
    await api(
      kernel.baseUrl,
      "/timer/start",
      {
        method: "POST",
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    // 3. Running (work)
    const running = await api<{
      state: "running";
      issueId: string;
      segmentType: "work" | "rest";
      elapsedSeconds: number;
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(running.state).toBe("running");
    expect(running.issueId).toBe(issue.id);
    expect(running.segmentType).toBe("work");

    await new Promise(r => setTimeout(r, 2000));

    // 3.5 Check elapsed time increased
    const running2 = await api<{
      state: "running";
      issueId: string;
      segmentType: "work" | "rest";
      elapsedSeconds: number;
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(running2.elapsedSeconds).toBeGreaterThan(running.elapsedSeconds);

    // 4. Pause (work → rest)
    await api(
      kernel.baseUrl,
      "/timer/pause",
      {
        method: "POST",
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    const paused = await api<{
      state: "paused";
      segmentType: "rest";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(paused.state).toBe("paused");
    expect(paused.segmentType).toBe("rest");

    // 5. Resume (rest → work)
    await api(
      kernel.baseUrl,
      "/timer/resume",
      {
        method: "POST",
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    const resumed = await api<{
      state: "running";
      segmentType: "work";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(resumed.state).toBe("running");
    expect(resumed.segmentType).toBe("work");

    // 6. End (session closed)
    await api(
      kernel.baseUrl,
      "/timer/end",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "content-type": "application/json",
        },
        body: JSON.stringify({
          commitMessage: "Worked on timer e2e tests",
        })
      }
    );

    const ended = await api<{ state: "idle" }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(ended.state).toBe("idle");

    // fetch the sessions for the issue to verify
    const sessions = await api<Session[]>(
      kernel.baseUrl,
      `/sessions?issueId=${issue.id}`,
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(sessions.length).toBe(1);
    const session = sessions[0];
    const parsedNotes = SessionNotesParser.parse(session?.notes);
    expect(parsedNotes.commit).toBe("Worked on timer e2e tests");

    // ammend last session commit message
    const ammendedMessage = "Finalized timer e2e tests";
    await api(
      kernel.baseUrl,
      `/sessions/note`,
      {
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "content-type": "application/json",
        },
        body: JSON.stringify({
          note: ammendedMessage,
        })
      }
    );

    const updatedSessions = await api<Session[]>(
      kernel.baseUrl,
      `/sessions?issueId=${issue.id}`,
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(updatedSessions.length).toBe(1);
    const updatedSession = updatedSessions[0];
    const updatedParsedNotes = SessionNotesParser.parse(updatedSession?.notes);
    expect(updatedParsedNotes.commit).toBe(ammendedMessage);

    // 7. Events observed
    const eventTypes = events().map(e => e.event);
    expect(eventTypes).toContain("timer.state");

    close();
  });

  it("should not start timer without active issue context", async () => {
    // Clear issue from context
    await api(
      kernel.baseUrl,
      "/context/issue",
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    const res = await fetch(`${kernel.baseUrl}/timer/start`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${kernel.token}`,
      },
    });

    expect(res.status).toBe(500);

    const body = await res.json();
    expect(body.message).toContain(
      "No issue specified and no active issue in context"
    );
  });
});
