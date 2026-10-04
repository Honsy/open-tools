export const adminApp = import.meta.env.BASE_URL.startsWith("/admin");

export function adminPath(tail = "") {
  const prefix = adminApp ? "" : "/admin";
  if (!tail) return prefix || "/";
  return `${prefix}/${tail}`;
}
