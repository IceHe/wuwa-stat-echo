<template>
  <div class="container">
    <div>
      <Echo />
      <EchoLogs />
    </div>
    <div>
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
  display: flex;
  //justify-content: space-between; /* 左右两端对齐 */
  justify-content: flex-start;
  gap: 20px; /* 可选：设置组件间距 */
}
</style>
