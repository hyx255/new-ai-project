<template>
  <Teleport to="body">
    <Transition name="toast">
      <div v-if="visible" class="toast" :class="toastClass">
        <span class="toast-icon">{{ icon }}</span>
        <span class="toast-message">{{ message }}</span>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch, ref } from "vue";

const props = defineProps<{
  visible: boolean;
  message: string;
  type?: "success" | "error" | "warning" | "info";
  duration?: number;
}>();

const emit = defineEmits<{
  close: [];
}>();

const autoCloseTimer = ref<number | null>(null);

const toastClass = computed(() => `toast-${props.type || "info"}`);

const icon = computed(() => {
  switch (props.type) {
    case "success":
      return "?";
    case "error":
      return "?";
    case "warning":
      return "?";
    default:
      return "?";
  }
});

watch(
  () => props.visible,
  (newVal) => {
    if (autoCloseTimer.value) {
      clearTimeout(autoCloseTimer.value);
      autoCloseTimer.value = null;
    }
    if (newVal && props.duration !== 0) {
      autoCloseTimer.value = window.setTimeout(() => {
        emit("close");
      }, props.duration || 3000);
    }
  }
);
</script>

<style scoped>
.toast {
  position: fixed;
  top: 20px;
  right: 20px;
  padding: 12px 20px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  z-index: 2000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  max-width: 400px;
}

.toast-success {
  background: #dcfce7;
  color: #166534;
  border: 1px solid #86efac;
}

.toast-error {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fca5a5;
}

.toast-warning {
  background: #fef3c7;
  color: #92400e;
  border: 1px solid #fcd34d;
}

.toast-info {
  background: #dbeafe;
  color: #1e40af;
  border: 1px solid #93c5fd;
}

.toast-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.toast-message {
  line-height: 1.4;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
