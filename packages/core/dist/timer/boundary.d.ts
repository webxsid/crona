import type { SessionSegmentType } from "../domain";
import type { CoreSettingsTable } from "../storage";
export type BoundaryResult = {
    nextSegment: SessionSegmentType;
    afterMinutes: number;
} | null;
/**
 * Decide the next boundary from current segment + settings
 */
export declare function computeNextBoundary(current: SessionSegmentType, settings: CoreSettingsTable, completedWorkCycles: number): BoundaryResult;
