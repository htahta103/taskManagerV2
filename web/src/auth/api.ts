import type { User } from "./types";
import { readStoredToken, writeStoredToken } from "./token";

const API_PREFIX = "/api/v1";

export class ApiError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(message: string, status: number, body: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

function apiBase(): string {
  return import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") ?? "";
}

function mockAuthEnabled(): boolean {
  return import.meta.env.VITE_MOCK_AUTH === "true";
}

type Json = Record<string, unknown>;

async function parseJson(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function messageFromBody(body: unknown, fallback: string): string {
  if (body && typeof body === "object" && "error" in body) {
    const err = (body as Json).error;
    if (typeof err === "string") return err;
  }
  if (body && typeof body === "object" && "message" in body) {
    const msg = (body as Json).message;
    if (typeof msg === "string") return msg;
  }
  return fallback;
}

export async function apiFetch(
  path: string,
  init: RequestInit & { json?: unknown } = {},
): Promise<Response> {
  const url = `${apiBase()}${API_PREFIX}${path}`;
  const headers = new Headers(init.headers);
  if (!headers.has("Content-Type") && init.json !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  const token = readStoredToken();
  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const { json, ...rest } = init;
  const body = json !== undefined ? JSON.stringify(json) : rest.body;
  return fetch(url, {
    ...rest,
    headers,
    body,
    credentials: "include",
  });
}

function persistAccessTokenFromAuthBody(body: unknown): void {
  if (!body || typeof body !== "object") return;
  const o = body as Json;
  const access =
    typeof o.access_token === "string"
      ? o.access_token
      : typeof o.token === "string"
        ? o.token
        : null;
  if (access) writeStoredToken(access);
}

function userFromJson(raw: unknown): User | null {
  if (!raw || typeof raw !== "object") return null;
  const o = raw as Json;
  const id = o.id;
  const email = o.email;
  const name = o.name;
  if (typeof id !== "string" || typeof email !== "string") return null;
  return {
    id,
    email,
    name: typeof name === "string" ? name : "",
  };
}

export async function fetchMe(): Promise<User | null> {
  if (mockAuthEnabled()) {
    const token = readStoredToken();
    if (!token) return null;
    return {
      id: "mock-user",
      email: "you@example.com",
      name: "Demo user",
    };
  }
  const res = await apiFetch("/me", { method: "GET" });
  if (res.status === 401) {
    writeStoredToken(null);
    return null;
  }
  if (!res.ok) {
    const body = await parseJson(res);
    throw new ApiError(messageFromBody(body, "Could not load profile"), res.status, body);
  }
  const body = await parseJson(res);
  return userFromJson(body);
}

export async function login(email: string, password: string): Promise<User> {
  if (mockAuthEnabled()) {
    writeStoredToken("mock-token");
    return { id: "mock-user", email, name: email.split("@")[0] ?? "User" };
  }
  const res = await apiFetch("/auth/login", {
    method: "POST",
    json: { email, password },
  });
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Sign in failed"), res.status, body);
  }
  persistAccessTokenFromAuthBody(body);
  const user = userFromJson(body && typeof body === "object" ? (body as Json).user : null);
  if (!user) throw new ApiError("Invalid response from server", res.status, body);
  return user;
}

export async function signup(input: {
  email: string;
  password: string;
  name: string;
}): Promise<User> {
  if (mockAuthEnabled()) {
    writeStoredToken("mock-token");
    return { id: "mock-user", email: input.email, name: input.name || "User" };
  }
  const res = await apiFetch("/auth/register", {
    method: "POST",
    json: input,
  });
  const body = await parseJson(res);
  if (!res.ok) {
    throw new ApiError(messageFromBody(body, "Could not create account"), res.status, body);
  }
  persistAccessTokenFromAuthBody(body);
  const user = userFromJson(body && typeof body === "object" ? (body as Json).user : null);
  if (!user) throw new ApiError("Invalid response from server", res.status, body);
  return user;
}

export async function logout(): Promise<void> {
  if (mockAuthEnabled()) {
    writeStoredToken(null);
    return;
  }
  try {
    await apiFetch("/auth/logout", { method: "POST" });
  } finally {
    writeStoredToken(null);
  }
}
