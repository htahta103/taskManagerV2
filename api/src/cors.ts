/** JSON responses use this content type everywhere task handlers respond with JSON. */
export const JSON_CONTENT_TYPE = "application/json; charset=utf-8";

const DEFAULT_ALLOW_METHODS = "GET, POST, PATCH, DELETE, OPTIONS";
const DEFAULT_ALLOW_HEADERS = "Authorization, Content-Type";

export type CorsConfig = {
  /** Origins allowed for `Access-Control-Allow-Origin`. Use env-specific list in deploy. */
  allowedOrigins: string[];
  allowCredentials?: boolean;
};

export function taskApiPathPrefix(): string {
  return "/api/v1/tasks";
}

/** True for `/api/v1/tasks` and `/api/v1/tasks/...` (including tag sub-routes). */
export function isTaskEndpointPath(pathname: string): boolean {
  const base = taskApiPathPrefix();
  return pathname === base || pathname.startsWith(`${base}/`);
}

export function resolveAllowOrigin(request: Request, config: CorsConfig): string | null {
  const origin = request.headers.get("Origin");
  if (!origin) {
    if (config.allowedOrigins.includes("*")) {
      return config.allowCredentials ? null : "*";
    }
    return null;
  }
  if (config.allowedOrigins.includes("*")) {
    return config.allowCredentials ? null : "*";
  }
  return config.allowedOrigins.includes(origin) ? origin : null;
}

/** CORS headers for a single response (merge into your `Response` headers). */
export function corsHeadersForRequest(request: Request, config: CorsConfig): Headers {
  const headers = new Headers();
  const allowOrigin = resolveAllowOrigin(request, config);
  if (allowOrigin) {
    headers.set("Access-Control-Allow-Origin", allowOrigin);
  }
  if (config.allowCredentials) {
    headers.set("Access-Control-Allow-Credentials", "true");
  }
  headers.set("Access-Control-Allow-Methods", DEFAULT_ALLOW_METHODS);
  headers.set("Access-Control-Allow-Headers", DEFAULT_ALLOW_HEADERS);
  headers.set("Access-Control-Max-Age", "86400");
  return headers;
}

/**
 * For task routes: respond to CORS preflight with 204 and full CORS headers.
 * Returns `null` when the request is not an OPTIONS preflight for a task path.
 */
export function taskOptionsResponse(
  request: Request,
  pathname: string,
  config: CorsConfig,
): Response | null {
  if (request.method !== "OPTIONS" || !isTaskEndpointPath(pathname)) {
    return null;
  }
  return new Response(null, {
    status: 204,
    headers: corsHeadersForRequest(request, config),
  });
}
