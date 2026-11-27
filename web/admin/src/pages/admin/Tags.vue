<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useConfirm } from "primevue/useconfirm";
import { useToast } from "primevue/usetoast";
import { getJSON } from "../../api";

type Row = { id: number; group: string; name: string; sort: number };

const toast = useToast();
const confirm = useConfirm();
const drawer = ref(false);
const loading = ref(true);
const saving = ref(false);
const formError = ref("");
const rows = ref<Row[]>([]);
const error = ref("");
const editing = ref(0);
const form = reactive({ group: "", name: "", sort: 1 });

function reset() {
  editing.value = 0;
  form.group = "";
  form.name = "";
  form.sort = 1;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    rows.value = await getJSON<Row[]>("/api/admin/tags");
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
    await getJSON("/api/admin/tags" + (editing.value ? "/" + editing.value : ""), {
      method: editing.value ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    drawer.value = false;
    toast.add({ severity: "success", summary: "标签已保存", life: 2400 });
    await load();
  } catch (failure) {
    formError.value = failure instanceof Error ? failure.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

function remove(row: Row) {
  confirm.require({
    header: "删除标签",
    message: "确定删除「" + row.name + "」？",
    acceptLabel: "删除", rejectLabel: "取消",
    acceptProps: { severity: "danger" },
    accept: async () => {
      error.value = "";
      try {
        await getJSON("/api/admin/tags/" + row.id, { method: "DELETE" });
        toast.add({ severity: "success", summary: "标签已删除", life: 2400 });
        await load();
      } catch (failure) {
        error.value = failure instanceof Error ? failure.message : "删除失败";
      }
    },
  });
}

onMounted(load);
</script>

<template>
  <header class="admin-page-header"><div><h1>标签管理</h1><p>维护网址集使用的标签和分组。</p></div><PButton label="添加标签" icon="pi pi-plus" @click="add" /></header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <PDataTable :value="rows" :loading="loading" data-key="id" class="admin-table" scrollable table-style="min-width: 460px">
        <template #empty>暂无标签。</template>
        <PColumn field="name" header="标签名称" /><PColumn field="group" header="分组" /><PColumn field="sort" header="排序" />
        <PColumn header="操作" style="min-width: 145px"><template #body="{ data }"><div class="admin-actions"><PButton size="small" text label="编辑" icon="pi pi-pencil" @click="edit(data)" /><PButton size="small" text severity="danger" label="删除" @click="remove(data)" /></div></template></PColumn>
      </PDataTable>
    </template>
  </PCard>
  <PDrawer aria-label="标签编辑" v-model:visible="drawer" :header="editing ? '编辑标签' : '添加标签'" position="right" class="admin-drawer" :dismissable="false">
    <PMessage v-if="formError" severity="error" class="admin-alert">{{ formError }}</PMessage>
    <form id="admin-tag-editor" class="admin-form-grid" @submit.prevent="save">
      <div class="admin-field"><label for="tag-name">标签名称</label><PInputText id="tag-name" v-model="form.name" required /></div>
      <div class="admin-field"><label for="tag-group">分组</label><PInputText id="tag-group" v-model="form.group" required /></div>
      <div class="admin-field"><label for="tag-sort">排序</label><PInputNumber input-id="tag-sort" v-model="form.sort" :use-grouping="false" /></div>
    </form>
    <template #footer><div class="admin-actions admin-drawer-actions"><PButton label="取消" severity="secondary" outlined :disabled="saving" @click="drawer = false" /><PButton type="submit" form="admin-tag-editor" label="保存标签" :loading="saving" /></div></template>
  </PDrawer>
</template>
