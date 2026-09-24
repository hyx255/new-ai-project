<template>
  <span class="capability-tag">{{ label }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Capability } from "@/types/device";
import { CAPABILITY_LABELS } from "@/types/device";

const props = defineProps<{
  capability: Capability;
}>();

const label = computed(
  () => CAPABILITY_LABELS[props.capability] || props.capability
);
</script>

<style scoped>
.capability-tag {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  background: #ede9fe;
  color: #5b21b6;
  margin-right: 4px;
  margin-bottom: 2px;
}
</style>
