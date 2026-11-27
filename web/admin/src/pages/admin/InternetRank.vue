<script setup lang="ts">
import { onMounted, ref, shallowRef } from "vue";
import { useConfirm } from "primevue/useconfirm";
import { useToast } from "primevue/usetoast";
import { getJSON } from "../../api";

type RankMonth = { month: string; source: string; entries: number; updatedAt: string };
type RankConfig = { months: RankMonth[]; blocked: { domain: string }[] };
type Preview = {
  month: string; total: number; excluded: number; version: string;
  entries: { Domain: string; Name: string; Category: string; Visits: string }[];
  excludedEntries: { domain: string; reason: string }[];
};
const confirm = useConfirm();
const toast = useToast();
const config = ref<RankConfig>({ months: [], blocked: [] });
const domain = ref("");
const busy = ref(false);
const error = ref("");
const previewError = ref("");
const preview = ref<Preview>();
const uploadForm = ref<HTMLFormElement>();
const draft = shallowRef<FormData>();
const showPreview = ref(false);

async function load() {
  config.value = await getJSON<RankConfig>("/api/admin/internet-rank");
}

async function run(action: () => Promise<void>) {
  if (busy.value) return;
  busy.value = true;
  error.value = "";
  try {
    await action();
    await load();
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "操作失败";
  } finally {
    busy.value = false;
  }
}

async function reviewUpload() {
  if (!uploadForm.value || busy.value) return;
  busy.value = true;
  error.value = "";
  previewError.value = "";
  try {
    const data = new FormData(uploadForm.value);
    const response = await getJSON<Preview>("/api/admin/internet-rank/preview", { method: "POST", body: data });
    data.set("previewVersion", response.version);
    draft.value = data;
    preview.value = response;
    showPreview.value = true;
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "预览失败";
  } finally {
    busy.value = false;
  }
}

async function publish() {
  if (busy.value || !draft.value) return;
  busy.value = true;
  previewError.value = "";
  try {
    const report = await getJSON<RankMonth>("/api/admin/internet-rank", { method: "POST", body: draft.value });
    uploadForm.value?.reset();
    showPreview.value = false;
    draft.value = undefined;
    toast.add({ severity: "success", summary: "月榜已发布", detail: report.month, life: 2600 });
    await load().catch(() => { error.value = "榜单已发布，列表暂未刷新。"; });
  } catch (failure) {
    previewError.value = failure instanceof Error ? failure.message : "发布失败";
  } finally {
    busy.value = false;
  }
}

function addBlock() {
  return run(async () => {
    await getJSON("/api/admin/internet-rank/blocked", {
      method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ domain: domain.value.trim() }),
    });
    domain.value = "";
    toast.add({ severity: "success", summary: "已屏蔽该域名及其子域名", life: 2500 });
  });
}

function removeBlock(name: string) {
  confirm.require({
    header: "解除月榜屏蔽",
    message: `解除 ${name} 的手动屏蔽后，符合条件的条目会重新显示。`,
    acceptLabel: "解除屏蔽", rejectLabel: "取消",
    accept: () => run(async () => {
      await getJSON(`/api/admin/internet-rank/blocked/${encodeURIComponent(name)}`, { method: "DELETE" });
      toast.add({ severity: "success", summary: "已解除屏蔽", life: 2500 });
    }),
  });
}

onMounted(() => load().catch(failure => {
  error.value = failure instanceof Error ? failure.message : "读取月榜失败";
}));
</script>

