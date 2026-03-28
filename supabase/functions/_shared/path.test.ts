import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import {
  extractTasksSuffix,
  parseTasksPathSuffix,
} from "./path.ts";

Deno.test("extractTasksSuffix from Supabase URL", () => {
  assertEquals(extractTasksSuffix("/functions/v1/tasks"), "/");
  assertEquals(extractTasksSuffix("/functions/v1/tasks/"), "/");
  assertEquals(
    extractTasksSuffix("/functions/v1/tasks/550e8400-e29b-41d4-a716-446655440000"),
    "/550e8400-e29b-41d4-a716-446655440000",
  );
  assertEquals(
    extractTasksSuffix("/functions/v1/tasks/clear/done"),
    "/clear/done",
  );
});

Deno.test("parseTasksPathSuffix collection", () => {
  assertEquals(parseTasksPathSuffix("/"), { kind: "collection" });
});

Deno.test("parseTasksPathSuffix by id", () => {
  assertEquals(
    parseTasksPathSuffix("/550e8400-e29b-41d4-a716-446655440000"),
    { kind: "by_id", id: "550e8400-e29b-41d4-a716-446655440000" },
  );
});

Deno.test("parseTasksPathSuffix clear done", () => {
  assertEquals(parseTasksPathSuffix("/clear/done"), { kind: "clear_done" });
});

Deno.test("parseTasksPathSuffix not found", () => {
  assertEquals(parseTasksPathSuffix("/nope"), { kind: "not_found" });
});
