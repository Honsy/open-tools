<script setup lang="ts">
import { useRoute } from "vue-router";
import { adminPath } from "../../adminPath";

const route = useRoute();
const items = [
  { to: adminPath(), name: "网址审核" },
  { to: adminPath("links"), name: "网址" },
  { to: adminPath("internet-rank"), name: "互联网月榜" },
  { to: adminPath("categories"), name: "分类" },
  { to: adminPath("articles"), name: "文章" },
  { to: adminPath("tags"), name: "标签" },
  { to: adminPath("feedback"), name: "反馈" },
  { to: adminPath("users"), name: "用户" },
];
</script>

<template>
  <div class="adm">
    <header class="adm-top">
      <strong>开物录管理</strong>
      <div>
        <a href="/">返回前台</a>
        <form method="post" action="/logout"><button type="submit">退出</button></form>
      </div>
    </header>
    <div class="adm-body">
      <nav class="adm-side" aria-label="管理导航">
        <RouterLink v-for="item in items" :key="item.to" :to="item.to" :class="{ on: route.path === item.to }" :aria-current="route.path === item.to ? 'page' : undefined">
          {{ item.name }}
        </RouterLink>
      </nav>
      <main class="adm-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>
