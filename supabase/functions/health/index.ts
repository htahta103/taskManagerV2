import { corsHeaders } from "../_shared/cors.ts";

const jsonHeaders: Record<string, string> = {
  ...corsHeaders,
  "Content-Type": "application/json",
};

Deno.serve(async (req: Request) => {
  if (req.method === "OPTIONS") {
    return new Response(null, {
      status: 204,
      headers: {
        ...corsHeaders,
        "Access-Control-Allow-Methods": "GET, OPTIONS",
        "Access-Control-Max-Age": "86400",
      },
    });
  }

  if (req.method !== "GET") {
    return new Response(
      JSON.stringify({ error: "Method not allowed" }),
      { status: 405, headers: jsonHeaders },
    );
  }

  const body = JSON.stringify({
    status: "ok",
    timestamp: new Date().toISOString(),
  });

  return new Response(body, { status: 200, headers: jsonHeaders });
});
