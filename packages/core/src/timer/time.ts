/**
 * Compute elapsed seconds between two ISO timestamps
 */
export function elapsedSeconds(
  startTime: string,
  now: string
): number {
  return Math.max(
    0,
    Math.floor(
      (Date.parse(now) - Date.parse(startTime)) / 1000
    )
  );
}
