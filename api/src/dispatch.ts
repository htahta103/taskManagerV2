import type { CorsConfig } from "./cors.js";
import { taskOptionsResponse } from "./cors.js";
import { notFoundJsonResponse } from "./response.js";

export type ApiFetchInner = (
  request: Request,
  url: URL,
) => Promise<Response | null> | Response | null;

/**
 * Shared fetch-style dispatch: task CORS preflight (OPTIONS), then inner handler,
 * then JSON 404 with CORS for unknown routes (per task endpoint consistency AC).
 */
export async function dispatchApiFetch(
  request: Request,
  config: CorsConfig,
  inner: ApiFetchInner,
): Promise<Response> {
  const url = new URL(request.url);
  const preflight = taskOptionsResponse(request, url.pathname, config);
  if (preflight) {
    return preflight;
  }
  const resolved = await inner(request, url);
  if (resolved) {
    return resolved;
  }
  return notFoundJsonResponse(request, config);
}
