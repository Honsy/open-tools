const result = document.querySelector("[data-lab-result]");
const error = document.querySelector("[data-lab-error]");

function el(tag, text, css) {
  const item = document.createElement(tag);
  if (text !== undefined) item.textContent = text;
  if (css) item.className = css;
  return item;
}

function show(...nodes) {
  result.replaceChildren(...nodes);
  result.hidden = false;
}

function summary(rows) {
  const list = el("dl", undefined, "grid gap-3 sm:grid-cols-2");
  for (const [label, value] of rows) {
    const item = el("div", undefined, "border border-base-300 rounded p-3");
    item.append(el("dt", label, "text-sm opacity-70"), el("dd", value, "whitespace-pre-wrap"));
    list.append(item);
  }
  return list;
}

const resultLabels = {
  zodiac: "星座", date: "日期", overall: "综合解读", love: "感情", career: "事业", wealth: "财运", lucky: "幸运提示",
  score: "评分", summary: "总结", zodiacA: "星座一", zodiacB: "星座二", loveScore: "感情评分", workScore: "工作评分",
  friendScore: "朋友评分", strengths: "优势", challenges: "挑战", advice: "建议", category: "主题", verse: "签文",
  interpret: "解签", luckyColor: "幸运色", luckyNumber: "幸运数字", luckyDir: "幸运方向", question: "问题",
  name: "牌名", meaning: "寓意", position: "位置", psychological: "心理视角", lifeAdvice: "生活启示", tags: "关键词",
  teamA: "球队一", teamB: "球队二", teamAWinProb: "球队一胜率", teamBWinProb: "球队二胜率", drawProb: "平局概率",
  predictedScore: "模拟比分", analysis: "分析", keyFactors: "关键因素", stage: "比赛阶段", confidence: "置信度",
};

function textResults(data) {
  const entries = Object.entries(data).filter(([key, value]) => resultLabels[key] && value !== "");
  const rows = entries.map(([key, value]) => [resultLabels[key], key.endsWith("Prob") ? value + "%" : String(value)]);
  const cards = Array.isArray(data.cards) ? data.cards.map(card => summary(
    Object.entries(card).filter(([key]) => resultLabels[key]).map(([key, value]) => [resultLabels[key], String(value)])
  )) : [];
  show(...cards, summary(rows));
}

function imageResults(images) {
  const gallery = el("div", undefined, "grid gap-3 sm:grid-cols-2");
  for (const image of images) {
    const url = new URL(image);
    if (url.protocol !== "https:") continue;
    const tile = el("a", undefined, "link link-hover");
    tile.href = url.href;
    tile.target = "_blank";
    tile.rel = "noopener noreferrer";
    const preview = el("img", undefined, "rounded w-full");
    preview.src = url.href;
    preview.alt = "生成结果";
    preview.referrerPolicy = "no-referrer";
    tile.append(preview, el("span", "查看原图"));
    gallery.append(tile);
  }
  show(gallery);
}

const aiForm = document.querySelector("[data-ai-tool]");
if (aiForm) {
  aiForm.addEventListener("submit", async event => {
    event.preventDefault();
    const button = aiForm.querySelector("button");
    button.disabled = true;
    button.textContent = "生成中…";
    error.hidden = result.hidden = true;
    const values = new FormData(aiForm);
    const multipart = aiForm.enctype === "multipart/form-data";
    try {
      const response = await fetch("/api/tools/" + aiForm.dataset.aiTool, {
        method: "POST",
        headers: multipart ? {} : {"Content-Type": "application/json"},
        body: multipart ? values : JSON.stringify(Object.fromEntries(values)),
      });
      let data = await response.json();
      if (!response.ok) throw new Error(data.error || "工具暂时不可用");
      if (data.job) data = await waitForImage(data.job);
      if (Array.isArray(data.images)) imageResults(data.images);
      else textResults(data);
    } catch (failure) {
      error.textContent = failure.message || "请求失败，请重试";
      error.hidden = false;
    } finally {
      button.disabled = false;
      button.textContent = "生成";
    }
  });
}

async function waitForImage(id) {
  for (let attempt = 0; attempt < 84; attempt++) {
    await new Promise(resolve => setTimeout(resolve, 3000));
    const response = await fetch("/api/tools/image-jobs/" + encodeURIComponent(id));
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || "图片生成失败");
    if (response.status !== 202) return data;
  }
  throw new Error("图片生成超时，请稍后重试");
}

export function loanSchedule(amount, annualRate, months, repayment) {
  const rate = annualRate / 1200;
  const fixedPayment = rate === 0 ? amount / months : amount * rate / -Math.expm1(-months * Math.log1p(rate));
  let remaining = amount;
  return Array.from({length: months}, (_, index) => {
    const interest = remaining * rate;
    const principal = index === months - 1 ? remaining : repayment === "principal" ? amount / months : fixedPayment - interest;
    remaining = Math.max(0, remaining - principal);
    return {payment: principal + interest, principal, interest, remaining};
  });
}

