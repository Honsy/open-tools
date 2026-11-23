<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { getJSON } from "../../api";

type Row = { id: number; group: string; name: string; sort: number };

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
  rows.value = await getJSON<Row[]>("/api/admin/tags");
}

function edit(row: Row) {
  editing.value = row.id;
  Object.assign(form, row);
}

async function save() {
  error.value = "";
  try {
    await getJSON("/api/admin/tags" + (editing.value ? "/" + editing.value : ""), {
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
  if (!window.confirm("删除标签「" + row.name + "」？")) return;
  await getJSON("/api/admin/tags/" + row.id, { method: "DELETE" });
  await load();
}

onMounted(load);
</script>

<template>
  <h1>标签</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <form class="adm-form" @submit.prevent="save">
    <label>分组<input v-model="form.group" required /></label>
    <label>名字<input v-model="form.name" required /></label>
    <label>排序<input v-model.number="form.sort" type="number" /></label>
    <button class="btn primary" type="submit">{{ editing ? "保存" : "添加" }}</button>
    <button v-if="editing" class="btn" type="button" @click="reset">取消</button>
  </form>
  <table>
    <thead>
      <tr><th>分组</th><th>名字</th><th>排序</th><th></th></tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ row.group }}</td>
        <td>{{ row.name }}</td>
        <td>{{ row.sort }}</td>
        <td>
          <button class="btn" type="button" @click="edit(row)">编辑</button>
          <button class="btn danger" type="button" @click="remove(row)">删除</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
