# Flare

Challenge all bookmarking apps and websites directories, Aim to Be a best performance monster.

🚧 **Code is being prepared and refactored, commits are slow.**

## 关于这个 Fork

[soulteary/flare](https://github.com/soulteary/flare) 是我很喜欢的导航 / 书签页面 —— 它简洁、高效、又轻量。

考虑到家庭环境部署时**局域网和公网访问同一个 bookmark 的链接往往不同**（例如 NAS 在内网是 `http://nas.local:5000`，从外网就得是 `https://nas.example.com`），手改来改去很烦。借助 AI 的帮助，本 fork 加了一个**内外网链接切换**的小能力：每条 bookmark 可以同时维护 LAN / WAN 两个 URL，页面左下角有一个小按钮随时切换。原始项目依旧在 upstream 维护，切换逻辑只在本 fork 启用。

## Feature

**Simple**, **Fast**, **Lightweight** and super **Easy** to install and use.

- Written in Go (Golang) and a little Modern vanilla Javascript only.
- HTTP stack: [Echo](https://echo.labstack.com/) v5.
- Doesn't depend on any database or any complicated framework.
- Single executable, no dependencies required, good docker support.
- You can choose whether to enable various functions according to your needs: offline mode, weather, editor, account, and so on.

## ScreenShot

TBD

## Documentation

TBD

- Browse automatically generated program documentation:
    - `godoc --http=localhost:8080`



## Directory

```bash
├── build                   build script
├── cmd                     user cli/env parser
├── config                  config for app
│   ├── data                    data for app running
│   ├── define                  define for app launch
│   └── model                   data model for app
├── docker                  docker
├── embed                   resource (assets, template) for web
├── internal
│   ├── auth                user login
│   ├── fn                  fn utils
│   ├── logger              logger
│   ├── misc
│   │   ├── deprecated
│   │   ├── health
│   │   └── redir
│   ├── pages
│   │   ├── editor
│   │   ├── guide
│   │   └── home
│   ├── resources           static resource after minify
│   ├── server
│   ├── settings
│   └── version
└── main.go
```