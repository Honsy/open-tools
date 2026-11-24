<script setup lang="ts">
import { onMounted, ref } from "vue";
import { getJSON } from "../../api";

type RankMonth = { month: string; source: string; entries: number };
type RankConfig = { months: RankMonth[]; blocked: { domain: string }[] };

const config = ref<RankConfig>({ months: [], blocked: [] });
const domain = ref("");
const busy = ref(false);
const error = ref("");
const message = ref("");

async function run(action: () => Promise<void>) {
  busy.value = true;
  error.value = "";
  message.value = "";
  try {
    await action();
    config.value = await getJSON<RankConfig>("/api/admin/internet-rank");
  } catch (failure) {
    error.value = failure instanceof Error ? failure.message : "操作失败";
  } finally {
    busy.value = false;
  }
}

async function upload(event: Event) {
  const form = event.currentTarget as HTMLFormElement;
  const body = new FormData(form);
  await run(async () => {
    const data = await getJSON<RankMonth>("/api/admin/internet-rank", { method: "POST", body });
    form.reset();
    message.value = `已发布 ${data.month} 月榜，共导入 ${data.entries} 个域名。`;
  });
}

async function addBlock() {
  await run(async () => {
    await getJSON("/api/admin/internet-rank/blocked", {
      method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ domain: domain.value }),
    });
    domain.value = "";
    message.value = "已屏蔽该域名及其子域名。";
  });
}

async function removeBlock(name: string) {
  await run(async () => {
    await getJSON(`/api/admin/internet-rank/blocked/${encodeURIComponent(name)}`, { method: "DELETE" });
    message.value = `已解除对 ${name} 的手动屏蔽。`;
  });
}

onMounted(() => run(async () => {}));
</script>

<template>
  <div class="adm-page-heading">
    <h1>互联网月榜</h1>
    <a class="btn" href="/rank/monthly" target="_blank" rel="noopener">查看公开榜单</a>
  </div>
  <p v-if="error" class="err" role="alert">{{ error }}</p>
  <p v-if="message" class="adm-message" role="status">{{ message }}</p>

  <section class="adm-card adm-stack">
    <h2>导入月榜数据</h2>
    <p>使用获准公开展示的第三方数据。每个 CSV 对应一个已结束月份，统计范围为全球桌面和移动网页。同月重新导入会替换原有当月榜单。</p>
    <p>必填列 <code>domain,category,visits</code>。可选 <code>name,month,risk</code>，风险值 <code>blocked</code> 的记录会被隐藏。</p>
    <form class="adm-form" @submit.prevent="upload">
      <label>统计月份<input name="month" type="month" required /></label>
      <label class="wide">数据来源<input name="source" required maxlength="128" /></label>
      <label class="wide">来源说明地址<input name="sourceUrl" type="url" placeholder="https://" required maxlength="512" /></label>
      <label>CSV 文件<input name="file" type="file" accept=".csv,text/csv" required /></label>
      <button class="btn primary" type="submit" :disabled="busy">导入并发布</button>
    </form>
    <ul v-if="config.months.length" class="adm-simple-list">
      <li v-for="item in config.months" :key="item.month">{{ item.month }} · {{ item.source }} · {{ item.entries }} 条导入记录</li>
    </ul>
    <p v-else-if="!busy">暂无已发布月份。</p>
  </section>

  <section class="adm-card adm-stack">
    <h2>月榜风险域名</h2>
    <p>名单中的域名及其子域名会从所有月份的榜单中移除。</p>
    <form class="adm-form" @submit.prevent="addBlock">
      <label class="wide">域名<input v-model.trim="domain" placeholder="example.com" required maxlength="253" /></label>
      <button class="btn" type="submit" :disabled="busy">屏蔽</button>
    </form>
    <ul v-if="config.blocked.length" class="adm-simple-list">
      <li v-for="item in config.blocked" :key="item.domain" class="adm-list-action">
        <span>{{ item.domain }}</span><button class="btn" type="button" :disabled="busy" @click="removeBlock(item.domain)">解除</button>
      </li>
    </ul>
  </section>
</template>
