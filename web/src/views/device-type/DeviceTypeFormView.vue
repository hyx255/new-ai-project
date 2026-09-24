<template>
  <div class="device-type-form-page">
    <div class="page-header">
      <h2>{{ isEdit ? "编辑设备类型" : "新建设备类型" }}</h2>
      <router-link to="/device-types" class="btn-link">返回列表</router-link>
    </div>

    <PageLoading v-if="loadingExisting" text="加载设备类型..." />

    <form v-else class="form-card" @submit.prevent="handleSubmit">
      <div class="form-group">
        <label class="form-label">名称 <span class="required">*</span></label>
        <input
          v-model="form.name"
          type="text"
          class="form-input"
          placeholder="例如：IP功放"
          required
        />
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">厂商 <span class="required">*</span></label>
          <input
            v-model="form.vendor"
            type="text"
            class="form-input"
            placeholder="例如：DSPPA"
            required
          />
        </div>
        <div class="form-group">
          <label class="form-label">型号 <span class="required">*</span></label>
          <input
            v-model="form.model"
            type="text"
            class="form-input"
            placeholder="例如：MP2806"
            required
          />
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">描述</label>
        <textarea
          v-model="form.description"
          class="form-input form-textarea"
          placeholder="设备类型描述（可选）"
          rows="3"
        ></textarea>
      </div>

      <div class="form-group">
        <label class="form-label">能力</label>
        <div class="capability-options">
          <label
            v-for="cap in ALL_CAPABILITIES"
            :key="cap"
            class="capability-option"
          >
            <input
              type="checkbox"
              :value="cap"
              v-model="form.capabilities"
            />
            <span class="capability-label">{{ CAPABILITY_LABELS[cap] }}</span>
            <span class="capability-code">{{ cap }}</span>
          </label>
        </div>
      </div>

      <div class="form-actions">
        <router-link to="/device-types" class="btn btn-secondary">
          取消
        </router-link>
        <button type="submit" class="btn btn-primary" :disabled="submitting">
          {{ submitting ? "保存中..." : isEdit ? "保存修改" : "创建" }}
        </button>
      </div>
    </form>

    <ToastNotification
      :visible="toast.visible"
      :message="toast.message"
      :type="toast.type"
      @close="toast.visible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  getDeviceType,
  createDeviceType,
  updateDeviceType,
} from "@/api/device-type";
import { extractErrorMessage } from "@/api/client";
import { ALL_CAPABILITIES, CAPABILITY_LABELS } from "@/types/device";
import type { Capability } from "@/types/device";
import PageLoading from "@/components/PageLoading.vue";
import ToastNotification from "@/components/ToastNotification.vue";

const route = useRoute();
const router = useRouter();

const isEdit = computed(() => !!route.params.id);
const loadingExisting = ref(false);
const submitting = ref(false);

const form = reactive({
  name: "",
  vendor: "",
  model: "",
  description: "",
  capabilities: [] as Capability[],
});

const toast = reactive({
  visible: false,
  message: "",
  type: "success" as "success" | "error",
});

async function loadExisting() {
  const id = route.params.id as string;
  if (!id) return;

  loadingExisting.value = true;
  const resp = await getDeviceType(id);

  if (resp.code === 0 && resp.data) {
    form.name = resp.data.name;
    form.vendor = resp.data.vendor;
    form.model = resp.data.model;
    form.description = resp.data.description;
    form.capabilities = [...(resp.data.capabilities || [])];
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
    setTimeout(() => router.push("/device-types"), 1500);
  }

  loadingExisting.value = false;
}

async function handleSubmit() {
  // Basic input validation (frontend only)
  if (!form.name.trim()) {
    toast.message = "请输入名称";
    toast.type = "error";
    toast.visible = true;
    return;
  }
  if (!form.vendor.trim()) {
    toast.message = "请输入厂商";
    toast.type = "error";
    toast.visible = true;
    return;
  }
  if (!form.model.trim()) {
    toast.message = "请输入型号";
    toast.type = "error";
    toast.visible = true;
    return;
  }

  submitting.value = true;

  const data = {
    name: form.name.trim(),
    vendor: form.vendor.trim(),
    model: form.model.trim(),
    description: form.description.trim(),
    capabilities: form.capabilities,
  };

  const resp = isEdit.value
    ? await updateDeviceType(route.params.id as string, data)
    : await createDeviceType(data);

  if (resp.code === 0) {
    toast.message = isEdit.value ? "设备类型已更新" : "设备类型已创建";
    toast.type = "success";
    toast.visible = true;
    setTimeout(() => router.push("/device-types"), 1000);
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
  }

  submitting.value = false;
}

onMounted(() => {
  if (isEdit.value) {
    loadExisting();
  }
});
</script>

<style scoped>
.device-type-form-page {
  max-width: 640px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  font-size: 22px;
}

.form-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 24px;
}

.form-group {
  margin-bottom: 18px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 6px;
}

.required {
  color: var(--color-danger);
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  color: var(--color-text);
  transition: border-color 0.15s;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.capability-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.capability-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}

.capability-option:hover {
  background: #f8fafc;
}

.capability-option input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary);
}

.capability-label {
  font-size: 14px;
  font-weight: 500;
  flex: 1;
}

.capability-code {
  font-size: 11px;
  color: var(--color-text-secondary);
  font-family: monospace;
  background: #f1f5f9;
  padding: 1px 6px;
  border-radius: 3px;
}

.form-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 24px;
  padding-top: 18px;
  border-top: 1px solid var(--color-border);
}
</style>
