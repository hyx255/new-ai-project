<template>
  <span class="status-badge" :class="statusClass">{{ label }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { DeviceStatus } from "@/types/device";
import { STATUS_LABELS } from "@/types/device";

const props = defineProps<{
  status: DeviceStatus;
}>();

const label = computed(() => STATUS_LABELS[props.status] || props.status);

const statusClass = computed(() => {
  switch (props.status) {
    case "ACTIVE":
      return "status-active";
    case "REGISTERED":
      return "status-registered";
    case "OFFLINE":
      return "status-offline";
    case "DISABLED":
      return "status-disabled";
    default:
      return "";
  }
});
</script>

<style scoped>
.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.5;
}

.status-active {
  background: #dcfce7;
  color: #166534;
}

.status-registered {
  background: #dbeafe;
  color: #1e40af;
}

.status-offline {
  background: #f3f4f6;
  color: #6b7280;
}

.status-disabled {
  background: #fee2e2;
  color: #991b1b;
}
</style>
