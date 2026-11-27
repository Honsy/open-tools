<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import { getJSON } from "../../api";

type Finding = {
  id: string;
  status: string;
  payload: {
    kind: string; name: string; official_url: string; evidence_urls: string[];
    published_on: string; category_slug: string; summary: string; body: string;
  };
  publication?: { summary: string; body: string };
  before?: { desc: string; body: string; slug: string };
};

const endpoint = "/api/admin/agent-discovery";
const categories: Record<string, string> = { "ai-ae0faa3203a9": "图像与设计", draw: "视频与短剧", "ai-audio": "音乐与语音" };
const tabs = [
  { label: "待审核", value: "pending" },
  { label: "已发布", value: "published" },
  { label: "已忽略", value: "ignored" },
];
const phases: Record<string, string> = {
  list_launch_sources: "读取信源", search_launches: "搜索发布线索",
  read_official_page: "核实官网与发布公告", lookup_catalog: "查询开物录已有条目",
};
const configured = ref(false);
const tab = ref("pending");
const items = ref<Finding[]>([]);
const loading = ref(false);
const sending = ref(false);
const page = ref(1);
const more = ref(false);
const error = ref("");
const taskStatus = ref("");
const selection = ref<Finding>();
const dialog = ref(false);
const summary = ref("");
const body = ref("");
const saving = ref(false);
let stream: EventSource | undefined;
let latest = 0;
let requestKey = "";

function message(cause: unknown) {
  return cause instanceof Error ? cause.message : "请求失败，请重试。";
}

function httpsLink(value: string) {
  try {
    const address = new URL(value);
    return address.protocol === "https:" ? address.href : undefined;
  } catch {
    return undefined;
  }
}

async function load(selectedPage = 1) {
  const request = ++latest;
  loading.value = true;
  error.value = "";
  try {
    const result = await getJSON<{ items: Finding[]; more: boolean }>(
      endpoint + "/findings?" + new URLSearchParams({ status: tab.value, page: String(selectedPage) }),
    );
    if (request !== latest) return;
    items.value = result.items;
    page.value = selectedPage;
    more.value = result.more;
  } catch (cause) {
    if (request === latest) error.value = message(cause);
  } finally {
    if (request === latest) loading.value = false;
  }
}

async function run() {
  sending.value = true;
  error.value = "";
  requestKey ||= crypto.randomUUID();
  try {
    const result = await getJSON<{ id: string; status: string }>(endpoint + "/runs", {
      method: "POST",
      headers: { "Idempotency-Key": requestKey },
    });
    requestKey = "";
    taskStatus.value = "任务已排队。";
    stream = new EventSource(endpoint + "/runs/" + encodeURIComponent(result.id) + "/events");
    stream.addEventListener("status", (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      if (data.status === "running") taskStatus.value = "Agent 正在运行。";
      if (data.status === "completed" || data.status === "failed") {
        stream?.close();
        sending.value = false;
        taskStatus.value = data.status === "completed" ? "执行完成，可以查看待审核候选。" : "任务失败，请在 Kaigong 管理台查看状态。";
        if (data.status === "completed") {
          tab.value = "pending";
          void load();
        }
      }
    });
    stream.addEventListener("tool_started", (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      taskStatus.value = phases[data.tool] || "Agent 正在调用工具。";
    });
    stream.onerror = () => {
      if (stream?.readyState === EventSource.CLOSED) {
        sending.value = false;
        taskStatus.value = "进度连接已关闭，请刷新候选或在 Kaigong 管理台查看任务。";
      } else {
        taskStatus.value = "正在重新连接任务进度。";
      }
    };
  } catch (cause) {
    error.value = message(cause);
    sending.value = false;
  }
}

function open(row: Finding) {
  selection.value = row;
  summary.value = row.publication?.summary || row.payload.summary;
  body.value = row.publication?.body || row.payload.body;
  error.value = "";
  dialog.value = true;
}

async function review(action: "published" | "ignored") {
  const row = selection.value;
  if (!row) return;
  saving.value = true;
  error.value = "";
  try {
    await getJSON(endpoint + "/findings/" + row.id + "/review", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action, summary: summary.value, body: body.value,
        previousDesc: row.before?.desc || "", previousBody: row.before?.body || "" }),
    });
    dialog.value = false;
    await load(page.value);
  } catch (cause) {
    error.value = message(cause);
  } finally {
    saving.value = false;
  }
}

