import { describe, expect, it } from "vitest";
import { dispatchApiFetch } from "./dispatch.js";
import { jsonResponse } from "./response.js";

const config = { allowedOrigins: ["http://localhost:5173"] };

describe("dispatchApiFetch", () => {
  it("returns JSON 404 for unknown routes", async () => {
    const req = new Request("http://localhost/api/v1/unknown", {
      headers: { Origin: "http://localhost:5173" },
    });
    const res = await dispatchApiFetch(req, config, () => null);
    expect(res.status).toBe(404);
    expect(res.headers.get("Content-Type")).toContain("application/json");
    const body = (await res.json()) as { error: string };
    expect(body.error).toBe("Not Found");
    expect(res.headers.get("Access-Control-Allow-Origin")).toBe("http://localhost:5173");
  });

  it("delegates when inner returns response", async () => {
    const req = new Request("http://localhost/api/v1/tasks", {
      headers: { Origin: "http://localhost:5173" },
    });
    const res = await dispatchApiFetch(req, config, () =>
      jsonResponse(req, config, 200, { items: [], next_cursor: null }),
    );
    expect(res.status).toBe(200);
    expect(res.headers.get("Content-Type")).toContain("application/json");
  });
});
