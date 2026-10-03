<script setup lang="ts">
import { reactive, ref } from "vue";
import Shell from "../components/Shell.vue";
import { getJSON } from "../api";

const form = reactive({ name: "", url: "", desc: "" });
const message = ref("");
const error = ref("");
const sending = ref(false);

async function send() {
  message.value = "";
  error.value = "";
  sending.value = true;
  try {
    const data = await getJSON<{ message: string }>("/api/submit", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(form),
    });
    message.value = data.message;
    form.name = "";
    form.url = "";
    form.desc = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "没有提交上";
  } finally {
    sending.value = false;
  }
}
</script>

<template>
  <Shell>
    <form class="form" @submit.prevent="send">
      <h1>提交网站</h1>
      <p class="kicker">先收进待审，不会马上出现在首页。</p>
      <label>名字<input v-model="form.name" required maxlength="40" /></label>
      <label>网址<input v-model="form.url" required placeholder="https://" /></label>
      <label>一句简介<textarea v-model="form.desc" rows="3" maxlength="120" /></label>
      <button type="submit" :disabled="sending">{{ sending ? "提交中" : "提交" }}</button>
      <p v-if="message" class="ok">{{ message }}</p>
      <p v-if="error" class="banner">{{ error }}</p>
    </form>
  </Shell>
</template>
