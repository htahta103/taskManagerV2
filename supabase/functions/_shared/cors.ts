/**
 * CORS for Edge Functions: localhost:5173 (Vite), any *.pages.dev (Cloudflare Pages),
 * plus optional CORS_ALLOWED_ORIGINS (comma-separated).
 */

const DEFAULT_ORIGINS = ["http://localhost:5173"];

function extraOrigins(): string[] {
  const raw = Deno.env.get("CORS_ALLOWED_ORIGINS");
  if (!raw) return [];
  return raw.split(",").map((s) => s.trim()).filter(Boolean);
}

export function isCloudflarePagesOrigin(origin: string): boolean {
  return /^https:\/\/[a-z0-9-]+\.pages\.dev$/i.test(origin);
}

export function resolveCorsOrigin(req: Request): string | null {
  const origin = req.headers.get("Origin");
  if (!origin) return null;
  if (DEFAULT_ORIGINS.includes(origin)) return origin;
  if (isCloudflarePagesOrigin(origin)) return origin;
  if (extraOrigins().includes(origin)) return origin;
  return null;
}

/** Headers to merge on JSON responses and preflight. */
export function corsHeaders(req: Request): Record<string, string> {
  const allow = resolveCorsOrigin(req);
  const headers: Record<string, string> = {
    "Access-Control-Allow-Headers":
      "authorization, x-client-info, apikey, content-type",
    "Access-Control-Allow-Methods": "GET, POST, PATCH, DELETE, OPTIONS",
    "Access-Control-Max-Age": "86400",
  };
  if (allow) {
    headers["Access-Control-Allow-Origin"] = allow;
    headers["Vary"] = "Origin";
  }
  return headers;
}
