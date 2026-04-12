<template>
  <div class="pity-page">
    <div class="toolbar">
      <button @click="fetchReport" :disabled="loading">刷新报告</button>
      <span class="toolbar-meta">统计截止于：{{ formatTime(report.generated_at) }}</span>
    </div>

    <div class="summary-grid">
      <div v-for="item in report.summaries" :key="item.label" class="summary-card">
        <div class="summary-label">{{ item.label }}</div>
        <div class="summary-value">{{ item.value }}</div>
      </div>
    </div>

    <section class="section-card">
      <h2>定义</h2>
      <ul class="bullet-list">
        <li v-for="line in report.definition_notes" :key="line">{{ line }}</li>
      </ul>
    </section>

    <section class="section-card">
      <h2>结论摘要</h2>
      <ul class="bullet-list">
        <li v-for="line in report.conclusions" :key="line">{{ line }}</li>
      </ul>
    </section>

    <section class="section-card">
      <h2>方法说明</h2>
      <ul class="bullet-list">
        <li v-for="line in report.method_notes" :key="line">{{ line }}</li>
      </ul>
      <div class="method-meta">
        <span>声骸总数：{{ report.echo_total }}</span>
        <span>玩家总数：{{ report.user_total }}</span>
        <span>`id` 与 `tuned_at` 顺序冲突数：{{ report.time_order_mismatch }}</span>
      </div>
    </section>

    <section class="section-card">
      <h2>硬保底反证</h2>
      <p class="section-copy">以下是同一玩家内部、按已记录声骸序列计算的最大内部间隔。若存在“单玩家 K 个已记录声骸内必出”的硬保底，则对应间隔不可能大于 K。</p>
      <div class="table-wrap">
        <table class="report-table">
          <thead>
          <tr>
            <th>目标</th>
            <th>最大内部间隔</th>
            <th>玩家 ID</th>
            <th>起点声骸 ID</th>
            <th>终点声骸 ID</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="row in report.max_gap_rows" :key="row.label">
            <td>{{ row.label }}</td>
            <td class="number-cell strong">{{ row.max_gap }}</td>
            <td class="number-cell">{{ formatEdge(row.user_id) }}</td>
            <td class="number-cell">{{ formatEdge(row.start_echo_id) }}</td>
            <td class="number-cell">{{ formatEdge(row.end_echo_id) }}</td>
          </tr>
          </tbody>
        </table>
      </div>
      <div class="table-wrap">
        <table class="report-table">
          <thead>
          <tr>
            <th>补双暴路径</th>
            <th>样本数</th>
            <th>至少开到了下一关键孔位</th>
            <th>最终开满 5 孔</th>
            <th>最终双暴数</th>
            <th>最终双暴率</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="row in report.double_crit_path_rows" :key="row.path_label">
            <td>{{ row.path_label }}</td>
            <td class="number-cell">{{ row.sample_count }}</td>
            <td class="number-cell">{{ row.eligible_count }}</td>
            <td class="number-cell">{{ row.completed_count }}</td>
            <td class="number-cell">{{ row.final_double_crit_count }}</td>
            <td class="number-cell strong">{{ percent(row.final_double_crit_rate) }}</td>
          </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="section-card">
      <h2>停手偏差 / 删失偏差</h2>
      <ul class="bullet-list">
        <li v-for="line in report.selection_bias_notes" :key="line">{{ line }}</li>
      </ul>
      <div class="summary-grid nested-summary">
        <div v-for="item in report.stage_summaries" :key="item.label" class="summary-card subtle">
          <div class="summary-label">{{ item.label }}</div>
          <div class="summary-value compact">{{ item.value }}</div>
        </div>
      </div>
      <div class="chart-section">
        <div class="chart-card">
          <h3>继续开孔率</h3>
          <div v-for="row in continuationChartRows" :key="`chart-continue-${row.label}`" class="bar-row">
            <div class="bar-label">{{ row.label }}</div>
            <div class="bar-track">
              <div class="bar-fill continue-bar" :style="{ width: `${row.rate * 100}%` }"></div>
            </div>
            <div class="bar-value">{{ percent(row.rate) }}</div>
          </div>
        </div>
        <div class="chart-card">
          <h3>补双暴率</h3>
          <div v-for="row in doubleCritChartRows" :key="`chart-dcrit-${row.label}`" class="bar-row">
            <div class="bar-label">{{ row.label }}</div>
            <div class="bar-track">
              <div class="bar-fill dcrit-bar" :style="{ width: `${row.rate * 100}%` }"></div>
            </div>
            <div class="bar-value">{{ percent(row.rate) }}</div>
          </div>
        </div>
      </div>
      <div class="table-wrap">
        <table class="report-table">
          <thead>
          <tr>
            <th>已开孔位</th>
            <th>前缀状态</th>
            <th>样本数</th>
            <th>继续开下一孔</th>
            <th>停手</th>
            <th>继续率</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="row in report.continuation_rows" :key="`continue-${row.stage_opened}-${row.prefix_category}`">
            <td class="number-cell">{{ row.stage_opened }}</td>
            <td>{{ row.prefix_category }}</td>
            <td class="number-cell">{{ row.sample_count }}</td>
            <td class="number-cell">{{ row.continue_count }}</td>
            <td class="number-cell">{{ row.stop_count }}</td>
            <td class="number-cell strong">{{ percent(row.continue_rate) }}</td>
          </tr>
          </tbody>
        </table>
      </div>
      <div class="table-wrap">
        <table class="report-table">
          <thead>
          <tr>
            <th>已开孔位</th>
            <th>前缀状态</th>
            <th>样本数</th>
            <th>最终开满 5 孔</th>
            <th>开满率</th>
            <th>最终双暴数</th>
            <th>最终双暴率</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="row in report.double_crit_future_rows" :key="`future-${row.stage_opened}-${row.prefix_category}`">
            <td class="number-cell">{{ row.stage_opened }}</td>
            <td>{{ row.prefix_category }}</td>
            <td class="number-cell">{{ row.sample_count }}</td>
            <td class="number-cell">{{ row.completed_count }}</td>
            <td class="number-cell">{{ percent(row.completed_rate) }}</td>
            <td class="number-cell">{{ row.final_double_crit_count }}</td>
            <td class="number-cell strong">{{ percent(row.final_double_crit_rate) }}</td>
          </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-for="event in report.events" :key="event.key" class="section-card">
      <h2>{{ event.label }}</h2>
      <div class="event-head">
        <div class="event-stat">
          <span class="event-label">总体命中率</span>
          <span class="event-value">{{ percent(event.base_rate) }}</span>
        </div>
        <div class="event-stat">
          <span class="event-label">内部样本命中率</span>
          <span class="event-value">{{ percent(event.internal_base_rate) }}</span>
        </div>
        <div class="event-stat">
          <span class="event-label">样本成功数</span>
          <span class="event-value">{{ event.success_count }}</span>
        </div>
      </div>
      <ul class="bullet-list compact">
        <li>{{ event.hard_pity_summary }}</li>
        <li>{{ event.soft_pity_summary }}</li>
      </ul>
      <div class="table-wrap">
        <table class="report-table">
          <thead>
          <tr>
            <th>当前 gap</th>
            <th>样本数</th>
            <th>成功数</th>
            <th>actual_rate</th>
            <th>expected_rate</th>
            <th>delta</th>
            <th>internal_rate</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="bucket in event.buckets" :key="`${event.key}-${bucket.gap_label}`">
            <td>{{ bucket.gap_label }}</td>
            <td class="number-cell">{{ bucket.trials }}</td>
            <td class="number-cell">{{ bucket.successes }}</td>
            <td class="number-cell">{{ percent(bucket.actual_rate) }}</td>
            <td class="number-cell">{{ percent(bucket.expected_rate) }}</td>
            <td class="number-cell" :class="bucket.delta_rate >= 0 ? 'delta-up' : 'delta-down'">
              {{ signedPercent(bucket.delta_rate) }}
            </td>
            <td class="number-cell">{{ percent(bucket.internal_rate) }}</td>
          </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="section-card">
      <h2>人话总结</h2>
      <div class="plain-language">
        <p v-for="line in plainLanguageSummary" :key="line">{{ line }}</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, reactive, ref } from 'vue'

