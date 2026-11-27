<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useToast } from "primevue/usetoast";
import { useConfirm } from "primevue/useconfirm";
import { getJSON, categoryLabel, type AdminCategory, type AdminPage } from "../../api";

const empty = {
  slug: "", name: "", url: "", desc: "", tags: "", alias: "", lang: "", region: "",
  spare: "", icp: "", body: "", categorySlug: "", status: "online", sort: 1, pinSort: 0, pinned: false,
};
type Row = typeof empty & { id: number; clicks: number; views: number; checkFails?: number };

const statuses = [
  { name: "上架", value: "online", color: "success" },
  { name: "待审", value: "pending", color: "warn" },
  { name: "下架", value: "off", color: "secondary" },
];
const toast = useToast();
const confirm = useConfirm();
const categories = ref<AdminCategory[]>([]);
const list = ref<AdminPage<Row>>({ items: [], total: 0, page: 1, pageSize: 25 });
const filters = reactive({ q: "", status: "", category: "" });
const applied = reactive({ ...filters });
const loading = ref(true);
const saving = ref(false);
const drawer = ref(false);
const editing = ref(0);
const error = ref("");
const formError = ref("");
const form = reactive({ ...empty });
let latestRequest = 0;

const options = computed(() => categories.value.filter(item => item.kind === "tab").map(item => ({
  value: item.slug, name: categoryLabel(item, categories.value),
})));
const categoryName = (slug: string) => options.value.find(item => item.value === slug)?.name || (slug === "inbox" ? "待分配" : slug);
const statusInfo = (value: string) => statuses.find(item => item.value === value);

async function load(page = list.value.page) {
  const request = ++latestRequest;
  loading.value = true;
  error.value = "";
  try {
    const query = new URLSearchParams({ ...applied, page: String(page) });
    const response = await getJSON<AdminPage<Row>>("/api/admin/links?" + query);
    if (request === latestRequest) list.value = response;
  } catch (failure) {
    if (request === latestRequest) error.value = failure instanceof Error ? failure.message : "读取网址失败";
  } finally {
    if (request === latestRequest) loading.value = false;
  }
}

function search() {
  Object.assign(applied, { q: filters.q, category: filters.category || "", status: filters.status || "" });
  void load(1);
}

function openEditor(row?: Row) {
  editing.value = row?.id ?? 0;
  formError.value = "";
  Object.assign(form, empty, row ?? { categorySlug: applied.category, status: "online" });
  drawer.value = true;
}

