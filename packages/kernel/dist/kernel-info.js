"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.writeKernelInfo = writeKernelInfo;
exports.readKernelInfo = readKernelInfo;
exports.clearKernelInfo = clearKernelInfo;
const promises_1 = __importDefault(require("fs/promises"));
const path_1 = __importDefault(require("path"));
const os_1 = __importDefault(require("os"));
const CRONA_DIR = path_1.default.join(os_1.default.homedir(), ".crona");
const KERNEL_INFO_FILE = path_1.default.join(CRONA_DIR, "kernel.json");
/**
 * Ensure ~/.crona exists with safe permissions
 */
async function ensureDir() {
    await promises_1.default.mkdir(CRONA_DIR, { recursive: true, mode: 0o700 });
}
/**
 * Write kernel connection info
 * Overwrites existing file atomically
 */
async function writeKernelInfo(info) {
    await ensureDir();
    const tempFile = `${KERNEL_INFO_FILE}.tmp`;
    await promises_1.default.writeFile(tempFile, JSON.stringify(info, null, 2), { mode: 0o600 });
    await promises_1.default.rename(tempFile, KERNEL_INFO_FILE);
}
/**
 * Read kernel info if present
 */
async function readKernelInfo() {
    try {
        const raw = await promises_1.default.readFile(KERNEL_INFO_FILE, "utf8");
        return JSON.parse(raw);
    }
    catch {
        return null;
    }
}
/**
 * Remove kernel info (on shutdown / crash recovery)
 */
async function clearKernelInfo() {
    try {
        await promises_1.default.unlink(KERNEL_INFO_FILE);
    }
    catch (err) {
        // ignore
        console.error("Failed to clear kernel info:", err);
    }
}
