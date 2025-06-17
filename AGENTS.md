# Repo Guidelines

- Run `go vet ./...` before committing Go code.
- After changes, run `go fmt ./...`.
- For Python code, run `python -m py_compile` on modified files.
- To build the Camoufox container and login, follow these steps:
  1. Run `scripts/bootstrap_login.sh` with `TARGET_URL` env set to login and generate `data/cookie.json`.
  2. Use `docker-compose up -d` to start the `camoufox` browser and Go server.
