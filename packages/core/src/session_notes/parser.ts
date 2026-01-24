import type { SessionSegment } from "../domain";
import type { ParsedSessionNotes, SessionNoteSection, SessionWorkSummary } from "./types";

export class SessionNotesParser {
  private static SECTION_PREFIX = "::";

  static parse(raw: string | null | undefined): ParsedSessionNotes {
    if (!raw) return {};

    const lines = raw.split("\n");
    const result: ParsedSessionNotes = {};

    let currentSection: SessionNoteSection | null = null;
    let buffer: string[] = [];

    const flush = () => {
      if (currentSection) {
        result[currentSection] = buffer.join("\n").trim();
      }
      buffer = [];
    };

    for (const line of lines) {
      if (line.startsWith(this.SECTION_PREFIX)) {
        flush();
        const section = line
          .slice(this.SECTION_PREFIX.length)
          .trim() as SessionNoteSection;

        currentSection = section;
        continue;
      }

      buffer.push(line);
    }

    flush();
    return result;
  }

  static serialize(sections: ParsedSessionNotes): string {
    const ordered: SessionNoteSection[] = [
      "commit",
      "context",
      "work",
      "notes",
    ];

    const blocks: string[] = [];

    for (const key of ordered) {
      const value = sections[key];
      if (!value) continue;

      blocks.push(`::${key}\n${value.trim()}`);
    }

    return blocks.join("\n\n");
  }

  static assertCommitMessage(
    notes: string | null | undefined
  ): void {
    const parsed = this.parse(notes);
    if (!parsed.commit) {
      throw new Error("Commit message is required in session notes");
    }
  }
  static generateDefaultSessionNotes(input: {
    commit?: string | undefined,
    repoId?: string | undefined,
    streamId?: string | undefined
    issueId?: string | undefined
    workSummary?: string[] | undefined
  }
  ): string {
    return this.serialize({
      commit: input.commit || "Work Session",
      context: [
        input.repoId ? `Repo ID: ${input.repoId}` : null,
        input.streamId ? `Stream ID: ${input.streamId}` : null,
        input.issueId ? `Issue ID: ${input.issueId}` : null,
      ]
        .filter(Boolean)
        .join("\n"),
      work: input.workSummary ? input.workSummary.join("\n") : "",
    })
  }

  static ammendCommitMessage(
    notes: string | null | undefined,
    additionalMessage: string
  ): string {
    const parsed = this.parse(notes);
    const existingCommit = parsed.commit || "";
    const newCommit = existingCommit
      ? `${additionalMessage}`.trim()
      : additionalMessage;

    parsed.commit = newCommit;
    return this.serialize(parsed);
  }

  static computeWorkSummary(
    segments: SessionSegment[]
  ): SessionWorkSummary {
    let workSeconds = 0;
    let restSeconds = 0;
    let workSegments = 0;
    let restSegments = 0;

    for (const segment of segments) {
      const duration =
        (Date.parse(segment.endTime?.toString() || "") - Date.parse(segment.startTime.toString())) / 1000;

      if (segment.segmentType === "work") {
        workSeconds += duration;
        workSegments += 1;
      } else {
        restSeconds += duration;
        restSegments += 1;
      }
    }

    return {
      workSeconds,
      restSeconds,
      workSegments,
      restSegments,
      totalSeconds: workSeconds + restSeconds,
    };
  }

  static formatWorkSummary(summary: SessionWorkSummary): string[] {

    return [
      `Work: ${this.formatDuration(summary.workSeconds)} (${summary.workSegments} segments)`,
      `Rest: ${this.formatDuration(summary.restSeconds)} (${summary.restSegments} segments)`,
      `Total: ${this.formatDuration(summary.totalSeconds)}`,
    ]
  }

  private static convertTimeInUnits(seconds: number, unit: 'auto' | 'seconds' | 'minutes' | 'hours' | 'days'): { value: number, unit: string } {
    switch (unit) {
      case 'seconds':
        return { value: seconds, unit: 'seconds' };
      case 'minutes':
        return { value: seconds / 60, unit: 'minutes' };
      case 'hours':
        return { value: seconds / 3600, unit: 'hours' };
      case 'days':
        return { value: seconds / 86400, unit: 'days' };
      case 'auto':
      default:
        if (seconds < 60) {
          return { value: seconds, unit: 'seconds' };
        } else if (seconds < 3600) {
          return { value: seconds / 60, unit: 'minutes' };
        } else if (seconds < 86400) {
          return { value: seconds / 3600, unit: 'hours' };
        } else {
          return { value: seconds / 86400, unit: 'days' };
        }
    }
  }

  private static formatDuration(
    seconds: number,
    unit: 'auto' | 'seconds' | 'minutes' | 'hours' | 'days' = 'auto',
    precision: number = 2
  ): string {
    const { value, unit: usedUnit } = this.convertTimeInUnits(seconds, unit);
    return `${value.toFixed(precision)}${this.timeUnitShortHandMap[usedUnit as keyof typeof this.timeUnitShortHandMap]}`;
  }

  private static timeUnitShortHandMap = {
    seconds: 's',
    minutes: 'm',
    hours: 'h',
    days: 'd',
  }
}
