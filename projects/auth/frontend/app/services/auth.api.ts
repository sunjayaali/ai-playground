import type { AuthError, UserInfo } from "../types/auth";

// Builtin $fetch (Nuxt global), injected for testability.
// Minimal structural type — no ofetch import.
export type FetchFn = <T = unknown>(
  request: string,
  opts?: Record<string, unknown>,
) => Promise<T>;

interface FetchLikeError {
  status?: number;
  statusCode?: number;
  data?: AuthError | undefined;
  message?: string;
}

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
  }
}

export interface AuthApi {
  login(username: string, password: string): Promise<void>;
  register(username: string, password: string): Promise<void>;
  getMe(): Promise<UserInfo | null>;
  refresh(): Promise<boolean>;
  logout(): Promise<void>;
}

function normalizeBaseUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, "");
}

// Builtin $fetch rejects on non-2xx; duck-type the error shape
// (status + optional body) back to the ApiError the UI catches.
function toApiError(error: unknown): ApiError | null {
  if (typeof error !== "object" || error === null) {
    return null;
  }
  const e = error as FetchLikeError;
  const status = e.status ?? e.statusCode ?? 0;
  if (!status) {
    return null;
  }
  return new ApiError(
    status,
    e.data?.error || e.message || `request failed with status ${status}`,
  );
}

export class AuthApiClient implements AuthApi {
  private refreshing: Promise<boolean> | null = null;

  constructor(
    readonly base: string,
    private fetcher: FetchFn = $fetch,
  ) {
    this.base = normalizeBaseUrl(base);
  }

  async refresh(): Promise<boolean> {
    try {
      await this.fetcher(`${this.base}/auth/refresh`, {
        method: "POST",
        credentials: "include",
      });
      return true;
    } catch (error) {
      if (toApiError(error)) {
        return false;
      }
      throw error;
    } finally {
      this.refreshing = null;
    }
  }

  async login(username: string, password: string): Promise<void> {
    await this.postJson("/auth/login", { username, password });
  }

  async register(username: string, password: string): Promise<void> {
    await this.postJson("/auth/register", { username, password });
  }

  async getMe(): Promise<UserInfo | null> {
    try {
      return await this.authedFetch<UserInfo>(`${this.base}/auth/me`);
    } catch (error) {
      if (toApiError(error)?.status === 401) {
        return null;
      }
      throw error;
    }
  }

  async logout(): Promise<void> {
    // Best effort: the client session ends even if the request fails.
    try {
      await this.authedFetch(`${this.base}/auth/logout`, { method: "POST" });
    } catch {
      // ignore
    }
  }

  private async ensureRefreshed(): Promise<boolean> {
    this.refreshing ??= this.refresh();
    return this.refreshing;
  }

  private async authedFetch<T = unknown>(
    url: string,
    init?: Record<string, unknown>,
  ): Promise<T> {
    try {
      return await this.fetcher<T>(url, {
        ...init,
        credentials: "include",
      });
    } catch (error) {
      if (toApiError(error)?.status !== 401) {
        throw error;
      }
      if (!(await this.ensureRefreshed())) {
        throw error;
      }
      return this.fetcher<T>(url, { ...init, credentials: "include" });
    }
  }

  private async postJson(path: string, body: unknown): Promise<void> {
    try {
      await this.fetcher(`${this.base}${path}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body,
        credentials: "include",
      });
    } catch (error) {
      const apiError = toApiError(error);
      if (apiError) {
        throw apiError;
      }
      throw error;
    }
  }
}

export function createAuthApi(baseUrl: string, fetchFn?: FetchFn): AuthApi {
  return new AuthApiClient(baseUrl, fetchFn);
}
