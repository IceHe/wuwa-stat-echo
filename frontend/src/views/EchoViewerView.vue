<template>
  <div class="container">
    <div class="left-column">
      <Echo ref="echoRef" :sync-echo-id-query="false" />
      <EchoLogs />
    </div>
    <aside class="right-column">
      <div class="viewer-header">
        <h2>声骸实时查看器</h2>
        <p class="operator-info">当前查看用户: {{ operatorId }}</p>
        <p class="connection-status" :class="{ connected: isConnected }">
          {{ isConnected ? '✅ 已连接' : '🔴 连接断开，正在重连...' }}
        </p>
      </div>

      <div class="ws-log-panel">
        <h3>WebSocket 消息日志</h3>
        <div class="log-content" ref="logContentRef">
          <div class="log-item" v-for="(log, index) in wsLogs" :key="index">
            <span class="log-time">{{ log.time }}</span>
            <span class="log-type" :class="log.type">{{ log.type }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import Echo from '@/components/Echo.vue'
import EchoLogs from '@/components/EchoLogs.vue'
import { API_BASE_URL } from '@/stores/constants'
import emitter from '@/stores/eventBus'
import {
  subscribeScoreTemplateChange,
  type ScoreTemplateChangeEvent,
} from '@/stores/scoreTemplateSync'

const route = useRoute()
const operatorId = ref('')
const isConnected = ref(false)
const ws = ref<WebSocket | null>(null)
const echoRef = ref<any>(null)
const wsLogs = ref<Array<{time: string, type: 'info' | 'error' | 'message', message: string}>>([])
const logContentRef = ref<HTMLElement | null>(null)
let refreshTimer: number | null = null
let unsubscribeScoreTemplateSync: (() => void) | null = null

const extractEchoId = (message: any): number => {
  const echoIdFromEchoLog = Number(message?.data?.echo_log?.id || 0)
  if (echoIdFromEchoLog > 0) {
    return echoIdFromEchoLog
  }
  const echoIdFromData = Number(message?.data?.id || 0)
  if (echoIdFromData > 0) {
    return echoIdFromData
  }
  return 0
}

const refreshViewerData = (message: any) => {
  if (refreshTimer !== null) {
    window.clearTimeout(refreshTimer)
  }

  refreshTimer = window.setTimeout(() => {
    emitter.emit('refreshEchoLogs')
    emitter.emit('refreshSubstatLogs')

    const echoId = extractEchoId(message)
    if (echoRef.value?.getEchoLog) {
      echoRef.value.getEchoLog(echoId, { silent: true })
    } else {
      emitter.emit('getEchoLog', echoId)
    }
  }, 150)
}

const addLog = (type: 'info' | 'error' | 'message', message: string) => {
  const now = new Date()
  const time = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`

  wsLogs.value.push({
    time,
    type,
    message
  })

  // 自动滚动到底部
  nextTick(() => {
    if (logContentRef.value) {
      logContentRef.value.scrollTop = logContentRef.value.scrollHeight
    }
  })
}

const handleScoreTemplateChanged = (payload: ScoreTemplateChangeEvent) => {
  const label = payload.field === 'resonator' ? '评分模板' : 'Cost主词条'
  const templateName = payload.resonator || '未设置'
  const cost = payload.cost || '未设置'
  addLog('info', `${label}切换为 ${payload.value}，当前评分上下文: ${templateName} / ${cost}`)
}

const applyRemoteScoreTemplateChange = (payload: ScoreTemplateChangeEvent) => {
  if (echoRef.value?.applyScoreTemplate) {
    echoRef.value.applyScoreTemplate(payload)
  }
  handleScoreTemplateChanged(payload)
}

const connectWebSocket = () => {
  if (!operatorId.value) {
    addLog('error', 'URL参数中缺少 operator_id')
    console.error('operator_id is required')
    return
  }

  // 构建 WebSocket 地址
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsBaseUrl = API_BASE_URL.replace(/^https?:\/\//, '')
  const wsUrl = `${wsProtocol}//${wsBaseUrl}/ws?operator_id=${operatorId.value}`

  addLog('info', `正在连接: ${wsUrl}`)
  console.log('Connecting to WebSocket:', wsUrl)

  ws.value = new WebSocket(wsUrl)

  ws.value.onopen = () => {
    addLog('info', 'WebSocket 连接成功')
    console.log('WebSocket connected successfully')
    isConnected.value = true
  }

  ws.value.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data)
      addLog('message', `收到消息: ${message.type} - ${JSON.stringify(message.data)}`)
      console.log('Received WebSocket message:', message)

      if (message.type === 'score_template_changed') {
        applyRemoteScoreTemplateChange(message.data)
        return
      }

      // 局部刷新数据，避免整页 reload 闪屏
      addLog('info', '正在局部刷新数据（不重载页面）')
      refreshViewerData(message)
    } catch (e) {
      addLog('error', `消息解析失败: ${e} - 原始内容: ${event.data}`)
      console.error('Failed to parse WebSocket message:', e)
    }
  }

  ws.value.onerror = (error) => {
    addLog('error', `连接错误: ${JSON.stringify(error)}`)
    console.error('WebSocket error:', error)
    isConnected.value = false
  }

  ws.value.onclose = (event) => {
    addLog('info', `连接断开 (code: ${event.code}, reason: ${event.reason})，3秒后自动重连...`)
    console.log('WebSocket disconnected', event)
    isConnected.value = false
    // 自动重连
    setTimeout(connectWebSocket, 3000)
  }
}

