import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import { escapeILikeLiteral, parseTaskListQuery } from "./query.ts";

Deno.test("parseTaskListQuery accepts valid filters", () => {
  const u = new URL("https://x/tasks?status=todo&priority=high&search=foo");
  const r = parseTaskListQuery(u);
  assertEquals(r, {
    ok: true,
    value: { status: "todo", priority: "high", search: "foo" },
  });
});

Deno.test("parseTaskListQuery rejects bad status", () => {
  const u = new URL("https://x/tasks?status=nope");
  const r = parseTaskListQuery(u);
  assertEquals(r.ok, false);
});

Deno.test("parseTaskListQuery rejects bad priority", () => {
  const u = new URL("https://x/tasks?priority=urgent");
  const r = parseTaskListQuery(u);
  assertEquals(r.ok, false);
});

Deno.test("parseTaskListQuery trims search", () => {
  const u = new URL("https://x/tasks?search=%20%20bar%20");
  const r = parseTaskListQuery(u);
  assertEquals(r, { ok: true, value: { search: "bar" } });
});

Deno.test("escapeILikeLiteral escapes wildcards", () => {
  assertEquals(escapeILikeLiteral("100%_done"), "100\\%\\_done");
});
