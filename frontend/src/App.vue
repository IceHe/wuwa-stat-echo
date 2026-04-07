<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { authState, clearStoredAuthToken, restoreAuthSession } from '@/auth'


const route = useRoute()
const router = useRouter()

const showShell = computed(() => route.name !== 'login')
const currentUserName = computed(() => authState.user?.name || '已登录')

const handleLogout = async () => {
  clearStoredAuthToken()
  await router.push('/login')
}

onMounted(async () => {
  if (authState.token && !authState.user) {
    await restoreAuthSession()
  }
})
</script>

<template>
  <template v-if="showShell">
    <div class="shell">
      <header class="topbar">
        <nav class="nav">
          <RouterLink to="/home">首页</RouterLink>
          <RouterLink to="/echo">声骸录入</RouterLink>
          <RouterLink :to="`/echo-viewer?operator_id=${authState.user?.id || ''}`" target="_blank">实时查看</RouterLink>
          <RouterLink to="/analysis">统计分析</RouterLink>
          <RouterLink to="/decision-lab">Decision Lab</RouterLink>
          <RouterLink to="/simulator">Simulator</RouterLink>
          <RouterLink to="/echo_dcrit_count">双暴统计</RouterLink>
        </nav>
        <div class="session">
          <span>{{ currentUserName }}</span>
          <button type="button" @click="handleLogout">退出</button>
        </div>
      </header>
      <main class="page">
        <RouterView />
      </main>
    </div>
  </template>
  <RouterView v-else />
</template>

<style scoped>
.shell {
  min-height: 100vh;
  background: linear-gradient(180deg, #f4f0e7 0%, #f8fbfd 100%);
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 20px;
  border-bottom: 1px solid rgba(17, 24, 39, 0.08);
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(12px);
}

.nav {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.nav a {
  padding: 8px 12px;
  border-radius: 999px;
  color: #23404b;
  text-decoration: none;
  background: rgba(35, 64, 75, 0.08);
}

.nav a.router-link-exact-active {
  color: #fff;
  background: #23404b;
}

.session {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #425466;
  font-size: 14px;
}

.session button {
  border: 0;
  border-radius: 999px;
  padding: 8px 14px;
  color: #fff;
  background: #23404b;
  cursor: pointer;
}

.page {
  padding: 20px;
}

@media (max-width: 720px) {
  .topbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .session {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
