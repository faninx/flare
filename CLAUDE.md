# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

**Flare** 是一个用 Go（Golang）编写的自托管书签 / 起始页应用，前端只使用少量原生 JavaScript。HTTP 层使用 **Echo v5** 框架，以单一可执行文件形式分发，不依赖任何数据库（状态以 YAML 文件保存），并将**性能**（QPS、内存分配、延迟）作为核心设计约束。

- 模块路径：`github.com/faninx/flare`
- Go 版本：`1.26.0`（CI 中 `GO_VERSION: "1.26.0"`）

## 常用命令

所有命令都已在 `Makefile` 中封装，也可直接用 `go` 执行。

```bash
# 构建
make build                    # 等价 go build ./...
go run build/build.go         # 资源生成流水线（首次构建前必须执行）
go mod tidy                   # 运行 build.go 之后必须执行以同步依赖

# 测试
make test                     # 等价 go test ./... -count=1
go test -race -coverprofile=coverage.out -covermode=atomic ./...   # CI 使用的变体
go test -run TestName ./internal/server                          # 运行单个测试
go test -v ./internal/pages/home                                 # 运行某个包的全部测试

# 覆盖率
make coverage                 # 生成 coverage.html

# 格式与静态检查
make fmt                      # 等价 gofmt -s -l .（CI 中用 gofmt -s -d .）
make vet                      # 等价 go vet ./...
```

CI（`/.github/workflows/ci.yml`）按以下顺序跑四个 Job：`fmt` → `vet` → `lint`（golangci-lint）→ `test`（带 race + 覆盖率上传）。`test` Job 在开始前会先执行 `go run build/build.go` 和 `go mod tidy`。

## 代码架构

### 入口与生命周期

`main.go` 故意保持极简，只做两件事：

```go
flags := cmd.Parse()           // env -> envfile -> CLI flag 合并，应用默认值
server.StartDaemon(&flags)     // 构建路由，启动 http.Server，处理 SIGINT/SIGTERM
```

`cmd.Parse()`（`cmd/cmd.go`）是配置的汇聚点。优先级为：代码内默认值（`config/define/cmd.go`）→ env 文件（`cmd/envfile.go`）→ 环境变量（`cmd/env.go`，通过 `caarlos0/env` 解析）→ CLI 参数（`cmd/cli.go`，通过 `spf13/pflag` 解析）。CLI 优先，缺失则降级到下一层。`define.AppFlags` 是运行时的唯一真实来源。

### 三层 `config/` 包

- **`config/model/`** — 纯数据结构（无逻辑、无 I/O）。包含 `Flags`、`Envs`、`EnvFile`、`Application`、`Bookmark` / `Category` / `Bookmarks`、`Theme` / `Palette`、`Weather`、`Page` / `API` / `RouteMaps`。**根据 `CONVENTIONS-OF-CODE.md`，先定义模型，再实现功能。**
- **`config/define/`** — 编译期常量与路由表。`RegularPages`、`SettingPages`、`SettingPagesAPI`、`MiscPages` 在 `router.go` 中构建一次后全局共享。`define.AppFlags` 全局变量保存运行时状态；`server.NewRouter` 内部会调用 `define.Init()`。
- **`config/data/`** — 所有文件系统 I/O：负责 `config.yml`、`apps.yml`、`bookmarks.yml` 的 YAML 读写。使用 `sync.RWMutex` 保护的内存缓存（`cache.go` 中 `cachedConfig`、`cachedApps`、`cachedBookmarks`），以文件名为 key，保存时失效。`editor.go` 处理书签编辑器的 CSV 导入/导出（基于 `jszwec/csvutil`）。

### 路由（`internal/server/router.go`）

