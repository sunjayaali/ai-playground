export default defineNuxtRouteMiddleware(async (_to, _from) => {
  const authStore = useAuthStore();
  const { user, isAuthenticated } = storeToRefs(authStore);
  const { fetchUser } = authStore;
  if (!user.value) {
    try {
      await fetchUser();
    } catch {}
  }

  if (isAuthenticated.value) {
    return navigateTo("/", { replace: true });
  }
});
