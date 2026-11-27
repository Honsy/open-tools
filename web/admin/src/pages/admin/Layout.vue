<script setup lang="ts">
import { ref } from "vue";
import { adminName } from "../../api";
import AdminNav from "./AdminNav.vue";

const navigationOpen = ref(false);
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <RouterLink to="/" class="admin-brand">
        <span class="admin-brand-icon"><i class="pi pi-compass" aria-hidden="true"></i></span>
        <span><strong>开物录</strong><small>管理后台</small></span>
      </RouterLink>
      <AdminNav />
    </aside>
    <div class="admin-workspace">
      <header class="admin-topbar">
        <div class="admin-topbar-left">
          <PButton class="admin-mobile-menu" icon="pi pi-bars" text aria-label="打开管理导航" @click="navigationOpen = true" />
          <span>管理后台</span>
        </div>
        <div class="admin-topbar-actions">
          <PButton as="a" href="/" text severity="secondary" label="查看网站" icon="pi pi-external-link" icon-pos="right" />
          <span class="admin-account" :title="adminName">
            <span class="admin-avatar" aria-hidden="true">{{ adminName.slice(0, 1).toUpperCase() }}</span>
            <span class="admin-account-name">{{ adminName }}</span>
          </span>
          <form method="post" action="/logout"><PButton type="submit" text severity="secondary" icon="pi pi-sign-out" aria-label="退出登录" /></form>
        </div>
      </header>
      <main class="admin-main"><RouterView /></main>
    </div>
    <PDrawer aria-label="开物录管理后台" v-model:visible="navigationOpen" header="开物录管理后台" class="admin-nav-drawer">
      <AdminNav @select="navigationOpen = false" />
    </PDrawer>
  </div>
</template>
