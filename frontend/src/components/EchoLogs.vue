<template>
  <div style="min-width: 480px">
    <button @click="fetchEchoLogs()">声骸列表 - 刷新</button>
    &nbsp;
    <button @click="forceRefreshTemplates()" :disabled="scoreTemplateState.loading">模板强刷</button>
    &nbsp;
    <span>声骸数量：{{echoTotal}}</span>
    &nbsp;
    <span class="template-context-text">
      评分上下文：{{ scoreTemplateContext.resonator || '未选模板' }} / {{ scoreTemplateContext.cost || '未选Cost' }}
    </span>
    &nbsp;
    <span class="template-context-text">
      模板源：{{ templateSourceLabel }}
    </span>
    <table class="my-table">
      <thead>
      <tr style="text-align: left;">
        <th>
          <div>玩家/声骸</div>
          <div style="font-size: 10px; color: #888; font-weight: normal;">自己的声骸或管理员可修改</div>
        </th>
        <th>套装</th>
        <th>词条1</th>
        <th>词条2</th>
        <th>词条3</th>
        <th>词条4</th>
        <th>词条5</th>
        <th>
          <div>当前分</div>
          <div style="font-size: 10px; color: #888; font-weight: normal;">理论最高</div>
        </th>
        <th>记录于</th>
        <th>操作</th>
      </tr>
      </thead>
      <tbody>
      <EchoLogRow
        v-for="echoLog in echoLogs"
        :key="echoLog.id + echoLog.updated_at + echoLog.deleted"
        :echoLog="echoLog"
        :operatorId="operatorId"
        :canManage="canManage"
      />
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import {ref, onMounted, computed} from 'vue'
import EchoLogRow from "@/components/EchoLogRow.vue";
import axios from 'axios'
import emitter from '../stores/eventBus'
import {API_BASE_URL} from '@/stores/constants.js'
import {authState} from '@/auth'
import {
  ensureScoreTemplatesLoaded,
  refreshScoreTemplates,
  scoreTemplateContext,
  scoreTemplateState,
} from '@/stores/scoreTemplates.ts'

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
  created() {
    emitter.on('refreshEchoLogs', (message) => {
      this.fetchEchoLogs()
    })
    emitter.on('syncEchoLog', (echoLog) => {
      this.upsertEchoLog(echoLog)
    })
  },
  setup(props) {
    const echoLogs = ref([])
    const echoTotal = ref(0)
    const templateSourceLabel = computed(() => {
      if (scoreTemplateState.source === 'remote') {
        return `远端 ${scoreTemplateState.version}`
      }
      if (scoreTemplateState.source === 'cache') {
        return `缓存 ${scoreTemplateState.version}`
      }
      return '内置'
    })

    const upsertEchoLog = (echoLog) => {
      if (!echoLog?.id) {
        return
      }

      const nextEchoLog = {
        ...echoLog,
      }
      const index = echoLogs.value.findIndex((item) => item.id === nextEchoLog.id)
      if (index >= 0) {
        echoLogs.value[index] = {
          ...echoLogs.value[index],
          ...nextEchoLog,
        }
        if (index > 0) {
          const [currentEcho] = echoLogs.value.splice(index, 1)
          echoLogs.value.unshift(currentEcho)
        }
        return
      }

      echoLogs.value.unshift(nextEchoLog)
      echoTotal.value += 1
      if (echoLogs.value.length > props.defaultSize) {
        echoLogs.value.length = props.defaultSize
      }
    }

    const fetchEchoLogs = async (pageSize: number | undefined = 0) => {
      const normalized = Number(pageSize)
      let size = Number.isNaN(normalized) ? props.defaultSize : normalized
      if (size <= 0) {
        size = props.defaultSize
      }
      try {
        const response = await axios.get(`${API_BASE_URL}/echo_logs?page_size=${size}`)
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

    // 页面加载时自动请求数据
    onMounted(() => {
      fetchEchoLogs()
      ensureScoreTemplatesLoaded()
    })

    // 返回模板需要的数据和方法
    const operatorId = computed(() => authState.user?.id)
    const canManage = computed(() => authState.user?.permissions?.includes('manage') ?? false)

    return {
      echoLogs,
      echoTotal,
      fetchEchoLogs,
      upsertEchoLog,
      operatorId,
      canManage,
      scoreTemplateContext,
      scoreTemplateState,
      templateSourceLabel,
      forceRefreshTemplates: () => refreshScoreTemplates(true),
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

.template-context-text {
  font-size: 12px;
  color: #666;
}
</style>