import { API_BASE_URL } from '@/stores/constants'

type SummaryItem = {
  label: string
  value: string
}

type MaxGapRow = {
  label: string
  max_gap: number
  user_id: number
  start_echo_id: number
  end_echo_id: number
}

type BucketRow = {
  gap_label: string
  trials: number
  successes: number
  actual_rate: number
  expected_rate: number
  delta_rate: number
  internal_rate: number
}

type EventReport = {
  key: string
  label: string
  success_count: number
  base_rate: number
  internal_base_rate: number
  hard_pity_summary: string
  soft_pity_summary: string
  buckets: BucketRow[]
}

type ReportResponse = {
  generated_at: string
  echo_total: number
  user_total: number
  time_order_mismatch: number
  summaries: SummaryItem[]
  definition_notes: string[]
  method_notes: string[]
  conclusions: string[]
  selection_bias_notes: string[]
  stage_summaries: SummaryItem[]
  max_gap_rows: MaxGapRow[]
  events: EventReport[]
  continuation_rows: Array<{
    stage_opened: number
    prefix_category: string
    sample_count: number
    continue_count: number
    stop_count: number
    continue_rate: number
  }>
  double_crit_future_rows: Array<{
    stage_opened: number
    prefix_category: string
    sample_count: number
    completed_count: number
    completed_rate: number
    final_double_crit_count: number
    final_double_crit_rate: number
  }>
  double_crit_path_rows: Array<{
    path_label: string
    sample_count: number
    eligible_count: number
    completed_count: number
    final_double_crit_count: number
    final_double_crit_rate: number
  }>
}

