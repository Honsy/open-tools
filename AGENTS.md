# 开物录

网址导航。游客不用登录就能看收录。没登录时收藏和最近使用记在浏览器 `localStorage`（`ot-custom`、`ot-recent`），登录后改记在账号上。前台和后台共用账号和登录会话，见下面「权限」。

站点名是开物录。`kaiwulu.com` 已启用，公开入口通过 Nginx 反代到上海服务器 `8.153.64.49`；旧的“域名未注册、现网用 IP”记录已过期。

## 结构

- `server/` — Go（Gin + GORM）。公开页是 `html/template`，改模板、CSS、JS 要重新编译。
- `web/` — Vite 6 + Vue 3。只做 `/admin`。开发时公开 HTML 由 Vite 转到本机 `:8080`。
- `legacy/` — 更早的工具站（Koa + React SSR）。和现网无关，不要再部署。

## 数据

库是独立的 MySQL `opentools`。不要把表写进 YouGo 的 `laoyouba`。

- 本机开发：`127.0.0.1:3306`，库 `opentools`。`server/config.yaml` 已忽略，不要提交，也不要指到线上。
- 新站写进 `links`（后台「网址」）。启动时 `EnsureCatalog` 只补还没有的名字，并只在旧地址还没被改过时修一批已知坏链。已经在库里的行不要靠改 seed 去覆盖。
- 首页点击数是种子基线，不是实时流量。新站 `clicks`、`views` 从 0 起。详情页加浏览，`/go/:id` 加点击。不要编访问量。

## 设计

公开页样式是 Tailwind CSS 4 加 daisyUI 5，主题按仓库根目录 `DESIGN.md`（蓝白统一导航）。源文件在 `ui/public.css`，`npm run build` 写到 `server/api/static/site.css`。改颜色、圆角、间距先改 `ui/public.css` 里的 daisyUI 主题，组件用 daisyUI 的 `btn`、`card`、`menu`、`navbar`、`tabs`、`badge`，布局用 Tailwind。不要再手写一套组件样式。

- 字体 IBM Plex Sans。中文回退 PingFang SC、Microsoft YaHei。标题 600，正文 14px，标签 12px / 500。
- 主色 `#0f62fe`，按下 `#004ccd`。页面底 `#fbf9f8`，卡片表面 `#f5f3f3` / `#ffffff`，字 `#1b1c1c`，次级字 `#424656`。成功绿 `#198038` 只用于确认，不拿来做装饰。
- 层次靠细边框和表面色差，不要重阴影。控件圆角约 0.25rem，大容器最多 0.75rem。
- 同目录的设计稿 HTML 只参考版式。站点、分类、热榜、介绍以库里的现有内容为准，不要换成稿子里的假站、广告或备案号。
- 邮箱、网盘、翻译、地图收在统一工具菜单中。

## 权限

所有账号存放在 `users`，默认 `is_admin = false`。首次初始化的后台管理员和旧后台迁入的账号标记为 `is_admin = true`。迁移保留已有用户 ID 和旧管理员密码。用户名重名时保留原读者账号，旧管理员使用迁移日志记录的新用户名。

不登录就能看、就能用：首页、分类、站点介绍、搜索、热榜、排行、公告、打开网站。这时收藏和最近使用只在这台浏览器。

读者在站内点「登录」只出登录框，点「注册」只出注册框。登录框和 `/login` 页里有「没有账号？去注册」，注册那边不倒回登录。直接打开 `/login`、`/register` 各是单独的一页，不套首页的顶栏、侧栏和页脚。

读者登录之后才做、才把内容写进页面：

- 提交网站必须登录。记录写在 `links.user_id`，状态仍是待审，不直接出现在首页。点用户名进 `/mine`：这一页不放分类侧栏，上面是用户名，下面是收藏、最近使用和自己的提交。
- 标明「登录后才展示」的区块。未登录时整段不要渲染进 HTML，页面上只留登录入口。不要用 CSS 把正文藏起来。
- 收藏和最近使用登录后写在 `user_links`，按 `user_id` 分开。每人最多 16 个收藏、12 条最近使用。登录时把这台浏览器里还没同步的记录并进账号，然后清掉本地那份。退出后再点的星和最近使用重新只留在这台浏览器。

