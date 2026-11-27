<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useToast } from "primevue/usetoast";
import { formatDate, getJSON } from "../../api";

type Feedback = { id: number; userId: number; body: string; status: string; createdAt: string };
const toast = useToast();
const rows = ref<Feedback[]>([]);
const error = ref("");
const loading = ref(true);
const saving = ref<number[]>([]);
const filter = ref("");
const choices = ref<Record<number, string>>({});
const statuses = [{ value: "pending", label: "待处理" }, { value: "done", label: "已处理" }, { value: "rejected", label: "已拒绝" }];
const options = [{ value: "", label: "全部状态" }, ...statuses];
const visibleRows = computed(() => filter.value ? rows.value.filter(row => row.status === filter.value) : rows.value);

async function load() {
  loading.value = true;
  error.value = "";
  try {
    rows.value = await getJSON<Feedback[]>("/api/admin/feedback");
    choices.value = Object.fromEntries(rows.value.map(row => [row.id, row.status]));
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "读取失败";
  } finally {
    loading.value = false;
  }
}

async function update(row: Feedback, status: string) {
  if (row.status === status) return;
  saving.value.push(row.id);
  error.value = "";
  try {
    const updated = await getJSON<Feedback>("/api/admin/feedback/" + row.id, {
      method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ status }),
    });
    row.status = updated.status;
    choices.value[row.id] = updated.status;
    toast.add({ severity: "success", summary: "反馈状态已更新", life: 2400 });
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "更新失败";
    choices.value[row.id] = row.status;
  } finally {
    saving.value = saving.value.filter(id => id !== row.id);
  }
}

onMounted(load);
</script>

<template>
  <header class="admin-page-header"><div><h1>站点反馈</h1><p>查看留言并跟进处理状态。</p></div></header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <div class="admin-toolbar"><PSelect v-model="filter" :options="options" option-label="label" option-value="value" aria-label="反馈状态筛选" /></div>
      <PDataTable :value="visibleRows" :loading="loading" data-key="id" class="admin-table" scrollable table-style="min-width: 620px">
        <template #empty>没有符合条件的反馈。</template>
        <PColumn header="日期" style="min-width: 100px"><template #body="{ data }">{{ formatDate(data.createdAt) }}</template></PColumn>
        <PColumn header="反馈内容" style="min-width: 270px"><template #body="{ data }"><div class="admin-feedback-text">{{ data.body }}</div></template></PColumn>
        <PColumn header="账号" style="min-width: 80px"><template #body="{ data }">{{ data.userId ? '#' + data.userId : '—' }}</template></PColumn>
        <PColumn header="状态" style="min-width: 145px"><template #body="{ data }">
          <PSelect v-model="choices[data.id]" :options="statuses" option-label="label" option-value="value" aria-label="处理状态" :disabled="saving.includes(data.id)" @change="update(data, $event.value)" />
        </template></PColumn>
      </PDataTable>
    </template>
  </PCard>
</template>
