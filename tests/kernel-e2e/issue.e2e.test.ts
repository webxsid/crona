import { describe, it, beforeAll, afterAll, expect } from "vitest";
import { startTestKernel, stopTestKernel, type IKernelTestHandle } from "./helpers/kernel";
import { api } from "./helpers/http";
import type { Repo, Stream, Issue, IssueStatus, IssueWithMeta } from "@crona/core";

describe("@issue @e2e", () => {
  let kernel: IKernelTestHandle;
  let repo: Repo;
  let stream: Stream;

  beforeAll(async () => {
    kernel = await startTestKernel();

    // Create repo
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

    // Create stream
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
          repoId: repo.id,
          name: "Backend",
        }),
      }
    );
  });

  afterAll(async () => {
    await stopTestKernel(kernel);
  });

  it("creates an issue", async () => {
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Implement timer boundaries",
          estimateMinutes: 90,
        }),
      }
    );

    expect(issue).toHaveProperty("id");
    expect(issue.title).toBe("Implement timer boundaries");
    expect(issue.streamId).toBe(stream.id);
    expect(issue.estimateMinutes).toBe(90);
  });

  it("lists issues by stream", async () => {
    const issues = await api<Issue[]>(
      kernel.baseUrl,
      `/issues?streamId=${stream.id}`,
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(Array.isArray(issues)).toBe(true);
    expect(issues.length).toBeGreaterThan(0);

    const issue = issues[0];
    expect(issue).toHaveProperty("id");
    expect(issue).toHaveProperty("title");
    expect(issue?.streamId).toBe(stream.id);
  });

  it("lists all issues with repo and stream meta", async () => {
    const issues = await api<IssueWithMeta[]>(
      kernel.baseUrl,
      "/issues/all",
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(Array.isArray(issues)).toBe(true);
    expect(issues.length).toBeGreaterThan(0);

    const issue = issues[0];
    expect(issue).toHaveProperty("id");
    expect(issue).toHaveProperty("repoId");
    expect(issue).toHaveProperty("repoName");
    expect(issue).toHaveProperty("streamName");
    expect(issue.repoId).toBe(repo.id);
    expect(issue.repoName).toBe(repo.name);
    expect(issue.streamName).toBe(stream.name);
    expect(issue.streamId).toBe(stream.id);
  });

  it("updates an issue", async () => {
    // create
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Old title",
        }),
      }
    );

    // update
    const updated = await api<Issue>(
      kernel.baseUrl,
      `/issue/${issue.id}`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          title: "New title",
          estimateMinutes: 45,
          notes: "Refined scope",
        }),
      }
    );

    expect(updated.id).toBe(issue.id);
    expect(updated.title).toBe("New title");
    expect(updated.estimateMinutes).toBe(45);
    expect(updated.notes).toBe("Refined scope");
  });

  it("changes issue status [should fail]", async () => {
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Status change test",
        }),
      }
    );
    try {

      await api<Issue>(
        kernel.baseUrl,
        `/issue/${issue.id}/status`,
        {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${kernel.token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            status: "done" as IssueStatus,
          }),
        }
      );
    } catch (error) {
      expect(error).toBeDefined();
    }

  });

  it("fails to update a non-existent issue", async () => {
    const nonExistentId = "non-existent-issue-id";

    try {
      await api<Issue>(
        kernel.baseUrl,
        `/issue/${nonExistentId}`,
        {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${kernel.token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            title: "Should fail",
          }),
        }
      );
    } catch (error) {
      expect(error).toBeDefined();
    }
  });

  // valid state transitionn
  it("changes issue status [valid transition]", async () => {
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Status change valid test",
        }),
      }
    );

    const updated = await api<Issue>(
      kernel.baseUrl,
      `/issue/${issue.id}/status`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          status: "active" as IssueStatus,
        }),
      }
    );

    expect(updated.id).toBe(issue.id);
    expect(updated.status).toBe("active");
  });

  it("deletes an issue", async () => {
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "To be deleted",
        }),
      }
    );

    const res = await api<{ ok: boolean }>(
      kernel.baseUrl,
      `/issue/${issue.id}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(res.ok).toBe(true);

    // Verify deletion
    const issues = await api<Issue[]>(
      kernel.baseUrl,
      `/issues?streamId=${stream.id}`,
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(issues.find(i => i.id === issue.id)).toBeUndefined();
  });

  it("marks an issue as todo for today", async () => {
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Todo for today test",
        }),
      }
    );

    const updated = await api<Issue>(
      kernel.baseUrl,
      `/issue/${issue.id}/todo`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(updated.id).toBe(issue.id);
    expect(updated.todoForDate).toBe(new Date().toISOString().split("T")[0]);
  });

  it("clears todo for date", async () => {
    const issue = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Clear todo for date test",
          todoForDate: new Date().toISOString().split("T")[0],
        }),
      }
    );

    const updated = await api<Issue>(
      kernel.baseUrl,
      `/issue/${issue.id}/todo/clear`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(updated.id).toBe(issue.id);
    expect(updated.todoForDate).toBeUndefined();
  });

  it("compute a daily issue summary", async () => {
    const today = new Date().toISOString().split("T")[0];

    // Create issues with todoForDate set to today
    const issue1 = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Daily summary issue 1",
          estimateMinutes: 30,
        }),
      }
    );

    const issue2 = await api<Issue>(
      kernel.baseUrl,
      "/issue",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          streamId: stream.id,
          title: "Daily summary issue 2",
          estimateMinutes: 45,
        }),
      }
    );

    // mark issues for today
    await api<Issue>(
      kernel.baseUrl,
      `/issue/${issue1.id}/todo`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    await api<Issue>(
      kernel.baseUrl,
      `/issue/${issue2.id}/todo`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    const summary = await api<{
      date: string;
      totalIssues: number;
      issues: Issue[];
      totalEstimatedMinutes: number;
    }>(
      kernel.baseUrl,
      `/issues/summary/today`,
      {
        headers: {
          Authorization: `Bearer ${kernel.token}`,
        },
      }
    );

    expect(summary.date).toBe(today);
    expect(summary.totalIssues).toBeGreaterThanOrEqual(2);
    expect(summary.issues.length).toBeGreaterThanOrEqual(2);
    expect(summary.totalEstimatedMinutes).toBeGreaterThanOrEqual(75);

  });

});
