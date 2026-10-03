<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import Shell from "../components/Shell.vue";
import LinkTile from "../components/LinkTile.vue";
import SectionView from "../components/SectionView.vue";
import { getJSON, type HomeData } from "../api";
import { useDeskStore } from "../stores/desk";

const route = useRoute();
const desk = useDeskStore();
const data = ref<HomeData | null>(null);
const error = ref("");
const active = reactive<Record<string, string>>({});
const shortTab = ref<"hot" | "recent" | "edit">("hot");
const latestTab = ref<"latest" | "popular">("latest");
const boardIndex = ref(0);
let timer = 0;

async function load() {
  error.value = "";
  try {
    data.value = await getJSON<HomeData>("/api/home");
    desk.refresh(collectLinks(data.value));
    for (const section of data.value.sections) {
      if (section.tabs?.length && !active[section.slug]) active[section.slug] = section.tabs[0].slug;
    }
    await nextTick();
    focusHash(route.hash);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "首页没有加载出来";
  }
}

function collectLinks(home: HomeData) {
  const links = [...home.pinned, ...home.latest, ...home.popular];
  for (const section of home.sections) {
    for (const tab of section.tabs || []) links.push(...(tab.links || []));
  }
  return links;
}

function focusHash(hash: string) {
  if (!hash || !data.value) return;
  const id = decodeURIComponent(hash.slice(1));
  let target = id;
  if (id.startsWith("tab-")) {
    const slug = id.slice(4);
    const section = data.value.sections.find((item) => item.tabs?.some((tab) => tab.slug === slug));
    if (section) {
      active[section.slug] = slug;
      target = "sec-" + section.slug;
    }
  }
  nextTick(() => {
    document.getElementById(target)?.scrollIntoView({ behavior: "smooth", block: "start" });
  });
}

onMounted(() => {
  document.title = "开物录 - 网址导航";
  load();
  timer = window.setInterval(() => {
    const count = data.value?.boards.length || 0;
    if (count > 1) boardIndex.value = (boardIndex.value + 1) % count;
  }, 5000);
});
onUnmounted(() => window.clearInterval(timer));
watch(() => route.hash, focusHash);
</script>

<template>
  <Shell>
    <p v-if="error" class="banner">{{ error }}。先在 server 目录启动接口。</p>
    <p v-else-if="!data" class="banner">正在加载</p>
    <template v-else>
      <div class="hero">
        <section class="panel" id="sec-today">
          <h2>{{ data.todayLabel }}</h2>
          <p class="kicker">历史上的今天</p>
          <ul class="today">
            <li v-for="(event, index) in data.today" :key="index">
              <b v-if="event.year">{{ event.year }}</b>
              <span>{{ event.text }}</span>
            </li>
          </ul>
        </section>

        <section class="panel" id="sec-shortcuts">
          <div class="tabs">
            <button type="button" :class="{ on: shortTab === 'hot' }" @click="shortTab = 'hot'">热门推荐</button>
            <button type="button" :class="{ on: shortTab === 'recent' }" @click="shortTab = 'recent'">最近使用</button>
            <button type="button" :class="{ on: shortTab === 'edit' }" @click="shortTab = 'edit'">编辑</button>
          </div>
          <div v-if="shortTab === 'hot'" class="sc-list">
            <LinkTile v-for="link in data.pinned" :key="link.id" :link="link" layout="tile" />
          </div>
          <div v-else-if="shortTab === 'recent'">
            <p v-if="!desk.recent.length" class="empty">还没有点过。点过的外链会出现在这里。</p>
            <div v-else class="sc-list">
              <LinkTile v-for="link in desk.recent" :key="link.id" :link="link" layout="tile" />
            </div>
          </div>
          <div v-else>
            <p class="empty">点卡片右上角的星，收藏到这台浏览器。只存在本地。</p>
            <div v-if="desk.custom.length" class="sc-list">
              <LinkTile v-for="link in desk.custom" :key="link.id" :link="link" layout="tile" />
            </div>
          </div>
        </section>

        <section class="panel" id="sec-boards">
          <div class="tabs">
            <button
              v-for="(board, index) in data.boards"
              :key="board.key"
              type="button"
              :class="{ on: boardIndex === index }"
              @click="boardIndex = index"
            >
              {{ board.name }}
            </button>
          </div>
          <ol class="board">
            <li v-for="(item, index) in data.boards[boardIndex]?.items || []" :key="item.href">
              <em>{{ index + 1 }}</em>
              <a :href="item.href" :rel="item.external ? 'nofollow noopener' : undefined" :target="item.external ? '_blank' : undefined">{{ item.title }}</a>
              <span>{{ item.heat }}</span>
            </li>
          </ol>
          <p v-if="data.liveHot" class="kicker">热搜来自微博和百度，大约每十分钟更新一次。</p>
          <p v-else class="kicker">实时热搜暂时取不到，这里是站内点击和刚收录的网站。</p>
        </section>
      </div>

      <SectionView
        v-for="section in data.sections"
        :key="section.slug"
        :section="section"
        :layout="section.slug === 'fun' ? 'tile' : 'row'"
        :active="active[section.slug]"
        @update:active="active[section.slug] = $event"
      />

      <section class="block" id="sec-latest">
        <div class="sec-head">
          <h2>最新收录</h2>
          <div class="tabs">
            <button type="button" :class="{ on: latestTab === 'latest' }" @click="latestTab = 'latest'">最新收录</button>
            <button type="button" :class="{ on: latestTab === 'popular' }" @click="latestTab = 'popular'">热门网址</button>
          </div>
        </div>
        <div class="link-grid mini">
          <LinkTile
            v-for="link in latestTab === 'latest' ? data.latest : data.popular"
            :key="link.id"
            :link="link"
            layout="mini"
          />
        </div>
      </section>
    </template>
  </Shell>
</template>
