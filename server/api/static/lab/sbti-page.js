import { buildShuffledQuestionList, computeSbtiResult, getDimExplanation, getVisibleQuestions } from "./sbti-result.js";
import { dimensionMeta, dimensionOrder } from "./sbti-data.js";

const form = document.querySelector("#sbti-form");
const progress = form.querySelector("progress");
const prompt = form.querySelector("[data-progress]");
const list = form.querySelector("[data-questions]");
const submit = form.querySelector("button");
const panel = document.querySelector("[data-lab-result]");
let questions = [];
let answers = {};

function make(tag, text, classes) {
  const node = document.createElement(tag);
  if (text !== undefined) node.textContent = text;
  if (classes) node.className = classes;
  return node;
}

function showQuestions() {
  const visible = getVisibleQuestions(questions, answers);
  const count = visible.filter(question => answers[question.id] !== undefined).length;
  progress.max = visible.length;
  progress.value = count;
  prompt.textContent = "已回答 " + count + " / " + visible.length;
  submit.disabled = count !== visible.length;
  const items = visible.map((question, index) => {
    const item = make("fieldset", undefined, "border border-base-300 rounded p-3");
    item.append(make("legend", "第 " + (index + 1) + " 题"), make("p", question.text, "mb-2"));
    for (const option of question.options) {
      const label = make("label", undefined, "flex gap-2 items-center py-1");
      const radio = make("input", undefined, "radio radio-sm");
      radio.type = "radio";
      radio.name = question.id;
      radio.value = option.value;
      radio.checked = answers[question.id] === option.value;
      label.append(radio, make("span", option.label));
      item.append(label);
    }
    return item;
  });
  list.replaceChildren(...items);
}

form.addEventListener("change", event => {
  const answer = event.target;
  if (answer.type !== "radio") return;
  answers[answer.name] = Number(answer.value);
  if (answer.name === "drink_gate_q1" && answers[answer.name] !== 3) delete answers.drink_gate_q2;
  showQuestions();
  Array.from(form.elements).find(item => item.name === answer.name && item.value === answer.value)?.focus({preventScroll: true});
});

function start() {
  questions = buildShuffledQuestionList();
  answers = {};
  form.hidden = false;
  panel.hidden = true;
  showQuestions();
}

form.addEventListener("submit", event => {
  event.preventDefault();
  if (submit.disabled) return;
  const result = computeSbtiResult(answers);
  const headline = make("h2", result.finalType.code + "（" + result.finalType.cn + "）", "text-xl font-semibold");
  const dimensions = make("details", undefined, "border border-base-300 rounded p-3");
  dimensions.append(make("summary", "查看各项得分", "cursor-pointer font-semibold"));
  for (const key of dimensionOrder) {
    dimensions.append(make("h3", dimensionMeta[key].name + " · " + result.rawScores[key] + " 分", "font-semibold mt-3"),
      make("p", getDimExplanation(key, result.levels[key])));
  }
  const credit = make("p", "测试仅供娱乐，题库及人格库源自 SBTI（原作者抖音号 48669855166）。", "text-sm opacity-70");
  const restart = make("button", "重新测试", "btn btn-primary justify-self-start");
  restart.addEventListener("click", start);
  panel.replaceChildren(headline, make("p", result.badge), make("p", result.finalType.intro), make("p", result.sub), make("p", result.finalType.desc));
  if (result.secondaryType) panel.append(make("p", "次要类型：" + result.secondaryType.code + "（" + result.secondaryType.cn + "）"));
  panel.append(dimensions, credit, restart);
  form.hidden = true;
  panel.hidden = false;
  panel.scrollIntoView({block: "start"});
});

start();
