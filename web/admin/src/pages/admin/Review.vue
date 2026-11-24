<script setup lang="ts">
import { onMounted, ref } from "vue";
import { getJSON } from "../../api";

type Row = {
  id: number;
  name: string;
  url: string;
  desc: string;
  categorySlug: string;
};
type Category = { slug: string; name: string; parentSlug: string; kind: string; contentKind: string };
type Stats = { online: number; pending: number; off: number };
type Decision = "approve" | "reject" | "block";

const stats = ref<Stats | null>(null);
const rows = ref<Row[]>([]);
const categories = ref<Category[]>([]);
const selections = ref<Record<number, string>>({});
const loading = ref(true);
const busy = ref(0);
const error = ref("");
const message = ref("");

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const [summary, pending, catalog] = await Promise.all([
      getJSON<Stats>("/api/admin/stats"),
      getJSON<Row[]>("/api/admin/links?status=pending"),
      getJSON<Category[]>("/api/admin/categories"),
    ]);
    stats.value = summary;
    rows.value = pending;
    categories.value = catalog;
    selections.value = Object.fromEntries(pending.map(row => [row.id, selections.value[row.id] || row.categorySlug]));
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "读取待审网址失败";
  } finally {
    loading.value = false;
  }
}

function catLabel(cat: Category) {
  const parent = categories.value.find(item => item.slug === cat.parentSlug);
  return parent ? `${parent.name} / ${cat.name}` : cat.name;
}

async function review(row: Row, decision: Decision) {
  busy.value = row.id;
  error.value = "";
  message.value = "";
  try {
    await getJSON(`/api/admin/submissions/${row.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decision, categorySlug: selections.value[row.id] }),
    });
    message.value = decision === "approve" ? `已上架「${row.name}」` : decision === "block" ? `已拒绝「${row.name}」并加入月榜风险名单` : `已拒绝「${row.name}」`;
    await load();
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "审核失败";
  } finally {
    busy.value = 0;
  }
}

onMounted(load);
</script>

<template>
  <div class="adm-page-heading">
    <h1>网址审核</h1>
    <a class="btn" href="/submit" target="_blank" rel="noopener">公开提交入口</a>
  </div>
  <p v-if="error" class="err" role="alert">{{ error }}</p>
  <p v-if="message" class="adm-message" role="status">{{ message }}</p>
  <div v-if="stats" class="counts">
    <span>待审 {{ stats.pending }}</span>
    <span>上架 {{ stats.online }}</span>
    <span>下架 {{ stats.off }}</span>
  </div>
  <p v-if="loading" class="adm-card" role="status">正在读取待审网址…</p>
  <p v-else-if="!rows.length && !error" class="adm-card">没有待审网址。</p>
  <article v-for="row in rows" :key="row.id" class="adm-card adm-review-row">
    <div>
      <h2>{{ row.name }}</h2>
      <p class="adm-url">{{ row.url }}</p>
      <p v-if="row.desc">{{ row.desc }}</p>
    </div>
    <form class="adm-form" @submit.prevent="review(row, 'approve')">
      <label>发布到
        <select v-model="selections[row.id]" required :disabled="busy !== 0">
          <option value="inbox" disabled>先选择网址分类</option>
          <option v-for="cat in categories.filter(item => item.kind === 'tab' && item.contentKind === 'links')" :key="cat.slug" :value="cat.slug">{{ catLabel(cat) }}</option>
        </select>
      </label>
      <button class="btn primary" type="submit" :disabled="busy !== 0 || !selections[row.id] || selections[row.id] === 'inbox'">通过并上架</button>
      <button class="btn" type="button" :disabled="busy !== 0" @click="review(row, 'reject')">拒绝</button>
      <button class="btn danger" type="button" :disabled="busy !== 0" @click="review(row, 'block')">拒绝并屏蔽月榜域名</button>
    </form>
  </article>
</template>
