export default defineNuxtPlugin(() => {
  const authStore = useAuthStore();
  const { isAuthenticated } = storeToRefs(authStore);

  watch(isAuthenticated, async (authed) => {
    if (authed) {
      return;
    }
    const route = useRoute();
    if (route.path === "/login" || route.path === "/register") {
      return;
    }
    await navigateTo("/login", { replace: true });
  });
});
