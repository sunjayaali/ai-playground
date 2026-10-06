# Chat

A Nuxt 4 + Nuxt UI frontend for the chat project.

## Layout

- `app/` — Nuxt app: pages, components, composables
- `backend/` (sibling, `../backend`) — Go Fiber WebSocket server

## Run

Two terminals.

```sh
# 1. backend on :3001
cd ../backend
go run . -addr :3001 -tick 5s
```

`-tick` is the interval for the `server_time` push; `0`
(the default) disables it.

```sh
# 2. frontend on :3000
npm run dev
```

Open http://localhost:3000.

## Protocol

Envelope on the wire:

```json
{ "type": "echo | system | pong", "text": "...", "ts": 1759700000000 }
```

- `GET /ws` — upgrade; server pushes `system` on join/leave
- `GET /health` — `{"ok":true,"clients":2}`

Client → server frames: `{"type":"ping"}` (answered with `pong`, used for RTT)
or `{"text":"..."}` (echoed to every client). Non-JSON frames are treated as
plain text.

## Notes

- Writes are serialized under one mutex, so an echo and a pong can never
  interleave frames on the same connection.
- The client reconnects with exponential backoff (500ms → 10s cap) and keeps
  the last 200 messages.
- The frontend derives its socket URL from `NUXT_PUBLIC_WS_URL` and
  `NUXT_PUBLIC_API_BASE`, defaulting to `localhost:3001`.
