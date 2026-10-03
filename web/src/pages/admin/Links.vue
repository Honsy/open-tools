<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { getJSON } from "../../api";

type Cat = { slug: string; name: string; kind: string; parentSlug: string };
type Row = {
  id: number;
  slug?: string;
  name: string;
  url: string;
  desc: string;
  tags: string;
  alias: string;
  lang: string;
  region: string;
  spare: string;
  icp: string;
  body: string;
  categorySlug: string;
  status: string;
  sort: number;
  pinSort: number;
  pinned: boolean;
  clicks: number;
  views: number;
};

const cats = ref<Cat[]>([]);
const rows = ref<Row[]>([]);
const status = ref("");
const q = ref("");
const error = ref("");
const editing = ref(0);
const form = reactive({
  slug: "",
  name: "",
  url: "",
  desc: "",
  tags: "",
  alias: "",
  lang: "",
  region: "",
  spare: "",
  icp: "",
  body: "",
  categorySlug: "",
  status: "online",
  sort: 1,
  pinSort: 0,
  pinned: false,
});

function catName(slug: string) {
  const cat = cats.value.find((item) => item.slug === slug);
  return cat ? catLabel(cat) : slug;
}

function catLabel(cat: Cat) {
  const parent = cats.value.find((item) => item.slug === cat.parentSlug);
  return (parent ? parent.name + " / " : "") + cat.name;
}

function reset() {
  editing.value = 0;
  form.slug = "";
  form.name = "";
  form.url = "";
  form.desc = "";
  form.tags = "";
  form.alias = "";
  form.lang = "";
  form.region = "";
  form.spare = "";
  form.icp = "";
  form.body = "";
  form.categorySlug = cats.value.find((c) => c.kind === "tab")?.slug || "";
  form.status = "online";
  form.sort = 1;
  form.pinSort = 0;
  form.pinned = false;
}

async function load() {
  error.value = "";
  cats.value = await getJSON<Cat[]>("/api/admin/categories");
  const query = new URLSearchParams();
  if (status.value) query.set("status", status.value);
  if (q.value.trim()) query.set("q", q.value.trim());
  rows.value = await getJSON<Row[]>("/api/admin/links?" + query.toString());
  if (!form.categorySlug) reset();
}

function edit(row: Row) {
  editing.value = row.id;
  Object.assign(form, row);
}

async function save() {
  error.value = "";
  try {
    await getJSON("/api/admin/links" + (editing.value ? "/" + editing.value : ""), {
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
  if (!window.confirm("删除「" + row.name + "」？")) return;
  await getJSON("/api/admin/links/" + row.id, { method: "DELETE" });
  await load();
}

onMounted(load);
</script>

<template>
  <h1>网址</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <form class="adm-form" @submit.prevent="save">
    <label>名字<input v-model="form.name" required /></label>
    <label class="wide">网址<input v-model="form.url" required placeholder="https://" /></label>
    <label class="wide">简介<input v-model="form.desc" maxlength="80" placeholder="卡片上的一句话" /></label>
    <label class="wide">标签<input v-model="form.tags" maxlength="40" placeholder="弹幕,视频" /></label>
    <label>别名<input v-model="form.alias" maxlength="20" placeholder="B站" /></label>
    <label>语言<input v-model="form.lang" maxlength="20" placeholder="中文" /></label>
    <label>地区<input v-model="form.region" maxlength="20" placeholder="中国" /></label>
    <label class="wide">备用地址<input v-model="form.spare" placeholder="https://" /></label>
    <label>备案<input v-model="form.icp" maxlength="40" placeholder="选填" /></label>
    <label class="full">网站介绍<textarea v-model="form.body" rows="8" maxlength="4000" placeholder="详情页正文，空一行分段" /></label>
    <label>分类
      <select v-model="form.categorySlug">
        <option v-for="cat in cats.filter((c) => c.kind === 'tab')" :key="cat.slug" :value="cat.slug">
          {{ catLabel(cat) }}
        </option>
      </select>
    </label>
    <label>状态
      <select v-model="form.status">
        <option value="online">上架</option>
        <option value="pending">待审</option>
        <option value="off">下架</option>
      </select>
    </label>
    <label>排序<input v-model.number="form.sort" type="number" /></label>
    <label>常用排序<input v-model.number="form.pinSort" type="number" /></label>
    <label><input v-model="form.pinned" type="checkbox" /> 放进热门推荐</label>
    <button class="btn primary" type="submit">{{ editing ? "保存" : "添加" }}</button>
    <button v-if="editing" class="btn" type="button" @click="reset">取消</button>
    <a v-if="editing && form.slug" :href="'/site/' + form.slug" target="_blank" rel="noopener">看详情页</a>
  </form>
  <div class="adm-filters">
    <select v-model="status" @change="load">
      <option value="">全部状态</option>
      <option value="online">上架</option>
      <option value="pending">待审</option>
      <option value="off">下架</option>
    </select>
    <input v-model="q" placeholder="搜名字或网址" @keyup.enter="load" />
    <button class="btn" type="button" @click="load">筛选</button>
  </div>
  <table>
    <thead>
      <tr><th>名字</th><th>分类</th><th>状态</th><th>打开 / 浏览</th><th></th></tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ row.name }}<br />{{ row.url }}</td>
        <td>{{ catName(row.categorySlug) }}</td>
        <td>{{ row.status === "online" ? "上架" : row.status === "pending" ? "待审" : "下架" }}</td>
        <td>{{ row.clicks }} / {{ row.views }}</td>
        <td>
          <button class="btn" type="button" @click="edit(row)">编辑</button>
          <button class="btn danger" type="button" @click="remove(row)">删除</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
