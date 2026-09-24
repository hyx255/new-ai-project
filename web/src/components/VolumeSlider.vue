<template>
  <div class="volume-slider">
    <label class="slider-label">
      目标音量: <span class="volume-value">{{ modelValue }}</span>
    </label>
    <input
      type="range"
      min="0"
      max="100"
      :value="modelValue"
      @input="handleInput"
      class="slider-input"
    />
    <div class="slider-marks">
      <span>0</span>
      <span>25</span>
      <span>50</span>
      <span>75</span>
      <span>100</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: number;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: number];
}>();

function handleInput(event: Event) {
  const target = event.target as HTMLInputElement;
  emit("update:modelValue", parseInt(target.value, 10));
}
</script>

<style scoped>
.volume-slider {
  width: 100%;
}

.slider-label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--color-text);
}

.volume-value {
  font-weight: 600;
  color: var(--color-primary);
  font-size: 16px;
}

.slider-input {
  width: 100%;
  height: 8px;
  border-radius: 4px;
  background: var(--color-border);
  outline: none;
  -webkit-appearance: none;
  appearance: none;
}

.slider-input::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--color-primary);
  cursor: pointer;
}

.slider-input::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--color-primary);
  cursor: pointer;
  border: none;
}

.slider-marks {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 11px;
  color: var(--color-text-secondary);
}
</style>