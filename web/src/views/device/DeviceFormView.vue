<template>
  <div class="device-form-page">
    <div class="page-header">
      <h2>{{ isEdit ? "编辑设备" : "新建设备" }}</h2>
      <router-link to="/devices" class="btn-link">返回列表</router-link>
    </div>

    <PageLoading v-if="loadingExisting" text="加载设备信息..." />

    <form v-else class="form-card" @submit.prevent="handleSubmit">
      <div class="form-group">
        <label class="form-label">设备名称 <span class="required">*</span></label>
        <input
          v-model="form.name"
          type="text"
          class="form-input"
          placeholder="例如：1号教学楼功放"
          required
        />
      </div>

      <div class="form-group">
        <label class="form-label">设备类型 <span class="required">*</span></label>
        <select v-model="form.device_type_id" class="form-input" required>
          <option value="" disabled>请选择设备类型</option>
          <option
            v-for="dt in deviceTypes"
            :key="dt.id"
            :value="dt.id"
          >
            {{ dt.name }} ({{ dt.vendor }} - {{ dt.model }})
          </option>
        </select>
        <p v-if="deviceTypes.length === 0" class="form-hint">
          暂无设备类型，请先
          <router-link to="/device-types/create">创建设备类型</router-link>
        </p>
      </div>

      <div class="form-group">
        <label class="form-label">地址</label>
        <input
          v-model="form.address"
          type="text"
          class="form-input"
          placeholder="例如：192.168.1.100 或 simulator://localhost:9001"
        />
        <p class="form-hint">设备的网络地址或标识（可选）</p>
      </div>

      <div class="form-actions">
        <router-link to="/devices" class="btn btn-secondary">
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
import { getDevice, createDevice, updateDevice } from "@/api/device";
import { listDeviceTypes } from "@/api/device-type";
import { extractErrorMessage } from "@/api/client";
import type { DeviceType } from "@/types/device";
import PageLoading from "@/components/PageLoading.vue";
import ToastNotification from "@/components/ToastNotification.vue";

const route = useRoute();
const router = useRouter();

const isEdit = computed(() => !!route.params.id);
const loadingExisting = ref(false);
const submitting = ref(false);
const deviceTypes = ref<DeviceType[]>([]);

const form = reactive({
  name: "",
  device_type_id: "",
  address: "",
});

const toast = reactive({
  visible: false,
  message: "",
  type: "success" as "success" | "error",
});

async function loadDeviceTypes() {
  const resp = await listDeviceTypes();
  if (resp.code === 0 && resp.data) {
    deviceTypes.value = resp.data.items || [];
  }
}

async function loadExisting() {
  const id = route.params.id as string;
  if (!id) return;

  loadingExisting.value = true;

  // Load device types first
  await loadDeviceTypes();

  // Then load the device
  const resp = await getDevice(id);

  if (resp.code === 0 && resp.data) {
    form.name = resp.data.name;
    form.device_type_id = resp.data.device_type_id;
    form.address = resp.data.address || "";

    // If the current device_type_id is not in the list (edge case), add it
    if (
      form.device_type_id &&
      !deviceTypes.value.find((dt) => dt.id === form.device_type_id) &&
      resp.data.device_type
    ) {
      deviceTypes.value.unshift(resp.data.device_type);
    }
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
    setTimeout(() => router.push("/devices"), 1500);
  }

  loadingExisting.value = false;
}

async function handleSubmit() {
  // Basic input validation
  if (!form.name.trim()) {
    toast.message = "请输入设备名称";
    toast.type = "error";
    toast.visible = true;
    return;
  }
  if (!form.device_type_id) {
    toast.message = "请选择设备类型";
    toast.type = "error";
    toast.visible = true;
    return;
  }

  submitting.value = true;

  const data = {
    name: form.name.trim(),
    device_type_id: form.device_type_id,
    address: form.address.trim(),
  };

  const resp = isEdit.value
    ? await updateDevice(route.params.id as string, data)
    : await createDevice(data);

  if (resp.code === 0) {
    toast.message = isEdit.value ? "设备已更新" : "设备已创建";
    toast.type = "success";
    toast.visible = true;
    setTimeout(() => router.push("/devices"), 1000);
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
  }

  submitting.value = false;
}

onMounted(async () => {
  if (isEdit.value) {
    await loadExisting();
  } else {
    await loadDeviceTypes();
  }
});
</script>

<style scoped>
.device-form-page {
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

.form-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 4px;
}

.form-hint a {
  color: var(--color-primary);
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