const loading = ref(false)
const report = reactive<ReportResponse>({
  generated_at: '',
  echo_total: 0,
  user_total: 0,
  time_order_mismatch: 0,
  summaries: [],
  definition_notes: [],
  method_notes: [],
  conclusions: [],
  selection_bias_notes: [],
  stage_summaries: [],
  max_gap_rows: [],
  events: [],
  continuation_rows: [],
  double_crit_future_rows: [],
  double_crit_path_rows: [],
})

const applyReport = (payload?: Partial<ReportResponse>) => {
  report.generated_at = payload?.generated_at || ''
  report.echo_total = Number(payload?.echo_total || 0)
  report.user_total = Number(payload?.user_total || 0)
  report.time_order_mismatch = Number(payload?.time_order_mismatch || 0)
  report.summaries = Array.isArray(payload?.summaries) ? payload!.summaries as SummaryItem[] : []
  report.definition_notes = Array.isArray(payload?.definition_notes) ? payload!.definition_notes as string[] : []
  report.method_notes = Array.isArray(payload?.method_notes) ? payload!.method_notes as string[] : []
  report.conclusions = Array.isArray(payload?.conclusions) ? payload!.conclusions as string[] : []
  report.selection_bias_notes = Array.isArray(payload?.selection_bias_notes) ? payload!.selection_bias_notes as string[] : []
  report.stage_summaries = Array.isArray(payload?.stage_summaries) ? payload!.stage_summaries as SummaryItem[] : []
  report.max_gap_rows = Array.isArray(payload?.max_gap_rows) ? payload!.max_gap_rows as MaxGapRow[] : []
  report.events = Array.isArray(payload?.events) ? payload!.events as EventReport[] : []
  report.continuation_rows = Array.isArray(payload?.continuation_rows) ? payload!.continuation_rows as ReportResponse['continuation_rows'] : []
  report.double_crit_future_rows = Array.isArray(payload?.double_crit_future_rows) ? payload!.double_crit_future_rows as ReportResponse['double_crit_future_rows'] : []
  report.double_crit_path_rows = Array.isArray(payload?.double_crit_path_rows) ? payload!.double_crit_path_rows as ReportResponse['double_crit_path_rows'] : []
}

