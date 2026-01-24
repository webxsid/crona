"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.elapsedSeconds = elapsedSeconds;
/**
 * Compute elapsed seconds between two ISO timestamps
 */
function elapsedSeconds(startTime, now) {
    return Math.max(0, Math.floor((Date.parse(now) - Date.parse(startTime)) / 1000));
}
