<template>
  <div class="max-gap-page">
    <div class="toolbar">
      <div class="toolbar-left">
        <label class="field">
          <span>玩家 ID</span>
          <input
            v-model.trim="form.userId"
            type="text"
            inputmode="numeric"
            placeholder="留空查看全部玩家"
            @keydown.enter.prevent="fetchStats(false)"
          >
        </label>
        <button @click="fetchStats(false)" :disabled="loading">查询</button>
        <button @click="clearUserId" :disabled="loading">清空</button>
        <button @click="fetchStats(true)" :disabled="loading">强制刷新</button>
      </div>
      <div class="toolbar-right">
        <span class="hint">首次查询会生成缓存；强制刷新每个范围 1 小时一次。</span>
      </div>
    </div>

    <div class="summary-card">
      <div class="summary-row">
        <span>统计范围：{{ stats.scope_label || '全部玩家' }}</span>
        <span>词条总数：{{ stats.tune_log_total }}</span>
      </div>
      <div class="summary-row">
        <span>统计截止于：{{ formatTime(stats.generated_at) }}</span>
        <span>数据来源：{{ stats.cache_hit ? '缓存' : '即时统计' }}</span>
      </div>
      <div class="summary-row">
        <span>上次强制刷新：{{ formatTime(stats.last_forced_refresh_at) }}</span>
        <span>下次可强刷：{{ formatTime(stats.refresh_available_at) }}</span>
      </div>
      <div class="summary-row">
        <span>以下最大间隔均按单个玩家内部序列计算。</span>
      </div>
      <div v-if="stats.refresh_blocked" class="summary-alert">当前范围仍在强制刷新冷却中，展示的是现有缓存。</div>
      <div v-else-if="stats.force_applied" class="summary-ok">已按当前范围重新统计并更新缓存。</div>
    </div>

    <div class="table-wrap">
      <table class="my-table">
        <thead>
        <tr>
          <th>副词条</th>
          <th>最大间隔</th>
          <th>最大间隔起点 ID</th>
          <th>最大间隔终点 ID</th>
          <th v-if="showOwnerColumn">玩家 ID</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="row in stats.rows" :key="row.substat">
          <td :style="`color: ${getSubstatColor(1 << row.substat)}; font-weight: 700;`">{{ row.name_cn }}</td>
          <td class="number-cell">{{ row.max_gap }}</td>
          <td class="number-cell">{{ formatGapEdge(row.max_gap_start_id) }}</td>
          <td class="number-cell">{{ formatGapEdge(row.max_gap_end_id) }}</td>
          <td v-if="showOwnerColumn" class="number-cell">{{ row.owner_user_id || '-' }}</td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { API_BASE_URL, getSubstatColor } from '@/stores/constants'

type MaxGapRow = {
  substat: number
  name: string
  name_cn: string
  owner_user_id: number
  max_gap: number
  occurrence_count: number
  leading_gap: number
  trailing_gap: number
  max_gap_start_id: number
  max_gap_end_id: number
}

type MaxGapResponse = {
  user_id: number
  scope_label: string
  tune_log_total: number
  generated_at: string
  last_forced_refresh_at: string
  refresh_available_at: string
  cache_hit: boolean
  force_applied: boolean
  refresh_blocked: boolean
  rows: MaxGapRow[]
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const form = reactive({
  userId: typeof route.query.user_id === 'string' ? route.query.user_id : '',
})
const stats = reactive<MaxGapResponse>({
  user_id: 0,
  scope_label: '',
  tune_log_total: 0,
  generated_at: '',
  last_forced_refresh_at: '',
  refresh_available_at: '',
  cache_hit: false,
  force_applied: false,
  refresh_blocked: false,
  rows: [],
})

const showOwnerColumn = computed(() => stats.user_id === 0)

const syncQuery = () => {
  const userId = form.userId.trim()
  router.replace({
    query: {
      ...route.query,
      user_id: userId || undefined,
    },
  })
}

const applyStats = (payload?: Partial<MaxGapResponse>) => {
  stats.user_id = Number(payload?.user_id || 0)
  stats.scope_label = payload?.scope_label || ''
  stats.tune_log_total = Number(payload?.tune_log_total || 0)
  stats.generated_at = payload?.generated_at || ''
  stats.last_forced_refresh_at = payload?.last_forced_refresh_at || ''
  stats.refresh_available_at = payload?.refresh_available_at || ''
  stats.cache_hit = Boolean(payload?.cache_hit)
  stats.force_applied = Boolean(payload?.force_applied)
  stats.refresh_blocked = Boolean(payload?.refresh_blocked)
  stats.rows = Array.isArray(payload?.rows) ? payload!.rows as MaxGapRow[] : []
}

const fetchStats = async (force: boolean) => {
  loading.value = true
  syncQuery()
  try {
    const userId = form.userId.trim()
    const params = new URLSearchParams()
    if (userId) {
      params.set('user_id', userId)
    }
    if (force) {
      params.set('force', '1')
    }
    const response = await axios.get(`${API_BASE_URL}/stats/substat_max_gap?${params.toString()}`)
    if (response.data?.code === 200) {
      applyStats(response.data.data)
    } else {
      alert('获取副词条最大间隔失败')
    }
  } catch (error) {
    console.error('获取副词条最大间隔失败:', error)
    alert('获取副词条最大间隔失败')
  } finally {
    loading.value = false
  }
}

const clearUserId = () => {
  form.userId = ''
  fetchStats(false)
}

const formatTime = (value?: string) => {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  const pad = (num: number) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const formatGapEdge = (value?: number) => {
  if (value == null || value < 0) {
    return '-'
  }
  return String(value)
}

onMounted(() => {
  fetchStats(false)
})
</script>

<style scoped>
.max-gap-page {
  display: grid;
  gap: 12px;
  width: 100%;
  max-width: none;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  align-items: end;
  gap: 10px;
  flex-wrap: wrap;
}

.toolbar-right {
  display: flex;
  align-items: center;
}

.field {
  display: grid;
  gap: 6px;
  color: #425466;
  font-size: 13px;
}

.field input {
  width: 220px;
  padding: 8px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  background: #fff;
}

button {
  border: 0;
  border-radius: 999px;
  padding: 9px 14px;
  color: #fff;
  background: #23404b;
  cursor: pointer;
}

button:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.hint {
  color: #64748b;
  font-size: 12px;
}

.summary-card {
  display: grid;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #f8fafc;
  border: 1px solid #d7dde5;
}

.summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  color: #334155;
  font-size: 13px;
}

.summary-alert {
  color: #9a3412;
  font-weight: 700;
}

.summary-ok {
  color: #166534;
  font-weight: 700;
}

.my-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #d7dde5;
  background: #fff;
}

.table-wrap {
  width: 100%;
  overflow-x: auto;
}

.my-table th,
.my-table td {
  border: 1px solid #d7dde5;
  padding: 8px 8px;
}

.my-table thead th {
  background: #f5efe2;
  text-align: left;
  white-space: nowrap;
}

.number-cell {
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .toolbar-left {
    width: 100%;
  }

  .field input {
    width: min(100%, 280px);
  }

  .my-table {
    min-width: 980px;
  }
}

@media (max-width: 720px) {
  .toolbar-right,
  .summary-row {
    width: 100%;
  }

  .my-table {
    min-width: 920px;
  }
}
</style>
