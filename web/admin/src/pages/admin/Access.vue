<script setup lang="ts">
import { useRoute } from "vue-router";

const route = useRoute();
</script>

<template>
  <main class="admin-access">
    <template v-if="route.query.unavailable === '1'">
      <h1>暂时无法连接后台</h1>
      <a href="/admin/">重试</a>
    </template>
    <template v-else>
      <h1>当前账号无后台权限</h1>
      <form method="post" action="/logout">
        <input type="hidden" name="next" value="/login?next=/admin/" />
        <button type="submit">切换账号</button>
      </form>
    </template>
    <a href="/">返回前台</a>
  </main>
</template>
