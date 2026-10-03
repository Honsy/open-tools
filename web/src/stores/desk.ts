import { defineStore } from "pinia";
import type { LinkItem } from "../api";

const recentKey = "ot-recent";
const customKey = "ot-custom";

function load(key: string): LinkItem[] {
  try {
    const raw = localStorage.getItem(key);
    const parsed = raw ? (JSON.parse(raw) as LinkItem[]) : [];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export const useDeskStore = defineStore("desk", {
  state: () => ({
    recent: load(recentKey),
    custom: load(customKey),
  }),
  actions: {
    remember(link: LinkItem) {
      this.recent = [link, ...this.recent.filter((item) => item.id !== link.id)].slice(0, 12);
      localStorage.setItem(recentKey, JSON.stringify(this.recent));
    },
    toggle(link: LinkItem) {
      const exists = this.custom.some((item) => item.id === link.id);
      this.custom = exists
        ? this.custom.filter((item) => item.id !== link.id)
        : [...this.custom, link].slice(-16);
      localStorage.setItem(customKey, JSON.stringify(this.custom));
    },
    isCustom(id: number) {
      return this.custom.some((item) => item.id === id);
    },
    refresh(links: LinkItem[]) {
      const byId = new Map(links.map((link) => [link.id, link]));
      const merge = (list: LinkItem[]) =>
        list.map((item) => {
          const fresh = byId.get(item.id);
          return fresh ? { ...item, ...fresh } : item;
        });
      this.recent = merge(this.recent);
      this.custom = merge(this.custom);
      localStorage.setItem(recentKey, JSON.stringify(this.recent));
      localStorage.setItem(customKey, JSON.stringify(this.custom));
    },
  },
});
