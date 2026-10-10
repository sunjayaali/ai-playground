import type { AuthApi } from "~/services/auth.api";
import { ApiError } from "~/services/auth.api";
import type { UserInfo } from "~/types/auth";

export const useAuthStore = defineStore("auth", () => {
  const { $authApi } = useNuxtApp();
  const api: AuthApi = $authApi;

  const user = ref<UserInfo | null>(null);
  const isAuthenticated = computed(() => !!user.value);

  // Session refresh: succeeds only when the cookie-based refresh
  // actually worked. The store keeps no tokens — the backend sets
  // them as httpOnly cookies.
  async function refreshSession(): Promise<boolean> {
    const ok = await api.refresh();
    if (!ok) {
      user.value = null;
    }
    return ok;
  }

  async function login(username: string, password: string) {
    await api.login(username, password);

    // Cookies are set, but the current user only comes from /auth/me.
    const me = await api.getMe();
    if (!me) {
      throw new ApiError(502, "could not load user after login");
    }
    user.value = me;
    return me;
  }

  async function register(username: string, password: string) {
    await api.register(username, password);
  }

  async function fetchUser(): Promise<UserInfo | null> {
    const me = await api.getMe();
    user.value = me;
    return me;
  }

  async function signOut(): Promise<void> {
    await api.logout();
    user.value = null;
  }

  return {
    user,
    isAuthenticated,
    login,
    register,
    signOut,
    fetchUser,
    refreshSession,
  };
});
