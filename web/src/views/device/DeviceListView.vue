<template>
  <div class="device-list-page">
    <div class="page-header">
      <h2>设备管理</h2>
      <div class="header-actions">
        <button
          v-if="selectedIds.length > 0"
          class="btn btn-primary"
          @click="showBatchDialog = true"
        >
          批量操作 ({{ selectedIds.length }})
        </button>
        <router-link to="/devices/create" class="btn btn-primary">
          + 新建设备
        </router-link>
      </div>
    </div>

    <div class="search-bar">
      <input
        v-model="searchQuery"
        type="text"
        class="form-input search-input"
        placeholder="搜索设备名称..."
        @keyup.enter="handleSearch"
      />
      <button class="btn btn-secondary" @click="handleSearch">搜索</button>
      <button
        v-if="searchQuery"
        class="btn-link"
        @click="clearSearch"
      >
        清除
      </button>
    </div>

    <PageLoading v-if="loading" text="加载设备列表..." />

    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="fetchData">重试</button>
    </div>

    <template v-else>
      <div v-if="devices.length === 0" class="empty-state">
        <p>{{ searchQuery ? "未找到匹配的设备" : "暂无设备" }}</p>
        <router-link v-if="!searchQuery" to="/devices/create" class="btn btn-primary">
          创建第一个设备
        </router-link>
      </div>

      <div v-else class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-check">
                <input
                  type="checkbox"
                  :checked="allSelected"
                  :indeterminate="someSelected"
                  @change="toggleSelectAll"
                />
              </th>
              <th>设备名称</th>
              <th>设备类型</th>
              <th>地址</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="device in devices"
              :key="device.id"
              :class="{ selected: selectedIds.includes(device.id) }"
            >
              <td class="col-check">
                <input
                  type="checkbox"
                  :checked="selectedIds.includes(device.id)"
                  @change="toggleSelect(device.id)"
                />
              </td>
              <td class="cell-name">
                <router-link :to="`/devices/${device.id}`">
                  {{ device.name }}
                </router-link>
              </td>
              <td>{{ device.device_type?.name || device.device_type_id }}</td>
              <td class="cell-address">{{ device.address || "-" }}</td>
              <td>
                <StatusBadge :status="device.status" />
              </td>
              <td class="cell-time">{{ formatTime(device.created_at) }}</td>
              <td class="cell-actions">
                <router-link :to="`/devices/${device.id}`" class="btn-link">
                  查看
                </router-link>
                <router-link
                  :to="`/devices/${device.id}/edit`"
                  class="btn-link"
                >
                  编辑
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="pagination.total > pagination.pageSize" class="pagination">
          <button
            class="btn btn-secondary btn-sm"
            :disabled="pagination.page <= 1"
            @click="goToPage(pagination.page - 1)"
          >
            上一页
          </button>
          <span class="page-info">
            第 {{ pagination.page }} 页 / 共 {{ totalPages }} 页（{{ pagination.total }} 条）
          </span>
          <button
            class="btn btn-secondary btn-sm"
            :disabled="pagination.page >= totalPages"
            @click="goToPage(pagination.page + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </template>

    <BatchOperationDialog
      :visible="showBatchDialog"
      :device-count="selectedIds.length"
      ref="batchDialogRef"
      @cancel="showBatchDialog = false"
      @submit="handleBatchSubmit"
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
import { useRouter } from "vue-router";
import { listDevices } from "@/api/device";
import { createBatchOperation } from "@/api/batch-operation";
import { extractErrorMessage } from "@/api/client";
import type { Device } from "@/types/device";
import StatusBadge from "@/components/StatusBadge.vue";
import PageLoading from "@/components/PageLoading.vue";
import ToastNotification from "@/components/ToastNotification.vue";
import BatchOperationDialog from "@/components/BatchOperationDialog.vue";

const router = useRouter();

const loading = ref(true);
const error = ref<string | null>(null);
const devices = ref<Device[]>([]);
const searchQuery = ref("");
const selectedIds = ref<string[]>([]);
const showBatchDialog = ref(false);
const batchDialogRef = ref<InstanceType<typeof BatchOperationDialog> | null>(null);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
});

const totalPages = computed(() =>
  Math.ceil(pagination.total / pagination.pageSize)
);

const allSelected = computed(() =>
  devices.value.length > 0 && selectedIds.value.length === devices.value.length
);

const someSelected = computed(() =>
  selectedIds.value.length > 0 && selectedIds.value.length < devices.value.length
);

const toast = reactive({
  visible: false,
  message: "",
  type: "success" as "success" | "error",
});

function toggleSelect(id: string) {
  const idx = selectedIds.value.indexOf(id);
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1);
  } else {
    selectedIds.value.push(id);
  }
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedIds.value = [];
  } else {
    selectedIds.value = devices.value.map((d) => d.id);
  }
}

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

  const resp = await listDevices(
    pagination.page,
    pagination.pageSize,
    searchQuery.value || undefined
  );

  if (resp.code === 0 && resp.data) {
    devices.value = resp.data.items || [];
    pagination.total = resp.data.total;
    pagination.page = resp.data.page;
    pagination.pageSize = resp.data.page_size;
  } else {
    error.value = extractErrorMessage(resp);
  }

  loading.value = false;
}

function handleSearch() {
  pagination.page = 1;
  selectedIds.value = [];
  fetchData();
}

function clearSearch() {
  searchQuery.value = "";
  pagination.page = 1;
  selectedIds.value = [];
  fetchData();
}

function goToPage(page: number) {
  pagination.page = page;
  selectedIds.value = [];
  fetchData();
}

async function handleBatchSubmit(payload: { type: string; parameters?: Record<string, unknown> }) {
  const resp = await createBatchOperation({
    device_ids: selectedIds.value,
    type: payload.type as "QUERY_STATUS" | "SET_VOLUME" | "RESTART",
    parameters: payload.parameters,
  });

  if (resp.code === 0 && resp.data) {
    showBatchDialog.value = false;
    selectedIds.value = [];
    // Navigate to batch operation detail page
    router.push(`/batch-operations/${resp.data.operation.id}`);
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
    batchDialogRef.value?.resetSubmitting();
  }
}

onMounted(fetchData);
</script>

<style scoped>
.device-list-page {
  max-width: 1100px;
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
  gap: 8px;
  align-items: center;
}

.search-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 16px;
}

.search-input {
  max-width: 300px;
}

.table-container {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  overflow: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table th {
  text-align: left;
  padding: 12px 16px;
  background: #f8fafc;
  color: var(--color-text-secondary);
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--color-border);
}

.data-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
}

.data-table tbody tr:hover {
  background: #f8fafc;
}

.data-table tbody tr.selected {
  background: #eff6ff;
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.col-check {
  width: 40px;
  text-align: center;
}

.col-check input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.cell-name a {
  font-weight: 600;
  color: var(--color-primary);
}

.cell-address {
  font-family: monospace;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.cell-time {
  color: var(--color-text-secondary);
  font-size: 13px;
  white-space: nowrap;
}

.cell-actions {
  white-space: nowrap;
}

.cell-actions .btn-link {
  margin-right: 12px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 16px;
  border-top: 1px solid var(--color-border);
}

.page-info {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.error-state {
  background: #fee2e2;
  color: #991b1b;
  padding: 20px;
  border-radius: var(--radius);
  text-align: center;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--color-text-secondary);
}

.empty-state p {
  margin-bottom: 16px;
}
</style>