<template>
  <Teleport to="body">
    <div v-if="visible" class="dialog-overlay" @click.self="$emit('cancel')">
      <div class="dialog-box batch-dialog">
        <h3 class="dialog-title">批量操作</h3>

        <div class="device-count">
          已选择 <strong>{{ deviceCount }}</strong> 台设备
        </div>

        <div class="op-type-group">
          <label class="op-type-label">操作类型:</label>
          <div class="op-type-options">
            <label
              v-for="opt in operationOptions"
              :key="opt.value"
              class="op-type-radio"
              :class="{ selected: selectedType === opt.value }"
            >
              <input
                type="radio"
                :value="opt.value"
                v-model="selectedType"
                name="opType"
              />
              <span>{{ opt.label }}</span>
            </label>
          </div>
        </div>

        <div v-if="selectedType === 'SET_VOLUME'" class="volume-section">
          <VolumeSlider v-model="volume" />
        </div>

        <div v-if="selectedType === 'RESTART'" class="restart-warning">
          <p>⚠️ 重启期间设备将暂时离线</p>
        </div>

        <div class="dialog-actions">
          <button class="btn btn-secondary" @click="$emit('cancel')">取消</button>
          <button
            class="btn btn-primary"
            :disabled="submitting"
            @click="handleSubmit"
          >
            {{ submitting ? "提交中..." : "确认执行" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import VolumeSlider from "./VolumeSlider.vue";

const props = defineProps<{
  visible: boolean;
  deviceCount: number;
}>();

const emit = defineEmits<{
  cancel: [];
  submit: [payload: { type: string; parameters?: Record<string, unknown> }];
}>();

const operationOptions = [
  { value: "QUERY_STATUS", label: "查询状态" },
  { value: "SET_VOLUME", label: "设置音量" },
  { value: "RESTART", label: "重启设备" },
];

const selectedType = ref("QUERY_STATUS");
const volume = ref(50);
const submitting = ref(false);

// Reset when dialog opens
watch(
  () => props.visible,
  (v) => {
    if (v) {
      selectedType.value = "QUERY_STATUS";
      volume.value = 50;
      submitting.value = false;
    }
  }
);

function handleSubmit() {
  submitting.value = true;

  const payload: { type: string; parameters?: Record<string, unknown> } = {
    type: selectedType.value,
  };

  if (selectedType.value === "SET_VOLUME") {
    payload.parameters = { volume: volume.value };
  }

  emit("submit", payload);
}

defineExpose({
  resetSubmitting: () => {
    submitting.value = false;
  },
});
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.batch-dialog {
  background: white;
  border-radius: 8px;
  padding: 24px;
  min-width: 420px;
  max-width: 520px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.dialog-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
}

.device-count {
  margin-bottom: 20px;
  font-size: 14px;
  color: var(--color-text-secondary, #6b7280);
}

.op-type-group {
  margin-bottom: 20px;
}

.op-type-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
}

.op-type-options {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.op-type-radio {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 2px solid var(--color-border, #e5e7eb);
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.15s;
}

.op-type-radio:hover {
  border-color: var(--color-primary, #3b82f6);
}

.op-type-radio.selected {
  border-color: var(--color-primary, #3b82f6);
  background: #eff6ff;
  color: var(--color-primary, #3b82f6);
  font-weight: 500;
}

.op-type-radio input {
  display: none;
}

.volume-section {
  margin-bottom: 20px;
}

.restart-warning {
  margin-bottom: 20px;
  padding: 12px;
  background: #fef3c7;
  border-radius: 6px;
  font-size: 13px;
  color: #92400e;
}

.dialog-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>