<template>
  <div class="device-detail-page">
    <div class="page-header">
      <h2>设备详情</h2>
      <div class="header-actions">
        <router-link to="/devices" class="btn-link">返回列表</router-link>
        <router-link :to="`/devices/${deviceId}/edit`" class="btn btn-secondary">
          编辑
        </router-link>
      </div>
    </div>

    <PageLoading v-if="loading" text="加载设备信息..." />

    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="fetchData">重试</button>
    </div>

    <template v-else-if="device">
      <div class="detail-card">
        <div class="detail-header">
          <h3 class="device-name">{{ device.name }}</h3>
          <StatusBadge :status="device.status" />
        </div>

        <div class="detail-grid">
          <div class="detail-item">
            <span class="detail-label">设备 ID</span>
            <span class="detail-value mono">{{ device.id }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">设备类型</span>
            <span class="detail-value">
              <template v-if="device.device_type">
                {{ device.device_type.name }}
                <span class="text-secondary">
                  ({{ device.device_type.vendor }} - {{ device.device_type.model }})
                </span>
              </template>
              <template v-else>
                {{ device.device_type_id }}
              </template>
            </span>
          </div>
          <div class="detail-item">
            <span class="detail-label">地址</span>
            <span class="detail-value mono">{{ device.address || "-" }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">创建时间</span>
            <span class="detail-value">{{ formatTime(device.created_at) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">更新时间</span>
            <span class="detail-value">{{ formatTime(device.updated_at) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">最后上线</span>
            <span class="detail-value">
              {{ device.last_online_at ? formatTime(device.last_online_at) : "-" }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="device.device_type" class="capabilities-card">
        <h4>设备能力</h4>
        <div class="capabilities-list">
          <CapabilityTag
            v-for="cap in device.device_type.capabilities"
            :key="cap"
            :capability="cap"
          />
          <span v-if="!device.device_type.capabilities?.length" class="text-secondary">
            无
          </span>
        </div>
      </div>

      <!-- Operation Panel -->
      <div class="operation-panel">
        <h4>设备操作</h4>
        <div class="action-buttons">
          <button
            class="btn btn-primary"
            :disabled="!canQueryStatus || operationLoading"
            @click="executeQueryStatus"
          >
            {{ operationLoading && pendingOpType === 'QUERY_STATUS' ? '执行中...' : '查询状态' }}
          </button>
          <button
            class="btn btn-primary"
            :disabled="!canSetVolume || operationLoading"
            @click="openSetVolumeDialog"
          >
            {{ operationLoading && pendingOpType === 'SET_VOLUME' ? '执行中...' : '设置音量' }}
          </button>
          <button
            class="btn btn-danger"
            :disabled="!canRestart || operationLoading"
            @click="openRestartDialog"
          >
            {{ operationLoading && pendingOpType === 'RESTART' ? '执行中...' : '重启设备' }}
          </button>
        </div>
        <div v-if="operationDisabledHint" class="operation-hint">
          {{ operationDisabledHint }}
        </div>
      </div>

      <!-- Last Operation Result -->
      <div v-if="lastOperation && lastExecution" class="result-section">
        <h4>最近操作结果</h4>
        <OperationResultCard
          :operation="lastOperation"
          :execution="lastExecution"
        />
      </div>

      <!-- Status Actions (enable/disable) -->
      <div class="actions-card">
        <h4>状态操作</h4>
        <div class="action-buttons">
          <button
            v-if="canDisable"
            class="btn btn-warning"
            :disabled="statusActionLoading"
            @click="confirmAction('disable')"
          >
            {{ statusActionLoading && pendingAction === 'disable' ? '处理中...' : '禁用设备' }}
          </button>
          <button
            v-if="canEnable"
            class="btn btn-success"
            :disabled="statusActionLoading"
            @click="confirmAction('enable')"
          >
            {{ statusActionLoading && pendingAction === 'enable' ? '处理中...' : '启用设备' }}
          </button>
          <p v-if="!canDisable && !canEnable" class="action-hint">
            当前状态不支持手动操作（{{ statusHint }}）
          </p>
        </div>
      </div>

      <!-- Operation History -->
      <OperationHistory ref="historyRef" :device-id="deviceId" />
    </template>

    <!-- Enable/Disable Confirm Dialog -->
    <ConfirmDialog
      :visible="confirmDialog.visible"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :confirm-text="confirmDialog.confirmText"
      :confirm-class="confirmDialog.confirmClass"
      :loading="statusActionLoading"
      @confirm="handleAction"
      @cancel="cancelAction"
    />

    <!-- SetVolume Dialog -->
    <ConfirmDialog
      :visible="volumeDialog.visible"
      title="设置音量"
      confirm-text="确认设置"
      confirm-class="btn-primary"
      :loading="operationLoading"
      @confirm="executeSetVolume"
      @cancel="closeVolumeDialog"
    >
      <VolumeSlider v-model="volumeDialog.volume" />
    </ConfirmDialog>

    <!-- Restart Confirm Dialog -->
    <ConfirmDialog
      :visible="restartDialog.visible"
      title="重启设备"
      message="确定要重启该设备吗？重启期间设备将暂时离线。"
      confirm-text="确认重启"
      confirm-class="btn-danger"
      :loading="operationLoading"
      @confirm="executeRestart"
      @cancel="closeRestartDialog"
    />

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
import { useRoute } from "vue-router";
import { getDevice, updateDeviceStatus } from "@/api/device";
import { createOperation } from "@/api/operation";
import { extractErrorMessage } from "@/api/client";
import type { Device } from "@/types/device";
import type { Operation, Execution } from "@/types/operation";
import StatusBadge from "@/components/StatusBadge.vue";
import CapabilityTag from "@/components/CapabilityTag.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ToastNotification from "@/components/ToastNotification.vue";
import PageLoading from "@/components/PageLoading.vue";
import OperationResultCard from "@/components/OperationResultCard.vue";
import VolumeSlider from "@/components/VolumeSlider.vue";
import OperationHistory from "./OperationHistory.vue";

const route = useRoute();
const deviceId = computed(() => route.params.id as string);

const loading = ref(false);
const error = ref<string | null>(null);
const device = ref<Device | null>(null);
const statusActionLoading = ref(false);
const pendingAction = ref<"enable" | "disable" | null>(null);

// Operation state
const operationLoading = ref(false);
const pendingOpType = ref<string | null>(null);
const lastOperation = ref<Operation | null>(null);
const lastExecution = ref<Execution | null>(null);
const historyRef = ref<InstanceType<typeof OperationHistory> | null>(null);

// SetVolume dialog
const volumeDialog = reactive({
  visible: false,
  volume: 50,
});

// Restart dialog
const restartDialog = reactive({
  visible: false,
});

// Enable/Disable computed
const canDisable = computed(() => {
  if (!device.value) return false;
  return ["REGISTERED", "ACTIVE", "OFFLINE"].includes(device.value.status);
});

const canEnable = computed(() => {
  if (!device.value) return false;
  return device.value.status === "DISABLED";
});

const statusHint = computed(() => {
  if (!device.value) return "";
  switch (device.value.status) {
    case "ACTIVE":
    case "OFFLINE":
      return "在线/离线状态由系统自动管理";
    default:
      return "";
  }
});

// Capability helpers
const deviceCapabilities = computed(() => {
  return device.value?.device_type?.capabilities || [];
});

const hasCapability = (cap: string) => deviceCapabilities.value.includes(cap as never);

const canQueryStatus = computed(() => {
  if (!device.value) return false;
  if (device.value.status === "DISABLED") return false;
  return hasCapability("QUERY_STATUS");
});

const canSetVolume = computed(() => {
  if (!device.value) return false;
  if (!["REGISTERED", "ACTIVE"].includes(device.value.status)) return false;
  return hasCapability("SET_VOLUME");
});

const canRestart = computed(() => {
  if (!device.value) return false;
  if (!["REGISTERED", "ACTIVE"].includes(device.value.status)) return false;
  return hasCapability("RESTART");
});

const operationDisabledHint = computed(() => {
  if (!device.value) return "";
  if (device.value.status === "DISABLED") return "设备已禁用，无法执行操作";
  if (device.value.status === "OFFLINE") return "设备离线，仅支持状态查询";
  return "";
});

// Enable/Disable dialog
const confirmDialog = reactive({
  visible: false,
  title: "",
  message: "",
  confirmText: "",
  confirmClass: "",
});

const toast = reactive({
  visible: false,
  message: "",
  type: "success" as "success" | "error",
});

function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

async function fetchData() {
  loading.value = true;
  error.value = null;

  const resp = await getDevice(deviceId.value);

  if (resp.code === 0 && resp.data) {
    device.value = resp.data;
  } else {
    error.value = extractErrorMessage(resp);
  }

  loading.value = false;
}

// --- Enable/Disable actions ---
function confirmAction(action: "enable" | "disable") {
  pendingAction.value = action;

  if (action === "disable") {
    confirmDialog.title = "禁用设备";
    confirmDialog.message = `确定要禁用设备「${device.value?.name}」吗？禁用后设备将无法执行操作。`;
    confirmDialog.confirmText = "禁用";
    confirmDialog.confirmClass = "btn-warning";
  } else {
    confirmDialog.title = "启用设备";
    confirmDialog.message = `确定要启用设备「${device.value?.name}」吗？启用后设备将恢复为已注册状态。`;
    confirmDialog.confirmText = "启用";
    confirmDialog.confirmClass = "btn-success";
  }

  confirmDialog.visible = true;
}

function cancelAction() {
  confirmDialog.visible = false;
  pendingAction.value = null;
}

async function handleAction() {
  if (!pendingAction.value || !device.value) return;

  statusActionLoading.value = true;

  const resp = await updateDeviceStatus(device.value.id, {
    action: pendingAction.value,
  });

  if (resp.code === 0 && resp.data) {
    device.value = resp.data;
    toast.message = pendingAction.value === "disable" ? "设备已禁用" : "设备已启用";
    toast.type = "success";
    toast.visible = true;
    confirmDialog.visible = false;
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
  }

  statusActionLoading.value = false;
  pendingAction.value = null;
}

// --- Operation actions ---
async function executeQueryStatus() {
  if (!device.value) return;
  await runOperation("QUERY_STATUS");
}

function openSetVolumeDialog() {
  volumeDialog.volume = 50;
  volumeDialog.visible = true;
}

function closeVolumeDialog() {
  volumeDialog.visible = false;
}

async function executeSetVolume() {
  if (!device.value) return;
  await runOperation("SET_VOLUME", { volume: volumeDialog.volume });
  volumeDialog.visible = false;
}

function openRestartDialog() {
  restartDialog.visible = true;
}

function closeRestartDialog() {
  restartDialog.visible = false;
}

async function executeRestart() {
  if (!device.value) return;
  await runOperation("RESTART");
  restartDialog.visible = false;
}

async function runOperation(
  type: "QUERY_STATUS" | "SET_VOLUME" | "RESTART",
  parameters?: Record<string, unknown>
) {
  operationLoading.value = true;
  pendingOpType.value = type;

  const resp = await createOperation({
    device_id: device.value!.id,
    type,
    parameters,
  });

  if (resp.code === 0 && resp.data) {
    lastOperation.value = resp.data.operation;
    lastExecution.value = resp.data.execution;

    // Refresh device data (status may have changed after QUERY_STATUS)
    await fetchData();

    // Refresh operation history
    historyRef.value?.refresh();

    // Show toast based on result
    if (resp.data.execution.status === "SUCCESS") {
      toast.message = getSuccessMessage(type, resp.data.execution);
      toast.type = "success";
    } else {
      toast.message = resp.data.execution.error || "操作执行失败";
      toast.type = "error";
    }
    toast.visible = true;
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
  }

  operationLoading.value = false;
  pendingOpType.value = null;
}

function getSuccessMessage(type: string, execution: Execution): string {
  switch (type) {
    case "QUERY_STATUS":
      return "状态查询成功";
    case "SET_VOLUME": {
      const vol = execution.result?.volume;
      return vol !== undefined ? `音量已设置为 ${vol}` : "音量设置成功";
    }
    case "RESTART":
      return "设备重启成功";
    default:
      return "操作执行成功";
  }
}

onMounted(fetchData);
</script>

<style scoped>
.device-detail-page {
  max-width: 800px;
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

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.detail-card,
.capabilities-card,
.operation-panel,
.result-section,
.actions-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 24px;
  margin-bottom: 16px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--color-border);
}

.device-name {
  font-size: 20px;
  font-weight: 600;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 12px;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-value {
  font-size: 14px;
  font-weight: 500;
}

.detail-value.mono {
  font-family: monospace;
  font-size: 13px;
  word-break: break-all;
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: 13px;
}

.capabilities-card h4,
.operation-panel h4,
.result-section h4,
.actions-card h4 {
  font-size: 16px;
  margin-bottom: 12px;
}

.capabilities-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.action-hint {
  font-size: 13px;
  color: var(--color-text-secondary);
  font-style: italic;
}

.operation-hint {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
  font-style: italic;
}

.error-state {
  background: #fee2e2;
  color: #991b1b;
  padding: 20px;
  border-radius: var(--radius);
  text-align: center;
}

@media (max-width: 640px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>