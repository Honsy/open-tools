<script setup lang="ts">
import { onMounted, ref } from "vue";
import { formatDate, getJSON } from "../../api";

type Feedback = { id: number; userId: number; body: string; status: string; createdAt: string };
const rows = ref<Feedback[]>([]);
const error = ref("");
const statuses = { pending: "待处理", done: "已处理", rejected: "已拒绝" };

async function load() {
  try {
    rows.value = await getJSON<Feedback[]>("/api/admin/feedback");
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "读取失败";
  }
}

async function update(row: Feedback, event: Event) {
  const select = event.target as HTMLSelectElement;
  const status = select.value;
  error.value = "";
  try {
    const updated = await getJSON<Feedback>("/api/admin/feedback/" + row.id, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    });
    row.status = updated.status;
  } catch (failure) {
    select.value = row.status;
    error.value = failure instanceof Error ? failure.message : "更新失败";
  }
}

onMounted(load);
</script>

<template>
  <h1>站点反馈</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <p v-else-if="!rows.length">暂无反馈</p>
  <table v-else>
    <thead><tr><th>日期</th><th>反馈内容</th><th>状态</th></tr></thead>
    <tbody>
      <tr v-for="row in rows" :key="row.id">
        <td>{{ formatDate(row.createdAt) }}</td>
        <td style="white-space: pre-wrap">{{ row.body }}</td>
        <td><select :value="row.status" aria-label="处理状态" @change="update(row, $event)">
          <option v-for="(label, value) in statuses" :key="value" :value="value">{{ label }}</option>
        </select></td>
      </tr>
    </tbody>
  </table>
</template>
