<script setup lang="ts">
import { computed, ref } from "vue";
import { useDeskStore } from "../stores/desk";
import type { LinkItem } from "../api";

const props = withDefaults(defineProps<{ link: LinkItem; layout?: "tile" | "row" | "mini" }>(), {
  layout: "row",
});
const desk = useDeskStore();
const failed = ref(false);
const external = computed(() => /^https?:\/\//i.test(props.link.url));
const href = computed(() => {
  if (props.link.slug) return `/site/${props.link.slug}`;
  return external.value ? `/go/${props.link.id}` : props.link.url;
});
const detail = computed(() => Boolean(props.link.slug) || external.value);
const letter = computed(() => [...props.link.name][0] || "站");
const host = computed(() => {
  try {
    return new URL(props.link.url).hostname;
  } catch {
    return "";
  }
});
const icon = computed(() => (host.value ? `/ico?host=${host.value}` : ""));

function mark() {
  if (external.value) desk.remember(props.link);
}
</script>

<template>
  <component
    :is="detail ? 'a' : 'router-link'"
    class="link"
    :class="layout"
    :href="detail ? href : undefined"
    :to="detail ? undefined : href"
    :title="link.desc"
    @click="mark"
  >
    <span class="mark">
      <img v-if="icon && !failed" :src="icon" alt="" @error="failed = true" />
      <span v-else>{{ letter }}</span>
    </span>
    <span class="meta">
      <span class="name">{{ link.name }}</span>
      <span v-if="layout === 'row'" class="desc">{{ link.desc }}</span>
    </span>
    <button
      class="star"
      :class="{ on: desk.isCustom(link.id) }"
      type="button"
      :aria-label="desk.isCustom(link.id) ? '取消收藏' : '收藏'"
      @click.prevent.stop="desk.toggle(link)"
    >
      ★
    </button>
  </component>
</template>
