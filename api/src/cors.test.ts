import { describe, expect, it } from "vitest";
import {
  corsHeadersForRequest,
  isTaskEndpointPath,
  JSON_CONTENT_TYPE,
  resolveAllowOrigin,
  taskOptionsResponse,
} from "./cors.js";

const config = { allowedOrigins: ["http://localhost:5173"] };

describe("cors", () => {
  it("identifies task paths", () => {
    expect(isTaskEndpointPath("/api/v1/tasks")).toBe(true);
    expect(isTaskEndpointPath("/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000")).toBe(true);
    expect(isTaskEndpointPath("/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000/tags")).toBe(true);
    expect(isTaskEndpointPath("/api/v1/projects")).toBe(false);
  });

  it("returns OPTIONS 204 for task paths with CORS headers", () => {
    const req = new Request("http://localhost/api/v1/tasks", {
      method: "OPTIONS",
      headers: { Origin: "http://localhost:5173" },
    });
    const res = taskOptionsResponse(req, new URL(req.url).pathname, config);
    expect(res).not.toBeNull();
    if (res === null) {
      return;
    }
    expect(res.status).toBe(204);
    expect(res.headers.get("Access-Control-Allow-Origin")).toBe("http://localhost:5173");
    expect(res.headers.get("Access-Control-Allow-Methods")).toContain("OPTIONS");
  });

  it("returns null for OPTIONS on non-task paths", () => {
    const req = new Request("http://localhost/api/v1/projects", { method: "OPTIONS" });
    expect(taskOptionsResponse(req, new URL(req.url).pathname, config)).toBeNull();
  });

  it("sets JSON content type constant", () => {
    expect(JSON_CONTENT_TYPE).toContain("application/json");
  });

  it("resolveAllowOrigin rejects unknown origin", () => {
    const req = new Request("http://x", { headers: { Origin: "https://evil.test" } });
    expect(resolveAllowOrigin(req, config)).toBeNull();
  });

  it("merges cors headers for responses", () => {
    const req = new Request("http://x", { headers: { Origin: "http://localhost:5173" } });
    const h = corsHeadersForRequest(req, config);
    expect(h.get("Access-Control-Allow-Origin")).toBe("http://localhost:5173");
  });
});
