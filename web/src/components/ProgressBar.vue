<template>
  <div class="progress-bar-container">
    <div class="progress-bar" :style="{ width: percentage + '%' }" :class="barClass"></div>
    <span class="progress-text">{{ percentage }}%</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  total: number;
  success: number;
  failed: number;
  skipped: number;
}>();

const percentage = computed(() => {
  if (props.total === 0) return 0;
  const done = props.success + props.failed + props.skipped;
  return Math.round((done / props.total) * 100);
});

const barClass = computed(() => {
  if (percentage.value < 100) return "in-progress";
  if (props.failed > 0 && props.success > 0) return "partial";
  if (props.failed > 0 && props.success === 0) return "failed";
  return "completed";
});
</script>

<style scoped>
.progress-bar-container {
  position: relative;
  width: 100%;
  height: 24px;
  background: #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  border-radius: 12px;
  transition: width 0.5s ease;
}

.progress-bar.in-progress {
  background: var(--color-primary, #3b82f6);
}

.progress-bar.completed {
  background: var(--color-success, #22c55e);
}

.progress-bar.partial {
  background: var(--color-warning, #f59e0b);
}

.progress-bar.failed {
  background: var(--color-danger, #ef4444);
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: 600;
  color: #1f2937;
}
</style>