onMounted(() => {
  emitter.on('scoreTemplateChanged', handleScoreTemplateChanged)
  unsubscribeScoreTemplateSync = subscribeScoreTemplateChange(applyRemoteScoreTemplateChange)
  operatorId.value = route.query.operator_id as string
  if (operatorId.value) {
    connectWebSocket()
  } else {
    addLog('error', 'URL参数中缺少 operator_id')
    console.error('No operator_id provided in URL')
  }
})

onUnmounted(() => {
  emitter.off('scoreTemplateChanged', handleScoreTemplateChanged)
  if (unsubscribeScoreTemplateSync) {
    unsubscribeScoreTemplateSync()
    unsubscribeScoreTemplateSync = null
  }
  if (ws.value) {
    ws.value.close()
  }
  if (refreshTimer !== null) {
    window.clearTimeout(refreshTimer)
  }
})
</script>

<style scoped>
.container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 20px;
  display: flex;
  align-items: flex-start;
  gap: 20px;
}

.left-column {
  flex: 1;
  min-width: 0;
}

.right-column {
  width: 360px;
  flex: 0 0 360px;
  position: sticky;
  top: 20px;
}

.viewer-header {
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
  margin-bottom: 12px;
}

.viewer-header h2 {
  margin: 0 0 8px 0;
  font-size: 20px;
}

.operator-info {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #666;
}

.connection-status {
  margin: 0;
  font-size: 14px;
  color: #dc2626;
}

.connection-status.connected {
  color: #059669;
}

.ws-log-panel {
  background: #1e1e1e;
  border-radius: 8px;
  padding: 16px;
  font-family: monospace;
}

.ws-log-panel h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #fff;
}

.log-content {
  height: 300px;
  overflow-y: auto;
  background: #000;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
}

.log-item {
  display: flex;
  gap: 8px;
  margin-bottom: 2px;
  flex-wrap: wrap;
}

.log-time {
  color: #888;
  min-width: 70px;
}

.log-type {
  min-width: 60px;
  font-weight: bold;
}

.log-type.info {
  color: #4ec9b0;
}

.log-type.error {
  color: #f44747;
}

.log-type.message {
  color: #dcdcaa;
}

.log-message {
  color: #d4d4d4;
  word-break: break-all;
}

/* 滚动条样式 */
.log-content::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.log-content::-webkit-scrollbar-track {
  background: #333;
  border-radius: 3px;
}

.log-content::-webkit-scrollbar-thumb {
  background: #666;
  border-radius: 3px;
}

.log-content::-webkit-scrollbar-thumb:hover {
  background: #888;
}

@media (max-width: 1100px) {
  .container {
    display: block;
  }

  .right-column {
    width: auto;
    position: static;
    margin-bottom: 20px;
  }
}
</style>
