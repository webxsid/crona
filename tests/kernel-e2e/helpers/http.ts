import fetch from "node-fetch";
import type { RequestInit } from "node-fetch";

export async function api<R>(
  baseUrl: string,
  path: string,
  options: RequestInit = {}
): Promise<R> {
  let appendContentType = false;
  if (options.body && typeof options.body === "object" && !(options.body instanceof Buffer)) {
    options.body = JSON.stringify(options.body);
    appendContentType = true;
  }
  const res = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers: {
      ...(appendContentType ? { "Content-Type": "application/json" } : {}),
      ...(options.headers || {})
    }
  });

  if (!res.ok) {
    const text = await res.text();
    console.error("API request failed:", { path, status: res.status, body: text });
    throw new Error(`${res.status}: ${text}`);
  }

  return res.json() as R;
}