const numberFormat = new Intl.NumberFormat("zh-CN", {minimumFractionDigits: 2, maximumFractionDigits: 2});
const mortgage = document.querySelector("#mortgage-form");
if (mortgage) {
  const kind = mortgage.elements.namedItem("type");
  kind.addEventListener("change", () => {
    const commercial = mortgage.querySelector("[data-commercial]");
    const fund = mortgage.querySelector("[data-fund]");
    commercial.hidden = commercial.disabled = kind.value === "fund";
    fund.hidden = fund.disabled = kind.value === "commercial";
    result.hidden = true;
  });
  mortgage.addEventListener("submit", event => {
    event.preventDefault();
    const form = new FormData(mortgage);
    const months = Number(form.get("years")) * 12;
    const loanTypes = ["commercial", "fund"].filter(key => form.has(key));
    const plans = loanTypes.map(key => loanSchedule(Number(form.get(key)) * 10000, Number(form.get(key + "Rate")), months, form.get("repayment")));
    const schedule = plans[0].map((row, index) => {
      const second = plans[1]?.[index];
      return second ? Object.fromEntries(Object.keys(row).map(key => [key, row[key] + second[key]])) : row;
    });
    const total = key => schedule.reduce((value, row) => value + row[key], 0);
    const money = value => numberFormat.format(value) + " 元";
    const totals = [["首月月供", money(schedule[0].payment)], ["贷款本金", money(total("principal"))], ["总利息", money(total("interest"))], ["本息合计", money(total("payment"))]];
    if (form.get("repayment") === "principal") totals.push(["每月递减", money(schedule[0].payment - schedule[1].payment)]);
    const details = el("details");
    details.append(el("summary", "查看每月还款明细", "cursor-pointer font-medium"));
    const wrapper = el("div", undefined, "overflow-x-auto");
    const table = el("table", undefined, "table table-sm");
    const headers = table.createTHead().insertRow();
    for (const name of ["月份", "月供（元）", "本金（元）", "利息（元）", "剩余本金（元）"]) headers.append(el("th", name));
    const body = table.createTBody();
    schedule.forEach((row, index) => {
      const tr = body.insertRow();
      tr.insertCell().textContent = index + 1;
      for (const key of ["payment", "principal", "interest", "remaining"]) tr.insertCell().textContent = numberFormat.format(row[key]);
    });
    wrapper.append(table);
    details.append(wrapper);
    show(summary(totals), details);
  });
}

const retirement = document.querySelector("#retirement-form");
const asDate = index => Math.floor(index / 12) + " 年 " + (index % 12 + 1) + " 月";
function duration(months) {
  return Math.floor(months / 12) + " 年" + (months % 12 ? " " + (months % 12) + " 个月" : "");
}

export function retirementPlan(year, month, originalAge) {
  const originalDate = (year + originalAge) * 12 + month - 1;
  const elapsed = originalDate - 2025 * 12;
  const delay = elapsed < 0 ? 0 : Math.min(originalAge === 50 ? 60 : 36, Math.floor(elapsed / (originalAge === 50 ? 2 : 4)) + 1);
  const date = originalDate + delay;
  const minPaid = 180 + Math.min(10, Math.max(0, Math.floor(date / 12) - 2029)) * 6;
  return {date, delay, age: originalAge * 12 + delay, minPaid, earliest: Math.max(originalDate, date - 36), latest: date + 36};
}

if (retirement) {
  const now = new Date();
  retirement.elements.namedItem("year").max = now.getFullYear();
  retirement.addEventListener("submit", event => {
    event.preventDefault();
    const form = new FormData(retirement);
    const plan = retirementPlan(Number(form.get("year")), Number(form.get("month")), Number(form.get("category")));
    const totals = [["法定退休时间", asDate(plan.date)], ["法定退休年龄", duration(plan.age)], ["延迟月数", plan.delay + " 个月"], ["届时最低缴费年限", duration(plan.minPaid)]];
    if (plan.date >= 2025 * 12) totals.push(["弹性提前退休最早月份", asDate(plan.earliest)], ["弹性延迟退休最晚月份", asDate(plan.latest)]);
    if (form.get("paid") !== "") totals.push(["距最低缴费年限还差", duration(Math.max(0, plan.minPaid - Number(form.get("paid")) * 12))]);
    const monthsLeft = plan.date - (now.getFullYear() * 12 + now.getMonth());
    if (monthsLeft > 0) totals.push(["距法定退休还有", duration(monthsLeft)]);
    show(summary(totals), el("p", "弹性提前退休需达到最低缴费年限，弹性延迟退休需与单位协商。", "text-sm opacity-70"));
  });
}
