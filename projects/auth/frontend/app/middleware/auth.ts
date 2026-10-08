export default defineNuxtRouteMiddleware(async (to, _) => {
  const { fetchUser, isAuthenticated } = useAuth();
  await fetchUser();

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
