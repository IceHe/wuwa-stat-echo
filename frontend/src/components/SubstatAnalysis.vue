<template>
  <section class="analysis-page">
    <header class="hero">
      <div>
        <p class="eyebrow">Advanced Stats</p>
        <h1>高级统计面板</h1>
        <p class="hero-copy">
          同时查看调谐分布、双暴率、目标命中率，以及个人 vs 全站的偏差和显著性。
        </p>
      </div>
      <form class="control-panel" @submit.prevent="applyFilters">
        <label class="field">
          <span>用户 ID</span>
          <input v-model.trim="userIdInput" inputmode="numeric" placeholder="全站留空" />
        </label>
        <label class="field">
          <span>统计窗口</span>
          <select v-model="windowInput">
            <option v-for="option in windowOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>
        <label class="field">
          <span>目标词条</span>
          <select v-model.number="targetBitsInput">
            <option :value="3">双暴</option>
            <option :value="7">双暴 + 攻击</option>
            <option :value="259">双暴 + 共效</option>
            <option :value="2051">双暴 + 共解</option>
          </select>
        </label>
        <div class="control-actions">
          <button type="submit">刷新</button>
          <span class="hint">当前窗口：{{ tuneStats?.window ?? windowInput }}</span>
        </div>
      </form>
    </header>

    <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

    <section class="summary-grid">
      <article class="summary-card accent-gold">
        <p class="summary-label">调谐样本</p>
        <p class="summary-value">{{ tuneStats?.data_total ?? 0 }}</p>
        <p class="summary-meta">
          {{ formatWindowLabel(tuneStats?.window ?? windowInput) }}
          <span v-if="tuneStats?.baseline_compare"> · 对比已启用</span>
        </p>
      </article>
      <article class="summary-card accent-ink">
        <p class="summary-label">双暴率</p>
        <p class="summary-value">{{ formatRate(dcritStats?.dcrit_rate_stats) }}</p>
        <p class="summary-meta">{{ formatConfidence(dcritStats?.dcrit_rate_stats) }}</p>
        <p v-if="dcritComparison?.bias_hint" class="summary-hint">
          {{ dcritComparison.bias_hint.message }}
        </p>
      </article>
      <article class="summary-card accent-teal">
        <p class="summary-label">目标命中率</p>
        <p class="summary-value">{{ formatRate(echoAnalysis?.target_rate_stats) }}</p>
        <p class="summary-meta">{{ formatConfidence(echoAnalysis?.target_rate_stats) }}</p>
        <p v-if="echoComparison?.bias_hint" class="summary-hint">
          {{ echoComparison.bias_hint.message }}
        </p>
      </article>
      <article class="summary-card accent-coral">
        <p class="summary-label">窗口声骸数</p>
        <p class="summary-value">{{ echoAnalysis?.sample_size ?? dcritStats?.echo_count ?? 0 }}</p>
        <p class="summary-meta">
          目标命中 {{ echoAnalysis?.target ?? 0 }} 次 · 双暴 {{ dcritStats?.dcrit_total ?? 0 }} 次
        </p>
      </article>
    </section>

    <section class="comparison-grid">
      <article class="panel">
        <div class="panel-head">
          <div>
            <p class="panel-kicker">Echo Analysis</p>
            <h2>目标命中对比</h2>
          </div>
          <span class="tag">{{ formatWindowLabel(echoAnalysis?.window ?? windowInput) }}</span>
        </div>
        <template v-if="echoComparison">
          <div class="compare-strip">
            <div>
              <span class="compare-label">个人</span>
              <strong>{{ formatRate(echoComparison.user) }}</strong>
            </div>
            <div>
              <span class="compare-label">全站</span>
              <strong>{{ formatRate(echoComparison.global) }}</strong>
            </div>
            <div>
              <span class="compare-label">差值</span>
              <strong :class="rateClass(echoComparison.delta_rate)">
                {{ formatSignedRate(echoComparison.delta_rate) }}
              </strong>
            </div>
          </div>
          <div class="insight-list">
            <p>
              <strong>显著性：</strong>
              {{ formatSignificance(echoComparison.significance) }}
            </p>
            <p>
              <strong>提示：</strong>
              {{ echoComparison.bias_hint?.message ?? '暂无' }}
            </p>
          </div>
        </template>
        <p v-else class="empty-state">传入 `user_id` 后显示个人 vs 全站对比。</p>
      </article>

      <article class="panel">
        <div class="panel-head">
          <div>
            <p class="panel-kicker">Dcrit</p>
            <h2>双暴率对比</h2>
          </div>
          <span class="tag">{{ formatWindowLabel(dcritStats?.window ?? windowInput) }}</span>
        </div>
        <template v-if="dcritComparison">
          <div class="compare-strip">
            <div>
              <span class="compare-label">个人</span>
              <strong>{{ formatRate(dcritComparison.user) }}</strong>
            </div>
            <div>
              <span class="compare-label">全站</span>
              <strong>{{ formatRate(dcritComparison.global) }}</strong>
            </div>
            <div>
              <span class="compare-label">差值</span>
              <strong :class="rateClass(dcritComparison.delta_rate)">
                {{ formatSignedRate(dcritComparison.delta_rate) }}
              </strong>
            </div>
          </div>
          <div class="insight-list">
            <p>
              <strong>显著性：</strong>
              {{ formatSignificance(dcritComparison.significance) }}
            </p>
            <p>
              <strong>提示：</strong>
              {{ dcritComparison.bias_hint?.message ?? '暂无' }}
            </p>
          </div>
        </template>
        <p v-else class="empty-state">传入 `user_id` 后显示个人 vs 全站对比。</p>
      </article>
    </section>

    <section class="comparison-grid">
      <article class="panel">
        <div class="panel-head">
          <div>
            <p class="panel-kicker">Echo Summary</p>
            <h2>目标养成成本</h2>
          </div>
        </div>
        <div class="metric-list">
          <div class="metric-item">
            <span>目标命中距离</span>
            <strong>{{ echoAnalysis?.target_echo_distance ?? 0 }} 声骸</strong>
          </div>
          <div class="metric-item">
            <span>目标副词条距离</span>
            <strong>{{ echoAnalysis?.target_substat_distance ?? 0 }} 次</strong>
          </div>
          <div class="metric-item">
            <span>调谐器消耗</span>
            <strong>{{ echoAnalysis?.tuner_consumed ?? 0 }}</strong>
          </div>
          <div class="metric-item">
            <span>经验消耗</span>
            <strong>{{ echoAnalysis?.exp_consumed ?? 0 }}</strong>
          </div>
        </div>
      </article>

      <article class="panel">
        <div class="panel-head">
          <div>
            <p class="panel-kicker">Tune Baseline</p>
            <h2>偏差高亮</h2>
          </div>
        </div>
        <ul v-if="tuneHighlights.length" class="highlight-list">
          <li v-for="item in tuneHighlights" :key="item.substat" class="highlight-item">
            <div>
              <strong>{{ item.name_cn }}</strong>
              <p>{{ item.comparison?.bias_hint?.message ?? '暂无提示' }}</p>
            </div>
            <span :class="rateClass(item.comparison?.delta_rate ?? 0)">
              {{ formatSignedRate(item.comparison?.delta_rate ?? 0) }}
            </span>
          </li>
        </ul>
        <p v-else class="empty-state">
          当前个人和全站差异不大，或没有传入 `user_id`。
        </p>
      </article>
    </section>

    <section class="panel">
      <div class="panel-head">
        <div>
          <p class="panel-kicker">Tune Table</p>
          <h2>副词条分布明细</h2>
        </div>
      </div>
      <div class="table-wrap">
        <table class="stats-table">
          <thead>
            <tr>
              <th rowspan="2">词条</th>
              <th rowspan="2">档位</th>
              <template v-for="i in 5" :key="`head-${i}`">
                <th colspan="2">孔{{ i }}</th>
              </template>
              <th colspan="2">所有孔位</th>
            </tr>
            <tr>
              <template v-for="i in 5" :key="`subhead-${i}`">
                <th>次数</th>
                <th>占比</th>
              </template>
              <th>总次数</th>
              <th>总占比</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="substat in SUBSTAT" :key="substat.num">
              <tr
                v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
                :key="`${substat.num}-${value.value_number}`"
                :style="substat.font_color ? `color: ${substat.font_color}` : ''"
              >
                <td v-if="value.value_number === 0" :rowspan="SUBSTAT_VALUE_MAP[substat.num].length + 1">
                  <div class="substat-name">
                    <span>{{ substat.name }}</span>
                    <small>{{ formatCompactComparison(substat.num) }}</small>
                  </div>
                </td>
                <td>{{ value.desc }}</td>
                <template v-for="i in 5" :key="`cell-${substat.num}-${value.value_number}-${i}`">
                  <td>{{ tuneStats?.substat_dict?.[substat.num]?.value_dict?.[value.value_number]?.position_dict?.[i - 1]?.total ?? 0 }}</td>
                  <td>{{ formatPercentCell(tuneStats?.substat_dict?.[substat.num]?.value_dict?.[value.value_number]?.position_dict?.[i - 1]?.percent) }}</td>
                </template>
                <td>{{ tuneStats?.substat_dict?.[substat.num]?.value_dict?.[value.value_number]?.total ?? 0 }}</td>
                <td>{{ formatPercentCell(tuneStats?.substat_dict?.[substat.num]?.value_dict?.[value.value_number]?.percent) }}</td>
              </tr>
              <tr class="summary-row" :style="substat.font_color ? `color: ${substat.font_color}` : ''">
                <td>全档位</td>
                <template v-for="i in 5" :key="`all-${substat.num}-${i}`">
                  <td>{{ tuneStats?.substat_dict?.[substat.num]?.value_dict?.all?.position_dict?.[i - 1]?.total ?? 0 }}</td>
                  <td>{{ formatPercentCell(tuneStats?.substat_dict?.[substat.num]?.value_dict?.all?.position_dict?.[i - 1]?.percent) }}</td>
                </template>
                <td>{{ tuneStats?.substat_dict?.[substat.num]?.value_dict?.all?.total ?? 0 }}</td>
                <td>{{ formatPercentCell(tuneStats?.substat_dict?.[substat.num]?.value_dict?.all?.percent) }}</td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { API_BASE_URL, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants'

type ProportionStat = {
  count: number
  total: number
  rate: number
  ci95_low: number
  ci95_high: number
}

type SignificanceSummary = {
  significant: boolean
  sample_enough: boolean
  p_value: number
  z_score: number
  effect_size_pp: number
  direction: string
}

type BiasHintSummary = {
  code: string
  message: string
}

type RateComparison = {
  user: ProportionStat
  global: ProportionStat
  delta_rate: number
  delta_count: number
  significance?: SignificanceSummary
  bias_hint?: BiasHintSummary
}

type TunePositionStat = {
  total: number
  percent: number | string
}

type TuneValueStat = {
  total: number
  percent: number | string
  position_dict: Record<string, TunePositionStat>
}

type TuneSubstatItem = {
  value_dict: Record<string, TuneValueStat>
}

type TuneHighlightItem = {
  substat: string
  name_cn: string
  comparison?: RateComparison
}

type TuneStatsResponse = {
  data_total: number
  window?: string
  substat_dict: Record<string, TuneSubstatItem>
  baseline_compare?: {
    user_sample_size: number
    global_sample_size: number
    substat_rate_delta: Record<string, RateComparison>
    highlights?: TuneHighlightItem[]
  }
}

type EchoDcritResponse = {
  echo_count: number
  dcrit_total: number
  window?: string
  dcrit_rate_stats?: ProportionStat
  baseline_compare?: {
    dcrit_rate?: RateComparison
  }
  counts: Record<string, Record<string, number>>
}

type EchoAnalysisResponse = {
  sample_size: number
  target: number
  target_echo_distance: number
  target_substat_distance: number
  tuner_consumed: number
  exp_consumed: number
  window?: string
  target_rate_stats?: ProportionStat
  baseline_compare?: {
    target_rate?: RateComparison
  }
}

type WindowOption = {
  value: string
  label: string
}

const route = useRoute()
const router = useRouter()

const tuneStats = ref<TuneStatsResponse | null>(null)
const dcritStats = ref<EchoDcritResponse | null>(null)
const echoAnalysis = ref<EchoAnalysisResponse | null>(null)
const errorMessage = ref('')

const userIdInput = ref('')
const windowInput = ref('all')
const targetBitsInput = ref(3)

const windowOptions: WindowOption[] = [
  { value: 'all', label: '全量' },
  { value: 'last_100', label: '最近 100' },
  { value: 'last_500', label: '最近 500' },
  { value: 'last_1000', label: '最近 1000' },
  { value: 'day_7', label: '近 7 天' },
  { value: 'day_30', label: '近 30 天' },
]

const dcritComparison = computed(() => dcritStats.value?.baseline_compare?.dcrit_rate ?? null)
const echoComparison = computed(() => echoAnalysis.value?.baseline_compare?.target_rate ?? null)
const tuneHighlights = computed(() => tuneStats.value?.baseline_compare?.highlights ?? [])

function queryValue(value: unknown): string {
  return Array.isArray(value) ? `${value[0] ?? ''}` : `${value ?? ''}`
}

function syncFiltersFromRoute() {
  userIdInput.value = queryValue(route.query.user_id)
  windowInput.value = queryValue(route.query.window) || 'all'
  const rawTargetBits = Number.parseInt(queryValue(route.query.target_bits) || '3', 10)
  targetBitsInput.value = Number.isFinite(rawTargetBits) ? rawTargetBits : 3
}

async function fetchDashboard() {
  errorMessage.value = ''
  const userId = userIdInput.value.trim()
  const window = windowInput.value || 'all'
  const targetBits = targetBitsInput.value || 3

  const tuneParams = new URLSearchParams()
  const dcritParams = new URLSearchParams()
  const analysisParams = new URLSearchParams()
  if (userId) {
    tuneParams.set('user_id', userId)
    dcritParams.set('user_id', userId)
    analysisParams.set('user_id', userId)
  }
  if (window && window !== 'all') {
    tuneParams.set('window', window)
    dcritParams.set('window', window)
    analysisParams.set('window', window)
  }
  analysisParams.set('target_bits', `${targetBits}`)

  try {
    const [tuneResponse, dcritResponse, analysisResponse] = await Promise.all([
      axios.get(`${API_BASE_URL}/tune_stats?${tuneParams.toString()}`),
      axios.get(`${API_BASE_URL}/counts/echo_dcrit?${dcritParams.toString()}`),
      axios.get(`${API_BASE_URL}/echo_logs/analysis?${analysisParams.toString()}`),
    ])
    tuneStats.value = tuneResponse.data.data
    dcritStats.value = dcritResponse.data.data
    echoAnalysis.value = analysisResponse.data.data
  } catch (error) {
    console.error('请求失败:', error)
    errorMessage.value = '高级统计数据加载失败，请检查后端服务或筛选条件。'
  }
}

async function applyFilters() {
  const query: Record<string, string> = {}
  if (userIdInput.value.trim()) {
    query.user_id = userIdInput.value.trim()
  }
  if (windowInput.value && windowInput.value !== 'all') {
    query.window = windowInput.value
  }
  if (targetBitsInput.value && targetBitsInput.value !== 3) {
    query.target_bits = `${targetBitsInput.value}`
  }
  await router.replace({ query })
}

watch(
  () => route.fullPath,
  async () => {
    syncFiltersFromRoute()
    await fetchDashboard()
  },
  { immediate: true },
)

function formatWindowLabel(window: string | undefined) {
  const matched = windowOptions.find((option) => option.value === (window || 'all'))
  return matched?.label ?? '全量'
}

function formatRate(stat?: ProportionStat | null) {
  if (!stat) return '-'
  return `${stat.rate.toFixed(2)}%`
}

function formatSignedRate(rate: number) {
  if (!Number.isFinite(rate)) return '-'
  return `${rate > 0 ? '+' : ''}${rate.toFixed(2)}pp`
}

function formatConfidence(stat?: ProportionStat | null) {
  if (!stat) return '95% CI -'
  return `95% CI ${stat.ci95_low.toFixed(2)}% ~ ${stat.ci95_high.toFixed(2)}%`
}

function formatPercentCell(value: number | string | undefined) {
  if (value === undefined || value === null || value === '') return '-'
  return `${value}%`
}

function rateClass(rate: number) {
  if (rate > 0) return 'rate-positive'
  if (rate < 0) return 'rate-negative'
  return 'rate-neutral'
}

function formatSignificance(significance?: SignificanceSummary) {
  if (!significance) return '暂无'
  const signal = significance.significant ? '显著' : '未显著'
  const sample = significance.sample_enough ? '样本充足' : '样本偏小'
  return `${signal} · ${sample} · p=${significance.p_value.toFixed(4)} · z=${significance.z_score.toFixed(3)}`
}

function formatCompactComparison(substatNum: number) {
  const comparison = tuneStats.value?.baseline_compare?.substat_rate_delta?.[`${substatNum}`]
  if (!comparison) return ''
  return `${formatSignedRate(comparison.delta_rate)} / ${comparison.bias_hint?.message ?? '暂无提示'}`
}
</script>

<style scoped>
.analysis-page {
  display: grid;
  gap: 20px;
  min-width: 980px;
  color: #182128;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.3fr) minmax(360px, 0.9fr);
  gap: 18px;
  padding: 24px;
  border-radius: 28px;
  background:
    radial-gradient(circle at top left, rgba(255, 222, 173, 0.8), transparent 42%),
    linear-gradient(135deg, #f6efe0 0%, #edf7f3 55%, #e7eff8 100%);
  box-shadow: 0 18px 40px rgba(24, 33, 40, 0.12);
}

.eyebrow {
  margin: 0 0 8px;
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #8f5b2e;
}

h1 {
  margin: 0;
  font-size: 34px;
  line-height: 1.05;
}

.hero-copy {
  max-width: 48ch;
  margin: 12px 0 0;
  color: #4c5a64;
}

.control-panel {
  display: grid;
  gap: 12px;
  padding: 16px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.76);
  border: 1px solid rgba(24, 33, 40, 0.08);
}

.field {
  display: grid;
  gap: 6px;
}

.field span {
  font-size: 12px;
  font-weight: 700;
  color: #46545d;
}

.field input,
.field select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid rgba(24, 33, 40, 0.15);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.96);
  color: #182128;
}

