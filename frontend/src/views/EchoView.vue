<template>
  <div class="container">
    <div class="left-column">
      <Echo />
      <EchoLogs />
    </div>
    <div class="right-column">
      <FindEcho />
      <SubstatLogs :default-size="52" />
    </div>
  </div>
</template>

<script setup lang="ts">
import Echo from '@/components/Echo.vue'
import EchoLogs from '@/components/EchoLogs.vue'
import SubstatLogs from '@/components/SubstatLogs.vue'
import FindEcho from "@/components/FindEcho.vue";
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
</style>
