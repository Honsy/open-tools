import { createReadStream, readFileSync } from "node:fs";
import { stat } from "node:fs/promises";
import http from "node:http";
import https from "node:https";
import { extname, join } from "node:path";
import { pipeline } from "node:stream/promises";
import { fileURLToPath } from "node:url";
import { renderPage } from "./render.js";

const pageMediaType = "application/vnd.kaiwulu.page.v1+json";
const pageLimit = 16 * 1024 * 1024;
const privateSiteAPI = "/_site";
const staticDefault = fileURLToPath(new URL("./static/", import.meta.url));
const assetTypes = {
  ".css": "text/css; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".svg": "image/svg+xml",
};
const hopHeaders = new Set([
  "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
  "te", "trailer", "transfer-encoding", "upgrade",
]);

export function createSite({
  apiOrigin = process.env.SITE_API_ORIGIN || "http://127.0.0.1:8080",
  staticDir = process.env.SITE_STATIC_DIR || staticDefault,
} = {}) {
  const api = new URL(apiOrigin);
  if (!["http:", "https:"].includes(api.protocol) || api.username || api.password ||
      api.pathname !== "/" || api.search || api.hash) {
    throw new Error("SITE_API_ORIGIN must be an HTTP(S) origin");
  }
  let assets;
  try {
    assets = JSON.parse(readFileSync(join(staticDir, "fingerprinted", "manifest.json"), "utf8"));
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
  }
  return http.createServer({ requestTimeout: 120_000 }, async (req, res) => {
    let pathname = "";
    try {
      const url = new URL(req.url, "http://localhost");
      pathname = decodeURIComponent(url.pathname);
      if (pathname === privateSiteAPI || pathname.startsWith(privateSiteAPI + "/")) {
        return send(res, 404, "404 page not found\n");
      }
      if (pathname === "/admin" || pathname === "/tools") return redirect(res, pathname + "/" + url.search, 301);
      if (pathname === "/favicon.ico") return redirect(res, "/favicon.svg", 302);
      if (pathname === "/robots.txt") return send(res, 200, robots(origin(req)));
      if (pathname === "/favicon.svg" || pathname === "/baidu_verify_codeva-LbSF6vH8BX.html") {
        return await serveAsset(req, res, staticDir, pathname.slice(1));
      }
      if (pathname.startsWith("/static/")) {
        return await serveAsset(req, res, staticDir, pathname.slice("/static/".length));
      }
      const forwardAPI = pathname.startsWith("/api/") || pathname.startsWith("/go/") || pathname === "/ico" || pathname === "/health";
      await serveUpstream(req, res, api, url, !forwardAPI, assets);
    } catch (error) {
      if (res.destroyed) return;
      if (res.headersSent) res.destroy();
      else {
        console.error("Public site request failed", req.method, pathname, error.message);
        send(res, error instanceof URIError ? 400 : 502, "页面暂时不可用\n");
      }
    }
  });
}

function origin(req) {
  const protocol = firstHeader(req.headers["x-forwarded-proto"]);
  const scheme = protocol === "https" || protocol === "http" ? protocol : "http";
  const host = firstHeader(req.headers["x-forwarded-host"]) || req.headers.host;
  return scheme + "://" + host;
}

function firstHeader(value) {
  return typeof value === "string" ? value.split(",", 1)[0].trim() : "";
}

function copyHeaders(headers) {
  const connection = new Set(firstHeaderList(headers.connection));
  return Object.fromEntries(Object.entries(headers).filter(([key]) => !hopHeaders.has(key) && !connection.has(key)));
}

function firstHeaderList(value) {
  return typeof value === "string" ? value.toLowerCase().split(",").map(part => part.trim()) : [];
}

