<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import Shell from "../components/Shell.vue";
import { tools } from "../tools";

const route = useRoute();
const tool = computed(() => tools[String(route.params.slug)] || null);
</script>

<template>
  <Shell>
    <article v-if="tool" class="article">
      <p class="kicker">已收进 legacy</p>
      <h1>{{ tool.name }}</h1>
      <p class="lead">{{ tool.desc }}</p>
      <p class="body">这个工具还没迁到新界面。原来的页面在仓库的 {{ tool.legacy }}。首页「网页工具」先把它当作入口留着。</p>
    </article>
    <p v-else class="banner">没有这个工具。</p>
  </Shell>
</template>
