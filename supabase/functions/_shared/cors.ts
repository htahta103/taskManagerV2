/**
 * Shared CORS headers for Edge Functions so all routes behave consistently
 * for browser clients and preflight checks.
 */
export const corsHeaders: Record<string, string> = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers":
    "authorization, x-client-info, apikey, content-type",
};
