<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import Shell from "../components/Shell.vue";
import LinkTile from "../components/LinkTile.vue";
import { getJSON, type LinkItem } from "../api";

const route = useRoute();
const links = ref<LinkItem[]>([]);
const q = ref("");
const error = ref("");
const ready = ref(false);

async function run() {
  q.value = typeof route.query.q === "string" ? route.query.q : "";
  ready.value = false;
  error.value = "";
  document.title = (q.value ? q.value + " - " : "") + "搜索 - 开物";
  try {
    const data = await getJSON<{ links: LinkItem[] }>("/api/search?q=" + encodeURIComponent(q.value));
    links.value = data.links;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "搜索失败";
  } finally {
    ready.value = true;
  }
}

watch(() => route.query.q, run, { immediate: true });
</script>

<template>
  <Shell>
    <section class="block">
      <div class="sec-head">
        <h2>搜索</h2>
      </div>
      <p v-if="error" class="banner">{{ error }}</p>
      <p v-else-if="!q" class="empty">在顶栏输入关键词，搜名字和简介。</p>
      <p v-else-if="!ready" class="empty">正在搜索</p>
      <p v-else-if="!links.length" class="empty">没有和「{{ q }}」对应的网址。</p>
      <div v-else class="link-grid row">
        <LinkTile v-for="link in links" :key="link.id" :link="link" layout="row" />
      </div>
    </section>
  </Shell>
</template>
