<template>
  <nav class="app-sidebar">
    <ul class="sidebar-menu">
      <li>
        <router-link to="/dashboard" class="menu-item" active-class="active" exact>
          <span class="menu-icon">&#x1F4CA;</span>
          <span class="menu-label">Dashboard</span>
        </router-link>
      </li>
      <li>
        <router-link
          to="/device-types"
          class="menu-item"
          active-class="active"
          :class="{ active: isDeviceTypeActive }"
        >
          <span class="menu-icon">&#x1F4F1;</span>
          <span class="menu-label">设备类型</span>
        </router-link>
      </li>
      <li>
        <router-link
          to="/devices"
          class="menu-item"
          active-class="active"
          :class="{ active: isDeviceActive }"
        >
          <span class="menu-icon">&#x1F4E1;</span>
          <span class="menu-label">设备管理</span>
        </router-link>
      </li>
      <li class="menu-disabled">
        <span class="menu-item disabled">
          <span class="menu-icon">&#x1F4CB;</span>
          <span class="menu-label">Operations</span>
          <span class="menu-badge">TBD</span>
        </span>
      </li>
      <li class="menu-disabled">
        <span class="menu-item disabled">
          <span class="menu-icon">&#x1F6A8;</span>
          <span class="menu-label">Alarms</span>
          <span class="menu-badge">TBD</span>
        </span>
      </li>
      <li class="menu-disabled">
        <span class="menu-item disabled">
          <span class="menu-icon">&#x1F916;</span>
          <span class="menu-label">Agent</span>
          <span class="menu-badge">TBD</span>
        </span>
      </li>
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const isDeviceTypeActive = computed(() =>
  route.path.startsWith("/device-types")
);

const isDeviceActive = computed(() =>
  route.path.startsWith("/devices")
);
</script>

<style scoped>
.app-sidebar {
  width: 220px;
  background: var(--color-sidebar-bg);
  border-right: 1px solid var(--color-border);
  padding: 16px 0;
  flex-shrink: 0;
}

.sidebar-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;
  color: var(--color-text);
  font-size: 14px;
  transition: background 0.15s, color 0.15s;
  cursor: pointer;
  text-decoration: none;
}

.menu-item:hover:not(.disabled) {
  background: rgba(37, 99, 235, 0.05);
  color: var(--color-primary);
}

.menu-item.active {
  background: rgba(37, 99, 235, 0.1);
  color: var(--color-primary);
  font-weight: 600;
  border-right: 3px solid var(--color-primary);
}

.menu-item.disabled {
  color: var(--color-text-secondary);
  opacity: 0.6;
  cursor: default;
}

.menu-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
}

.menu-label {
  flex: 1;
}

.menu-badge {
  font-size: 10px;
  background: var(--color-border);
  color: var(--color-text-secondary);
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 500;
}
</style>
