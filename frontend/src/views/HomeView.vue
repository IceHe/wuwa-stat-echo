<template>
  <div class="home-page">
    <div class="home-card">
      <h1>鸣潮声骸工具</h1>
      <p>选择要进入的子页面。</p>
      <nav class="home-links">
        <RouterLink class="home-link" to="/echo">
          <span class="home-link-title">声骸录入</span>
          <span class="home-link-path">/echo</span>
        </RouterLink>
        <RouterLink class="home-link" to="/analysis">
          <span class="home-link-title">高级统计</span>
          <span class="home-link-path">/analysis</span>
        </RouterLink>
        <RouterLink class="home-link" to="/decision-lab">
          <span class="home-link-title">Decision Lab</span>
          <span class="home-link-path">/decision-lab</span>
        </RouterLink>
        <RouterLink class="home-link" to="/simulator">
          <span class="home-link-title">Simulator</span>
          <span class="home-link-path">/simulator</span>
        </RouterLink>
        <RouterLink class="home-link" to="/echo_dcrit_count">
          <span class="home-link-title">双暴统计</span>
          <span class="home-link-path">/echo_dcrit_count</span>
        </RouterLink>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<style scoped>
.home-page {
  display: flex;
  justify-content: center;
  padding: 48px 20px;
}

.home-card {
  width: min(720px, 100%);
  padding: 28px;
  border-radius: 18px;
  background: linear-gradient(135deg, #f5efe2 0%, #eef6f4 100%);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
}

h1 {
  margin: 0 0 8px;
  font-size: 30px;
}

p {
  margin: 0 0 20px;
  color: #555;
}

.home-links {
  display: grid;
  gap: 12px;
}

.home-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.9);
  color: #222;
  text-decoration: none;
  border: 1px solid rgba(0, 0, 0, 0.08);
  transition: transform 120ms ease, box-shadow 120ms ease;
}

.home-link:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.08);
}

.home-link-title {
  font-size: 18px;
  font-weight: 800;
}

.home-link-path {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  color: #666;
}
</style>