onMounted(async () => {
  try {
    configured.value = (await getJSON<{ configured: boolean }>(endpoint + "/config")).configured;
    if (configured.value) await load();
  } catch (cause) {
    error.value = message(cause);
  }
});
onBeforeUnmount(() => stream?.close());
</script>

<template>
  <header class="admin-page-header">
    <div><h1>Agent 发现</h1><p>核对 AIGC 新品和功能更新的官方出处，确认后发布到开物录。</p></div>
    <PButton label="立即发现" icon="pi pi-play" :disabled="!configured" :loading="sending" @click="run" />
  </header>
  <PMessage v-if="!dialog && error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PMessage v-if="taskStatus" severity="info" class="admin-alert" role="status">{{ taskStatus }}</PMessage>
  <PMessage v-if="!configured" severity="info">开工 Agent 尚未接入。请在服务端配置项目密钥和 Agent ID。</PMessage>
  <PCard v-else class="admin-panel admin-table-panel"><template #content>
    <div class="admin-toolbar">
      <PSelectButton v-model="tab" :options="tabs" option-label="label" option-value="value" :allow-empty="false" aria-label="候选状态" @change="load()" />
      <PButton label="刷新" icon="pi pi-refresh" text @click="load(page)" />
    </div>
    <PDataTable :value="items" :loading="loading" data-key="id" class="admin-table" scrollable>
      <template #empty>当前状态没有候选。</template>
      <PColumn header="产品"><template #body="{ data }"><strong>{{ data.payload.name }}</strong><div class="admin-muted">{{ data.payload.summary }}</div></template></PColumn>
      <PColumn header="类型"><template #body="{ data }">{{ data.payload.kind === "new" ? "新增" : "更新" }}</template></PColumn>
      <PColumn header="分类"><template #body="{ data }">{{ categories[data.payload.category_slug] || data.payload.category_slug }}</template></PColumn>
      <PColumn header="发布日期"><template #body="{ data }">{{ data.payload.published_on }}</template></PColumn>
      <PColumn header="操作"><template #body="{ data }"><PButton :label="data.status === 'pending' ? '审核' : data.status === 'publishing' ? '继续发布' : '查看'" text size="small" @click="open(data)" /></template></PColumn>
    </PDataTable>
    <div v-if="page > 1 || more" class="admin-toolbar">
      <PButton label="上一页" text :disabled="page < 2 || loading" @click="load(page - 1)" />
      <span>第 {{ page }} 页</span>
      <PButton label="下一页" text :disabled="!more || loading" @click="load(page + 1)" />
    </div>
  </template></PCard>
  <PDialog v-model:visible="dialog" modal :header="selection?.payload.name || '候选详情'" :style="{ width: 'min(760px, 94vw)' }">
    <template v-if="selection">
      <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
      <p><a :href="httpsLink(selection.payload.official_url)" target="_blank" rel="noopener noreferrer">打开官方入口</a></p>
      <p>官方出处</p>
      <ul><li v-for="link in selection.payload.evidence_urls" :key="link"><a :href="httpsLink(link)" target="_blank" rel="noopener noreferrer" style="overflow-wrap:anywhere">{{ link }}</a></li></ul>
      <details v-if="selection.before">
        <summary>查看当前详情</summary>
        <p>{{ selection.before.desc }}</p>
        <pre style="white-space:pre-wrap">{{ selection.before.body }}</pre>
      </details>
      <p><label for="finding-summary">简介</label></p>
      <PTextarea id="finding-summary" v-model="summary" rows="2" maxlength="220" style="width:100%" :readonly="selection.status !== 'pending'" />
      <p><label for="finding-body">网站介绍（Markdown）</label></p>
      <PTextarea id="finding-body" v-model="body" rows="8" maxlength="3000" style="width:100%" :readonly="selection.status !== 'pending'" />
    </template>
    <template #footer>
      <PButton v-if="selection?.status === 'pending'" label="忽略" severity="secondary" text :disabled="saving" @click="review('ignored')" />
      <PButton v-if="selection?.status === 'pending' || selection?.status === 'publishing'"
        :label="selection.status === 'publishing' ? '继续发布' : '批准发布'" :loading="saving" @click="review('published')" />
    </template>
  </PDialog>
</template>