async function serveUpstream(req, res, api, publicURL, page, assets) {
  const target = new URL(api);
  target.pathname = (page ? privateSiteAPI : "") + publicURL.pathname;
  target.search = publicURL.search;
  const publicOrigin = new URL(origin(req));
  const headers = copyHeaders(req.headers);
  headers.host = api.host;
  headers["x-forwarded-host"] = publicOrigin.host;
  headers["x-forwarded-proto"] = publicOrigin.protocol.slice(0, -1);
  headers["x-forwarded-for"] = req.socket.remoteAddress;
  delete headers.forwarded;
  if (page) delete headers["accept-encoding"];

  const transport = api.protocol === "https:" ? https : http;
  let request;
  const response = await new Promise((resolve, reject) => {
    request = transport.request(target, { method: req.method, headers, timeout: 120_000 }, resolve);
    request.on("error", reject);
    request.on("timeout", () => request.destroy(new Error("API timeout")));
    res.once("close", () => { if (!res.writableEnded) request.destroy(); });
    req.pipe(request);
  });

  const responseHeaders = copyHeaders(response.headers);
  const mediaType = firstHeader(response.headers["content-type"]).split(";", 1)[0];
  if (!page || mediaType !== pageMediaType) {
    if (page && responseHeaders.location?.startsWith(privateSiteAPI + "/")) {
      responseHeaders.location = responseHeaders.location.slice(privateSiteAPI.length);
    }
    res.writeHead(response.statusCode, responseHeaders);
    await pipeline(response, res);
    return;
  }

  const data = await readPageData(response);
  const sitemap = data.template === "sitemap";
  const body = sitemap ? renderSitemap(data.data, publicOrigin.origin) : renderPage(data, assets);
  responseHeaders["content-type"] = sitemap ? "application/xml; charset=utf-8" : "text/html; charset=utf-8";
  responseHeaders["content-length"] = Buffer.byteLength(body);
  delete responseHeaders["content-encoding"];
  delete responseHeaders.etag;
  res.writeHead(response.statusCode, responseHeaders);
  res.end(req.method === "HEAD" ? undefined : body);
}

async function readPageData(response) {
  const chunks = [];
  let bytes = 0;
  for await (const chunk of response) {
    bytes += chunk.length;
    if (bytes > pageLimit) throw new Error("Page JSON exceeded the size limit");
    chunks.push(chunk);
  }
  return JSON.parse(Buffer.concat(chunks).toString(), (_key, value, context) => {
    if (typeof value === "number" && !Number.isSafeInteger(value) && /^-?\d+$/.test(context.source)) {
      return BigInt(context.source);
    }
    return value;
  });
}

async function serveAsset(req, res, directory, name) {
  const type = assetTypes[extname(name)];
  if (!type || name.includes("\\") || name.split("/").some(part => !part || part === "." || part === "..")) {
    return send(res, 404, "404 page not found\n");
  }
  if (req.method !== "GET" && req.method !== "HEAD") return send(res, 405, "Method not allowed\n");
  const file = join(directory, name);
  let metadata;
  try {
    metadata = await stat(file);
    if (!metadata.isFile()) return send(res, 404, "404 page not found\n");
  } catch (error) {
    if (error.code === "ENOENT" || error.code === "ENOTDIR") return send(res, 404, "404 page not found\n");
    throw error;
  }
  const maxAge = /^fingerprinted\/site\.[0-9a-f]{16}\.(css|js)$/.test(name) ? 31536000 : 3600;
  const headers = { "cache-control": `public, max-age=${maxAge}`, "last-modified": metadata.mtime.toUTCString() };
  if (Date.parse(req.headers["if-modified-since"]) >= Math.floor(metadata.mtimeMs / 1000) * 1000) {
    res.writeHead(304, headers).end();
    return;
  }
  res.writeHead(200, { ...headers, "content-type": type, "content-length": metadata.size });
  if (req.method === "HEAD") res.end();
  else await pipeline(createReadStream(file), res);
}

function renderSitemap(paths, publicOrigin) {
  const nodes = paths.filter(path => path.startsWith("/") && !path.startsWith("//"))
    .map(path => "<url><loc>" + escapeXml(publicOrigin + path) + "</loc></url>").join("");
  return '<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">' + nodes + "</urlset>";
}

function escapeXml(text) {
  const escape = { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&apos;" };
  return text.replace(/[&<>"']/g, character => escape[character]);
}

function robots(publicOrigin) {
  return "User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /api/\nDisallow: /go/\nDisallow: /login\nDisallow: /register\nDisallow: /forgot\nDisallow: /reset\nDisallow: /mine\nDisallow: /random\nDisallow: /tools/\nDisallow: /feedback\nSitemap: " + publicOrigin + "/sitemap.xml\n";
}

function send(res, status, body) {
  res.writeHead(status, { "content-type": "text/plain; charset=utf-8" });
  res.end(body);
}

function redirect(res, path, status) {
  res.writeHead(status, { location: path });
  res.end();
}
