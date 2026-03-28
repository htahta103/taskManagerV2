import { corsHeaders } from "../_shared/cors.ts";

const jsonContent = "application/json";

Deno.serve(async (req: Request) => {
  if (req.method === "OPTIONS") {
    return new Response(null, { status: 204, headers: corsHeaders(req) });
  }

  if (req.method !== "GET") {
    return new Response(
      JSON.stringify({ error: "Method not allowed" }),
      {
        status: 405,
        headers: {
          ...corsHeaders(req),
          "Content-Type": jsonContent,
        },
      },
    );
  }

  const body = JSON.stringify({
    status: "ok",
    timestamp: new Date().toISOString(),
  });

  return new Response(body, {
    status: 200,
    headers: {
      ...corsHeaders(req),
      "Content-Type": jsonContent,
    },
  });
});