`NewRouter` 构建 Echo 实例并按以下顺序挂载中间件与路由：Recover（可选请求日志中间件）→ `auth.RequestHandle` → 天气初始化 → 模板 → 静态资源 → misc/health → home → settings/* → theme/weather/search/appearance/others → MDI → redir → 可选 guide → 可选 editor → deprecated。模板 / MDI / guide / editor 初始化失败会返回 `error` 并中止启动。`Visibility`（`DEFAULT` 或 `PRIVATE`）决定 home/help/applications/bookmarks 是否挂载 `auth.AuthRequired` 中间件。

### 页面与功能模块

- `internal/pages/home/` — `home.html` 模板，三个页面处理器（home、help、applications、bookmarks），天气缓存（`_CACHE_WEATHER_DATA`），问候语解析器（按 5–10 / 11–13 / 14–18 / 其他 时段切分）。
- `internal/pages/editor/` — 基于 CSV 的书签编辑器（受 `EnableEditor` 开关控制）。
- `internal/pages/guide/` — 首次使用引导页（受 `EnableGuide` 开关控制）。
- `internal/settings/{theme,weather,search,appearance,others}/` — 每个设置页一个子包，各自注册路由。`settings.go` 本身只是 `/settings` → `/settings/theme` 的重定向。
- `internal/misc/{health,redir,deprecated}/` — `/ping`、`/redir`、遗留接口的弃用提示。

### 资源是“生成”出来的，不要手写

`embed/` 目录保存的是**源**资源（CSS、第三方 JS、MDI 速查表、favicon、模板）。`build/build.go` 调用 `build/builder` 包，生成 Go 源码到 `internal/resources/`（模板 → `internal/resources/templates/html/`、MDI → `internal/resources/mdi/icons.go`、favicon → `internal/resources/assets/favicon.ico` 等）以及 `internal/pages/{guide,editor}/...`。**修改 `embed/` 下任何内容后，必须执行 `go run build/build.go && go mod tidy`。** 生成的产物通过 `.gitignore` 中的规则（如 `internal/templates/html/*`、`internal/mdi/mdi-cheat-sheets`）排除在版本控制之外。

### 横切关注点

- `internal/pool/template_map.go` — 模板渲染 map 的 `sync.Pool`（容量 48）以及 `bytes.Buffer` 池（位于 `internal/resources/templates/templates.go`）。模板渲染是热点路径，所有需要渲染 HTML 的处理器都应使用 `pool.GetTemplateMap()` / `pool.PutTemplateMap(m)`（参考 `internal/pages/home/home.go`）。
- `internal/i18n/` — 通过 `embed.FS` 加载 `locales/{zh,en}.json`。对外暴露 `T(locale, key)`、`Tf(locale, key, args...)`、`Weekday(locale, w)`、`DateFormat(locale)`。未知 locale 回退到 `en`，再回退到 `zh`；缺失的 key 直接返回 key 本身。模板中通过 `template.FuncMap` 调用 `{{T "key"}}`。
- `internal/auth/` — 基于 cookie 的会话（`gorilla/sessions` + `labstack/echo-contrib/v5/session`），使用 `crypto/subtle` 做常量时间密码比较。提供 `AuthRequired` 中间件、`CheckUserIsLogin`、`GetUserName`、`GetUserLoginDate`。`DisableLoginMode=true` 时整个模块被跳过。
- `internal/logger/` — `*slog.Logger` 单例，`init()` 中初始化。日志级别由 `FLARE_DEBUG=on`（环境变量）或 `FLARE_DEBUG`（CLI 标志）控制。Echo 请求日志中间件在 `echo_handler.go`；设置 `FLARE_BASELINE=1` 可关闭请求日志，便于压测。
- `internal/version/` — 使用 `github.com/soulteary/version-kit`，发布时通过 `-ldflags` 注入。

### 性能约定（来自 `CONVENTIONS-OF-CODE.md`）

性能是项目明文写出的核心价值。约定如下：

1. **优先使用普通函数，而非带接收者的方法。** 搜索 `func pageXxx(...)` 模式，而非 `func (s *Server) pageXxx(...)`。
2. **避免使用指针接收者** — 优先考虑并发安全与值类型语义。
3. **尽量减少内存重新分配** — 使用 `sync.Pool`（模板 map、渲染 buffer）、预分配已知容量的 slice、避免在热点路径上反复 `append`。
4. **模型优先、测试优先** — 在 `config/model/` 中先定义结构体，再写测试，最后实现函数体。

CI 同时使用 `go test -race`，代码必须在 `-race` 下安全。

## 静态检查

`.golangci.yml` 使用 **v2** 格式，启用 `gofmt -s`（simplify 模式）、`errcheck`（含 type-assertion 与 blank 检查）以及 `govet enable-all`（关闭 `fieldalignment`）。`errcheck` 在测试文件、`build/`、`tools/` 以及 `cmd/cli.go` 中的特定 pflag 调用处被排除。`goimports` 的本地前缀：`github.com/faninx/flare`。

## 主要配置项（高频关注）

所有环境变量均以 `FLARE_` 为前缀，由 `caarlos0/env` 解析：

`FLARE_PORT`（5005）、`FLARE_GUIDE`（true）、`FLARE_EDITOR`（true）、`FLARE_OFFLINE`（false）、`FLARE_DEPRECATED_NOTICE`（true）、`FLARE_DISABLE_CSP`（false）、`FLARE_VISIBILITY`（DEFAULT）、`FLARE_DISABLE_LOGIN`（true）、`FLARE_USER`、`FLARE_PASS`、`FLARE_COOKIE_NAME`（`flare`）、`FLARE_COOKIE_SECRET`（`secret` — 生产环境必须覆盖）。

运行时数据文件（首次启动在当前工作目录创建）：`config.yml`、`apps.yml`、`bookmarks.yml`。默认值模板见 `config/data/fs.go: getConfigPath` 与 `data/config.go: initAppConfig`。

## 发布与 Docker

- `.goreleaser.yaml` + `docker/goreleaser/Dockerfile.{amd64,arm32v6,arm32v7,arm64v8}` — 多架构发布流水线。
- `docker/manual/` — 不依赖 GoReleaser 的多架构 Dockerfile。
- CI 在测试前调用 `build/build.go` 以重新生成 embed 资源。

## 基准测试

`scripts/baseline-metrics.sh`（说明见 `docs/baseline-metrics.md`）对已构建的二进制进行 HTTP 压测，输出 `baseline-report.md` 以及 CPU / Heap pprof 文件到 `.benchmark/baseline-<timestamp>/`。可通过 `REQUESTS`、`CONCURRENCY`、`WARMUP` 环境变量调整。`FLARE_BASELINE=1` 环境变量会关闭请求日志，以获得更干净的压测数据。
