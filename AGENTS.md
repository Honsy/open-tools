# 开物录

网址导航。公开页没有账号，收藏和最近使用记在浏览器 `localStorage`（`ot-custom`、`ot-recent`）。后台只有一个管理员。

站点名是开物录。`kaiwulu.com` / `.net` / `.cn` 在 2026-10-03 查过还没人注册，也还没买。现网用 IP 访问。

## 结构

- `server/` — Go（Gin + GORM）。公开页是 `html/template`，改模板、CSS、JS 要重新编译。
- `web/` — Vite 6 + Vue 3。只做 `/admin`。开发时公开 HTML 由 Vite 转到本机 `:8080`。
- `legacy/` — 更早的工具站（Koa + React SSR）。和现网无关，不要再部署。

## 数据

库是独立的 MySQL `opentools`。不要把表写进 YouGo 的 `laoyouba`。

- 本机开发：`127.0.0.1:3306`，库 `opentools`。`server/config.yaml` 已忽略，不要提交，也不要指到线上。
- 新站写进 `links`（后台「网址」）。启动时 `EnsureCatalog` 只补还没有的名字，并只在旧地址还没被改过时修一批已知坏链。已经在库里的行不要靠改 seed 去覆盖。
- 首页点击数是种子基线，不是实时流量。新站 `clicks`、`views` 从 0 起。详情页加浏览，`/go/:id` 加点击。不要编访问量。

## 页面

- 卡片进 `/site/:slug`。站名页可以看介绍，再点「打开网站」。
- 「打开网站」走 `GET /go/:id`：点击加一，然后 302。地址上没有 `from` 时补上当前访问的主机名（`X-Forwarded-Host`，否则 `Host`，去掉端口）。对方已经带了 `from` 就不覆盖。这条跳转 `noindex, nofollow`。
- 公开页服务端渲染，带独立 title、description、canonical。`/go`、`/search`、`/submit`、`/random`、`/tools/`、`/admin`、`/api` 不进收录。
- 图标走 `/ico?host=`，取目标站自己的图标。不要用 Google s2。

## 线上（上海，和 YouGo 同一台）

`8.153.64.49`。口令不要写进这个仓库。

| 项 | 值 |
|---|---|
| 进程 | systemd `kaiwu`，工作目录 `/data/app/kaiwu`，二进制 `/data/app/kaiwu/opentools`，听 `:8080` |
| 配置 | `/data/app/kaiwu/config.yaml`（权限 600）。环境变量盖过 yaml。管理员口令为空则起不来 |
| 入口 | Nginx `:80` 根路径反代到 `127.0.0.1:8080`，带 `X-Forwarded-Host` |
| 后台 | `/admin/`，静态在 `/data/app/kaiwu/admin`，构建用 `vite build --base /admin/` |
| YouGo | 同机 `/yougo/`，反代 `127.0.0.1:8000`。不要占 80 |
| MySQL | 本机 `127.0.0.1:13366`，库 `opentools`。YouGo 仍用库 `laoyouba` |

这台机器更新靠本机交叉编译后上传，不要在机器上 `git pull`。Linux 包：`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`。编完把 `GOOS` 设回 `windows`，否则下一发本机编译会变成 ELF。不要编 Windows 可执行文件来跑本地接口。

公开页源码改完要换二进制。后台页面改完才要重新构建 `web/` 并上传 `/admin/`。

## 本地

不主动起服务。要跑时在 `server/` 里 `go run .`（`:8080`），另开 `web/` 的 `npm run dev`（`:5173`）。本机 Go 若低于 `go.mod` 的 1.24，设 `GOTOOLCHAIN=auto`，代理用 `https://goproxy.cn,direct`。
