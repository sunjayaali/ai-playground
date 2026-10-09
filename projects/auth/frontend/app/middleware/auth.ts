export default defineNuxtRouteMiddleware(async (to, _) => {
  const authStore = useAuthStore();
  const { user, isAuthenticated } = storeToRefs(authStore);
  const { fetchUser } = authStore;
  if (!user.value) {
    try {
      await fetchUser();
    } catch {}
  }

  if (to.path === "/login") {
    if (isAuthenticated.value) {
      return navigateTo("/", { replace: true });
    }

    return;
  }

  if (!isAuthenticated.value) {
    return navigateTo("/login", { replace: true });
  }
});
