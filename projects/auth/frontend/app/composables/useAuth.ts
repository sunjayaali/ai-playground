type User = {
  id: string;
  username: string;
};

type AuthError = {
  error: string;
};

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
  }
}

let refreshing: Promise<boolean> | null = null;

export default () => {
  const config = useRuntimeConfig();
  const apiBase = config.public.apiBase;
  const user = useState<User | null>("auth:user", () => null);
  const isAuthenticated = computed(() => !!user.value);

  async function refresh(): Promise<boolean> {
    refreshing ??= refreshTokenPair();
    const ok = await refreshing;
    if (!ok) {
      user.value = null;
    }
    return ok;
  }

  async function refreshTokenPair(): Promise<boolean> {
    try {
      const res = await fetch(`${apiBase}/auth/refresh`, {
        method: "POST",
        credentials: "include",
      });
      return res.ok;
    } catch {
      return false;
    } finally {
      refreshing = null;
    }
  }

  async function authenticatedFetch(
    input: string,
    init?: RequestInit,
  ): Promise<Response> {
    const res = await fetch(input, { ...init, credentials: "include" });
    if (res.status !== 401) {
      return res;
    }

    if (!(await refresh())) {
      return res;
    }

    return fetch(input, { ...init, credentials: "include" });
  }

  async function login(username: string, password: string) {
    // Backend answers with an oauth2 token pair; cookies are
    // set, but the current user only comes from /auth/me.
    const res = await fetch(`${apiBase}/auth/login`, {
      method: "POST",
      body: JSON.stringify({
        username: username,
        password: password,
      }),
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    });
    if (!res.ok) {
      const errorData = (await res.json()) as AuthError;
      throw new ApiError(res.status, errorData.error);
    }

    const me = await fetchUser();
    if (!me) {
      throw new ApiError(502, "could not load user after login");
    }
    return me;
  }

  async function register(username: string, password: string) {
    const res = await fetch(`${apiBase}/auth/register`, {
      method: "POST",
      body: JSON.stringify({
        username: username,
        password: password,
      }),
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    });
    if (!res.ok) {
      const errorData = (await res.json()) as AuthError;
      throw new ApiError(res.status, errorData.error);
    }

    return res;
  }

  async function fetchUser(): Promise<User | null> {
    const res = await authenticatedFetch(`${apiBase}/auth/me`);
    if (!res.ok) {
      user.value = null;
      return null;
    }

    user.value = (await res.json()) as User;
    return user.value;
  }

  async function signOut(): Promise<void> {
    await authenticatedFetch(`${apiBase}/auth/logout`, {
      method: "POST",
    });

    user.value = null;
  }

  return {
    login,
    register,
    signOut,
    fetchUser,
    authenticatedFetch,
    refresh,
    user,
    isAuthenticated,
  };
};
