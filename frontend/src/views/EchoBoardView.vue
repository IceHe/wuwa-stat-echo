<template>
  <div class="container">
    <div>
      <EchoBoard />
      <EchoLogsWs />
    </div>
  </div>
</template>

<script setup lang="ts">
import EchoBoard from '@/components/EchoBoard.vue'
import EchoLogsWs from '@/components/EchoLogsWs.vue'
</script>

<style scoped>
.container {
  display: flex;
//justify-content: space-between; /* 左右两端对齐 */ justify-content: flex-start;
  gap: 20px; /* 可选：设置组件间距 */
}
</style>
