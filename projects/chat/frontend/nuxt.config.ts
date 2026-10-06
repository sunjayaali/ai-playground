// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: ["@nuxt/ui", "@nuxt/eslint"],
  css: ["~/assets/css/main.css"],
  runtimeConfig: {
    public: {
      // Dev-time only. In production set NUXT_PUBLIC_WS_URL.
      wsUrl: process.env.NUXT_PUBLIC_WS_URL ?? "ws://192.168.0.113:3001/ws",
      apiBase: process.env.NUXT_PUBLIC_API_BASE ?? "http://localhost:3001",
    },
  },
});
