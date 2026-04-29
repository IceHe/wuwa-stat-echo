<template>
  <div class="container">
    <div class="left-column">
      <Echo />
      <EchoLogs />
    </div>
    <div class="right-column">
      <div class="search-tabs">
        <div class="tab-header">
          <button
            class="tab-button"
            :class="{ active: activeTab === 'recent' }"
            @click="activeTab = 'recent'"
          >
            最近录入
          </button>
          <button
            class="tab-button"
            :class="{ active: activeTab === 'search' }"
            @click="activeTab = 'search'"
          >
            搜索声骸
          </button>
        </div>
        <div v-show="activeTab === 'search'">
          <FindEcho />
        </div>
        <div v-show="activeTab === 'recent'">
          <FindEcho
            title="最近录入的声骸"
            refresh-button-label="最近录入列表 - 刷新"
            :allow-empty-query="true"
            :auto-load="true"
            :sync-from-editor="false"
            refresh-event-name="refreshRecentEchoLogs"
          />
        </div>
      </div>
      <SubstatLogs :default-size="52" />
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import Echo from '@/components/Echo.vue'
import EchoLogs from '@/components/EchoLogs.vue'
import SubstatLogs from '@/components/SubstatLogs.vue'
import FindEcho from "@/components/FindEcho.vue";

const activeTab = ref<'search' | 'recent'>('search')
</script>

<style scoped>
.container {
  display: grid;
  grid-template-columns: minmax(760px, 1fr) 620px;
  align-items: start;
  gap: 20px; /* 可选：设置组件间距 */
}

.left-column,
.right-column {
  box-sizing: border-box;
  position: relative;
  isolation: isolate;
}

.left-column::before,
.right-column::before {
  content: "";
  position: absolute;
  inset: -8px;
  z-index: -1;
  pointer-events: none;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow:
    0 0 0 1px rgba(17, 24, 39, 0.08),
    0 12px 28px rgba(15, 23, 42, 0.06);
}

.left-column {
  min-width: 0;
}

.right-column {
  width: 620px;
  max-width: 620px;
}

.search-tabs {
  margin-bottom: 16px;
}

.tab-header {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.tab-button {
  min-width: 112px;
  padding: 10px 16px;
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.9);
  color: #475569;
  font-weight: 700;
  cursor: pointer;
}

.tab-button.active {
  background: #1d4ed8;
  border-color: #1d4ed8;
  color: #fff;
  box-shadow: 0 8px 18px rgba(29, 78, 216, 0.18);
}
</style>
