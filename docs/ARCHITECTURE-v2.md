# Camoufox Agent Architecture v2

This document summarises the high level structure of the project. The goal is to
run a background agent that connects to a Camoufox browser and keeps an
authenticated page alive.

The repository is organised as follows:

```
.
├── cmd/
│   ├── server/    # gateway HTTP and WebSocket server
│   └── agent/     # Camoufox agent process
├── internal/
│   ├── proxy/     # proxy server business logic
│   └── browser/   # Playwright driver utilities
└── docs/
    └── ARCHITECTURE-v2.md
```

For full details see the project guide.
