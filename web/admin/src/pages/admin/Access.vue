<script setup lang="ts">
import { useRoute } from "vue-router";
const route = useRoute();
</script>

<template>
  <main class="admin-access">
    <PCard class="admin-panel">
      <template #content>
        <h1>{{ route.query.unavailable === "1" ? "暂时无法连接后台" : "当前账号无后台权限" }}</h1>
        <div class="admin-actions">
          <PButton v-if="route.query.unavailable === '1'" as="a" href="/admin/" label="重试" icon="pi pi-refresh" />
          <form v-else method="post" action="/logout">
            <input type="hidden" name="next" value="/login?next=/admin/" />
            <PButton type="submit" label="切换账号" />
          </form>
          <PButton as="a" href="/" label="返回前台" severity="secondary" outlined />
        </div>
      </template>
    </PCard>
  </main>
</template>
