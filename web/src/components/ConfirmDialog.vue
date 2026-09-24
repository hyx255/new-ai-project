<template>
  <Teleport to="body">
    <div v-if="visible" class="dialog-overlay" @click.self="$emit('cancel')">
      <div class="dialog-box">
        <h3 class="dialog-title">{{ title }}</h3>
        <div v-if="$slots.default" class="dialog-content">
          <slot />
        </div>
        <p v-else-if="message" class="dialog-message">{{ message }}</p>
        <div class="dialog-actions">
          <button class="btn btn-secondary" @click="$emit('cancel')">
            {{ cancelText }}
          </button>
          <button
            class="btn"
            :class="confirmClass"
            :disabled="loading"
            @click="$emit('confirm')"
          >
            {{ loading ? '处理中...' : confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean;
  title: string;
  message?: string;
  confirmText?: string;
  cancelText?: string;
  confirmClass?: string;
  loading?: boolean;
}>();

defineEmits<{
  confirm: [];
  cancel: [];
}>();
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

.dialog-box {
  background: white;
  border-radius: 8px;
  padding: 24px;
  min-width: 360px;
  max-width: 480px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.dialog-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--color-text);
}

.dialog-message {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-bottom: 20px;
  line-height: 1.5;
}

.dialog-content {
  margin-bottom: 20px;
}

.dialog-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>