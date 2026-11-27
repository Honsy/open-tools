<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useConfirm } from "primevue/useconfirm";
import { useToast } from "primevue/usetoast";
import { getJSON } from "../../api";

const empty = { title: "", summary: "", body: "", format: "", category: "", coverUrl: "", videoUrl: "", featured: false, sort: 0, draft: false };
type Row = typeof empty & { id: number; views: number };

const toast = useToast();
const confirm = useConfirm();
const drawer = ref(false);
const loading = ref(true);
const saving = ref(false);
const formError = ref("");
const rows = ref<Row[]>([]);
const error = ref("");
const uploading = ref(false);
const editing = ref(0);
const form = reactive({ ...empty });

function reset() {
  editing.value = 0;
  Object.assign(form, empty);
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    rows.value = await getJSON<Row[]>("/api/admin/articles");
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
  const { title, summary, body, format, category, coverUrl, videoUrl, featured, sort, draft } = row;
  Object.assign(form, { title, summary, body, format, category, coverUrl, videoUrl, featured, sort, draft });
}

async function upload(event: Event, cover: boolean) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  const payload = new FormData();
  payload.set("image", file);
  formError.value = "";
  uploading.value = true;
  try {
    const image = await getJSON<{url: string}>("/api/admin/articles/images", { method: "POST", body: payload });
    if (cover) form.coverUrl = image.url;
    else if (form.format === "html") form.body += `\n<img src="${image.url}" alt="">`;
    else {
      form.format = "markdown";
      form.body += `\n\n![](${image.url})\n`;
    }
  } catch (failure) {
    formError.value = failure instanceof Error ? failure.message : "上传失败";
  } finally {
    uploading.value = false;
    input.value = "";
  }
}

async function save() {
  if (saving.value || uploading.value) return;
  saving.value = true;
  formError.value = "";
  try {
    await getJSON("/api/admin/articles" + (editing.value ? "/" + editing.value : ""), {
      method: editing.value ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    drawer.value = false;
    toast.add({ severity: "success", summary: "文章已保存", life: 2400 });
    await load();
  } catch (failure) {
    formError.value = failure instanceof Error ? failure.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

function remove(row: Row) {
  confirm.require({
    header: "删除文章",
    message: "确定删除「" + row.title + "」？",
    acceptLabel: "删除", rejectLabel: "取消",
    acceptProps: { severity: "danger" },
    accept: async () => {
      error.value = "";
      try {
        await getJSON("/api/admin/articles/" + row.id, { method: "DELETE" });
        toast.add({ severity: "success", summary: "文章已删除", life: 2400 });
        await load();
      } catch (failure) {
        error.value = failure instanceof Error ? failure.message : "删除失败";
      }
    },
  });
}

const formats = [
  { name: "纯文本", value: "" }, { name: "Markdown", value: "markdown" }, { name: "HTML", value: "html" },
];
onMounted(load);
</script>

<template>
  <header class="admin-page-header"><div><h1>文章管理</h1><p>编辑站点文章、封面和发布状态。</p></div><PButton label="新建文章" icon="pi pi-plus" @click="add" /></header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <PDataTable :value="rows" :loading="loading" data-key="id" class="admin-table" scrollable table-style="min-width: 590px">
        <template #empty>暂无文章。</template>
        <PColumn header="标题" style="min-width: 190px"><template #body="{ data }"><div class="admin-primary">{{ data.title }}</div><div class="admin-muted">{{ data.category }}</div></template></PColumn>
        <PColumn header="状态"><template #body="{ data }"><PTag :value="data.draft ? '草稿' : '公开'" :severity="data.draft ? 'secondary' : 'success'" /></template></PColumn>
        <PColumn field="views" header="阅读" />
        <PColumn header="操作" style="min-width: 145px"><template #body="{ data }"><div class="admin-actions"><PButton size="small" text label="编辑" icon="pi pi-pencil" @click="edit(data)" /><PButton size="small" text severity="danger" label="删除" @click="remove(data)" /></div></template></PColumn>
      </PDataTable>
    </template>
  </PCard>
  <PDrawer aria-label="文章编辑" v-model:visible="drawer" :header="editing ? '编辑文章' : '新建文章'" position="right" class="admin-drawer" :dismissable="false">
    <PMessage v-if="formError" severity="error" class="admin-alert">{{ formError }}</PMessage>
    <form id="admin-article-editor" class="admin-form-grid" @submit.prevent="save">
      <div class="admin-field admin-form-wide"><label for="article-title">标题</label><PInputText id="article-title" v-model="form.title" required maxlength="255" /></div>
      <div class="admin-field admin-form-wide"><label for="article-summary">摘要</label><PTextarea id="article-summary" v-model="form.summary" rows="2" /></div>
      <div class="admin-field"><label for="article-format">正文格式</label><PSelect input-id="article-format" aria-label="正文格式" v-model="form.format" :options="formats" option-label="name" option-value="value" /></div>
      <div class="admin-field"><label for="article-category">分类</label><PInputText id="article-category" v-model="form.category" maxlength="64" /></div>
      <div class="admin-field admin-form-wide"><label for="article-body">正文</label><PTextarea id="article-body" v-model="form.body" rows="10" required /></div>
      <div class="admin-field admin-form-wide"><label for="article-body-image">插入正文图片</label><input id="article-body-image" class="admin-file-input" type="file" accept="image/png,image/jpeg,image/gif,image/webp" :disabled="uploading" @change="upload($event, false)" /></div>
      <div class="admin-field admin-form-wide"><label for="article-cover">封面地址</label><PInputText id="article-cover" v-model="form.coverUrl" type="url" maxlength="512" /></div>
      <div class="admin-field admin-form-wide"><label for="article-cover-image">上传封面</label><input id="article-cover-image" class="admin-file-input" type="file" accept="image/png,image/jpeg,image/gif,image/webp" :disabled="uploading" @change="upload($event, true)" /></div>
      <div class="admin-field admin-form-wide"><label for="article-video">视频地址</label><PInputText id="article-video" v-model="form.videoUrl" type="url" maxlength="512" /></div>
      <div class="admin-field"><label for="article-sort">排序</label><PInputNumber input-id="article-sort" v-model="form.sort" :use-grouping="false" /></div>
      <div class="admin-actions admin-form-wide">
        <PCheckbox input-id="article-featured" v-model="form.featured" binary /><label for="article-featured">推荐</label>
        <PCheckbox input-id="article-draft" v-model="form.draft" binary /><label for="article-draft">草稿</label>
      </div>
    </form>
    <template #footer><div class="admin-actions admin-drawer-actions"><PButton label="取消" severity="secondary" outlined :disabled="saving || uploading" @click="drawer = false" /><PButton type="submit" form="admin-article-editor" label="保存文章" :loading="saving" :disabled="uploading" /></div></template>
  </PDrawer>
</template>
