<template>
  <div class="batch-operation-page">
    <div class="page-header">
      <h2>批量操作</h2>
      <div class="header-actions">
        <router-link to="/devices" class="btn-link">返回设备列表</router-link>
      </div>
    </div>

    <PageLoading v-if="loading" text="加载操作信息..." />

    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="fetchDetail">重试</button>
    </div>

    <template v-else-if="operation">
      <!-- Summary Card -->
      <div class="summary-card">
        <div class="summary-header">
          <h3>{{ operationTypeLabel }}</h3>
          <span class="op-id mono">{{ operation.id }}</span>
        </div>

        <div class="summary-status">
          <span class="status-badge" :class="statusClass">{{ statusLabel }}</span>
          <span class="op-time">{{ formatTime(operation.created_at) }}</span>
        </div>

        <ProgressBar
          :total="progress.total_count"
          :success="progress.success_count"
          :failed="progress.failed_count"
          :skipped="progress.skipped_count"
        />

        <div class="summary-counts">
          <div class="count-item">
            <span class="count-value">{{ progress.total_count }}</span>
            <span class="count-label">总数</span>
          </div>
          <div class="count-item success">
            <span class="count-value">{{ progress.success_count }}</span>
            <span class="count-label">成功</span>
          </div>
          <div class="count-item failed">
            <span class="count-value">{{ progress.failed_count }}</span>
            <span class="count-label">失败</span>
          </div>
          <div class="count-item skipped">
            <span class="count-value">{{ progress.skipped_count }}</span>
            <span class="count-label">跳过</span>
          </div>
          <div class="count-item running">
            <span class="count-value">{{ progress.running_count }}</span>
            <span class="count-label">执行中</span>
          </div>
        </div>
      </div>

      <!-- Execution Details -->
      <div class="executions-card">
        <h4>设备执行明细</h4>

        <div v-if="executions.length === 0" class="empty">
          暂无执行记录
        </div>

        <div v-else class="execution-list">
          <div
            v-for="exec in executions"
            :key="exec.id"
            class="execution-item"
            :class="exec.status.toLowerCase()"
          >
            <div class="exec-info">
              <span class="exec-device-id mono">{{ exec.device_id }}</span>
              <span class="exec-status" :class="exec.status.toLowerCase()">
                {{ executionStatusLabel(exec.status) }}
              </span>
            </div>

            <div v-if="exec.result" class="exec-result">
              <pre>{{ formatResult(exec.result) }}</pre>
            </div>

            <div v-if="exec.error" class="exec-error">
              {{ exec.error }}
            </div>

            <div v-if="exec.retry_count > 0" class="exec-retry">
              重试: {{ exec.retry_count }}/{{ exec.max_retries }}
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import { getBatchOperation, getBatchProgress } from "@/api/batch-operation";
import { extractErrorMessage } from "@/api/client";
import type { Operation, Execution } from "@/types/operation";
import type { BatchProgress } from "@/types/batch-operation";
import { OPERATION_TYPE_LABELS } from "@/types/operation";
import ProgressBar from "@/components/ProgressBar.vue";
import PageLoading from "@/components/PageLoading.vue";

const route = useRoute();
const operationId = computed(() => route.params.id as string);

const loading = ref(false);
const error = ref<string | null>(null);
const operation = ref<Operation | null>(null);
const executions = ref<Execution[]>([]);
const progress = ref<BatchProgress>({
  operation_id: "",
  status: "PENDING",
  total_count: 0,
  success_count: 0,
  failed_count: 0,
  skipped_count: 0,
  running_count: 0,
  is_complete: false,
});

let pollTimer: ReturnType<typeof setInterval> | null = null;

const operationTypeLabel = computed(() => {
  if (!operation.value) return "";
  return OPERATION_TYPE_LABELS[operation.value.type as keyof typeof OPERATION_TYPE_LABELS] || operation.value.type;
});

const statusClass = computed(() => {
  return progress.value.status.toLowerCase().replace("_", "-");
});

const statusLabel = computed(() => {
  switch (progress.value.status) {
    case "PENDING":
      return "待执行";
    case "IN_PROGRESS":
      return "执行中";
    case "COMPLETED":
      return "全部成功";
    case "FAILED":
      return "全部失败";
    case "PARTIAL_SUCCESS":
      return "部分成功";
    default:
      return progress.value.status;
  }
});

