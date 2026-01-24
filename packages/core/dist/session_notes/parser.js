"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SessionNotesParser = void 0;
class SessionNotesParser {
    static SECTION_PREFIX = "::";
    static parse(raw) {
        if (!raw)
            return {};
        const lines = raw.split("\n");
        const result = {};
        let currentSection = null;
        let buffer = [];
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
                    .trim();
                currentSection = section;
                continue;
            }
            buffer.push(line);
        }
        flush();
        return result;
    }
    static serialize(sections) {
        const ordered = [
            "commit",
            "context",
            "work",
            "notes",
        ];
        const blocks = [];
        for (const key of ordered) {
            const value = sections[key];
            if (!value)
                continue;
            blocks.push(`::${key}\n${value.trim()}`);
        }
        return blocks.join("\n\n");
    }
    static assertCommitMessage(notes) {
        const parsed = this.parse(notes);
        if (!parsed.commit) {
            throw new Error("Commit message is required in session notes");
        }
    }
    static generateDefaultSessionNotes(input) {
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
        });
    }
    static ammendCommitMessage(notes, additionalMessage) {
        const parsed = this.parse(notes);
        const existingCommit = parsed.commit || "";
        const newCommit = existingCommit
            ? `${additionalMessage}`.trim()
            : additionalMessage;
        parsed.commit = newCommit;
        return this.serialize(parsed);
    }
    static computeWorkSummary(segments) {
        let workSeconds = 0;
        let restSeconds = 0;
        let workSegments = 0;
        let restSegments = 0;
        for (const segment of segments) {
            const duration = (Date.parse(segment.endTime?.toString() || "") - Date.parse(segment.startTime.toString())) / 1000;
            if (segment.segmentType === "work") {
                workSeconds += duration;
                workSegments += 1;
            }
            else {
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
    static formatWorkSummary(summary) {
        return [
            `Work: ${this.formatDuration(summary.workSeconds)} (${summary.workSegments} segments)`,
            `Rest: ${this.formatDuration(summary.restSeconds)} (${summary.restSegments} segments)`,
            `Total: ${this.formatDuration(summary.totalSeconds)}`,
        ];
    }
    static convertTimeInUnits(seconds, unit) {
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
                }
                else if (seconds < 3600) {
                    return { value: seconds / 60, unit: 'minutes' };
                }
                else if (seconds < 86400) {
                    return { value: seconds / 3600, unit: 'hours' };
                }
                else {
                    return { value: seconds / 86400, unit: 'days' };
                }
        }
    }
    static formatDuration(seconds, unit = 'auto', precision = 2) {
        const { value, unit: usedUnit } = this.convertTimeInUnits(seconds, unit);
        return `${value.toFixed(precision)}${this.timeUnitShortHandMap[usedUnit]}`;
    }
    static timeUnitShortHandMap = {
        seconds: 's',
        minutes: 'm',
        hours: 'h',
        days: 'd',
    };
}
exports.SessionNotesParser = SessionNotesParser;
