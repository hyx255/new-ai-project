<template>
  <div class="operation-history">
    <h4>操作历史</h4>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="fetchHistory">重试</button>
    </div>

    <div v-else-if="operations.length === 0" class="empty">暂无操作记录</div>

    <template v-else>
      <div class="operation-list">
        <div
          v-for="op in operations"
          :key="op.id"
          class="operation-item"
          :class="op.status.toLowerCase()"
          @click="selectOperation(op)"
        >
          <div class="op-header">
            <span class="op-type">{{ OPERATION_TYPE_LABELS[op.type] }}</span>
            <span class="op-status" :class="op.status.toLowerCase()">
              {{ OPERATION_STATUS_LABELS[op.status] }}
            </span>
          </div>
          <div class="op-time">{{ formatTime(op.created_at) }}</div>
        </div>
      </div>

      <div v-if="pagination.total > pagination.pageSize" class="pagination">
        <button
          class="btn btn-secondary btn-sm"
          :disabled="pagination.page <= 1"
          @click="goToPage(pagination.page - 1)"
        >
          上一页
        </button>
        <span class="page-info">
          第 {{ pagination.page }} 页 / 共 {{ totalPages }} 页
        </span>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="pagination.page >= totalPages"
          @click="goToPage(pagination.page + 1)"
        >
          下一页
        </button>
      </div>
    </template>

    <div v-if="selectedOperation" class="operation-detail">
      <h5>操作详情</h5>
      <OperationResultCard
        :operation="selectedOperation"
        :execution="selectedExecution"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { listDeviceOperations, getOperation } from "@/api/operation";
import { extractErrorMessage } from "@/api/client";
import type { Operation, Execution } from "@/types/operation";
import {
  OPERATION_TYPE_LABELS,
  OPERATION_STATUS_LABELS,
} from "@/types/operation";
import OperationResultCard from "@/components/OperationResultCard.vue";

const props = defineProps<{
  deviceId: string;
}>();

const loading = ref(false);
const error = ref<string | null>(null);
const operations = ref<Operation[]>([]);
const selectedOperation = ref<Operation | null>(null);
const selectedExecution = ref<Execution | undefined>(undefined);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
});

const totalPages = computed(() =>
  Math.ceil(pagination.total / pagination.pageSize)
);

function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleString("zh-CN", {
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

async function fetchHistory() {
  loading.value = true;
  error.value = null;

  const resp = await listDeviceOperations(
    props.deviceId,
    pagination.page,
    pagination.pageSize
  );

  if (resp.code === 0 && resp.data) {
    operations.value = resp.data.items || [];
    pagination.total = resp.data.total;
  } else {
    error.value = extractErrorMessage(resp);
  }

  loading.value = false;
}

async function selectOperation(op: Operation) {
  const resp = await getOperation(op.id);
  if (resp.code === 0 && resp.data) {
    selectedOperation.value = resp.data.operation;
    selectedExecution.value = resp.data.executions?.[0] ?? undefined;
  }
}

function goToPage(page: number) {
  pagination.page = page;
  fetchHistory();
}

onMounted(fetchHistory);

defineExpose({
  refresh: fetchHistory,
});
</script>

<style scoped>
.operation-history {
  background: var(--color-surface);
  border-radius: var(--radius);
  padding: 20px;
  box-shadow: var(--shadow);
}

.operation-history h4 {
  margin-bottom: 16px;
  font-size: 16px;
}

.loading,
.empty {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}

.error {
  text-align: center;
  padding: 20px;
  color: var(--color-danger);
}

.operation-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.operation-item {
  padding: 12px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  cursor: pointer;
  transition: all 0.15s;
}

.operation-item:hover {
  background: #f8fafc;
  border-color: var(--color-primary);
}

.operation-item.completed {
  border-left: 3px solid var(--color-success);
}

.operation-item.failed {
  border-left: 3px solid var(--color-danger);
}

.operation-item.in_progress,
.operation-item.pending {
  border-left: 3px solid var(--color-primary);
}

.op-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.op-type {
  font-weight: 600;
  font-size: 14px;
}

.op-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.op-status.completed {
  background: #dcfce7;
  color: #166534;
}

.op-status.failed {
  background: #fee2e2;
  color: #991b1b;
}

.op-status.in_progress,
.op-status.pending {
  background: #dbeafe;
  color: #1e40af;
}

.op-time {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
}

.page-info {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.operation-detail {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--color-border);
}

.operation-detail h5 {
  margin-bottom: 12px;
  font-size: 14px;
  color: var(--color-text-secondary);
}
</style>