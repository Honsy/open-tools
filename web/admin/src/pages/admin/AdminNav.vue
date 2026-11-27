<script setup lang="ts">
import { useRoute } from "vue-router";

const emit = defineEmits<{ select: [] }>();
const route = useRoute();
const groups = [
  { name: "内容运营", items: [
    { path: "/", name: "网址审核", icon: "pi-inbox" },
    { path: "/links", name: "网址管理", icon: "pi-link" },
    { path: "/categories", name: "分类管理", icon: "pi-folder" },
    { path: "/articles", name: "文章管理", icon: "pi-file-edit" },
    { path: "/tags", name: "标签管理", icon: "pi-tags" },
  ] },
  { name: "站点管理", items: [
    { path: "/internet-rank", name: "互联网月榜", icon: "pi-chart-bar" },
    { path: "/users", name: "用户管理", icon: "pi-users" },
    { path: "/feedback", name: "站点反馈", icon: "pi-comments" },
  ] },
];
</script>

<template>
  <nav aria-label="管理导航">
    <section v-for="group in groups" :key="group.name" class="admin-menu-section">
      <h2>{{ group.name }}</h2>
      <RouterLink v-for="item in group.items" :key="item.path" :to="item.path" class="admin-menu-item"
        :class="{ 'is-active': route.path === item.path }" :aria-current="route.path === item.path ? 'page' : undefined" @click="emit('select')">
        <i class="pi" :class="item.icon" aria-hidden="true"></i><span>{{ item.name }}</span>
      </RouterLink>
    </section>
  </nav>
</template>
