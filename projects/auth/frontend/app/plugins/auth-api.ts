import { AuthApiClient } from "~/services/auth.api";

export default defineNuxtPlugin({
  name: "auth-api",
  setup() {
    const config = useRuntimeConfig();
    // No credentials here — auth.api.ts sets them per request.
    const fetcher = $fetch.create({});
    const authApi = new AuthApiClient(config.public.apiBase, fetcher);

    return {
      provide: {
        authApi,
      },
    };
  },
});
