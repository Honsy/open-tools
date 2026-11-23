import { readFileSync, readdirSync } from "node:fs";
import ejs from "ejs";

const templates = new Map();
const templateDir = new URL("./templates/", import.meta.url);
for (const file of readdirSync(templateDir)) {
  if (!file.endsWith(".ejs")) continue;
  const name = file.slice(0, -4);
  const source = readFileSync(new URL(file, templateDir), "utf8");
  templates.set(name, ejs.compile(source, { filename: file }));
}

export function renderPage({ template, shell, data }) {
  const body = renderTemplate(template, data);
  return template === "authpage" ? body : renderTemplate("layout", { ...shell, Content: body });
}

function renderTemplate(name, data) {
  const render = templates.get(name);
  if (!render) throw new Error("Unknown public template");
  return render({ page: data, partial: renderTemplate, shown, safeURL, date, query, jsonScript });
}

function shown(value) {
  return Array.isArray(value) ? value.length > 0 : Boolean(value);
}

function safeURL(value) {
  if (typeof value !== "string" || /[\\\x00-\x20]/.test(value)) return "#";
  if ((value.startsWith("/") && !value.startsWith("//")) || /^[?#]/.test(value)) return value;
  if (/^https?:\/\/[^/?#]+/i.test(value) || /^mailto:[^@]+@[^@]+$/i.test(value)) return value;
  return "#";
}

function date(value, full = false) {
  if (typeof value !== "string" || !/^\d{4}-\d{2}-\d{2}T/.test(value)) return "";
  return full ? value : value.slice(0, 10);
}

function query(text) {
  return encodeURIComponent(text);
}

function jsonScript(value) {
  const escape = { "<": "\\u003c", ">": "\\u003e", "&": "\\u0026", "\u2028": "\\u2028", "\u2029": "\\u2029" };
  return JSON.stringify(value, (_key, item) => typeof item === "bigint" ? String(item) : item)
    .replace(/[<>&\u2028\u2029]/g, character => escape[character]);
}
