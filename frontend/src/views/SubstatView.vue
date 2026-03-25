<template>
  <div class="container">
    <div>
      <Substat :position="0" />
      <SubstatLogs default-size="52"/>
    </div>
    <SubstatStats />
  </div>
</template>

<script setup lang="ts">
import Substat from '@/components/Substat.vue'
import SubstatLogs from '@/components/SubstatLogs.vue'
import SubstatStats from '@/components/SubstatAnalysis.vue'
</script>

<style scoped>
.container {
  display: flex;
  justify-content: space-between; /* 左右两端对齐 */
  gap: 20px; /* 可选：设置组件间距 */
}
</style>