<template>
  <header class="admin-page-header">
    <div><h1>互联网月榜</h1><p>检查月度数据、预览排名并管理风险域名。</p></div>
    <PButton as="a" href="/rank/monthly" target="_blank" rel="noopener" label="查看公开榜单" icon="pi pi-external-link" icon-pos="right" severity="secondary" outlined />
  </header>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <div class="admin-rank-layout">
    <PCard class="admin-panel">
      <template #title><h2 class="admin-subheading">导入月榜数据</h2></template>
      <template #content>
        <form ref="uploadForm" class="admin-form-grid" @submit.prevent="reviewUpload">
          <div class="admin-field"><label for="rank-month">统计月份</label><PInputText id="rank-month" name="month" type="month" required /></div>
          <div class="admin-field"><label for="rank-source">数据来源</label><PInputText id="rank-source" name="source" required maxlength="128" /></div>
          <div class="admin-field admin-form-wide"><label for="rank-source-url">来源说明地址</label><PInputText id="rank-source-url" name="sourceUrl" type="url" placeholder="https://" required maxlength="512" /></div>
          <div class="admin-field admin-form-wide">
            <label for="rank-file">CSV 文件</label>
            <input id="rank-file" class="admin-file-input" name="file" type="file" accept=".csv,text/csv" required />
            <p class="admin-muted">每个文件对应一个已结束月份。必填列 <code>domain, category, visits</code>。</p>
            <p class="admin-muted">可选 <code>name, month, risk</code>。标记为 <code>blocked</code> 的条目会被剔除。</p>
          </div>
          <div class="admin-actions admin-form-wide"><PButton type="submit" icon="pi pi-eye" label="检查并预览" :loading="busy" /></div>
        </form>
      </template>
    </PCard>
    <PCard class="admin-panel">
      <template #title><h2 class="admin-subheading">风险域名</h2></template>
      <template #content>
        <p class="admin-muted">名单中的域名及其子域名会从所有月份的榜单移除。</p>
        <form class="admin-actions admin-rank-block" @submit.prevent="addBlock">
          <PInputText v-model="domain" placeholder="example.com" aria-label="风险域名" required maxlength="253" />
          <PButton type="submit" label="屏蔽" severity="danger" outlined :disabled="busy" />
        </form>
        <PDataTable :value="config.blocked" data-key="domain" class="admin-table" size="small">
          <template #empty>暂无手动屏蔽域名。</template>
          <PColumn field="domain" header="域名"><template #body="{ data }"><span class="admin-wrap">{{ data.domain }}</span></template></PColumn>
          <PColumn header="操作"><template #body="{ data }"><PButton size="small" text label="解除" :disabled="busy" @click="removeBlock(data.domain)" /></template></PColumn>
        </PDataTable>
      </template>
    </PCard>
  </div>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <h2 class="admin-subheading admin-toolbar">已发布月份</h2>
      <PDataTable :value="config.months" data-key="month" class="admin-table">
        <template #empty>暂无已发布月份。</template>
        <PColumn field="month" header="月份" />
        <PColumn field="source" header="数据来源" />
        <PColumn field="entries" header="导入条目" />
        <PColumn header="操作"><template #body="{ data }"><PButton as="a" :href="'/rank/monthly?month=' + encodeURIComponent(data.month)" target="_blank" rel="noopener" size="small" text label="查看榜单" /></template></PColumn>
      </PDataTable>
    </template>
  </PCard>

  <PDialog v-model:visible="showPreview" modal :draggable="false" header="发布预览" :closable="!busy" :close-on-escape="!busy" :style="{ width: 'min(700px, 94vw)' }">
    <template v-if="preview">
      <p class="admin-preview-summary"><strong>{{ preview.month }}</strong> · 导入 {{ preview.total }} 条 · 剔除 {{ preview.excluded }} 条 · 展示前 {{ preview.entries.length }} 名</p>
      <PMessage v-if="preview.version !== 'new'" severity="warn" class="admin-alert">确认发布后将替换该月份现有榜单。</PMessage>
      <PMessage v-if="previewError" severity="error" class="admin-alert">{{ previewError }}</PMessage>
      <PDataTable :value="preview.entries" data-key="Domain" class="admin-table" size="small" scrollable table-style="min-width: 400px">
        <template #empty>该文件没有可上榜的条目。</template>
        <PColumn header="名次" style="width: 60px"><template #body="{ index }">{{ index + 1 }}</template></PColumn>
        <PColumn header="网站"><template #body="{ data }"><div class="admin-primary">{{ data.Name }}</div><div class="admin-muted">{{ data.Domain }}</div></template></PColumn>
        <PColumn field="Category" header="分类" />
        <PColumn header="估算访问量"><template #body="{ data }">{{ BigInt(data.Visits).toLocaleString('zh-CN') }}</template></PColumn>
      </PDataTable>
      <details v-if="preview.excludedEntries?.length" class="admin-preview-exclusions">
        <summary>查看被剔除的域名（最多 10 条）</summary>
        <ul><li v-for="item in preview.excludedEntries" :key="item.domain">{{ item.domain }} · {{ item.reason }}</li></ul>
      </details>
    </template>
    <template #footer>
      <PButton label="返回修改" severity="secondary" outlined :disabled="busy" @click="showPreview = false" />
      <PButton label="确认发布" icon="pi pi-check" :loading="busy" @click="publish" />
    </template>
  </PDialog>
</template>
