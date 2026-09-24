<template>
  <div class="operation-result-card" :class="statusClass">
    <div class="result-header">
      <span class="status-icon">{{ statusIcon }}</span>
      <span class="status-text">{{ statusText }}</span>
      <span class="operation-type">{{ operationTypeLabel }}</span>
    </div>

    <div v-if="execution?.result" class="result-data">
      <pre>{{ formatResult(execution.result) }}</pre>
    </div>

    <div v-if="execution?.error" class="result-error">
      <p>{{ execution.error }}</p>
    </div>

    <div v-if="execution?.retry_count" class="result-retry">
      <p>重试次数: {{ execution.retry_count }} / {{ execution.max_retries }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Operation, Execution } from "@/types/operation";
import { OPERATION_TYPE_LABELS } from "@/types/operation";

const props = defineProps<{
  operation: Operation;
  execution?: Execution;
}>();

const statusClass = computed(() => {
  if (!props.execution) return "pending";
  switch (props.execution.status) {
    case "SUCCESS":
      return "success";
    case "FAILED":
      return "failed";
    case "RUNNING":
    case "RETRYING":
      return "running";
    default:
      return "pending";
  }
});

const statusIcon = computed(() => {
  if (!props.execution) return "⏳";
  switch (props.execution.status) {
    case "SUCCESS":
      return "✅";
    case "FAILED":
      return "❌";
    case "RUNNING":
    case "RETRYING":
      return "⚙️";
    default:
      return "⏳";
  }
});

const statusText = computed(() => {
  if (!props.execution) return "待执行";
  switch (props.execution.status) {
    case "SUCCESS":
      return "执行成功";
    case "FAILED":
      return "执行失败";
    case "RUNNING":
      return "执行中";
    case "RETRYING":
      return "重试中";
    default:
      return "待执行";
  }
});

const operationTypeLabel = computed(() => {
  return OPERATION_TYPE_LABELS[props.operation.type] || props.operation.type;
});

function formatResult(result: Record<string, unknown>): string {
  return JSON.stringify(result, null, 2);
}
</script>

<style scoped>
.operation-result-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  padding: 16px;
  margin-bottom: 12px;
  border-left: 4px solid var(--color-border);
}

.operation-result-card.success {
  border-left-color: var(--color-success);
  background: #f0fdf4;
}

.operation-result-card.failed {
  border-left-color: var(--color-danger);
  background: #fef2f2;
}

.operation-result-card.running {
  border-left-color: var(--color-primary);
  background: #eff6ff;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-icon {
  font-size: 20px;
}

.status-text {
  font-weight: 600;
  font-size: 14px;
}

.operation-type {
  margin-left: auto;
  font-size: 12px;
  color: var(--color-text-secondary);
  background: var(--color-border);
  padding: 2px 8px;
  border-radius: 4px;
}

.result-data {
  margin-top: 12px;
}

.result-data pre {
  background: #f8fafc;
  padding: 12px;
  border-radius: 4px;
  font-size: 13px;
  overflow-x: auto;
}

.result-error {
  margin-top: 12px;
  padding: 12px;
  background: #fee2e2;
  border-radius: 4px;
  color: var(--color-danger);
  font-size: 13px;
}

.result-retry {
  margin-top: 8px;
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>