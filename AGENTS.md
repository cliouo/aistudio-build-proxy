# Contributor Guide – camoufox‑agent

Welcome! 这里是 **浏览器守护进程 (agent)** 与 **网关 (server)** 的协作仓库。请在提交代码前阅读下列规范。

---

## Dev Environment Tips

- **Go Workspaces**  
  在项目根运行
  ```bash
  go work use ./cmd/... ./internal/...
````

让 `gopls` 与 `go test` 只扫描我们用到的子模块。

* **快速跳转**
  用 `ghq list | fzf` 定位包，再 `cd` 进入目录；或 `go run ./tools/cdserver.go` 直接切换到 `cmd/server`。

* **安装前端依赖**
  若需调试 `web/`，推荐
  `pnpm install --filter web && pnpm --filter web dev`.

* **Playwright 浏览器缓存**
  初次运行 `scripts/start_camoufox.sh` 会自动 `camoufox fetch`。如需升级版本，删除 `~/.cache/camoufox` 后重跑脚本。

---

## Testing Instructions

| 场景             | 命令                                                          |
| -------------- | ----------------------------------------------------------- |
| 单元测试           | `go test ./...`                                             |
| 集成（含 Camoufox） | `make itest` <br>*(docker‑compose 拉起 camoufox → 运行 e2e 流程)* |
| 静态检查           | `golangci-lint run ./...`                                   |
| 仅跑 agent 逻辑    | `go test ./internal/browser -run TestAgent`                 |
| 前端单测           | `pnpm --filter web test`                                    |

* GitHub Actions 工作流定义见 `.github/workflows/ci.yml`。
* 合并前 **必须** 让 `make ci` 全绿。

---

## PR Instructions

**标题格式**

```
[agent]  修复 token 过期重连
[proxy]  支持多租户动态路由
[docker] bump camoufox -> v133.2
```

**提交原则**

1. 🚦 **一行代码 = 一行测试**：新功能或 Bugfix 必须带测试。
2. 📜 **保持 CHANGELOG**：在 `CHANGELOG.md` 写明外部可见的变更。
3. 🧹 **跑 `go mod tidy`**：删除未用依赖。
4. 🔐 **不要提交机密**：Google Cookie、私钥等请放 `.env` 或 CI Secret。

感谢你的贡献！ 🎉

```

---

### 结语

按照上面的 **目录拆分 → agent 编写 → Camoufox 容器化 → 环境变量 / CI 守护** 四个步骤改造后，你的 Go 服务就可以在 **完全无头** 的环境里稳固地打开并保持 Google 登录页面，前端 WebSocket 客户端会自动连回网关，无需再手工挂浏览器。祝编码顺利！