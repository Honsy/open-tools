<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getJSON, setAdminToken } from "../../api";

const route = useRoute();
const router = useRouter();
const username = ref("admin");
const password = ref("");
const error = ref("");

async function send() {
  error.value = "";
  try {
    const data = await getJSON<{ token: string }>("/api/admin/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: username.value, password: password.value }),
    });
    setAdminToken(data.token);
    const next = typeof route.query.next === "string" ? route.query.next : "/admin";
    router.push(next);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "登录失败";
  }
}
</script>

<template>
  <form class="login-box" @submit.prevent="send">
    <h1>开物管理</h1>
    <label>账号<input v-model="username" autocomplete="username" /></label>
    <label>密码<input v-model="password" type="password" autocomplete="current-password" /></label>
    <button type="submit">登录</button>
    <p v-if="error" class="banner">{{ error }}</p>
  </form>
</template>
