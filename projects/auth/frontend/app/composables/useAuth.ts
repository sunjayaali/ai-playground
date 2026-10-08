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

export default () => {
  const config = useRuntimeConfig();
  const apiBase = config.public.apiBase;
  const user = useState<User | null>("auth:user", () => null);
  const isAuthenticated = computed(() => !!user.value);

  async function login(username: string, password: string) {
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

    return res;
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
    const res = await fetch(`${apiBase}/auth/me`, {
      credentials: "include",
    });
    if (!res.ok) {
      user.value = null;
      return null;
    }

    user.value = (await res.json()) as User;
    return user.value;
  }

  async function signOut(): Promise<void> {
    const res = await fetch(`${apiBase}/auth/logout`, {
      method: "POST",
      credentials: "include",
    });

    user.value = null;
  }

  return {
    login,
    register,
    signOut,
    fetchUser,
    user,
    isAuthenticated,
  };
};
