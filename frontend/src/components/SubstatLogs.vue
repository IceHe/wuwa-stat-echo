<template>
  <div class="substat-log-panel">
    <button @click="fetchSubstatLogs()">词条列表 - 刷新</button>
    &nbsp;
    <span>词条数量：{{substatTotal}}</span>
    <table class="my-table">
      <thead>
      <tr style="text-align: left;">
        <th>声骸 / 孔位 / 用户</th>
        <th>副词条</th>
        <th>档位</th>
        <th>记录于</th>
      </tr>
      </thead>
      <tbody>
      <TuneLogRow
        v-for="tuneLog in substatLogs"
        :key="tuneLog.id"
        :tuneLog="tuneLog"
        :refresh-substat-logs="fetchSubstatLogs"
        :operatorId="operatorId"
      />
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import {ref, onMounted, computed} from 'vue'
import axios from 'axios'
import TuneLogRow from '@/components/SubstatLogRow.vue'
import emitter from '../stores/eventBus'
import {API_SERV} from '@/stores/constants.js'
import {authState} from '@/auth'

export default {
  name: 'TuneLogs',
  props: {
    defaultSize: {
      type: Number,
      required: false,
      defaultSize: 52,
    },
  },
  created() {
    emitter.on('refreshSubstatLogs', (message) => {
      this.fetchSubstatLogs()
    })
  },
  components: {TuneLogRow},
  setup(props) {
    const substatLogs = ref([])
    const substatTotal = ref(0)

    const fetchSubstatLogs = (pageSize: number | undefined = 0) => {
      const normalized = Number(pageSize)
      let size = Number.isNaN(normalized) ? props.defaultSize : normalized
      if (size <= 0) {
        size = props.defaultSize
      }
      axios
        .get(`http://${API_SERV}/substat_logs?page_size=${size}`)
        .then((response) => {
          // 更新用户数据
          console.log("tune logs: ", response.data) // DEBUG
          substatLogs.value = response.data.data
          substatTotal.value = response.data.data_total
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }

    // 页面加载时自动请求数据
    onMounted(fetchSubstatLogs)

    // 返回模板需要的数据和方法
    const operatorId = computed(() => authState.user?.id)
    return {
      substatLogs,
      substatTotal,
      fetchSubstatLogs,
      operatorId,
    }
  },
}
</script>

<style scoped>
.substat-log-panel {
  width: 100%;
  max-width: 620px;
}

.my-table {
  width: 100%;
  border-collapse: collapse; /* 关键：合并边框 */
  border: 1px solid #e0e0e0; /* 表格边框 */
  table-layout: fixed;
  font-size: 12px;
}

.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 6px;
  word-break: break-word;
}
</style>
