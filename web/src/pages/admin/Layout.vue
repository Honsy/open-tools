<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { setAdminToken } from "../../api";

const route = useRoute();
const router = useRouter();
const items = [
  { to: "/admin", name: "待审" },
  { to: "/admin/links", name: "网址" },
  { to: "/admin/categories", name: "分类" },
  { to: "/admin/articles", name: "文章" },
  { to: "/admin/tags", name: "标签" },
];

function logout() {
  setAdminToken("");
  router.push("/admin/login");
}
</script>

<template>
  <div class="adm">
    <header class="adm-top">
      <strong>开物管理</strong>
      <div>
        <a href="/">返回前台</a>
        <button type="button" @click="logout">退出</button>
      </div>
    </header>
    <div class="adm-body">
      <nav class="adm-side">
        <RouterLink v-for="item in items" :key="item.to" :to="item.to" :class="{ on: route.path === item.to }">
          {{ item.name }}
        </RouterLink>
      </nav>
      <main class="adm-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>
