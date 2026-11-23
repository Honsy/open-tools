<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { getJSON } from "../../api";

const empty = { title: "", summary: "", body: "", format: "", category: "", coverUrl: "", videoUrl: "", featured: false, sort: 0, draft: false };
type Row = typeof empty & { id: number; views: number };

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
  rows.value = await getJSON<Row[]>("/api/admin/articles");
}

function edit(row: Row) {
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
  error.value = "";
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
    error.value = failure instanceof Error ? failure.message : "上传失败";
  } finally {
    uploading.value = false;
    input.value = "";
  }
}

async function save() {
  error.value = "";
  try {
    await getJSON("/api/admin/articles" + (editing.value ? "/" + editing.value : ""), {
      method: editing.value ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    reset();
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "没有存上";
  }
}

async function remove(row: Row) {
  if (!window.confirm("删除这篇文章？")) return;
  await getJSON("/api/admin/articles/" + row.id, { method: "DELETE" });
  await load();
}

onMounted(load);
</script>

<template>
  <h1>文章</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <form class="adm-form" @submit.prevent="save">
    <label class="wide">标题<input v-model="form.title" maxlength="255" required /></label>
    <label class="wide">摘要<input v-model="form.summary" /></label>
    <label>正文格式<select v-model="form.format"><option value="">纯文本</option><option value="markdown">Markdown</option><option value="html">HTML</option></select></label>
    <label>分类<input v-model="form.category" maxlength="64" /></label>
    <label>排序<input v-model.number="form.sort" type="number" /></label>
    <label><input v-model="form.featured" type="checkbox" />推荐</label>
    <label><input v-model="form.draft" type="checkbox" />草稿</label>
    <label class="wide">封面网址<input v-model="form.coverUrl" type="url" maxlength="512" /></label>
    <label>上传封面<input type="file" accept="image/png,image/jpeg,image/gif,image/webp" :disabled="uploading" @change="upload($event, true)" /></label>
    <label class="wide">视频网址<input v-model="form.videoUrl" type="url" maxlength="512" /></label>
    <label class="wide">正文<textarea v-model="form.body" required /></label>
    <label>插入正文图片<input type="file" accept="image/png,image/jpeg,image/gif,image/webp" :disabled="uploading" @change="upload($event, false)" /></label>
    <button class="btn primary" type="submit" :disabled="uploading">{{ editing ? "保存" : "添加" }}</button>
    <button v-if="editing" class="btn" type="button" @click="reset">取消</button>
  </form>
  <table>
    <thead>
      <tr><th>标题</th><th>状态</th><th>阅读</th><th></th></tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ row.title }}</td>
        <td>{{ row.draft ? "草稿" : "公开" }}</td>
        <td>{{ row.views }}</td>
        <td>
          <button class="btn" type="button" @click="edit(row)">编辑</button>
          <button class="btn danger" type="button" @click="remove(row)">删除</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
