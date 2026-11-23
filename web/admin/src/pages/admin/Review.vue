<script setup lang="ts">
import { onMounted, ref } from "vue";
import { getJSON } from "../../api";

type Row = {
  id: number;
  name: string;
  url: string;
  desc: string;
  categorySlug: string;
  status: string;
  sort: number;
  pinSort: number;
  pinned: boolean;
};
type Stats = { online: number; pending: number; off: number; articles: number; tags: number };

const stats = ref<Stats | null>(null);
const rows = ref<Row[]>([]);
const error = ref("");

async function load() {
  error.value = "";
  stats.value = await getJSON<Stats>("/api/admin/stats");
  rows.value = await getJSON<Row[]>("/api/admin/links?status=pending");
}

async function setStatus(row: Row, status: string) {
  await getJSON("/api/admin/links/" + row.id, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...row, status }),
  });
  await load();
}

onMounted(load);
</script>

<template>
  <h1>待审</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <div v-if="stats" class="counts">
    <span>上架 {{ stats.online }}</span>
    <span>待审 {{ stats.pending }}</span>
    <span>下架 {{ stats.off }}</span>
    <span>文章 {{ stats.articles }}</span>
    <span>标签 {{ stats.tags }}</span>
  </div>
  <p v-if="!rows.length" class="empty">没有待审网址。</p>
  <table v-else>
    <thead>
      <tr><th>名字</th><th>网址</th><th>简介</th><th></th></tr>
    </thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ row.name }}</td>
        <td>{{ row.url }}</td>
        <td>{{ row.desc }}</td>
        <td>
          <button class="btn primary" type="button" @click="setStatus(row, 'online')">通过</button>
          <button class="btn" type="button" @click="setStatus(row, 'off')">不收</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
