import type { ScratchPadMeta } from "../domain";
import type { ICommandContext } from "./context";
export declare function registerScratchpad(ctx: ICommandContext, meta: ScratchPadMeta): Promise<void>;
export declare function listScratchpads(ctx: ICommandContext, options: {
    pinnedOnly?: boolean;
}): Promise<ScratchPadMeta[]>;
export declare function pinScratchpad(ctx: ICommandContext, path: string, pinned: boolean): Promise<void>;
export declare function removeScratchpad(ctx: ICommandContext, path: string): Promise<void>;