const fetchReport = async () => {
  loading.value = true
  try {
    const response = await axios.get(`${API_BASE_URL}/stats/pity_analysis`)
    if (response.data?.code === 200) {
      applyReport(response.data.data)
    } else {
      alert('获取保底论证失败')
    }
  } catch (error) {
    console.error('获取保底论证失败:', error)
    alert('获取保底论证失败')
  } finally {
    loading.value = false
  }
}

const percent = (value: number) => `${(Number(value || 0) * 100).toFixed(1)}%`

const continuationChartRows = computed(() =>
  report.continuation_rows
    .filter((row) => row.stage_opened === 4 || row.stage_opened === 3)
    .map((row) => ({
      label: `${row.stage_opened}孔 ${row.prefix_category}`,
      rate: Number(row.continue_rate || 0),
    }))
)

const doubleCritChartRows = computed(() =>
  report.double_crit_path_rows
    .map((row) => ({
      label: row.path_label.replace(' -> ', ' ->\n'),
      rate: Number(row.final_double_crit_rate || 0),
    }))
)

const findPathRate = (pattern: string) =>
  Number(report.double_crit_path_rows.find((row) => row.path_label.includes(pattern))?.final_double_crit_rate || 0)

const findContinuationRate = (stage: number, category: string) =>
  Number(report.continuation_rows.find((row) => row.stage_opened === stage && row.prefix_category === category)?.continue_rate || 0)

const plainLanguageSummary = computed(() => {
  const noCritAfter4 = findContinuationRate(4, '无双暴')
  const critOnlyAfter4 = findContinuationRate(4, '仅暴击')
  const cdmgOnlyAfter4 = findContinuationRate(4, '仅暴伤')
  const noDoubleCritAfter3 = findPathRate('前 3 孔无双暴')
  const highCritAfter4 = findPathRate('前 4 孔高档位仅暴击')
  const highDmgAfter4 = findPathRate('前 4 孔高档位仅暴伤')

  return [
    '先说最直白的结论：这份数据看不出一个像样的“越黑越补”的保底曲线，常见的小阈值硬保底也已经被长间隔样本直接顶穿了。',
    '但这里的“顶穿”有严格前提：说的是单个玩家自己的已记录声骸序列，不是全站玩家混在一起，也不是“所有尝试过但没记下来的声骸”。',
    `但这不等于可以直接说“游戏底层绝对没保底”。因为玩家会挑着开。比如开到第 4 孔如果还是无双暴，继续开第 5 孔只有 ${percent(noCritAfter4)}；如果第 4 孔已经是仅暴击或仅暴伤，继续率会抬到 ${percent(critOnlyAfter4)} / ${percent(cdmgOnlyAfter4)}。`,
    `也就是说，很多“看起来还有机会补双暴”的声骸，会被人在中途就停掉。前 3 孔无双暴的样本里，最后仍然补成双暴的比例虽然只有 ${percent(noDoubleCritAfter3)}，但它不是 0，说明“前面没出，后面也可能补”。`,
    `高质量单暴更明显。前 4 孔高档位仅暴击 / 仅暴伤，最后 1 孔补成双暴的比例还有 ${percent(highCritAfter4)} / ${percent(highDmgAfter4)}。这说明玩家不只是看有没有单暴，还会看单暴质量。`,
    '所以更准确的人话版本是：现有数据不支持“单玩家在当前记录口径下存在明显硬保底”，也不支持典型软保底；但完整成品样本被玩家自己的停手策略筛过，尤其双暴结论不能直接当成底层机制证明。'
  ]
})

