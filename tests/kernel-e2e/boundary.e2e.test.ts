import { describe, it, expect, beforeAll, afterAll } from "vitest";
import type { IKernelTestHandle } from "./helpers/kernel";
import { startTestKernel, stopTestKernel } from "./helpers/kernel";
import { api } from "./helpers/http";
import { listenEvents } from "./helpers/events";
import { type Repo, type Stream, type Issue, type CoreSettings, type Session, SessionNotesParser } from "@crona/core";

let kernel: IKernelTestHandle;

beforeAll(async () => {
  kernel = await startTestKernel();
});

afterAll(async () => {
  await stopTestKernel(kernel);
});

describe("@timer @boundary @e2e", () => {
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
          name: "focus",
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
          title: "Boundary test",
          repoId: repo.id,
          streamId: stream.id,
        }),
      }
    );

    // Context
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

    // 🔧 Override core settings for fast boundary
    await api(
      kernel.baseUrl,
      "/settings/core",
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          timerMode: "structured",
          breaksEnabled: 1 as unknown as Boolean,
          workDurationMinutes: 0.05, // ~3 seconds
          shortBreakMinutes: 0.05,
          longBreakEnabled: 1 as unknown as Boolean,
          longBreakMinutes: 0.1, // ~6 seconds
          autoStartWork: 1 as unknown as Boolean,
          autoStartBreaks: 1 as unknown as Boolean,
          cyclesBeforeLongBreak: 1,
        } as Partial<CoreSettings>),
      }
    );
  });

  it("automatically transitions work → rest at boundary", async () => {
    const { events, close } = listenEvents(kernel.baseUrl, kernel.token);

    // Start timer
    await api(
      kernel.baseUrl,
      "/timer/start",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    // Initial state should be running(work)
    const running = await api<{
      state: "running";
      segmentType: "work";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(running.state).toBe("running");
    expect(running.segmentType).toBe("work");

    // ⏳ wait slightly longer than workDurationMinutes
    await new Promise((r) => setTimeout(r, 3001));

    // After boundary → rest
    const afterBoundary = await api<{
      state: "paused";
      segmentType: "short_break";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(afterBoundary.state).toBe("paused");
    expect(afterBoundary.segmentType).toBe("short_break");

    // Verify boundary event emitted
    const boundaryEvents = events().filter(
      e => e.event === "timer.boundary"
    );

    expect(boundaryEvents.length).toBeGreaterThan(0);
    expect(boundaryEvents[0]?.data).toMatchObject({
      from: "work",
      to: "short_break",
    });

    close();
  });

  it("automatically transitions rest → work at boundary", async () => {
    const { events, close } = listenEvents(kernel.baseUrl, kernel.token);

    // Currently should be in paused(rest) state
    const before = await api<{
      state: "paused";
      segmentType: "short_break";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(before.state).toBe("paused");
    expect(before.segmentType).toBe("short_break");

    // ⏳ wait slightly longer than shortBreakMinutes
    await new Promise((r) => setTimeout(r, 3001));

    // After boundary → work (should auto-start)
    const afterBoundary = await api<{
      state: "running";
      segmentType: "work";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(afterBoundary.state).toBe("running");
    expect(afterBoundary.segmentType).toBe("work");

    // Verify boundary event emitted
    const boundaryEvents = events().filter(
      e => e.event === "timer.boundary"
    );

    expect(boundaryEvents.length).toBeGreaterThan(0);
    expect(boundaryEvents[0]?.data).toMatchObject({
      from: "short_break",
      to: "work",
    });

    close();
  });
  it("handles long break boundary correctly", async () => {
    const { events, close } = listenEvents(kernel.baseUrl, kernel.token);

    // Currently should be in running(work) state
    const before = await api<{
      state: "running";
      segmentType: "work";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(before.state).toBe("running");
    expect(before.segmentType).toBe("work");

    // ⏳ wait slightly longer than workDurationMinutes
    await new Promise((r) => setTimeout(r, 3001));

    // After boundary → long_break
    const afterWorkBoundary = await api<{
      state: "paused";
      segmentType: "long_break";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(afterWorkBoundary.state).toBe("paused");
    expect(afterWorkBoundary.segmentType).toBe("long_break");

    // ⏳ wait slightly longer than longBreakMinutes
    await new Promise((r) => setTimeout(r, 6000));

    // After boundary → work (should auto-start)
    const afterLongBreakBoundary = await api<{
      state: "running";
      segmentType: "work";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(afterLongBreakBoundary.state).toBe("running");
    expect(afterLongBreakBoundary.segmentType).toBe("work");

    // Verify boundary events emitted
    const boundaryEvents = events().filter(
      e => e.event === "timer.boundary"
    );

    const longBreakStartEvent = boundaryEvents.find(
      e => e.data.from === "work" && e.data.to === "long_break"
    );
    const longBreakEndEvent = boundaryEvents.find(
      e => e.data.from === "long_break" && e.data.to === "work"
    );

    expect(longBreakStartEvent).toBeDefined();
    expect(longBreakEndEvent).toBeDefined();

    close();
  });
  it("ends session correctly after boundaries", async () => {
    const { events, close } = listenEvents(kernel.baseUrl, kernel.token);

    // Currently should be in running(work) state
    const before = await api<{
      state: "running";
      segmentType: "work";
    }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(before.state).toBe("running");
    expect(before.segmentType).toBe("work");

    // End timer
    await api(
      kernel.baseUrl,
      "/timer/end",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    // After end → idle
    const afterEnd = await api<{ state: "idle" }>(
      kernel.baseUrl,
      "/timer/state",
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(afterEnd.state).toBe("idle");

    const sessions = await api<Session[]>(
      kernel.baseUrl,
      `/sessions?issueId=${issue.id}`,
      {
        headers: { Authorization: `Bearer ${kernel.token}` },
      }
    );

    expect(sessions.length).toBeGreaterThanOrEqual(1);
    const session = sessions[0];
    expect(session?.endTime).toBeDefined();
    expect(session?.durationSeconds).toBeGreaterThan(0);

    const parsedNotes = SessionNotesParser.parse(session?.notes);
    expect(parsedNotes.commit).toBe("Work Session");
    expect(parsedNotes.work).toBeDefined();

    const ammendedMessage = "Finalized boundary e2e tests";
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
    expect(updatedParsedNotes.work).toBeDefined();
    expect(updatedParsedNotes.work).toBe(parsedNotes.work);

    // Verify no boundary events emitted
    const boundaryEvents = events().filter(
      e => e.event === "timer.boundary"
    );

    expect(boundaryEvents.length).toBe(0);

    close();
  });
});
