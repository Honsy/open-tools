<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useConfirm } from "primevue/useconfirm";
import { useToast } from "primevue/usetoast";
import { getJSON, formatDate, categoryLabel, type AdminPage, type AdminCategory } from "../../api";

type Submission = { id: number; name: string; url: string; desc: string; categorySlug: string; userId: number; submitter: string; createdAt: string; duplicate: boolean };
type Decision = "approve" | "reject" | "block";
type History = { id: number; name: string; url: string; decision: Decision; categorySlug: string; reviewer: string; note: string; createdAt: string };
type Stats = { online: number; pending: number; off: number };
const tabs = [{ label: "待审投稿", value: "pending" }, { label: "审核记录", value: "history" }];
const decisions = { approve: { label: "已通过", color: "success" }, reject: { label: "已拒绝", color: "secondary" }, block: { label: "风险屏蔽", color: "danger" } };
const metrics = [
  { key: "pending", label: "待审核", icon: "pi-inbox" },
  { key: "online", label: "已上架", icon: "pi-check-circle" },
  { key: "off", label: "已下架", icon: "pi-minus-circle" },
] as const;
const toast = useToast();
const confirm = useConfirm();
const stats = ref<Stats>();
const categories = ref<AdminCategory[]>([]);
const tab = ref("pending");
const pending = ref<AdminPage<Submission>>({ items: [], total: 0, page: 1, pageSize: 25 });
const history = ref<AdminPage<History>>({ items: [], total: 0, page: 1, pageSize: 25 });
const selected = ref<Submission>();
const drawer = ref(false);
const category = ref("");
const note = ref("");
const loading = ref(true);
const busy = ref(false);
const error = ref("");
const reviewError = ref("");
let latest = 0;

const categoryOptions = computed(() => categories.value.filter(item => item.kind === "tab" && item.contentKind === "links").map(item => ({
  value: item.slug, name: categoryLabel(item, categories.value),
})));
const activePage = computed(() => tab.value === "pending" ? pending.value : history.value);
const selectedURL = computed(() => {
  try {
    const url = new URL(selected.value?.url || "");
    if (["http:", "https:"].includes(url.protocol) && !url.username && !url.password) return url.href;
  } catch {
    // Keep malformed historical URLs visible as text.
  }
  return "";
});

async function load(page = activePage.value.page) {
  const request = ++latest;
  const state = tab.value;
  loading.value = true;
  error.value = "";
  try {
    if (state === "pending") {
      const response = await getJSON<AdminPage<Submission>>(`/api/admin/submissions?page=${page}`);
      if (request === latest) pending.value = response;
    } else {
      const response = await getJSON<AdminPage<History>>(`/api/admin/submission-reviews?page=${page}`);
      if (request === latest) history.value = response;
    }
  } catch (failure) {
    if (request === latest) error.value = failure instanceof Error ? failure.message : "读取审核列表失败";
  } finally {
    if (request === latest) loading.value = false;
  }
}

async function loadStats() {
  try {
    stats.value = await getJSON<Stats>("/api/admin/stats");
  } catch (failure) {
    toast.add({ severity: "error", summary: "统计读取失败", detail: failure instanceof Error ? failure.message : undefined, life: 3500 });
  }
}

function openReview(row: Submission) {
  selected.value = row;
  category.value = categoryOptions.value.some(option => option.value === row.categorySlug) ? row.categorySlug : "";
  note.value = "";
  reviewError.value = "";
  drawer.value = true;
}

