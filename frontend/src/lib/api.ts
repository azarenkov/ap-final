

const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

let cachedToken: string | null = null;

export function setAuthToken(token: string | null) {
    cachedToken = token;
}

export function getAuthToken(): string | null {
    if (cachedToken) return cachedToken;
    if (typeof window !== "undefined") {
        const t = window.localStorage.getItem("ap2.token");
        if (t) cachedToken = t;
    }
    return cachedToken;
}

interface Options extends Omit<RequestInit, "body"> {
    body?: unknown;
    auth?: boolean;
}

export class ApiError extends Error {
    status: number;
    constructor(message: string, status: number) {
        super(message);
        this.status = status;
    }
}

export async function api<T>(path: string, opts: Options = {}): Promise<T> {
    const { body, auth = true, headers, ...rest } = opts;
    const h = new Headers(headers);
    if (body !== undefined) h.set("Content-Type", "application/json");
    if (auth) {
        const t = getAuthToken();
        if (t) h.set("Authorization", `Bearer ${t}`);
    }
    const res = await fetch(`${API_BASE}${path}`, {
        ...rest,
        headers: h,
        body: body !== undefined ? JSON.stringify(body) : undefined,
    });
    const text = await res.text();
    const data = text ? (() => { try { return JSON.parse(text); } catch { return text; } })() : null;
    if (!res.ok) {
        const msg = (data && typeof data === "object" && "error" in data) ? (data as any).error : res.statusText;
        throw new ApiError(msg ?? "Request failed", res.status);
    }
    return data as T;
}
