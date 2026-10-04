import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import App from "./App.vue";
import { routes } from "./router";
import { getJSON } from "./api";
import { adminApp, adminPath } from "./adminPath";
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
  const inAdmin = adminApp || to.path === "/admin" || to.path.startsWith("/admin/");
  if (!inAdmin) return true;
  const loginPath = adminPath("login");
  if (to.path === loginPath) {
    try {
      await getJSON("/api/admin/me");
      return adminPath();
    } catch {
      return true;
    }
  }
  try {
    await getJSON("/api/admin/me");
    return true;
  } catch {
    return { path: loginPath, query: { next: to.fullPath } };
  }
});

createApp(App).use(createPinia()).use(router).mount("#app");
