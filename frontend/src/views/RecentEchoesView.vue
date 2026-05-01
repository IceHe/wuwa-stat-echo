<template>
  <div class="recent-page">
    <div class="page-card">
      <div class="page-header">
        <div>
          <h1>最近声骸</h1>
          <p v-if="canManage" class="page-note">
            默认展示全局最近录入声骸。管理员可按玩家、套装、词条条件筛选，并继续加载更多。
          </p>
          <p v-else class="page-note">
            当前为只读观察模式，默认展示全局最近录入声骸，支持继续加载更多。
          </p>
        </div>
        <div class="page-summary">
          <span>已加载：{{ echoLogs.length }}</span>
          <button class="summary-button" @click="refreshRecentEchoes()" :disabled="loading || loadingMore">
            {{ loading ? '刷新中...' : '刷新' }}
          </button>
        </div>
      </div>

      <div v-if="canManage" class="find-panel">
        <div class="find-toolbar-row">
          <span class="name">玩家ID</span>
          <input
            class="button user-id-input"
            type="text"
            v-model="filters.user_id"
            placeholder="当前玩家ID"
            @change="setUserId(filters.user_id)"
          />
          <input
            class="button keyword-input"
            type="text"
            v-model.trim="filters.keyword"
            placeholder="玩家ID / 声骸ID / 套装 / 词条"
            @keyup.enter="searchRecentEchoes()"
          />
          <span class="clazz-chip" :style="`color: ${CLASS_COLORS[filters.clazz]};`">
            {{ filters.clazz.substring(0, 4) }}
          </span>
          <button class="button clear-button" @click="resetFilters()">清空</button>
          <button class="button search-button" @click="searchRecentEchoes()" :disabled="loading || loadingMore">
            搜索
          </button>
        </div>
        <div class="find-toolbar-row">
          <span class="name">搜索方式</span>
          <div class="mode-switch">
            <button
              class="button mode-button"
              @click="setSearchMode(SEARCH_MODE.POSITIONAL)"
              :class="{ 'mode-active': filters.search_mode === SEARCH_MODE.POSITIONAL }"
            >
              孔位搜索
            </button>
            <button
              class="button mode-button"
              @click="setSearchMode(SEARCH_MODE.SUBSTAT_SET)"
              :class="{ 'mode-active': filters.search_mode === SEARCH_MODE.SUBSTAT_SET }"
            >
              副词条搜索
            </button>
          </div>
        </div>
        <div v-if="filters.search_mode === SEARCH_MODE.SUBSTAT_SET" class="find-mode-hint">
          忽略孔位和档位，仅按包含哪些副词条搜索
        </div>
        <div class="suite-row">
          <span class="name">声骸套装</span>
          <div class="suite-scroll">
            <button class="button suite-button" @click="setClazz('')" :style="filters.clazz === '' ? 'background-color: yellow;' : ''">
              不限
            </button>
            <button
              v-for="clazz in CLASSES"
              :key="clazz"
              class="button suite-button"
              @click="setClazz(clazz)"
              :style="filters.clazz === clazz ? 'background-color: yellow;' : ''"
            >
              <span :style="`color: ${CLASS_COLORS[clazz]};`">{{ clazz.substring(0, 4) }}</span>
            </button>
          </div>
        </div>
        <template v-if="filters.search_mode === SEARCH_MODE.POSITIONAL">
          <div class="find-toolbar-row">
            <span class="name">当前孔位</span>
            <div class="find-position-row">
              <button class="substat" @click="filters.pos = 0" :style="filters.pos === 0 ? 'background-color: yellow; font-color: red' : ''">
                {{ filters.s1_desc ? filters.s1_desc : '1' }}
              </button>
              <button class="substat" @click="filters.pos = 1" :style="filters.pos === 1 ? 'background-color: yellow; font-color: red' : ''">
                {{ filters.s2_desc ? filters.s2_desc : '2' }}
              </button>
              <button class="substat" @click="filters.pos = 2" :style="filters.pos === 2 ? 'background-color: yellow; font-color: red' : ''">
                {{ filters.s3_desc ? filters.s3_desc : '3' }}
              </button>
              <button class="substat" @click="filters.pos = 3" :style="filters.pos === 3 ? 'background-color: yellow; font-color: red' : ''">
                {{ filters.s4_desc ? filters.s4_desc : '4' }}
              </button>
              <button class="substat" @click="filters.pos = 4" :style="filters.pos === 4 ? 'background-color: yellow; font-color: red' : ''">
                {{ filters.s5_desc ? filters.s5_desc : '5' }}
              </button>
            </div>
          </div>
          <div v-for="substat in SUBSTAT" :key="substat.num" class="find-substat-row">
            <span class="name" :style="`color: ${substat.font_color}; font-weight: bolder;`">{{ substat.name.substring(0, 4) }}</span>
            <div class="find-substat-buttons">
              <button class="button compact-button" @click="addAnyTuneToFilter(substat.num)" :style="`color: ${substat.font_color}`">
                不限
              </button>
              <button
                v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
                :key="value.value_number"
                class="button compact-button"
                @click="addTuneToFilter(value.substat_number, value.value_number)"
                :style="`color: ${substat.font_color}`"
              >
                {{ value.desc }}
              </button>
            </div>
          </div>
        </template>
        <template v-else>
          <div class="find-substat-set-row">
            <span class="name">副词条</span>
            <div class="find-set-buttons">
              <button
                v-for="substat in SUBSTAT"
                :key="substat.num"
                class="button set-substat-button"
                @click="toggleSubstatSet(substat.num)"
                :style="getSubstatSetButtonStyle(substat)"
              >
                {{ substat.name }}
              </button>
            </div>
          </div>
          <div class="find-toolbar-row">
            <span class="name">已选</span>
            <div class="selected-substat-summary">{{ selectedSubstatSummary() }}</div>
          </div>
        </template>
      </div>

      <div class="find-results">
        <table class="my-table">
          <thead>
            <tr style="text-align: left;">
              <th>
                <div>玩家/声骸</div>
                <div style="font-size: 10px; color: #888; font-weight: normal;">
                  管理员可删除，普通用户只读观察
                </div>
              </th>
              <th>套装</th>
              <th>词条1</th>
              <th>词条2</th>
              <th>词条3</th>
              <th>词条4</th>
              <th>词条5</th>
              <th>记录于</th>
              <th v-if="canManage">操作</th>
            </tr>
          </thead>
          <tbody>
            <EchoLogRow
              v-for="echoLog in echoLogs"
              :key="echoLog.id + echoLog.updated_at + echoLog.deleted"
              :echo-log="echoLog"
              :operator-id="operatorId"
              :can-manage="canManage"
              :show-actions="canManage"
              :show-score="false"
              :show-select-button="false"
            />
          </tbody>
        </table>

        <div v-if="!loading && echoLogs.length === 0" class="empty-state">
          暂无符合条件的声骸
        </div>

        <div class="load-more-row">
          <button class="load-more-button" @click="loadMore()" :disabled="loading || loadingMore || !hasMore">
            {{ loadingMore ? '加载中...' : (hasMore ? '加载更多' : '没有更多了') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'

import EchoLogRow from '@/components/EchoLogRow.vue'
import { authState } from '@/auth'
import emitter from '@/stores/eventBus'
import { API_BASE_URL, CLASS_COLORS, CLASSES, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants'

const PAGE_SIZE = 20
const SEARCH_MODE = {
  POSITIONAL: 'positional',
  SUBSTAT_SET: 'substat_set',
} as const

type SearchMode = (typeof SEARCH_MODE)[keyof typeof SEARCH_MODE]

type RecentEchoFilterState = {
  user_id: number
  keyword: string
  clazz: string
  search_mode: SearchMode
  substat_all_mask: number
  pos: number
  substat1: number
  substat2: number
  substat3: number
  substat4: number
  substat5: number
  s1_desc: string
  s2_desc: string
  s3_desc: string
  s4_desc: string
  s5_desc: string
}

const canManage = computed(() => authState.user?.permissions?.includes('manage') ?? false)
const operatorId = computed(() => authState.user?.id ?? null)

const buildEmptyFilters = (): RecentEchoFilterState => ({
  user_id: 0,
  keyword: '',
  clazz: '',
  search_mode: SEARCH_MODE.POSITIONAL,
  substat_all_mask: 0,
  pos: 0,
  substat1: 0,
  substat2: 0,
  substat3: 0,
  substat4: 0,
  substat5: 0,
  s1_desc: '',
  s2_desc: '',
  s3_desc: '',
  s4_desc: '',
  s5_desc: '',
})

const filters = ref<RecentEchoFilterState>(buildEmptyFilters())
const echoLogs = ref<any[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const hasMore = ref(false)
const nextCursorUpdatedAt = ref('')
const nextCursorId = ref(0)

const normalizeUserId = (userId: unknown) => {
  if (userId === '' || userId === null || userId === undefined) {
    return 0
  }
  const normalized = Number(userId)
  return Number.isNaN(normalized) ? 0 : normalized
}

const resetCursor = () => {
  nextCursorUpdatedAt.value = ''
  nextCursorId.value = 0
  hasMore.value = false
}

const buildRequestPayload = (append: boolean) => ({
  user_id: normalizeUserId(filters.value.user_id),
  clazz: filters.value.clazz,
  keyword: filters.value.keyword.trim(),
  search_mode: filters.value.search_mode,
  substat_all_mask: filters.value.search_mode === SEARCH_MODE.SUBSTAT_SET ? filters.value.substat_all_mask : 0,
  substat1: filters.value.search_mode === SEARCH_MODE.POSITIONAL ? filters.value.substat1 : 0,
  substat2: filters.value.search_mode === SEARCH_MODE.POSITIONAL ? filters.value.substat2 : 0,
  substat3: filters.value.search_mode === SEARCH_MODE.POSITIONAL ? filters.value.substat3 : 0,
  substat4: filters.value.search_mode === SEARCH_MODE.POSITIONAL ? filters.value.substat4 : 0,
  substat5: filters.value.search_mode === SEARCH_MODE.POSITIONAL ? filters.value.substat5 : 0,
  cursor_updated_at: append ? nextCursorUpdatedAt.value : '',
  cursor_id: append ? nextCursorId.value : 0,
  page_size: PAGE_SIZE,
})

const applySearchResponse = (data: any, append: boolean) => {
  const items = Array.isArray(data?.items) ? data.items : []
  echoLogs.value = append ? [...echoLogs.value, ...items] : items
  hasMore.value = Boolean(data?.has_more)
  nextCursorUpdatedAt.value = hasMore.value ? String(data?.next_cursor_updated_at || '') : ''
  nextCursorId.value = hasMore.value ? Number(data?.next_cursor_id || 0) : 0
}

const fetchRecentEchoes = async (append = false) => {
  if (append) {
    if (loading.value || loadingMore.value || !hasMore.value) {
      return
    }
    loadingMore.value = true
  } else {
    if (loading.value) {
      return
    }
    loading.value = true
    resetCursor()
  }

  try {
    const response = await axios.post(`${API_BASE_URL}/echo_log/recent_search`, buildRequestPayload(append))
    if (response.data.code === 200) {
      applySearchResponse(response.data.data, append)
      return
    }
    if (response.data.code === 403) {
      alert('仅管理员可进行条件搜索')
      return
    }
    alert('获取最近声骸失败')
  } catch (error) {
    console.error('获取最近声骸失败:', error)
    alert('获取最近声骸失败')
  } finally {
    if (append) {
      loadingMore.value = false
    } else {
      loading.value = false
    }
  }
}

const refreshRecentEchoes = () => fetchRecentEchoes(false)
const searchRecentEchoes = () => fetchRecentEchoes(false)
const loadMore = () => fetchRecentEchoes(true)

const resetFilters = () => {
  filters.value = buildEmptyFilters()
  void refreshRecentEchoes()
}

const setUserId = (userId: unknown) => {
  filters.value.user_id = normalizeUserId(userId)
}

const setClazz = (clazz: string) => {
  filters.value.clazz = clazz
}

const setSearchMode = (mode: SearchMode) => {
  filters.value.search_mode = mode
}

const hasDuplicatedSubstat = (substat: number) => (
  (1 << substat) & (
    (filters.value.pos !== 0 ? filters.value.substat1 : 0) |
    (filters.value.pos !== 1 ? filters.value.substat2 : 0) |
    (filters.value.pos !== 2 ? filters.value.substat3 : 0) |
    (filters.value.pos !== 3 ? filters.value.substat4 : 0) |
    (filters.value.pos !== 4 ? filters.value.substat5 : 0)
  )
)

const isSubstatSetSelected = (substat: number) => ((filters.value.substat_all_mask >> substat) & 1) === 1

const toggleSubstatSet = (substat: number) => {
  filters.value.substat_all_mask ^= 1 << substat
}

const selectedSubstatSummary = () => {
  const names = SUBSTAT
    .filter((substat) => isSubstatSetSelected(substat.num))
    .map((substat) => substat.name)
  return names.length > 0 ? names.join(' / ') : '不限'
}

const getSubstatSetButtonStyle = (substat: { num: number; font_color: string }) => (
  isSubstatSetSelected(substat.num)
    ? `background-color: #fff4a3; color: ${substat.font_color}; border-color: ${substat.font_color}; font-weight: 700;`
    : `color: ${substat.font_color};`
)

const setSubstatToCurrentPos = (substatBits: number, substatDesc: string) => {
  switch (filters.value.pos) {
    case 0:
      filters.value.substat1 = substatBits
      filters.value.s1_desc = substatDesc
      break
    case 1:
      filters.value.substat2 = substatBits
      filters.value.s2_desc = substatDesc
      break
    case 2:
      filters.value.substat3 = substatBits
      filters.value.s3_desc = substatDesc
      break
    case 3:
      filters.value.substat4 = substatBits
      filters.value.s4_desc = substatDesc
      break
    case 4:
      filters.value.substat5 = substatBits
      filters.value.s5_desc = substatDesc
      break
    default:
      alert('请先选择孔位')
      return false
  }
  return true
}

const addAnyTuneToFilter = (substat: number) => {
  if (hasDuplicatedSubstat(substat)) {
    alert('已存在相同词条，请检查')
    return
  }
  setSubstatToCurrentPos(1 << substat, `${SUBSTAT[substat].name} 不限`)
}

const addTuneToFilter = (substat: number, value: number) => {
  if (hasDuplicatedSubstat(substat)) {
    alert('已存在相同词条，请检查')
    return
  }
  const substatDesc = SUBSTAT_VALUE_MAP[substat][value].desc_full
  setSubstatToCurrentPos((1 << substat) | (1 << (value + 13)), substatDesc)
}

const handleRefreshRecentEchoLogs = () => {
  void refreshRecentEchoes()
}

onMounted(() => {
  ;(emitter as any).on('refreshRecentEchoLogs', handleRefreshRecentEchoLogs)
  void refreshRecentEchoes()
})

onUnmounted(() => {
  ;(emitter as any).off('refreshRecentEchoLogs', handleRefreshRecentEchoLogs)
})
</script>

<style scoped>
.recent-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-card {
  position: relative;
  padding: 20px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.78);
  box-shadow:
    0 0 0 1px rgba(17, 24, 39, 0.08),
    0 18px 36px rgba(15, 23, 42, 0.08);
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.page-header h1 {
  margin: 0 0 8px;
}

.page-note {
  margin: 0;
  color: #64748b;
  font-size: 14px;
}

.page-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  white-space: nowrap;
  font-weight: 700;
}

.summary-button,
.load-more-button {
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 999px;
  background: #fff;
  cursor: pointer;
}

.summary-button {
  padding: 8px 14px;
}

.find-panel {
  margin-bottom: 16px;
}

.find-toolbar-row,
.find-substat-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.find-substat-set-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.name {
  flex: 0 0 64px;
  min-width: 64px;
  padding-top: 10px;
}

.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
  padding: 0 10px;
  height: 40px;
  text-align: center;
}

.user-id-input {
  width: 120px;
  min-width: 120px;
  font-weight: bolder;
}

.keyword-input {
  flex: 1 1 180px;
  min-width: 180px;
  justify-content: flex-start;
}

.clazz-chip {
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  font-weight: bolder;
}

.clear-button {
  min-width: 64px;
  color: red;
}

.search-button {
  min-width: 64px;
}

.mode-switch {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mode-button {
  min-width: 96px;
}

.mode-active {
  background-color: #fff4a3;
  font-weight: 700;
}

.find-mode-hint {
  margin: -2px 0 8px 72px;
  color: #64748b;
  font-size: 12px;
}

.find-position-row,
.find-substat-buttons,
.find-set-buttons {
  min-width: 0;
}

.find-position-row {
  display: flex;
  flex-wrap: nowrap;
  gap: 0;
}

.find-substat-buttons {
  display: flex;
  flex-wrap: nowrap;
  gap: 0;
  overflow-x: visible;
}

.find-set-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.suite-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.suite-scroll {
  display: flex;
  gap: 0;
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 100%;
  padding-bottom: 6px;
}

.suite-button {
  flex: 0 0 42px;
  width: 42px;
  min-width: 42px;
  max-width: 42px;
  height: 84px;
}

.substat {
  width: 88px;
  min-width: 88px;
  max-width: 88px;
  height: 40px;
  text-align: center;
}

.compact-button {
  width: 54px;
  min-width: 54px;
  max-width: 54px;
  padding: 0 6px;
}

.set-substat-button {
  width: auto;
  min-width: 72px;
  max-width: none;
}

.selected-substat-summary {
  padding-top: 10px;
  color: #334155;
}

.my-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #e0e0e0;
  table-layout: fixed;
  font-size: 12px;
}

.my-table td,
.my-table th {
  border: 1px solid #ddd;
  padding: 6px;
  word-break: break-word;
}

.empty-state {
  padding: 18px 0;
  color: #64748b;
  text-align: center;
}

.load-more-row {
  display: flex;
  justify-content: center;
  padding-top: 16px;
}

.load-more-button {
  min-width: 140px;
  padding: 10px 18px;
}

@media (max-width: 960px) {
  .page-header {
    flex-direction: column;
  }

  .page-summary {
    white-space: normal;
  }
}
</style>
