<template>
  <header class="app-header">
    <div class="header-brand">
      <span class="header-icon">&#x1F4E1;</span>
      <h1 class="header-title">Smart Broadcast Platform</h1>
    </div>
    <div class="header-meta">
      <span class="header-version">v0.1.0-skeleton</span>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  background: var(--color-header-bg);
  color: white;
  padding: 0 24px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: var(--shadow);
  z-index: 10;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-icon {
  font-size: 24px;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -0.02em;
}

.header-meta {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-version {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
}
</style>

