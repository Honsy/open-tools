import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import App from "./App.vue";
import { routes } from "./router";
import { getJSON } from "./api";
import "./styles.css";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to) {
    if (to.hash) return false;
    return { top: 0 };
  },
});

router.beforeEach(async (to) => {
  if (!to.path.startsWith("/admin")) return true;
  if (to.path === "/admin/login") {
    try {
      await getJSON("/api/admin/me");
      return "/admin";
    } catch {
      return true;
    }
  }
  try {
    await getJSON("/api/admin/me");
    return true;
  } catch {
    return { path: "/admin/login", query: { next: to.fullPath } };
  }
});

createApp(App).use(createPinia()).use(router).mount("#app");
