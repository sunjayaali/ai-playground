export default defineNuxtRouteMiddleware(async (to, _) => {
  const { user, fetchUser, isAuthenticated } = useAuth();
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