function executionStatusLabel(status: string): string {
  switch (status) {
    case "PENDING":
      return "⏳ 待执行";
    case "RUNNING":
      return "⚙️ 执行中";
    case "RETRYING":
      return "🔄 重试中";
    case "SUCCESS":
      return "✅ 成功";
    case "FAILED":
      return "❌ 失败";
    case "SKIPPED":
      return "⏭️ 跳过";
    default:
      return status;
  }
}

function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleString("zh-CN");
  } catch {
    return iso;
  }
}

function formatResult(result: Record<string, unknown>): string {
  return JSON.stringify(result, null, 2);
}

async function fetchDetail() {
  loading.value = true;
  error.value = null;

  const resp = await getBatchOperation(operationId.value);

  if (resp.code === 0 && resp.data) {
    operation.value = resp.data.operation;
    executions.value = resp.data.executions || [];

    // Update progress from operation data
    progress.value = {
      operation_id: resp.data.operation.id,
      status: resp.data.operation.status,
      total_count: resp.data.operation.total_count,
      success_count: resp.data.operation.success_count,
      failed_count: resp.data.operation.failed_count,
      skipped_count: resp.data.operation.skipped_count,
      running_count: resp.data.operation.total_count -
        resp.data.operation.success_count -
        resp.data.operation.failed_count -
        resp.data.operation.skipped_count,
      is_complete: ["COMPLETED", "FAILED", "PARTIAL_SUCCESS"].includes(resp.data.operation.status),
    };

    if (!progress.value.is_complete) {
      startPolling();
    }
  } else {
    error.value = extractErrorMessage(resp);
  }

  loading.value = false;
}

async function pollProgress() {
  const resp = await getBatchProgress(operationId.value);

  if (resp.code === 0 && resp.data) {
    progress.value = resp.data;

    if (resp.data.is_complete) {
      stopPolling();
      // Refresh full detail
      await fetchDetail();
    }
  }
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(pollProgress, 2000);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

onMounted(fetchDetail);
onUnmounted(stopPolling);
</script>

<style scoped>
.batch-operation-page {
  max-width: 900px;
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

.summary-card,
.executions-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 24px;
  margin-bottom: 16px;
}

.summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.summary-header h3 {
  font-size: 18px;
}

.op-id {
  font-size: 12px;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.summary-status {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
}

.status-badge.in-progress {
  background: #dbeafe;
  color: #1e40af;
}

.status-badge.completed {
  background: #dcfce7;
  color: #166534;
}

.status-badge.failed {
  background: #fee2e2;
  color: #991b1b;
}

.status-badge.partial-success {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.pending {
  background: #f3f4f6;
  color: #374151;
}

.op-time {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.summary-counts {
  display: flex;
  justify-content: space-around;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.count-item {
  text-align: center;
}

.count-value {
  display: block;
  font-size: 24px;
  font-weight: 700;
}

.count-label {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.count-item.success .count-value {
  color: #16a34a;
}

.count-item.failed .count-value {
  color: #dc2626;
}

.count-item.skipped .count-value {
  color: #d97706;
}

.count-item.running .count-value {
  color: #2563eb;
}

.executions-card h4 {
  margin-bottom: 16px;
}

.execution-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.execution-item {
  padding: 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  border-left: 3px solid var(--color-border);
}

.execution-item.success {
  border-left-color: #22c55e;
}

.execution-item.failed {
  border-left-color: #ef4444;
}

.execution-item.skipped {
  border-left-color: #f59e0b;
  opacity: 0.8;
}

.execution-item.running,
.execution-item.retrying {
  border-left-color: #3b82f6;
}

.exec-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.exec-device-id {
  font-size: 13px;
  font-family: monospace;
  word-break: break-all;
}

.exec-status {
  font-size: 13px;
  white-space: nowrap;
}

.exec-result {
  margin-top: 8px;
}

.exec-result pre {
  background: #f8fafc;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  overflow-x: auto;
  margin: 0;
}

.exec-error {
  margin-top: 8px;
  padding: 8px;
  background: #fee2e2;
  border-radius: 4px;
  font-size: 12px;
  color: #991b1b;
}

.exec-retry {
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.empty {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}

.error-state {
  background: #fee2e2;
  color: #991b1b;
  padding: 20px;
  border-radius: var(--radius);
  text-align: center;
}

.mono {
  font-family: monospace;
}
</style>