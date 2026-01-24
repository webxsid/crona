export interface KernelInfo {
    port: number;
    token: string;
    pid: number;
    startedAt: string;
}
/**
 * Write kernel connection info
 * Overwrites existing file atomically
 */
export declare function writeKernelInfo(info: KernelInfo): Promise<void>;
/**
 * Read kernel info if present
 */
export declare function readKernelInfo(): Promise<KernelInfo | null>;
/**
 * Remove kernel info (on shutdown / crash recovery)
 */
export declare function clearKernelInfo(): Promise<void>;
