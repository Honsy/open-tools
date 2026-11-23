<script setup lang="ts">
import { onMounted, ref } from "vue";
import { formatDate, getJSON } from "../../api";

type User = { id: number; username: string; email: string; isAdmin: boolean; createdAt: string };
type UserPage = { items: User[]; total: number; page: number; pageCount: number };

const search = ref("");
const result = ref<UserPage>({ items: [], total: 0, page: 1, pageCount: 1 });
const loading = ref(false);
const error = ref("");
let activeSearch = "";

async function load(page: number, query = activeSearch) {
  loading.value = true;
  error.value = "";
  try {
    const params = new URLSearchParams({ q: query, page: String(page) });
    result.value = await getJSON<UserPage>(`/api/admin/users?${params}`);
    activeSearch = query;
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "读取失败";
  } finally {
    loading.value = false;
  }
}

onMounted(() => load(1));
</script>

<template>
  <h1>用户</h1>
  <form class="adm-filters" role="search" @submit.prevent="load(1, search.trim())">
    <input v-model="search" type="search" maxlength="254" aria-label="用户名或邮箱" placeholder="用户名或邮箱" />
    <button class="btn primary" type="submit" :disabled="loading">搜索</button>
  </form>
  <p v-if="error" class="err" role="alert">{{ error }}</p>
  <p v-else-if="loading" role="status">正在读取用户…</p>
  <p v-else-if="!result.items.length">没有找到用户</p>
  <template v-else>
    <div class="users-table">
      <table>
        <thead><tr><th>ID</th><th>用户名</th><th>邮箱</th><th>角色</th><th>注册日期</th></tr></thead>
        <tbody>
          <tr v-for="user in result.items" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.username }}</td>
            <td>{{ user.email || "未绑定" }}</td>
            <td>{{ user.isAdmin ? "管理员" : "普通用户" }}</td>
            <td>{{ formatDate(user.createdAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <nav class="adm-filters users-pagination" aria-label="用户分页">
      <span>共 {{ result.total }} 人 · 第 {{ result.page }} / {{ result.pageCount }} 页</span>
      <button class="btn" type="button" :disabled="result.page === 1" @click="load(result.page - 1)">上一页</button>
      <button class="btn" type="button" :disabled="result.page >= result.pageCount" @click="load(result.page + 1)">下一页</button>
    </nav>
  </template>
</template>

<style scoped>
.users-table { overflow-x: auto; }
.users-table table { min-width: 32rem; }
.adm-filters input { min-width: 0; }
.adm-filters button:disabled { opacity: 0.55; cursor: default; }
.users-pagination { align-items: center; flex-wrap: wrap; margin-top: 0.75rem; }
</style>
