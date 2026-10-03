<script setup lang="ts">
import { computed } from "vue";
import LinkTile from "./LinkTile.vue";
import { formatDate, getJSON, type LinkItem, type Section } from "../api";
import { useDeskStore } from "../stores/desk";

const props = withDefaults(
  defineProps<{ section: Section; layout?: "tile" | "row"; active?: string; showMore?: boolean }>(),
  { layout: "row", showMore: true },
);
const emit = defineEmits<{ (e: "update:active", value: string): void }>();
const desk = useDeskStore();

const current = computed(() => {
  const tabs = props.section.tabs || [];
  return tabs.find((tab) => tab.slug === props.active) || tabs[0];
});

function choose(slug: string) {
  emit("update:active", slug);
}

async function surprise() {
  const link = await getJSON<LinkItem>("/api/random");
  desk.remember(link);
  window.location.href = link.slug ? `/site/${link.slug}` : `/go/${link.id}`;
}
</script>

<template>
  <section :id="'sec-' + section.slug" class="block">
    <div class="sec-head">
      <h2>{{ section.name }}</h2>
      <div v-if="section.tabs?.length" class="tabs">
        <button
          v-for="tab in section.tabs"
          :id="'tab-' + tab.slug"
          :key="tab.slug"
          type="button"
          :class="{ on: current?.slug === tab.slug }"
          @click="choose(tab.slug)"
        >
          {{ tab.name }}
        </button>
      </div>
      <RouterLink v-if="showMore && section.kind !== 'tags'" class="more" :to="'/c/' + section.slug">更多</RouterLink>
    </div>

    <div v-if="section.kind === 'tags'" class="tag-board">
      <div v-for="group in section.groups" :key="group.name" class="tag-group">
        <h3>{{ group.name }}</h3>
        <div class="tags">
          <RouterLink v-for="tag in group.tags" :key="tag" :to="{ path: '/search', query: { q: tag } }">
            {{ tag }}
          </RouterLink>
        </div>
      </div>
      <button class="random" type="button" @click="surprise">随机网址</button>
    </div>

    <div v-else-if="current?.kind === 'articles'" class="posts">
      <RouterLink v-for="post in current.articles" :key="post.id" class="post" :to="'/a/' + post.id">
        <time>{{ formatDate(post.createdAt) }}</time>
        <strong>{{ post.title }}</strong>
        <p>{{ post.summary }}</p>
        <span>{{ post.views }} 次阅读</span>
      </RouterLink>
    </div>

    <div v-else class="link-grid" :class="layout">
      <LinkTile v-for="link in current?.links || []" :key="link.id" :link="link" :layout="layout" />
    </div>
  </section>
</template>
