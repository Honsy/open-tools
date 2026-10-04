<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getJSON, type SideItem } from "../api";

const route = useRoute();
const router = useRouter();
const name = ref("开物录");
const sidebar = ref<SideItem[]>([]);
const q = ref("");
const engine = ref("site");
const engines = [
  { id: "site", name: "站内" },
  { id: "baidu", name: "百度", tpl: "https://www.baidu.com/s?wd=" },
  { id: "so", name: "360", tpl: "https://www.so.com/s?q=" },
  { id: "sogou", name: "搜狗", tpl: "https://www.sogou.com/web?query=" },
  { id: "bing", name: "必应", tpl: "https://www.bing.com/search?q=" },
  { id: "google", name: "Google", tpl: "https://www.google.com/search?q=" },
  { id: "bili", name: "哔哩哔哩", tpl: "https://search.bilibili.com/all?keyword=" },
  { id: "zhihu", name: "知乎", tpl: "https://www.zhihu.com/search?type=content&q=" },
];

const utilities = [
  { name: "邮箱", href: "https://mail.qq.com" },
  { name: "网盘", href: "https://www.alipan.com" },
  { name: "翻译", href: "https://translate.google.com" },
  { name: "地图", href: "https://map.baidu.com" },
];

onMounted(async () => {
  try {
    const data = await getJSON<{ name: string; sidebar: SideItem[] }>("/api/nav");
    name.value = data.name;
    sidebar.value = data.sidebar;
  } catch {
    sidebar.value = [];
  }
});

function search() {
  const text = q.value.trim();
  if (!text) return;
  const picked = engines.find((item) => item.id === engine.value);
  if (!picked || picked.id === "site" || !picked.tpl) {
    router.push({ path: "/search", query: { q: text } });
    return;
  }
  window.open(picked.tpl + encodeURIComponent(text), "_blank", "noopener");
}

</script>

<template>
  <div class="app">
    <header class="top">
      <div class="wrap top-row">
        <RouterLink class="logo" to="/">{{ name }}</RouterLink>
        <nav class="top-links">
          <RouterLink to="/hot">今日热榜</RouterLink>
          <RouterLink to="/c/ai">AI 工具</RouterLink>
          <RouterLink to="/submit">提交网站</RouterLink>
        </nav>
        <div class="util">
          <a v-for="item in utilities" :key="item.name" :href="item.href" target="_blank" rel="noopener">{{ item.name }}</a>
        </div>
      </div>
      <form class="wrap search" @submit.prevent="search">
        <select v-model="engine" aria-label="搜索引擎">
          <option v-for="item in engines" :key="item.id" :value="item.id">{{ item.name }}</option>
        </select>
        <input v-model="q" type="search" placeholder="搜本站网址，或换一个引擎" />
        <button type="submit">搜索</button>
      </form>
    </header>
    <div class="wrap body">
      <aside>
        <div v-for="item in sidebar" :key="item.name" class="side-group">
          <RouterLink :class="{ on: route.path === item.href }" :to="item.href">{{ item.name }}</RouterLink>
          <RouterLink
            v-for="child in item.children"
            :key="child.href"
            class="child"
            :class="{ on: route.path === child.href }"
            :to="child.href"
          >
            {{ child.name }}
          </RouterLink>
        </div>
      </aside>
      <main>
        <slot />
        <footer><RouterLink to="/admin">管理</RouterLink></footer>
      </main>
    </div>
  </div>
</template>
