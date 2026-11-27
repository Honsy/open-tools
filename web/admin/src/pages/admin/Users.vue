<script setup lang="ts">
import { onMounted, ref } from "vue";
import { formatDate, getJSON } from "../../api";

type User = { id: number; username: string; email: string; isAdmin: boolean; createdAt: string };
type UserPage = { items: User[]; total: number; page: number; pageCount: number };

const search = ref("");
const result = ref<UserPage>({ items: [], total: 0, page: 1, pageCount: 1 });
const loading = ref(true);
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
  <header class="admin-page-header"><div><h1>用户管理</h1><p>按用户名或邮箱查询注册账号。</p></div></header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <form class="admin-toolbar" role="search" @submit.prevent="load(1, search.trim())">
        <PInputText v-model="search" type="search" maxlength="254" aria-label="用户名或邮箱" placeholder="搜索用户名或邮箱" />
        <PButton type="submit" label="搜索" icon="pi pi-search" severity="secondary" outlined :disabled="loading" />
      </form>
      <PDataTable :value="result.items" :loading="loading" data-key="id" class="admin-table" scrollable table-style="min-width: 620px">
        <template #empty>没有找到用户。</template>
        <PColumn field="id" header="ID" style="width: 80px" />
        <PColumn field="username" header="用户名"><template #body="{ data }"><strong class="admin-primary">{{ data.username }}</strong></template></PColumn>
        <PColumn field="email" header="邮箱"><template #body="{ data }">{{ data.email || "未绑定" }}</template></PColumn>
        <PColumn header="角色"><template #body="{ data }"><PTag :value="data.isAdmin ? '管理员' : '普通用户'" :severity="data.isAdmin ? 'info' : 'secondary'" /></template></PColumn>
        <PColumn header="注册日期"><template #body="{ data }">{{ formatDate(data.createdAt) }}</template></PColumn>
      </PDataTable>
      <div class="admin-pagination"><span class="admin-muted">共 {{ result.total }} 人</span><PPaginator :first="(result.page - 1) * 50" :rows="50" :total-records="result.total" :page-link-size="3" @page="load($event.page + 1)" /></div>
    </template>
  </PCard>
</template>
