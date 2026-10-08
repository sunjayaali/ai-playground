export default defineNuxtRouteMiddleware(async (to, from) => {
  if (import.meta.server) {
    return;
  }

  const { fetchUser, isAuthenticated } = useAuth();
  await fetchUser();

  if (isAuthenticated.value) {
    return navigateTo("/", { replace: true });
  }
});
