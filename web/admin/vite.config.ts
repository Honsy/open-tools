import http from "node:http";
import { defineConfig, type ViteDevServer } from "vite";
import vue from "@vitejs/plugin-vue";

function devGateway() {
  return {
    name: "kaiwulu-dev-gateway",
    configureServer(server: ViteDevServer) {
      server.middlewares.use((req, res, next) => {
        if (!req.url) {
          next();
          return;
        }
        const path = req.url.split("?")[0];
        if (path === "/admin" || path.startsWith("/admin/")) {
          next();
          return;
        }
        const headers = { ...req.headers, "x-forwarded-host": req.headers.host, "x-forwarded-proto": "http" };
        const upstream = http.request({
          hostname: "127.0.0.1",
          port: 8081,
          path: req.url,
          method: req.method,
          headers,
        }, incoming => {
          res.writeHead(incoming.statusCode || 502, incoming.headers);
          incoming.pipe(res);
        });
        upstream.on("error", () => {
          if (res.headersSent) res.destroy();
          else {
            res.statusCode = 502;
            res.end("公开站服务未启动");
          }
        });
        req.pipe(upstream);
      });
    },
  };
}

export default defineConfig({
  base: "/admin/",
  plugins: [devGateway(), vue()],
  server: { port: 5173, strictPort: true },
});
