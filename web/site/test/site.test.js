import assert from "node:assert/strict";
import { mkdtemp, mkdir, rm, writeFile } from "node:fs/promises";
import http from "node:http";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { renderPage } from "../render.js";
import { createSite } from "../server.js";

const shell = {
  Title: "开物录", Description: "网址导航", Path: "/", Canonical: "https://kaiwulu.example/",
  IconBase: "/ico?host=", JSONLD: { name: "<开物录>" },
};

async function listen(server, t) {
  await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
  t.after(() => { server.close(); });
  return "http://127.0.0.1:" + server.address().port;
}

function page(res, template, data, layout) {
  res.setHeader("content-type", "application/vnd.kaiwulu.page.v1+json");
  res.end(JSON.stringify({ template, shell: layout, data }));
}

test("the public Node process renders API views and forwards sessions without exposing the private routes", async t => {
  const hits = [];
  const api = await listen(http.createServer((req, res) => {
    hits.push(req.method + " " + req.url);
    assert.equal(req.headers["x-forwarded-host"], "kaiwulu.example");
    assert.equal(req.headers["x-forwarded-proto"], "https");
    switch (req.url.split("?")[0]) {
      case "/_site/login":
        res.setHeader("cache-control", "no-store");
        if (req.method === "POST") {
          res.setHeader("set-cookie", "ot_reader=test; Path=/; HttpOnly; Secure");
          if (req.headers["x-auth-modal"] === "1") {
            res.setHeader("content-type", "application/json");
            res.end('{"ok":true,"next":"/mine"}');
          } else {
            res.writeHead(303, { location: "/admin/links" }).end();
          }
        } else {
          page(res, "authpage", { Kind: "login", Title: "登录", Action: "/login", Next: "/admin/links", SwitchTo: "/register" });
        }
        break;
      case "/_site/sitemap.xml":
        page(res, "sitemap", ["/", "/c/ai", "/search?a=1&b=2"]);
        break;
      case "/_site/tools/sbti":
        page(res, "labTool", { Slug: "sbti", Name: "SBTI", Desc: "趣味测试" }, shell);
        break;
      case "/api/admin/me":
        assert.equal(req.headers.cookie, "ot_reader=test");
        res.setHeader("content-type", "application/json");
        res.end('{"username":"admin"}');
        break;
      case "/health":
        res.setHeader("content-type", "application/json");
        res.end('{"ok":true}');
        break;
      default:
        res.writeHead(404).end();
    }
  }), t);
  const site = await listen(createSite({ apiOrigin: api }), t);
  const request = (path, options = {}) => fetch(site + path, {
    ...options, redirect: "manual", headers: {
      "x-forwarded-proto": "https", "x-forwarded-host": "kaiwulu.example", cookie: "ot_reader=test",
      ...options.headers,
    },
  });
  const checks = [
    ["/login?next=/admin/links", 200, 'name="next" value="/admin/links"'],
    ["/sitemap.xml", 200, "https://kaiwulu.example/search?a=1&amp;b=2"],
    ["/tools/sbti", 200, "/static/lab/sbti-page.js"],
    ["/api/admin/me", 200, '{"username":"admin"}'],
    ["/health", 200, '{"ok":true}'],
    ["/static/site.js", 200, "ot-recent"],
    ["/robots.txt", 200, "Sitemap: https://kaiwulu.example/sitemap.xml"],
    ["/_site/login", 404, "404"],
  ];
  for (const [path, status, content] of checks) {
    const response = await request(path);
    assert.equal(response.status, status, path);
    assert.ok((await response.text()).includes(content), path);
  }
  const modal = await request("/login", { method: "POST", headers: { "x-auth-modal": "1" } });
  assert.deepEqual(await modal.json(), { ok: true, next: "/mine" });
  const redirect = await request("/login", { method: "POST" });
  assert.equal(redirect.status, 303);
  assert.equal(redirect.headers.get("location"), "/admin/links");
  assert.match(redirect.headers.get("set-cookie"), /HttpOnly; Secure/);
  assert.equal(redirect.headers.get("cache-control"), "no-store");
  assert.deepEqual(hits, [
    "GET /_site/login?next=/admin/links", "GET /_site/sitemap.xml", "GET /_site/tools/sbti",
    "GET /api/admin/me", "GET /health", "POST /_site/login", "POST /_site/login",
  ]);
});

test("the public site links to built assets and gives only fingerprinted files a long cache lifetime", async t => {
  const staticDir = await mkdtemp(join(tmpdir(), "kaiwulu-assets-"));
  t.after(() => rm(staticDir, { recursive: true, force: true }));
  const names = { css: "site.0123456789abcdef.css", js: "site.fedcba9876543210.js" };
  const assets = { css: "/static/fingerprinted/" + names.css, js: "/static/fingerprinted/" + names.js };
  await mkdir(join(staticDir, "fingerprinted"));
  await Promise.all([
    writeFile(join(staticDir, "fingerprinted", "manifest.json"), JSON.stringify(assets)),
    writeFile(join(staticDir, "fingerprinted", names.css), "body {}"),
    writeFile(join(staticDir, "fingerprinted", names.js), "void 0;"),
    writeFile(join(staticDir, "site.css"), "body {}"),
  ]);
  const api = await listen(http.createServer((req, res) => {
    if (req.url === "/_site/login") page(res, "authpage", { Kind: "login", Title: "登录" });
    else page(res, "home", {}, shell);
  }), t);
  const site = await listen(createSite({ apiOrigin: api, staticDir }), t);
  const home = await (await fetch(site + "/")).text();
  assert.ok(home.includes(`href="${assets.css}"`));
  assert.ok(home.includes(`src="${assets.js}"`));
  const login = await (await fetch(site + "/login")).text();
  assert.ok(login.includes(`href="${assets.css}"`));
  for (const path of [assets.css, assets.js, "/static/site.css"]) {
    const response = await fetch(site + path);
    assert.equal(response.status, 200, path);
    assert.equal(response.headers.get("cache-control"), `public, max-age=${path === "/static/site.css" ? 3600 : 31536000}`);
    await response.text();
  }
});

