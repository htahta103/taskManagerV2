import { createClient } from "https://esm.sh/@supabase/supabase-js@2.49.1";
import { corsHeaders } from "../_shared/cors.ts";
import { escapeILikeLiteral, parseTaskListQuery } from "../_shared/query.ts";

const jsonHeaders: Record<string, string> = {
  ...corsHeaders,
  "Content-Type": "application/json",
};

function jsonResponse(body: unknown, status: number) {
  return new Response(JSON.stringify(body), {
    status,
    headers: jsonHeaders,
  });
}

function isAuthRelatedError(message: string): boolean {
  const m = message.toLowerCase();
  return m.includes("jwt") ||
    (m.includes("expired") && m.includes("token")) ||
    (m.includes("invalid") && m.includes("token"));
}

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
    return jsonResponse({ error: "Method not allowed" }, 405);
  }

  const authHeader = req.headers.get("Authorization");
  if (!authHeader?.startsWith("Bearer ")) {
    return jsonResponse(
      { error: "Missing or invalid Authorization header" },
      401,
    );
  }

  const url = new URL(req.url);
  const parsed = parseTaskListQuery(url);
  if (!parsed.ok) {
    return jsonResponse(
      { error: parsed.error, code: "validation_error" },
      422,
    );
  }

  const supabaseUrl = Deno.env.get("SUPABASE_URL") ?? "";
  const supabaseAnonKey = Deno.env.get("SUPABASE_ANON_KEY") ?? "";
  const supabase = createClient(supabaseUrl, supabaseAnonKey, {
    global: { headers: { Authorization: authHeader } },
  });

  let q = supabase
    .from("tasks")
    .select("*", { count: "exact" })
    .order("created_at", { ascending: false });

  const { status, priority, search } = parsed.value;
  if (status) q = q.eq("status", status);
  if (priority) q = q.eq("priority", priority);
  if (search) {
    const pattern = `%${escapeILikeLiteral(search)}%`;
    q = q.ilike("title", pattern);
  }

  const { data, error, count } = await q;

  if (error) {
    if (isAuthRelatedError(error.message)) {
      return jsonResponse({ error: error.message }, 401);
    }
    return jsonResponse({ error: error.message }, 500);
  }

  return jsonResponse({ data: data ?? [], count: count ?? 0 }, 200);
});