.control-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

button {
  border: 0;
  border-radius: 999px;
  background: linear-gradient(135deg, #1f6d62, #174c7e);
  color: #fff;
  padding: 10px 18px;
  font-weight: 800;
  cursor: pointer;
}

.hint {
  color: #5f6b74;
  font-size: 12px;
}

.error-banner {
  margin: 0;
  padding: 12px 14px;
  border-radius: 16px;
  background: #fff1ef;
  color: #9f362d;
  border: 1px solid rgba(159, 54, 45, 0.18);
}

.summary-grid,
.comparison-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.comparison-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.summary-card,
.panel {
  padding: 18px;
  border-radius: 22px;
  background: #fff;
  border: 1px solid rgba(24, 33, 40, 0.08);
  box-shadow: 0 12px 28px rgba(24, 33, 40, 0.08);
}

.accent-gold {
  background: linear-gradient(135deg, #fff8eb, #fff1d6);
}

.accent-ink {
  background: linear-gradient(135deg, #eef3fb, #e2ebf7);
}

.accent-teal {
  background: linear-gradient(135deg, #ecfbf7, #dbf4ee);
}

.accent-coral {
  background: linear-gradient(135deg, #fff3ef, #ffe2d7);
}

.summary-label,
.panel-kicker {
  margin: 0 0 6px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: #7a5a33;
}

.summary-value {
  margin: 0;
  font-size: 28px;
  font-weight: 800;
}

.summary-meta,
.summary-hint {
  margin: 8px 0 0;
  color: #52616b;
  font-size: 12px;
}

.panel-head {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.panel-head h2 {
  margin: 0;
  font-size: 22px;
}

.tag {
  padding: 6px 10px;
  border-radius: 999px;
  background: #eef3f6;
  color: #4b5b66;
  font-size: 11px;
  font-weight: 700;
}

.compare-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.compare-label {
  display: block;
  margin-bottom: 4px;
  color: #667680;
  font-size: 12px;
}

.insight-list,
.metric-list {
  display: grid;
  gap: 10px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 14px;
  background: #f6f8f9;
}

.highlight-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 10px;
}

.highlight-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px;
  border-radius: 14px;
  background: #f6f8f9;
}

.highlight-item p {
  margin: 4px 0 0;
  color: #5f6d76;
  font-size: 12px;
}

.empty-state {
  margin: 0;
  color: #6a7881;
}

.rate-positive {
  color: #147a52;
}

.rate-negative {
  color: #bf4d32;
}

.rate-neutral {
  color: #63727c;
}

.table-wrap {
  overflow: auto;
}

.stats-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1020px;
}

.stats-table th,
.stats-table td {
  padding: 8px 10px;
  border: 1px solid #e3e8eb;
  text-align: center;
  font-size: 12px;
}

.stats-table thead th {
  position: sticky;
  top: 0;
  background: #f7fafb;
  z-index: 1;
}

.summary-row {
  font-weight: 700;
  background: rgba(17, 58, 87, 0.04);
}

.substat-name {
  display: grid;
  gap: 4px;
}

.substat-name small {
  color: #5e6b75;
  font-size: 10px;
  line-height: 1.3;
}

@media (max-width: 1100px) {
  .analysis-page {
    min-width: 0;
  }

  .hero,
  .summary-grid,
  .comparison-grid {
    grid-template-columns: 1fr;
  }
}
</style>
