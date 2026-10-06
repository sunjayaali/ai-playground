# ai-playground

Monorepo of AI experiments. Currently one project.

## Projects

- **chat** — Go Fiber WebSocket broadcast server + Nuxt 4 / Nuxt UI frontend

```
projects/chat/
├── backend/    # Go Fiber server (main.go)
└── frontend/   # Nuxt 4 app
```

## Run chat

```sh
# terminal 1 — backend on :3001
cd projects/chat/backend && go run . -addr :3001 -tick 5s

# terminal 2 — frontend on :3000
cd projects/chat/frontend && npm run dev
```

See [projects/chat/frontend/README.md](projects/chat/frontend/README.md) for the
WebSocket protocol.
