<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h2>Dashboard</h2>
      <p class="dashboard-subtitle">Smart Broadcast Platform Overview</p>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="status-card loading">
      <div class="spinner"></div>
      <span>Checking backend health...</span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="status-card error">
      <div class="status-icon">&#x274C;</div>
      <div class="status-content">
        <h3>Backend Unreachable</h3>
        <p>{{ error }}</p>
        <button class="retry-btn" @click="fetchHealth">Retry</button>
      </div>
    </div>

    <!-- Success State -->
    <template v-else-if="health">
      <div class="cards-grid">
        <!-- Backend Status -->
        <div class="status-card" :class="health.status === 'ok' ? 'success' : 'warning'">
          <div class="status-icon">
            {{ health.status === 'ok' ? '&#x2705;' : '&#x26A0;&#xFE0F;' }}
          </div>
          <div class="status-content">
            <h3>Backend Status</h3>
            <p class="status-value">{{ health.status.toUpperCase() }}</p>
          </div>
        </div>

        <!-- Database Status -->
        <div class="status-card" :class="dbStatusClass">
          <div class="status-icon">
            {{ dbStatusIcon }}
          </div>
          <div class="status-content">
            <h3>Database Status</h3>
            <p class="status-value">{{ health.database.toUpperCase() }}</p>
          </div>
        </div>

        <!-- Trace ID -->
        <div class="status-card info">
          <div class="status-icon">&#x1F517;</div>
          <div class="status-content">
            <h3>Trace ID</h3>
            <p class="status-value trace-id">{{ health.trace_id || "N/A" }}</p>
          </div>
        </div>

        <!-- API Version -->
        <div class="status-card info">
          <div class="status-icon">&#x1F4E6;</div>
          <div class="status-content">
            <h3>API Version</h3>
            <p class="status-value">v0.1.0-skeleton</p>
          </div>
        </div>
      </div>

      <!-- Architecture Info -->
      <div class="info-section">
        <h3>Architecture</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">Backend</span>
            <span class="info-value">Go + Gin (Modular Monolith)</span>
          </div>
          <div class="info-item">
            <span class="info-label">Frontend</span>
            <span class="info-value">Vue 3 + TypeScript + Vite</span>
          </div>
          <div class="info-item">
            <span class="info-label">API Style</span>
            <span class="info-value">REST (Unified Response)</span>
          </div>
          <div class="info-item">
            <span class="info-label">Response Format</span>
            <span class="info-value">{{ "{code, message, data}" }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { getHealth } from "@/api/health";
import type { HealthData } from "@/types/api";

const loading = ref(true);
const error = ref<string | null>(null);
const health = ref<HealthData | null>(null);

const dbStatusClass = computed(() => {
  if (!health.value) return "";
  switch (health.value.database) {
    case "healthy":
      return "success";
    case "unhealthy":
      return "error";
    default:
      return "warning";
  }
});

const dbStatusIcon = computed(() => {
  if (!health.value) return "";
  switch (health.value.database) {
    case "healthy":
      return "&#x2705;";
    case "unhealthy":
      return "&#x274C;";
    default:
      return "&#x26A0;&#xFE0F;";
  }
});

async function fetchHealth() {
  loading.value = true;
  error.value = null;

  const response = await getHealth();

  if (response.code === 0 && response.data) {
    health.value = response.data;
  } else {
    error.value = response.message || "Failed to fetch health status";
    health.value = null;
  }

  loading.value = false;
}

onMounted(fetchHealth);
</script>

<style scoped>
.dashboard {
  max-width: 900px;
}

.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h2 {
  font-size: 24px;
  color: var(--color-text);
}

.dashboard-subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin-top: 4px;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.status-card {
  background: var(--color-surface);
  border-radius: var(--radius);
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: var(--shadow);
  border-left: 4px solid transparent;
  transition: transform 0.15s;
}

.status-card:hover {
  transform: translateY(-1px);
}

.status-card.success {
  border-left-color: var(--color-success);
}

.status-card.warning {
  border-left-color: var(--color-warning);
}

.status-card.error {
  border-left-color: var(--color-danger);
  background: #fef2f2;
}

.status-card.info {
  border-left-color: var(--color-primary);
}

.status-card.loading {
  border-left-color: var(--color-primary);
  gap: 12px;
  color: var(--color-text-secondary);
}

.status-icon {
  font-size: 28px;
  flex-shrink: 0;
}

.status-content h3 {
  font-size: 13px;
  color: var(--color-text-secondary);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status-value {
  font-size: 18px;
  font-weight: 700;
  margin-top: 2px;
}

.trace-id {
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 13px;
  color: var(--color-primary);
  word-break: break-all;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.retry-btn {
  margin-top: 8px;
  padding: 6px 16px;
  background: var(--color-danger);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.retry-btn:hover {
  opacity: 0.9;
}

.info-section {
  background: var(--color-surface);
  border-radius: var(--radius);
  padding: 20px;
  box-shadow: var(--shadow);
}

.info-section h3 {
  font-size: 16px;
  margin-bottom: 16px;
  color: var(--color-text);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: 12px;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-value {
  font-size: 14px;
  font-weight: 500;
}

@media (max-width: 768px) {
  .cards-grid,
  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>

