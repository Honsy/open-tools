import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import App from "./App.vue";
import { routes } from "./router";
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
  const accessPath = adminPath("access");
  if (to.path === accessPath) return true;
  try {
    const response = await fetch("/api/admin/me");
    if (response.ok) return true;
    if (response.status === 401) {
      const next = router.resolve(to.fullPath).href;
      window.location.replace(`/login?${new URLSearchParams({ next })}`);
      return false;
    }
    if (response.status === 403) return accessPath;
  } catch {
    // The access page also handles an unavailable API.
  }
  return { path: accessPath, query: { unavailable: "1" } };
});

createApp(App).use(createPinia()).use(router).mount("#app");
