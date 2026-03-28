import type { CorsConfig } from "./cors.js";
import { JSON_CONTENT_TYPE, corsHeadersForRequest } from "./cors.js";

export type ApiErrorBody = {
  error: string;
  code?: string;
  details?: Record<string, unknown>;
};

function baseHeaders(request: Request, config: CorsConfig): Headers {
  const h = corsHeadersForRequest(request, config);
  h.set("Content-Type", JSON_CONTENT_TYPE);
  return h;
}

export function jsonResponse(
  request: Request,
  config: CorsConfig,
  status: number,
  body: unknown,
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: baseHeaders(request, config),
  });
}

export function jsonErrorResponse(
  request: Request,
  config: CorsConfig,
  status: number,
  body: ApiErrorBody,
): Response {
  return jsonResponse(request, config, status, body);
}

export function notFoundJsonResponse(request: Request, config: CorsConfig): Response {
  return jsonErrorResponse(request, config, 404, {
    error: "Not Found",
    code: "not_found",
  });
}

export function noContentResponse(request: Request, config: CorsConfig): Response {
  const headers = corsHeadersForRequest(request, config);
  return new Response(null, { status: 204, headers });
}