后台未登录时转到统一的 `/login`，登录后回到原后台页面。`/api/admin` 每次请求都检查账号的 `is_admin`，管理员可以审核提交、改网址、分类、文章、标签。前台和后台退出时清除同一个会话。

## 页面

- 卡片进 `/site/:slug`。站名页可以看介绍，再点「打开网站」。
- 「打开网站」走 `GET /go/:id`：点击加一，然后 302。地址上没有 `from` 时补上当前访问的主机名（`X-Forwarded-Host`，否则 `Host`，去掉端口）。对方已经带了 `from` 就不覆盖。这条跳转 `noindex, nofollow`。
- 公开页服务端渲染，带独立 title、description、canonical。`/go`、`/search`、`/submit`、`/random`、`/tools/`、`/admin`、`/api` 不进收录。
- 图标走 `/ico?host=`，取目标站自己的图标。不要用 Google s2。

## 从 LaoYou8 迁入的内容

- AI 网址进入现有 `/c/ai`，文章进入 `/c/daily`，AI 排行复用 `/rank?category=ai`。文章支持 Markdown、清理后的 HTML、封面、视频入口、草稿和评论。旧文章访问通过 `/ai/articles/:id` 查映射。
- 非游戏实验室工具走 `/tools/`。趣味测试、房贷和退休计算在浏览器运行。AI 工具使用开物录账号、`AI_BAILIAN_API_KEY` 和可选的 `AI_BAILIAN_CHAT_MODELS`。换背景的 PNG 和后台文章图片使用开物录的 R2 配置。
- 切换两站前，先让开物录启动一次，建好新字段和表。在 `server/` 运行 `go run ./cmd/import-laoyou8` 预览来源，再加 `-apply` 导入。`LAOYOU8_SOURCE_DSN` 使用指向 `laoyouba` 的只读账号或 SSH 隧道，目标按开物录配置连接 `opentools`。同网址或同标题、正文的已有内容会复用，已经迁入的记录可以重复导入核对。旧评论保留显示名，历史站点留言和 AI 投稿进入后台「反馈」。游戏记录、友情链接及其投稿留在 LaoYou8。
- 老文章媒体继续使用原有公开地址。LaoYou8 移除已迁走的公开入口、接口和任务。原日报和 GitHub 趋势任务用开物录的 `DAILY_ARTICLES=true` 启用，配置后每天北京时间 08:00 写入文章。

## 线上（上海，和 YouGo 同一台）

`8.153.64.49`。口令不要写进这个仓库。

| 项 | 值 |
|---|---|
| 进程 | systemd `kaiwu`，工作目录 `/data/app/kaiwu`，二进制 `/data/app/kaiwu/opentools`，听 `:8080` |
| 配置 | `/data/app/kaiwu/config.yaml`（权限 600）。会话密钥用 `sessionSecret`，兼容旧的 `adminSecret`。全新初始化时用 `adminPassword` 创建首个管理员 |
| 入口 | Nginx `:80` 根路径反代到 `127.0.0.1:8080`，带 `X-Forwarded-Host` |
| 后台 | `/admin/`，静态在 `/data/app/kaiwu/admin`，构建用 `vite build --base /admin/` |
| YouGo | 同机 `/yougo/`，反代 `127.0.0.1:8000`。不要占 80 |
| MySQL | 本机 `127.0.0.1:13366`，库 `opentools`。YouGo 仍用库 `laoyouba` |

这台机器更新靠本机交叉编译后上传，不要在机器上 `git pull`。Linux 包：`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`。编完把 `GOOS` 设回 `windows`，否则下一发本机编译会变成 ELF。不要编 Windows 可执行文件来跑本地接口。

公开页源码改完要换二进制。后台页面改完才要重新构建 `web/` 并上传 `/admin/`。

## 本地

不主动起服务。要跑时在 `server/` 里 `go run .`（`:8080`），另开 `web/` 的 `npm run dev`（`:5173`）。本机 Go 若低于 `go.mod` 的 1.24，设 `GOTOOLCHAIN=auto`，代理用 `https://goproxy.cn,direct`。
