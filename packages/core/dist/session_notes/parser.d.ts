import type { SessionSegment } from "../domain";
import type { ParsedSessionNotes, SessionWorkSummary } from "./types";
export declare class SessionNotesParser {
    private static SECTION_PREFIX;
    static parse(raw: string | null | undefined): ParsedSessionNotes;
    static serialize(sections: ParsedSessionNotes): string;
    static assertCommitMessage(notes: string | null | undefined): void;
    static generateDefaultSessionNotes(input: {
        commit?: string | undefined;
        repoId?: string | undefined;
        streamId?: string | undefined;
        issueId?: string | undefined;
        workSummary?: string[] | undefined;
    }): string;
    static ammendCommitMessage(notes: string | null | undefined, additionalMessage: string): string;
    static computeWorkSummary(segments: SessionSegment[]): SessionWorkSummary;
    static formatWorkSummary(summary: SessionWorkSummary): string[];
    private static convertTimeInUnits;
    private static formatDuration;
    private static timeUnitShortHandMap;
}
