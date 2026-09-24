<template>
  <div class="page-loading">
    <div class="spinner"></div>
    <span class="loading-text">{{ text || "加载中..." }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  text?: string;
}>();
</script>

<style scoped>
.page-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 16px;
  color: var(--color-text-secondary);
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-text {
  font-size: 14px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
