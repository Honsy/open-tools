<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useConfirm } from "primevue/useconfirm";
import { useToast } from "primevue/usetoast";
import { getJSON } from "../../api";

type Row = {
  id: number;
  slug: string;
  name: string;
  parentSlug: string;
  kind: string;
  contentKind: string;
  sort: number;
  showOnHome: boolean;
};

const toast = useToast();
const confirm = useConfirm();
const drawer = ref(false);
const loading = ref(true);
const saving = ref(false);
const formError = ref("");
const rows = ref<Row[]>([]);
const error = ref("");
const editing = ref(0);
const form = reactive({
  slug: "",
  name: "",
  parentSlug: "",
  kind: "tab",
  contentKind: "links",
  sort: 1,
  showOnHome: false,
});

function reset() {
  editing.value = 0;
  Object.assign(form, {
    slug: "",
    name: "",
    parentSlug: "",
    kind: "tab",
    contentKind: "links",
    sort: 1,
    showOnHome: false,
  });
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    rows.value = await getJSON<Row[]>("/api/admin/categories");
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "读取列表失败";
  } finally {
    loading.value = false;
  }
}

function add() {
  reset();
  formError.value = "";
  drawer.value = true;
}

function edit(row: Row) {
  formError.value = "";
  drawer.value = true;
  editing.value = row.id;
  Object.assign(form, row);
}

async function save() {
  if (saving.value) return;
  saving.value = true;
  formError.value = "";
  try {
    await getJSON("/api/admin/categories" + (editing.value ? "/" + editing.value : ""), {
      method: editing.value ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    drawer.value = false;
    toast.add({ severity: "success", summary: "分类已保存", life: 2400 });
    await load();
  } catch (failure) {
    formError.value = failure instanceof Error ? failure.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

function remove(row: Row) {
  confirm.require({
    header: "删除分类",
    message: "确定删除「" + row.name + "」？",
    acceptLabel: "删除", rejectLabel: "取消",
    acceptProps: { severity: "danger" },
    accept: async () => {
      error.value = "";
      try {
        await getJSON("/api/admin/categories/" + row.id, { method: "DELETE" });
        toast.add({ severity: "success", summary: "分类已删除", life: 2400 });
        await load();
      } catch (failure) {
        error.value = failure instanceof Error ? failure.message : "删除失败";
      }
    },
  });
}

const kinds = [{ name: "分区", value: "section" }, { name: "子分类", value: "tab" }];
const contentKinds = [
  { name: "网址", value: "links" }, { name: "文章", value: "articles" },
  { name: "标签", value: "tags" }, { name: "混合", value: "mixed" },
];
const parentName = (slug: string) => rows.value.find(row => row.slug === slug)?.name || slug || "—";
onMounted(load);
</script>

<template>
  <header class="admin-page-header"><div><h1>分类管理</h1><p>维护导航分区、子分类和首页展示顺序。</p></div><PButton label="添加分类" icon="pi pi-plus" @click="add" /></header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <PDataTable :value="rows" :loading="loading" data-key="id" class="admin-table" scrollable table-style="min-width: 710px">
        <template #empty>暂无分类。</template>
        <PColumn header="分类"><template #body="{ data }"><div class="admin-primary">{{ data.name }}</div><div class="admin-muted">{{ data.slug }}</div></template></PColumn>
        <PColumn header="上级"><template #body="{ data }">{{ parentName(data.parentSlug) }}</template></PColumn>
        <PColumn header="层级"><template #body="{ data }"><PTag :value="data.kind === 'section' ? '分区' : '子分类'" :severity="data.kind === 'section' ? 'info' : 'secondary'" /></template></PColumn>
        <PColumn header="内容"><template #body="{ data }">{{ contentKinds.find(item => item.value === data.contentKind)?.name ?? data.contentKind }}</template></PColumn>
        <PColumn field="sort" header="排序" />
        <PColumn header="首页"><template #body="{ data }">{{ data.showOnHome ? "展示" : "—" }}</template></PColumn>
        <PColumn header="操作" style="min-width: 145px"><template #body="{ data }"><div class="admin-actions"><PButton size="small" text label="编辑" icon="pi pi-pencil" @click="edit(data)" /><PButton size="small" text severity="danger" label="删除" @click="remove(data)" /></div></template></PColumn>
      </PDataTable>
    </template>
  </PCard>
  <PDrawer aria-label="分类编辑" v-model:visible="drawer" :header="editing ? '编辑分类' : '添加分类'" position="right" class="admin-drawer" :dismissable="false">
    <PMessage v-if="formError" severity="error" class="admin-alert">{{ formError }}</PMessage>
    <form id="admin-category-editor" class="admin-form-grid" @submit.prevent="save">
      <div class="admin-field"><label for="category-name">分类名称</label><PInputText id="category-name" v-model="form.name" required /></div>
      <div class="admin-field"><label for="category-slug">标识</label><PInputText id="category-slug" v-model="form.slug" :disabled="!!editing" required placeholder="video" /></div>
      <div class="admin-field"><label for="category-kind">层级</label><PSelect input-id="category-kind" aria-label="层级" v-model="form.kind" :options="kinds" option-label="name" option-value="value" /></div>
      <div class="admin-field"><label for="category-content">内容类型</label><PSelect input-id="category-content" aria-label="内容类型" v-model="form.contentKind" :options="contentKinds" option-label="name" option-value="value" /></div>
      <div class="admin-field"><label for="category-parent">上级标识</label><PInputText id="category-parent" v-model="form.parentSlug" placeholder="分区留空" /></div>
      <div class="admin-field"><label for="category-sort">排序</label><PInputNumber input-id="category-sort" v-model="form.sort" :use-grouping="false" /></div>
      <div class="admin-actions admin-form-wide"><PCheckbox input-id="category-home" v-model="form.showOnHome" binary /><label for="category-home">显示在首页</label></div>
    </form>
    <template #footer><div class="admin-actions admin-drawer-actions"><PButton label="取消" severity="secondary" outlined :disabled="saving" @click="drawer = false" /><PButton type="submit" form="admin-category-editor" label="保存分类" :loading="saving" /></div></template>
  </PDrawer>
</template>
