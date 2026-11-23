<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
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
  rows.value = await getJSON<Row[]>("/api/admin/categories");
}

function edit(row: Row) {
  editing.value = row.id;
  Object.assign(form, row);
}

async function save() {
  error.value = "";
  try {
    await getJSON("/api/admin/categories" + (editing.value ? "/" + editing.value : ""), {
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
  if (!window.confirm("删除分类「" + row.name + "」？")) return;
  error.value = "";
  try {
    await getJSON("/api/admin/categories/" + row.id, { method: "DELETE" });
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "没有删掉";
  }
}

onMounted(load);
</script>

<template>
  <h1>分类</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <form class="adm-form" @submit.prevent="save">
    <label>标识<input v-model="form.slug" :disabled="!!editing" required placeholder="video" /></label>
    <label>名字<input v-model="form.name" required /></label>
    <label>上级标识<input v-model="form.parentSlug" placeholder="分区留空" /></label>
    <label>层级
      <select v-model="form.kind">
        <option value="section">分区</option>
        <option value="tab">子类</option>
      </select>
    </label>
    <label>内容
      <select v-model="form.contentKind">
        <option value="links">网址</option>
        <option value="articles">文章</option>
        <option value="tags">标签</option>
        <option value="mixed">混合</option>
      </select>
    </label>
    <label>排序<input v-model.number="form.sort" type="number" /></label>
    <label><input v-model="form.showOnHome" type="checkbox" /> 出现在首页</label>
    <button class="btn primary" type="submit">{{ editing ? "保存" : "添加" }}</button>
    <button v-if="editing" class="btn" type="button" @click="reset">取消</button>
  </form>
  <table>
    <thead>
      <tr><th>标识</th><th>名字</th><th>上级</th><th>层级</th><th>内容</th><th>首页</th><th></th></tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ row.slug }}</td>
        <td>{{ row.name }}</td>
        <td>{{ row.parentSlug }}</td>
        <td>{{ row.kind === "section" ? "分区" : "子类" }}</td>
        <td>{{ row.contentKind }}</td>
        <td>{{ row.showOnHome ? "是" : "" }}</td>
        <td>
          <button class="btn" type="button" @click="edit(row)">编辑</button>
          <button class="btn danger" type="button" @click="remove(row)">删除</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
