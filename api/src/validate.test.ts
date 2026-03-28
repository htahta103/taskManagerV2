import { describe, expect, it } from "vitest";
import {
  parseTaskIdFromPath,
  parseTaskListQuery,
  validateTaskCreate,
  validateTaskPatch,
} from "./validate.js";

describe("validateTaskCreate", () => {
  it("accepts minimal create", () => {
    const r = validateTaskCreate({ title: "Hello" });
    expect(r.ok).toBe(true);
    if (r.ok) {
      expect(r.value.title).toBe("Hello");
    }
  });

  it("rejects empty title", () => {
    const r = validateTaskCreate({ title: "   " });
    expect(r.ok).toBe(false);
  });

  it("rejects title over max", () => {
    const r = validateTaskCreate({ title: "x".repeat(201) });
    expect(r.ok).toBe(false);
  });

  it("rejects invalid status", () => {
    const r = validateTaskCreate({ title: "a", status: "nope" });
    expect(r.ok).toBe(false);
  });
});

describe("validateTaskPatch", () => {
  it("requires at least one field", () => {
    const r = validateTaskPatch({});
    expect(r.ok).toBe(false);
  });

  it("accepts partial patch", () => {
    const r = validateTaskPatch({ status: "done" });
    expect(r.ok).toBe(true);
  });
});

describe("parseTaskListQuery", () => {
  it("applies default limit", () => {
    const r = parseTaskListQuery(new URLSearchParams());
    expect(r.ok).toBe(true);
    if (r.ok) {
      expect(r.value.limit).toBe(50);
    }
  });

  it("rejects bad limit", () => {
    const r = parseTaskListQuery(new URLSearchParams("limit=0"));
    expect(r.ok).toBe(false);
  });
});

describe("parseTaskIdFromPath", () => {
  it("parses task id", () => {
    const id = "550e8400-e29b-41d4-a716-446655440000";
    const r = parseTaskIdFromPath(`/api/v1/tasks/${id}`);
    expect(r.ok).toBe(true);
    if (r.ok) {
      expect(r.value).toBe(id);
    }
  });

  it("rejects invalid segment", () => {
    const r = parseTaskIdFromPath("/api/v1/tasks/not-a-uuid");
    expect(r.ok).toBe(false);
  });
});
