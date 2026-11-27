import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import PrimeVue from "primevue/config";
import Aura from "@primeuix/themes/aura";
import { definePreset } from "@primeuix/themes";
import Button from "primevue/button";
import Card from "primevue/card";
import Checkbox from "primevue/checkbox";
import Column from "primevue/column";
import ConfirmationService from "primevue/confirmationservice";
import ConfirmDialog from "primevue/confirmdialog";
import DataTable from "primevue/datatable";
import Dialog from "primevue/dialog";
import Drawer from "primevue/drawer";
import InputNumber from "primevue/inputnumber";
import InputText from "primevue/inputtext";
import Message from "primevue/message";
import Paginator from "primevue/paginator";
import Select from "primevue/select";
import SelectButton from "primevue/selectbutton";
import Tag from "primevue/tag";
import Textarea from "primevue/textarea";
import Toast from "primevue/toast";
import ToastService from "primevue/toastservice";
import "primeicons/primeicons.css";
import App from "./App.vue";
import { routes } from "./router";
import { adminPath } from "./adminPath";
import { adminName } from "./api";
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
  const accessPath = adminPath("access");
  if (to.path === accessPath) return true;
  try {
    const response = await fetch("/api/admin/me");
    if (response.ok) {
      adminName.value = (await response.json()).username;
      return true;
    }
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

const theme = definePreset(Aura, {
  semantic: {
    primary: {
      50: "{blue.50}", 100: "{blue.100}", 200: "{blue.200}", 300: "{blue.300}",
      400: "{blue.400}", 500: "#0f62fe", 600: "#004ccd", 700: "{blue.700}",
      800: "{blue.800}", 900: "{blue.900}", 950: "{blue.950}",
    },
  },
});
const app = createApp(App).use(router).use(PrimeVue, {
  theme: { preset: theme, options: { darkModeSelector: false } },
  locale: {
    accept: "确认", reject: "取消", close: "关闭", emptyMessage: "暂无记录",
    emptyFilterMessage: "没有匹配项", searchMessage: "找到 {0} 项",
  },
}).use(ToastService).use(ConfirmationService);
for (const [name, component] of Object.entries({
  PButton: Button, PCard: Card, PCheckbox: Checkbox, PColumn: Column,
  PConfirmDialog: ConfirmDialog, PDataTable: DataTable, PDialog: Dialog, PDrawer: Drawer,
  PInputNumber: InputNumber, PInputText: InputText, PMessage: Message, PPaginator: Paginator,
  PSelect: Select, PSelectButton: SelectButton, PTag: Tag, PTextarea: Textarea, PToast: Toast,
})) app.component(name, component);
app.mount("#app");
