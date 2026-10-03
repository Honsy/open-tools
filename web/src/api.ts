export type LinkItem = {
  id: number;
  name: string;
  url: string;
  desc: string;
  clicks: number;
  categorySlug: string;
  slug?: string;
};

export type ArticleCard = {
  id: number;
  title: string;
  summary: string;
  views: number;
  createdAt: string;
};

export type Tab = {
  slug: string;
  name: string;
  kind: string;
  links?: LinkItem[];
  articles?: ArticleCard[];
};

export type TagGroup = { name: string; tags: string[] };

export type Section = {
  slug: string;
  name: string;
  kind: string;
  tabs?: Tab[];
  groups?: TagGroup[];
};

export type BoardItem = { title: string; heat: string; href: string; external?: boolean };
export type Board = { key: string; name: string; items: BoardItem[] };
export type TodayEvent = { year: number; text: string };

export type HomeData = {
  todayLabel: string;
  today: TodayEvent[];
  pinned: LinkItem[];
  sections: Section[];
  latest: LinkItem[];
  popular: LinkItem[];
  boards: Board[];
  liveHot?: boolean;
};

export type SideItem = { name: string; href: string; children?: SideItem[] };

export type Article = ArticleCard & { body: string };

export function adminToken() {
  return sessionStorage.getItem("ot-admin") || "";
}

export function setAdminToken(token: string) {
  if (token) sessionStorage.setItem("ot-admin", token);
  else sessionStorage.removeItem("ot-admin");
}

export async function getJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  const token = adminToken();
  if (token) headers.set("Authorization", "Bearer " + token);
  const res = await fetch(url, { ...init, headers });
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
