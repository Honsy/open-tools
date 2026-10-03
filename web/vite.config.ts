import http from "node:http";
import { defineConfig, type ViteDevServer } from "vite";
import vue from "@vitejs/plugin-vue";

function fromGo(req: http.IncomingMessage, res: http.ServerResponse) {
  const headers = { ...req.headers };
  if (req.headers.host) headers["x-forwarded-host"] = req.headers.host;
  headers["x-forwarded-proto"] = "http";
  const upstream = http.request(
    {
      hostname: "127.0.0.1",
      port: 8080,
      path: req.url,
      method: req.method,
      headers,
    },
    (incoming) => {
      res.writeHead(incoming.statusCode || 502, incoming.headers);
      incoming.pipe(res);
    },
  );
  upstream.on("error", () => {
    if (!res.headersSent) res.statusCode = 502;
    res.end("接口没有起来");
  });
  req.pipe(upstream);
}

function isDocument(url: string) {
  const path = url.split("?")[0];
  return (
    path === "/" ||
    path.startsWith("/c/") ||
    path.startsWith("/a/") ||
    path.startsWith("/site/") ||
    path === "/search" ||
    path === "/submit" ||
    path.startsWith("/tools/") ||
    path.startsWith("/static/") ||
    path.startsWith("/ico") ||
    path === "/random" ||
    path === "/sitemap.xml" ||
    path === "/robots.txt" ||
    path.startsWith("/go/") ||
    path.startsWith("/api/")
  );
}

function ssrProxy() {
  return {
    name: "opentools-ssr",
    configureServer(server: ViteDevServer) {
      server.middlewares.use((req, res, next) => {
        if (!req.url || !isDocument(req.url)) {
          next();
          return;
        }
        fromGo(req, res);
      });
    },
  };
}

export default defineConfig({
  plugins: [ssrProxy(), vue()],
  server: { port: 5173 },
});