test("migrated templates preserve article HTML, feedback, pagination and escaped links", () => {
  const article = renderPage({ template: "article", shell, data: {
    Title: "标题 <检查>", ID: 9, Rich: true, HTML: "<h2>正文</h2>", Video: "javascript:alert(1)",
    SignedIn: true, UserID: 7,
    Comments: [{ ID: 8, UserID: 7, CanEdit: true, Author: "读者", Body: "评论", CreatedAt: "2026-06-07T08:00:00+08:00" }],
  } });
  for (const text of ["标题 &lt;检查&gt;", "<h2>正文</h2>", "2026-06-07", "编辑", '"name":"\\u003c开物录\\u003e"']) {
    assert.ok(article.includes(text), text);
  }
  assert.ok(!article.includes("javascript:alert"));
  const feedback = renderPage({ template: "feedback", shell, data: {
    SignedIn: true, Items: [{ body: "留言内容", status: "done", createdAt: "2026-06-07T08:00:00+08:00" }],
  } });
  assert.ok(feedback.includes("已处理"));
  assert.ok(feedback.includes("留言内容"));
  const category = renderPage({ template: "category", shell, data: {
    Name: "AI", Section: { Name: "AI", Tabs: [{ Kind: "links", On: true, PageCount: 2, Total: 29, Page: 1 }] },
  } });
  assert.ok(category.includes("共 29 个网站"));
});

test("site details publish the guide and its metadata in the server-rendered HTML", () => {
  const description = "Poe 是 Quora 提供的多模型 AI 聊天平台。";
  const guide = renderPage({ template: "site", shell: { ...shell, Title: "Poe｜多模型 AI 聊天", Description: description }, data: {
    Name: "Poe", Desc: "多模型 AI 聊天", ID: 29, External: true,
    BodyHTML: '<p>首段介绍</p><h3>Poe 免费吗？</h3><p><a href="https://poe.com/explore">机器人目录</a></p>',
  } });
  assert.ok(guide.includes(`<meta name="description" content="${description}">`));
  assert.ok(guide.includes("<h3>Poe 免费吗？</h3>"));
  assert.ok(guide.includes('href="https://poe.com/explore">机器人目录</a>'));
  assert.ok(guide.includes('href="/go/29"'));
  const short = renderPage({ template: "site", shell, data: { Name: "Poe", Desc: "多模型 AI 聊天" } });
  assert.ok(!short.includes('class="site-description"'));
});

test("internet ranking renders a complete month with its attribution and category filter", () => {
  const selected = { month: "2026-10", source: "测试数据源", sourceUrl: "https://example.org/methodology" };
  const page = renderPage({ template: "internetRank", shell, data: {
    Title: "互联网月访问量排行榜", Months: [selected], Selected: selected, Categories: ["搜索"], Category: "搜索",
    Entries: [{ Domain: "example.org", Name: "示例网站", Category: "搜索", Visits: 1493200 }],
  } });
  for (const text of ['value="2026-10" selected', 'value="搜索" selected', "1,493,200", "https://example.org/methodology", "https://example.org/", "全球桌面及移动网页"]) {
    assert.ok(page.includes(text), text);
  }
  const empty = renderPage({ template: "internetRank", shell, data: { Title: "互联网月访问量排行榜" } });
  assert.ok(empty.includes("月榜还没有发布"));
  assert.ok(!empty.includes("1,493,200"));
});

test("public navigation and board entries render from the Go JSON field names", () => {
  const navigation = {
    ...shell,
    Sidebar: [{ name: "AI 工具", href: "/c/ai", tool: true, icon: "tool" }],
    Launcher: [{ name: "简约模式", href: "#simple", icon: "simple" }],
  };
  const boards = [{ Key: "local", Name: "热搜", On: true, Items: [{
    title: "热搜条目", href: "https://example.test/news", heat: "99", external: true,
  }] }];
  const home = renderPage({ template: "home", shell: navigation, data: {
    Tasks: [{ name: "学习", href: "/c/learn" }],
    Boards: boards,
    Today: [{ year: 1984, text: "历史事件" }],
    Sections: [{ Slug: "collection", Groups: [{ name: "主题", tags: ["设计"] }] }],
    ArticleSections: [{ Slug: "daily", Name: "每日一篇", More: "/c/daily", Tabs: [{
      Slug: "daily", Kind: "articles", On: true, Posts: [{ ID: 42, Title: "一篇文章" }],
    }] }],
  } });
  assert.match(home, /href="\/c\/ai">AI 工具<\/a>/);
  assert.match(home, /data-simple>.*简约模式<\/button>/);
  assert.ok(home.includes('href="/c/learn">学习</a>'));
  assert.ok(home.includes("<b>1984</b> 历史事件"));
  assert.ok(home.includes('href="/search?q=%E8%AE%BE%E8%AE%A1">设计</a>'));
  for (const html of [home, renderPage({ template: "hot", shell: navigation, data: { Boards: boards } })]) {
    assert.match(html, /href="https:\/\/example\.test\/news"[^>]*>热搜条目<\/a>/);
  }
  assert.match(home, /id="sec-latest"[\s\S]*id="sec-daily"[\s\S]*href="\/a\/42"[\s\S]*<footer/);
});
