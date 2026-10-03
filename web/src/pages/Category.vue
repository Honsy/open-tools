<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import Shell from "../components/Shell.vue";
import SectionView from "../components/SectionView.vue";
import { getJSON, type Section } from "../api";

const route = useRoute();
const section = ref<Section | null>(null);
const error = ref("");
const active = ref("");

async function load() {
  error.value = "";
  section.value = null;
  try {
    const data = await getJSON<Section>("/api/categories/" + route.params.slug);
    section.value = data;
    active.value = data.tabs?.[0]?.slug || "";
    document.title = data.name + " - 开物录";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "没有这个分类";
    document.title = "开物录";
  }
}

watch(() => route.params.slug, load, { immediate: true });
</script>

<template>
  <Shell>
    <p v-if="error" class="banner">{{ error }}</p>
    <p v-else-if="!section" class="banner">正在加载</p>
    <SectionView
      v-else
      :section="section"
      :layout="section.slug === 'fun' ? 'tile' : 'row'"
      :active="active"
      :show-more="false"
      @update:active="active = $event"
    />
  </Shell>
</template>
