<template>
  <div style="min-width: 480px">
    <button @click="fetchEchoLogs()">声骸列表 - 刷新</button>
    &nbsp;
    <span>声骸数量：{{echoTotal}}</span>
    <table class="my-table">
      <thead>
      <tr style="text-align: left;">
        <th>玩家/声骸</th>
        <th>套装</th>
        <th>词条1</th>
        <th>词条2</th>
        <th>词条3</th>
        <th>词条4</th>
        <th>词条5</th>
        <th>记录于</th>
        <!--<th>操作</th>-->
      </tr>
      </thead>
      <tbody>
      <EchoLogRow
        v-for="echoLog in echoLogs"
        :key="echoLog.id + echoLog.updated_at + echoLog.deleted"
        :echoLog="echoLog"
        :can-edit="false"
      />
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import {ref, onMounted} from 'vue'
import EchoLogRow from "@/components/EchoLogRow.vue";
import axios from 'axios'
import {API_SERV} from '@/stores/constants.js'

export default {
  name: 'EchoLogs',
  props: {
    defaultSize: {
      type: Number,
      required: false,
      default: 35,
    },
  },
  components: {EchoLogRow},
  setup(props) {
    const echoLogs = ref([])
    const echoTotal = ref(0)

    const fetchEchoLogs = async (pageSize: number | undefined = 0) => {
      const normalized = Number(pageSize)
      let size = Number.isNaN(normalized) ? props.defaultSize : normalized
      if (size <= 0) {
        size = props.defaultSize
      }
      try {
        const response = await axios.get(`http://${API_SERV}/echo_logs?page_size=${size}`)
        console.log('listEchoLog: ', response.data)
        if (response.data.code === 200) {
          echoLogs.value = response.data.data
          echoTotal.value = response.data.data_total
        } else {
          console.error('获取声骸历史 失败:', response.data)
          alert('获取声骸历史 失败')
        }
      } catch (error) {
        console.error('获取声骸历史 请求失败:', error)
      }
    }

    onMounted(fetchEchoLogs)

    return {
      echoLogs,
      echoTotal,
      fetchEchoLogs,
    }
  },
}
</script>

<style scoped>
.my-table {
  width: 100%;
  border-collapse: collapse; /* 关键：合并边框 */
  border: 1px solid #e0e0e0; /* 表格边框 */
}

.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 8px;
}
</style>