async function submit(decision: Decision) {
  if (!selected.value || busy.value || (decision === "approve" && !category.value)) return;
  busy.value = true;
  reviewError.value = "";
  try {
    await getJSON(`/api/admin/submissions/${selected.value.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ decision, categorySlug: category.value, note: note.value }),
    });
    toast.add({ severity: "success", summary: decisions[decision].label, detail: selected.value.name, life: 2600 });
    drawer.value = false;
    await Promise.all([load(), loadStats()]);
  } catch (failure) {
    reviewError.value = failure instanceof Error ? failure.message : "审核失败";
  } finally {
    busy.value = false;
  }
}

function block() {
  confirm.require({
    header: "拒绝并屏蔽月榜域名",
    message: "这条投稿将被拒绝。该域名及其子域名也会从所有月份的互联网月榜移除。",
    acceptLabel: "确认屏蔽", rejectLabel: "取消",
    acceptProps: { severity: "danger" },
    accept: () => submit("block"),
  });
}

onMounted(() => {
  void load();
  void loadStats();
  void getJSON<AdminCategory[]>("/api/admin/categories").then(result => { categories.value = result; }).catch(failure => {
    toast.add({ severity: "error", summary: "分类读取失败", detail: failure instanceof Error ? failure.message : undefined, life: 3500 });
  });
});
</script>

<template>
  <header class="admin-page-header">
    <div><h1>网址审核</h1><p>核对用户提交的外链，选择分类后发布。</p></div>
    <PButton as="a" href="/submit" target="_blank" rel="noopener" label="公开提交入口" icon="pi pi-external-link" icon-pos="right" severity="secondary" outlined />
  </header>
  <div class="admin-stats">
    <section v-for="item in metrics" :key="item.key" class="admin-metric">
      <div><span class="admin-muted">{{ item.label }}</span><strong>{{ stats?.[item.key] ?? "—" }}</strong></div>
      <span class="admin-metric-icon" aria-hidden="true"><i class="pi" :class="item.icon"></i></span>
    </section>
  </div>
  <PMessage v-if="error" severity="error" class="admin-alert">{{ error }}</PMessage>
  <PCard class="admin-panel admin-table-panel">
    <template #content>
      <div class="admin-toolbar">
        <PSelectButton v-model="tab" :options="tabs" option-label="label" option-value="value" :allow-empty="false" aria-label="审核列表" @change="load()" />
        <PButton label="刷新" icon="pi pi-refresh" text severity="secondary" :disabled="loading" @click="load()" />
      </div>
      <PDataTable v-if="tab === 'pending'" :value="pending.items" :loading="loading" data-key="id" class="admin-table admin-adaptive" scrollable table-style="min-width: 690px">
        <template #empty>没有待审投稿。</template>
        <PColumn header="网站" header-class="admin-mobile-title" body-class="admin-mobile-title" style="min-width: 250px; max-width: 395px">
          <template #body="{ data }"><div class="admin-primary">{{ data.name }}</div><div class="admin-muted admin-truncate" :title="data.url">{{ data.url }}</div><div class="admin-mobile-meta admin-muted">{{ data.submitter || (data.userId ? '#' + data.userId : '—') }} · {{ formatDate(data.createdAt) }}<span v-if="data.duplicate"> · 有重复</span></div></template>
        </PColumn>
        <PColumn header="投稿人" header-class="admin-hide-mobile" body-class="admin-hide-mobile" style="min-width: 100px"><template #body="{ data }">{{ data.submitter || (data.userId ? '#' + data.userId : '—') }}</template></PColumn>
        <PColumn header="提交日期" header-class="admin-hide-mobile" body-class="admin-hide-mobile" style="min-width: 100px"><template #body="{ data }">{{ formatDate(data.createdAt) }}</template></PColumn>
        <PColumn header="核对" header-class="admin-hide-mobile" body-class="admin-hide-mobile" style="min-width: 120px"><template #body="{ data }"><PTag v-if="data.duplicate" value="已收录相同网址" severity="warn" /><span v-else class="admin-muted">—</span></template></PColumn>
        <PColumn header="操作" header-class="admin-mobile-action" body-class="admin-mobile-action" style="min-width: 95px"><template #body="{ data }"><PButton label="审核" icon="pi pi-arrow-right" icon-pos="right" text size="small" @click="openReview(data)" /></template></PColumn>
      </PDataTable>
      <PDataTable v-else :value="history.items" :loading="loading" data-key="id" class="admin-table" scrollable table-style="min-width: 720px">
        <template #empty>暂无审核记录。</template>
        <PColumn header="网站" style="min-width: 215px; max-width: 340px"><template #body="{ data }"><div class="admin-primary">{{ data.name }}</div><div class="admin-muted admin-truncate" :title="data.url">{{ data.url }}</div></template></PColumn>
        <PColumn header="结果" style="min-width: 115px"><template #body="{ data }"><PTag :value="decisions[data.decision as Decision]?.label" :severity="decisions[data.decision as Decision]?.color || 'secondary'" /><div v-if="data.categorySlug" class="admin-muted">{{ categoryOptions.find(item => item.value === data.categorySlug)?.name || data.categorySlug }}</div></template></PColumn>
        <PColumn field="reviewer" header="审核人" style="min-width: 100px" />
        <PColumn header="审核日期" style="min-width: 100px"><template #body="{ data }">{{ formatDate(data.createdAt) }}</template></PColumn>
        <PColumn header="备注" style="min-width: 160px"><template #body="{ data }"><div class="admin-wrap">{{ data.note || "—" }}</div></template></PColumn>
      </PDataTable>
      <div class="admin-pagination"><span class="admin-muted">共 {{ activePage.total }} 条</span><PPaginator :first="(activePage.page - 1) * activePage.pageSize" :rows="activePage.pageSize" :total-records="activePage.total" :page-link-size="3" @page="load($event.page + 1)" /></div>
    </template>
  </PCard>

  <PDrawer aria-label="审核投稿" v-model:visible="drawer" header="审核投稿" position="right" class="admin-drawer" :dismissable="false">
    <template v-if="selected">
      <div class="admin-detail">
        <h2>{{ selected.name }}</h2>
        <a v-if="selectedURL" :href="selectedURL" target="_blank" rel="noopener noreferrer" class="admin-wrap">{{ selected.url }}</a>
        <span v-else class="admin-wrap">{{ selected.url }}</span>
        <p v-if="selected.desc" class="admin-wrap">{{ selected.desc }}</p>
        <p class="admin-muted">投稿人 {{ selected.submitter || (selected.userId ? '#' + selected.userId : '—') }} · {{ formatDate(selected.createdAt) }}</p>
      </div>
      <PMessage v-if="selected.duplicate" severity="warn" class="admin-alert">站内已收录相同网址，请核对后再通过。</PMessage>
      <PMessage v-if="reviewError" severity="error" class="admin-alert">{{ reviewError }}</PMessage>
      <div class="admin-field"><label for="review-category">发布分类</label><PSelect input-id="review-category" aria-label="发布分类" v-model="category" :options="categoryOptions" option-label="name" option-value="value" filter placeholder="选择子分类" /></div>
      <div class="admin-field"><label for="review-note">审核备注</label><PTextarea id="review-note" v-model="note" rows="3" maxlength="300" placeholder="选填，便于回查审核原因" /></div>
    </template>
    <template #footer>
      <div class="admin-actions admin-drawer-actions">
        <PButton severity="danger" text label="拒绝并屏蔽月榜域名" :disabled="busy" @click="block" />
        <PButton severity="secondary" outlined label="拒绝投稿" :disabled="busy" @click="submit('reject')" />
        <PButton label="通过并上架" icon="pi pi-check" :loading="busy" :disabled="!category" @click="submit('approve')" />
      </div>
    </template>
  </PDrawer>
</template>
