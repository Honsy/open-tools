<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import Shell from "../components/Shell.vue";
import { formatDate, getJSON, type Article } from "../api";

const route = useRoute();
const article = ref<Article | null>(null);
const error = ref("");

async function load() {
  error.value = "";
  article.value = null;
  try {
    article.value = await getJSON<Article>("/api/articles/" + route.params.id);
    document.title = article.value.title + " - 开物录";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "没有这篇文章";
  }
}

watch(() => route.params.id, load, { immediate: true });
</script>

<template>
  <Shell>
    <p v-if="error" class="banner">{{ error }}</p>
    <article v-else-if="article" class="article">
      <p class="kicker">{{ formatDate(article.createdAt) }} · {{ article.views }} 次阅读</p>
      <h1>{{ article.title }}</h1>
      <p class="lead">{{ article.summary }}</p>
      <div class="body">{{ article.body }}</div>
    </article>
    <p v-else class="banner">正在加载</p>
  </Shell>
</template>
