# auth frontend

A Nuxt 4 + Nuxt UI demo for the Go auth service in
`../` (same stack as `projects/chat/frontend`).

## Layout

- `app/composables/useAuth.ts` — session state only:
  login/register/logout, `/auth/me` via
  `authenticatedFetch`, and `ready()`, which resolves once
  the boot-time session restore has finished. No token
  state at all — the tokens ride the backend's httpOnly
  cookies, which `credentials: "include"` replays on
  every call
- `app/middleware/auth.ts` — guards `/` (redirects to
  `/login` when logged out; awaits `ready()` first so a
  reload doesn't bounce an authenticated user)
- `app/middleware/guest.ts` — keeps logged-in users off
  `/login` and `/register`
- `app/pages/login.vue`, `register.vue` — `UAuthForm`
  in a `UPageCard`, Zod 4 validation, server messages
  surfaced via `#validation`
- `app/pages/index.vue` — the demo: session identity and
  a note on where the tokens actually live

## Run

Two terminals.

```sh
# 1. backend on :3002
cd ../backend
go run ./cmd/auth
```

```sh
# 2. frontend on :3000
pnpm dev
```

Open http://localhost:3000. Register a user, then:

- reload the page — the session survives, because the
  httpOnly cookies replay automatically
- sign out — the backend clears the cookies, ending the
  browser's session

## Notes

- The backend has a hand-rolled CORS middleware (no
  extra dependency) so `:3000` can call `:3002`.
- Every API call goes out with `credentials:
  "include"`, so the httpOnly cookies the backend sets
  are stored and replayed automatically.
- The frontend never handles a token. The backend sets
  them as httpOnly cookies and the browser carries them,
  so a page reload silently keeps the session — no
  re-login, and XSS has nothing to steal.
