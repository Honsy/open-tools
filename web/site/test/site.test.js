import assert from "node:assert/strict";
import http from "node:http";
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
