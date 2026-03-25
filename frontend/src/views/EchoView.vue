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

.left-column {
  min-width: 0;
}

.right-column {
  width: 620px;
  max-width: 620px;
}
</style>
