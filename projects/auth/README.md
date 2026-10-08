# auth

JWT auth service in Go, built on Fiber v3 (same stack as
`projects/chat`). Stateless HMAC-SHA256 access tokens plus
refresh tokens, both carried in httpOnly cookies — tokens
never reach JavaScript. There is no revocation store: tokens
stay valid until their TTL runs out, so logout is advisory.

## Endpoints

| Method | Route            | Auth    | Purpose                          |
| ------ | ---------------- | ------- | -------------------------------- |
| `GET`  | `/health`        | —       | liveness probe                   |
| `POST` | `/auth/register` | —       | create user, sets token cookies  |
| `POST` | `/auth/login`    | —       | verify credentials, token cookies |
| `POST` | `/auth/refresh`  | cookie  | rotate the refresh cookie        |
| `POST` | `/auth/logout`   | bearer  | clear the token cookies          |
| `GET`  | `/auth/me`       | bearer  | current user's id                |

### Request / response shapes

```sh
# register or login — same body, same response
curl -X POST :3002/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"supersecret"}'
# -> {"access_token":"eyJ...","token_type":"Bearer",
#     "refresh_token":"eyJ...","expiry":"2026-10-09T...Z","expires_in":86400}
#   + Set-Cookie: access_token=... (httpOnly)
#   + Set-Cookie: refresh_token=... (httpOnly)

# any protected route (cookies ride along with -c/-b)
curl :3002/auth/me -b cookies.txt
# -> {"id":"954fc95f-21f3-4198-87a6-f1a7fe404001"}

# rotate the refresh token from its cookie
curl -X POST :3002/auth/refresh -b cookies.txt -c cookies.txt
# -> {"access_token":"eyJ...","token_type":"Bearer",
#     "refresh_token":"eyJ...","expiry":"2026-10-09T...Z","expires_in":86400}
```

Validation: username ≥ 3 chars, password ≥ 8 chars.
Wrong username and wrong password return the same 401
(and take the same bcrypt time), so the response cannot
leak which half failed.

## Layout

```
backend/cmd/auth/main.go   entrypoint: flags, .env, Fiber wiring
backend/internal/auth/     TokenService, UserStore
```

## Run

```sh
cd projects/auth/backend
cp .env.example .env   # fill in a random secret
go run ./cmd/auth
# or: go run ./cmd/auth -secret <secret> -addr :3002
```

The secret is picked up in order of precedence: `-secret`
flag, `JWT_SECRET` env var, then `JWT_SECRET` in a `.env`
file (loaded with `github.com/joho/godotenv`).

Flags: `-addr` (default `:3002`), `-secret` (or
`JWT_SECRET` env / `.env` — required), `-access-ttl`
(24h), `-refresh-ttl` (7d), `-secure-cookies`
(false in dev; sets `Secure` + `SameSite=None`
for HTTPS deployments).

## Design notes

- **Token responses use the oauth2.Token shape.** The
  JSON body of register, login, and refresh mirrors
  `golang.org/x/oauth2`'s `Token` struct —
  `access_token`, `token_type` (`Bearer`),
  `refresh_token`, `expiry`, `expires_in` — so a
  client that speaks OAuth2 recognizes it. The
  response is descriptive only: the tokens themselves
  travel solely in the httpOnly cookies.
- **Token types are pinned.** The JWT header `typ` is
  `access` or `refresh`; an access token cannot be replayed
  to `/auth/refresh`, and vice versa.
- **Users are identified by UUID.** Every user gets a random
  UUID at registration (`github.com/google/uuid`). That UUID
  is the JWT `sub` and the key the `UserStore` is indexed by,
  with a separate username -> id map for login lookup. The
  username stays unique but is no longer the identity, so a
  session survives a username change and ids are unguessable.
  `/auth/me` returns the `id`, not the name.
- **alg is pinned to HMAC.** The keyfunc rejects anything
  but `SigningMethodHMAC`, so a forged `alg: none` or
  `alg: RS256` token fails signature verification.
- **Refresh rotates, without revocation.** Every
  refresh issues a fresh pair, but with no blocklist the
  old refresh token also keeps working until it expires.
  Add a revocation store (Redis, say) if a stolen token
  must stop working immediately.
- **Logout is advisory.** With no revocation store,
  `/auth/logout` acknowledges the request but the tokens
  keep working until their TTL runs out — the default
  access TTL is 24h. Clearing both cookies ends the
  browser's side of the session.

## Limits

Users live in an in-memory map (`store.go`). Swap
`UserStore` for a real database before using this anywhere
real, and add a revocation store if logout must take
effect before the token TTLs expire.