<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { getJSON } from "../../api";

type Row = { id: number; title: string; summary: string; body: string; views: number };

const rows = ref<Row[]>([]);
const error = ref("");
const editing = ref(0);
const form = reactive({ title: "", summary: "", body: "" });

function reset() {
  editing.value = 0;
  form.title = "";
  form.summary = "";
  form.body = "";
}

async function load() {
  rows.value = await getJSON<Row[]>("/api/admin/articles");
}

function edit(row: Row) {
  editing.value = row.id;
  Object.assign(form, row);
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
    <label class="wide">标题<input v-model="form.title" required /></label>
    <label class="wide">摘要<input v-model="form.summary" /></label>
    <label class="wide">正文<textarea v-model="form.body" required /></label>
    <button class="btn primary" type="submit">{{ editing ? "保存" : "添加" }}</button>
    <button v-if="editing" class="btn" type="button" @click="reset">取消</button>
  </form>
  <table>
    <thead>
      <tr><th>标题</th><th>阅读</th><th></th></tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ row.title }}</td>
        <td>{{ row.views }}</td>
        <td>
          <button class="btn" type="button" @click="edit(row)">编辑</button>
          <button class="btn danger" type="button" @click="remove(row)">删除</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