const signedPercent = (value: number) => {
  const number = Number(value || 0) * 100
  return `${number >= 0 ? '+' : ''}${number.toFixed(1)}%`
}

const formatEdge = (value?: number) => {
  if (!value || value < 0) {
    return '-'
  }
  return String(value)
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

onMounted(() => {
  fetchReport()
})
</script>

<style scoped>
.pity-page {
  display: grid;
  gap: 16px;
  width: 100%;
  max-width: none;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
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

.toolbar-meta {
  color: #475569;
  font-size: 13px;
}

.summary-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.summary-card,
.section-card {
  padding: 14px 16px;
  border: 1px solid #d7dde5;
  border-radius: 14px;
  background: #fff;
}

.summary-card.subtle {
  background: #f8fafc;
}

.summary-label {
  color: #64748b;
  font-size: 12px;
}

.summary-value {
  margin-top: 6px;
  color: #0f172a;
  font-size: 24px;
  font-weight: 800;
}

.summary-value.compact {
  font-size: 20px;
}

.section-card h2 {
  margin: 0 0 10px;
  font-size: 20px;
}

.nested-summary {
  margin: 14px 0;
}

.chart-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 14px;
  margin-bottom: 14px;
}

.chart-card {
  padding: 12px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #d7dde5;
}

.chart-card h3 {
  margin: 0 0 12px;
  font-size: 16px;
}

.bar-row {
  display: grid;
  grid-template-columns: minmax(120px, 1.6fr) minmax(140px, 3fr) 64px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.bar-label {
  color: #334155;
  font-size: 12px;
  line-height: 1.4;
  white-space: pre-line;
}

.bar-track {
  height: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: #e2e8f0;
}

.bar-fill {
  height: 100%;
  border-radius: 999px;
}

.continue-bar {
  background: linear-gradient(90deg, #0f766e 0%, #34d399 100%);
}

.dcrit-bar {
  background: linear-gradient(90deg, #b45309 0%, #f59e0b 100%);
}

.bar-value {
  text-align: right;
  color: #0f172a;
  font-size: 12px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.section-copy {
  margin: 0 0 12px;
  color: #475569;
  line-height: 1.6;
}

.method-meta,
.event-head {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.method-meta {
  margin-top: 12px;
  color: #475569;
  font-size: 13px;
}

.event-head {
  margin-bottom: 12px;
}

.event-stat {
  min-width: 140px;
  padding: 10px 12px;
  border-radius: 12px;
  background: #f8fafc;
}

.event-label {
  display: block;
  color: #64748b;
  font-size: 12px;
}

.event-value {
  display: block;
  margin-top: 6px;
  font-size: 20px;
  font-weight: 800;
}

.bullet-list {
  margin: 0;
  padding-left: 18px;
  color: #334155;
  line-height: 1.7;
}

.bullet-list.compact {
  margin-bottom: 12px;
}

.table-wrap {
  width: 100%;
  overflow-x: auto;
}

.report-table {
  width: 100%;
  min-width: 920px;
  border-collapse: collapse;
}

.report-table th,
.report-table td {
  padding: 9px 10px;
  border: 1px solid #d7dde5;
}

.report-table thead th {
  background: #f5efe2;
  text-align: left;
  white-space: nowrap;
}

.number-cell {
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.strong {
  font-weight: 800;
}

.delta-up {
  color: #b91c1c;
  font-weight: 700;
}

.delta-down {
  color: #15803d;
  font-weight: 700;
}

.plain-language p {
  margin: 0 0 10px;
  color: #334155;
  line-height: 1.8;
}

@media (max-width: 900px) {
  .report-table {
    min-width: 820px;
  }
}

@media (max-width: 720px) {
  .summary-grid {
    grid-template-columns: 1fr 1fr;
  }

  .report-table {
    min-width: 760px;
  }

  .bar-row {
    grid-template-columns: 1fr;
  }

  .bar-value {
    text-align: left;
  }
}
</style>
