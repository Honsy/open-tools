export async function getJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const message = typeof data.error === "string" ? data.error : "请求失败";
    throw new Error(message);
  }
  return data as T;
}

export function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const m = `${date.getMonth() + 1}`.padStart(2, "0");
  const d = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${m}-${d}`;
}
import { ref } from "vue";

export const adminName = ref("");

export type AdminPage<T> = { items: T[]; total: number; page: number; pageSize: number };
export type AdminCategory = { slug: string; name: string; parentSlug: string; kind: string; contentKind: string };

export function categoryLabel(category: AdminCategory, catalog: AdminCategory[]) {
  const parent = catalog.find(item => item.slug === category.parentSlug);
  return parent ? `${parent.name} / ${category.name}` : category.name;
}
