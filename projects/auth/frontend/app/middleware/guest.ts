export default defineNuxtRouteMiddleware(async (_to, _from) => {
  const { user, fetchUser, isAuthenticated } = useAuth();
  if (!user.value) {
    try {
      await fetchUser();
    } catch {}
  }

  if (isAuthenticated.value) {
    return navigateTo("/", { replace: true });
  }
});
