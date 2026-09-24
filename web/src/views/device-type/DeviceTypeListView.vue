<template>
  <div class="device-type-list-page">
    <div class="page-header">
      <h2>设备类型管理</h2>
      <router-link to="/device-types/create" class="btn btn-primary">
        + 新建类型
      </router-link>
    </div>

    <PageLoading v-if="loading" text="加载设备类型..." />

    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="fetchData">重试</button>
    </div>

    <template v-else>
      <div v-if="deviceTypes.length === 0" class="empty-state">
        <p>暂无设备类型</p>
        <router-link to="/device-types/create" class="btn btn-primary">
          创建第一个设备类型
        </router-link>
      </div>

      <div v-else class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>厂商</th>
              <th>型号</th>
              <th>描述</th>
              <th>能力</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="dt in deviceTypes" :key="dt.id">
              <td class="cell-name">{{ dt.name }}</td>
              <td>{{ dt.vendor }}</td>
              <td>{{ dt.model }}</td>
              <td class="cell-desc">{{ dt.description || "-" }}</td>
              <td class="cell-caps">
                <CapabilityTag
                  v-for="cap in dt.capabilities"
                  :key="cap"
                  :capability="cap"
                />
                <span v-if="!dt.capabilities?.length" class="text-muted">-</span>
              </td>
              <td class="cell-time">{{ formatTime(dt.created_at) }}</td>
              <td class="cell-actions">
                <router-link
                  :to="`/device-types/${dt.id}/edit`"
                  class="btn-link"
                >
                  编辑
                </router-link>
                <button
                  class="btn-link btn-danger"
                  @click="confirmDelete(dt)"
                >
                  删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <ConfirmDialog
      :visible="deleteDialog.visible"
      title="删除设备类型"
      :message="`确定要删除设备类型「${deleteDialog.target?.name}」吗？此操作不可撤销。`"
      confirm-text="删除"
      confirm-class="btn-danger"
      :loading="deleteDialog.loading"
      @confirm="handleDelete"
      @cancel="cancelDelete"
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
import { ref, reactive, onMounted } from "vue";
import { listDeviceTypes, deleteDeviceType } from "@/api/device-type";
import { extractErrorMessage } from "@/api/client";
import type { DeviceType } from "@/types/device";
import CapabilityTag from "@/components/CapabilityTag.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ToastNotification from "@/components/ToastNotification.vue";
import PageLoading from "@/components/PageLoading.vue";

const loading = ref(true);
const error = ref<string | null>(null);
const deviceTypes = ref<DeviceType[]>([]);

const deleteDialog = reactive({
  visible: false,
  target: null as DeviceType | null,
  loading: false,
});

const toast = reactive({
  visible: false,
  message: "",
  type: "success" as "success" | "error" | "warning" | "info",
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

  const resp = await listDeviceTypes();

  if (resp.code === 0 && resp.data) {
    deviceTypes.value = resp.data.items || [];
  } else {
    error.value = extractErrorMessage(resp);
  }

  loading.value = false;
}

function confirmDelete(dt: DeviceType) {
  deleteDialog.target = dt;
  deleteDialog.visible = true;
}

function cancelDelete() {
  deleteDialog.visible = false;
  deleteDialog.target = null;
}

async function handleDelete() {
  if (!deleteDialog.target) return;

  deleteDialog.loading = true;
  const resp = await deleteDeviceType(deleteDialog.target.id);

  if (resp.code === 0) {
    toast.message = `设备类型「${deleteDialog.target.name}」已删除`;
    toast.type = "success";
    toast.visible = true;
    deleteDialog.visible = false;
    deleteDialog.target = null;
    await fetchData();
  } else {
    toast.message = extractErrorMessage(resp);
    toast.type = "error";
    toast.visible = true;
  }

  deleteDialog.loading = false;
}

onMounted(fetchData);
</script>

<style scoped>
.device-type-list-page {
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
  color: var(--color-text);
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

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.cell-name {
  font-weight: 600;
  color: var(--color-primary);
}

.cell-desc {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cell-caps {
  white-space: nowrap;
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

.text-muted {
  color: var(--color-text-secondary);
}
</style>