async function save() {
  if (saving.value) return;
  if (!form.name.trim() || !form.url.trim() || !options.value.some(item => item.value === form.categorySlug)) {
    formError.value = "请填写名称和网址，并选择一个子分类。";
    return;
  }
  saving.value = true;
  formError.value = "";
  try {
    await getJSON("/api/admin/links" + (editing.value ? "/" + editing.value : ""), {
      method: editing.value ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    toast.add({ severity: "success", summary: editing.value ? "网址已保存" : "网址已添加", life: 2400 });
    drawer.value = false;
    await load(editing.value ? list.value.page : 1);
  } catch (failure) {
    formError.value = failure instanceof Error ? failure.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

function remove(row: Row) {
  confirm.require({
    header: "删除网址",
    message: `删除「${row.name}」后，前台详情地址将无法访问。`,
    icon: "pi pi-exclamation-triangle",
    acceptLabel: "删除", rejectLabel: "取消",
    acceptProps: { severity: "danger" }, rejectProps: { severity: "secondary", outlined: true },
    accept: async () => {
      try {
        await getJSON("/api/admin/links/" + row.id, { method: "DELETE" });
        toast.add({ severity: "success", summary: "网址已删除", life: 2400 });
        await load();
      } catch (failure) {
        error.value = failure instanceof Error ? failure.message : "删除失败";
      }
    },
  });
}

onMounted(() => {
  void getJSON<AdminCategory[]>("/api/admin/categories").then(value => { categories.value = value; }).catch(failure => {
    error.value = failure instanceof Error ? failure.message : "读取分类失败";
  });
  void load(1);
});
</script>

<template>
  <header class="admin-page-header">
    <div><h1>网址管理</h1><p>管理网站介绍、分类和上架状态。</p></div>
    <PButton label="添加网址" icon="pi pi-plus" @click="openEditor()" />
  </header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <form class="admin-toolbar" role="search" @submit.prevent="search">
        <PInputText v-model="filters.q" aria-label="搜索网址" placeholder="搜索名称或网址" type="search" maxlength="254" />
        <PSelect v-model="filters.category" :options="options" option-label="name" option-value="value" show-clear filter aria-label="分类" placeholder="全部分类" @change="search" />
        <PSelect v-model="filters.status" :options="statuses" option-label="name" option-value="value" show-clear aria-label="状态" placeholder="全部状态" @change="search" />
        <PButton type="submit" label="搜索" icon="pi pi-search" severity="secondary" outlined />
      </form>
      <PDataTable :value="list.items" :loading="loading" data-key="id" class="admin-table admin-adaptive" scrollable table-style="min-width: 740px">
        <template #empty>没有找到符合条件的网址。</template>
        <PColumn header="网站" header-class="admin-mobile-title" body-class="admin-mobile-title" style="min-width: 240px; max-width: 430px">
          <template #body="{ data }"><div class="admin-primary">{{ data.name }}</div><div class="admin-muted admin-truncate" :title="data.url">{{ data.url }}</div><div class="admin-mobile-meta admin-muted">{{ categoryName(data.categorySlug) }} · {{ statusInfo(data.status)?.name ?? data.status }}</div></template>
        </PColumn>
        <PColumn header="分类" header-class="admin-hide-mobile" body-class="admin-hide-mobile" style="min-width: 120px"><template #body="{ data }">{{ categoryName(data.categorySlug) }}</template></PColumn>
        <PColumn header="状态" header-class="admin-hide-mobile" body-class="admin-hide-mobile" style="width: 115px">
          <template #body="{ data }"><PTag :value="statusInfo(data.status)?.name ?? data.status" :severity="statusInfo(data.status)?.color ?? 'secondary'" /><div v-if="data.checkFails" class="admin-muted">失效 {{ data.checkFails }} 次</div></template>
        </PColumn>
        <PColumn header="打开 / 浏览" header-class="admin-hide-mobile" body-class="admin-hide-mobile" style="min-width: 105px"><template #body="{ data }">{{ data.clicks }} / {{ data.views }}</template></PColumn>
        <PColumn header="操作" header-class="admin-mobile-action" body-class="admin-mobile-action" style="min-width: 145px"><template #body="{ data }">
          <div class="admin-actions"><PButton size="small" text label="编辑" icon="pi pi-pencil" @click="openEditor(data)" /><PButton size="small" text severity="danger" label="删除" @click="remove(data)" /></div>
        </template></PColumn>
      </PDataTable>
      <div class="admin-pagination"><span class="admin-muted">共 {{ list.total }} 条</span><PPaginator :first="(list.page - 1) * list.pageSize" :rows="list.pageSize" :total-records="list.total" :page-link-size="3" @page="load($event.page + 1)" /></div>
    </template>
  </PCard>

  <PDrawer aria-label="编辑网址" v-model:visible="drawer" :header="editing ? '编辑网址' : '添加网址'" position="right" class="admin-drawer" :dismissable="false">
    <PMessage v-if="formError" severity="error" class="admin-alert">{{ formError }}</PMessage>
    <form id="admin-link-editor" class="admin-form-grid" @submit.prevent="save">
      <div class="admin-field admin-form-wide"><label for="link-name">网站名称</label><PInputText id="link-name" v-model="form.name" required maxlength="40" /></div>
      <div class="admin-field admin-form-wide"><label for="link-url">网站地址</label><PInputText id="link-url" v-model="form.url" required maxlength="512" placeholder="https://" /></div>
      <div class="admin-field"><label for="link-category">分类</label><PSelect input-id="link-category" aria-label="分类" v-model="form.categorySlug" :options="options" option-label="name" option-value="value" filter placeholder="选择子分类" /></div>
      <div class="admin-field"><label for="link-status">状态</label><PSelect input-id="link-status" aria-label="状态" v-model="form.status" :options="statuses" option-label="name" option-value="value" /></div>
      <div class="admin-field admin-form-wide"><label for="link-desc">卡片简介</label><PInputText id="link-desc" v-model="form.desc" maxlength="80" /></div>
      <div class="admin-field admin-form-wide"><label for="link-body">网站介绍</label><PTextarea id="link-body" v-model="form.body" rows="8" maxlength="4000" placeholder="首段写简介。支持 Markdown 小标题和链接。" /></div>
      <h3 class="admin-form-section">收录信息</h3>
      <div class="admin-field"><label for="link-tags">标签</label><PInputText id="link-tags" v-model="form.tags" maxlength="40" /></div>
      <div class="admin-field"><label for="link-alias">别名</label><PInputText id="link-alias" v-model="form.alias" maxlength="20" /></div>
      <div class="admin-field"><label for="link-lang">语言</label><PInputText id="link-lang" v-model="form.lang" maxlength="20" /></div>
      <div class="admin-field"><label for="link-region">地区</label><PInputText id="link-region" v-model="form.region" maxlength="20" /></div>
      <div class="admin-field"><label for="link-spare">备用地址</label><PInputText id="link-spare" v-model="form.spare" /></div>
      <div class="admin-field"><label for="link-icp">备案</label><PInputText id="link-icp" v-model="form.icp" maxlength="40" /></div>
      <div class="admin-field"><label for="link-sort">排序</label><PInputNumber input-id="link-sort" v-model="form.sort" :use-grouping="false" /></div>
      <div class="admin-field"><label for="link-pin-sort">常用排序</label><PInputNumber input-id="link-pin-sort" v-model="form.pinSort" :use-grouping="false" /></div>
      <div class="admin-actions admin-form-wide"><PCheckbox input-id="link-pinned" v-model="form.pinned" binary /><label for="link-pinned">放进热门推荐</label></div>
    </form>
    <template #footer>
      <div class="admin-actions admin-drawer-actions">
        <PButton v-if="editing && form.slug && form.status === 'online'" as="a" :href="'/site/' + encodeURIComponent(form.slug)" target="_blank" rel="noopener" text label="查看详情页" />
        <PButton label="取消" severity="secondary" outlined :disabled="saving" @click="drawer = false" />
        <PButton type="submit" form="admin-link-editor" label="保存网址" :loading="saving" />
      </div>
    </template>
  </PDrawer>
</template>
