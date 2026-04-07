<template>
  <section class="sim-page">
    <header class="hero">
      <div>
        <p class="eyebrow">Simulator</p>
        <h1>把未来路径摊开来看</h1>
        <p class="hero-copy">
          对比立刻止损、继续一手、继续到底三种策略，直接看达标率、资源成本和常见结局分布。
        </p>
      </div>
      <RouterLink class="hero-link" to="/decision-lab">切到 Decision Lab</RouterLink>
    </header>

    <div class="sim-layout">
      <form class="panel form-panel" @submit.prevent="runSimulator">
        <div class="panel-head">
          <div>
            <p class="panel-kicker">Simulation Input</p>
            <h2>模拟参数</h2>
          </div>
          <button type="submit" :disabled="loading">
            {{ loading ? '模拟中...' : '开始模拟' }}
          </button>
        </div>

        <div class="field-grid">
          <label class="field">
            <span>用户 ID</span>
            <input v-model.trim="form.userId" inputmode="numeric" placeholder="全站留空" />
          </label>
          <label class="field">
            <span>共鸣者</span>
            <select v-model="form.resonator">
              <option v-for="item in RESONATOR_OPTIONS" :key="item" :value="item">{{ item }}</option>
            </select>
          </label>
          <label class="field">
            <span>Cost</span>
            <select v-model="form.cost">
              <option v-for="item in COST_OPTIONS" :key="item" :value="item">{{ item }}</option>
            </select>
          </label>
          <label class="field">
            <span>目标档位</span>
            <select v-model="form.goal">
              <option v-for="item in GOAL_OPTIONS" :key="item" :value="item">{{ item }}</option>
            </select>
          </label>
          <label class="field">
            <span>窗口</span>
            <select v-model="form.window">
              <option v-for="item in WINDOW_OPTIONS" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>
          <label class="field">
            <span>模拟次数</span>
            <input v-model.number="form.trials" inputmode="numeric" min="100" max="50000" step="100" />
          </label>
          <label class="field field-wide">
            <span>目标词条</span>
            <select v-model.number="form.targetBits">
              <option v-for="item in TARGET_PRESETS" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>
        </div>

        <div class="slots">
          <div v-for="(slot, index) in form.slots" :key="`slot-${index}`" class="slot-card">
            <p class="slot-title">第 {{ index + 1 }} 词条</p>
            <select v-model="slot.substat">
              <option value="">未开启</option>
              <option v-for="item in substatOptions" :key="`${index}-${item.value}`" :value="item.value">
                {{ item.label }}
              </option>
            </select>
            <select v-model="slot.tier" :disabled="!slot.substat">
              <option v-for="item in tierOptionsForSubstat(slot.substat)" :key="`${index}-${item.value}`" :value="item.value">
                {{ item.label }}
              </option>
            </select>
          </div>
        </div>

        <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>
      </form>

      <div class="result-stack">
        <article class="panel overview-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Scenario</p>
              <h2>当前模拟设定</h2>
            </div>
          </div>
          <div class="overview-grid">
            <div class="overview-item">
              <span>目标词条</span>
              <strong>{{ describeTargetBits(form.targetBits) }}</strong>
            </div>
            <div class="overview-item">
              <span>模拟次数</span>
              <strong>{{ formatInteger(form.trials) }}</strong>
            </div>
          </div>
        </article>

        <article class="panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Strategy Compare</p>
              <h2>策略对比</h2>
            </div>
          </div>
          <div v-if="strategies.length" class="strategy-grid">
            <article v-for="item in strategies" :key="item.strategy" class="strategy-card">
              <div class="strategy-head">
                <h3>{{ strategyLabel(item.strategy) }}</h3>
                <span class="strategy-pill">{{ item.summary.trials }} 次</span>
              </div>
              <div class="metric-grid">
                <div class="metric-cell">
                  <span>达标率</span>
                  <strong>{{ formatPercent(item.summary.hit_prob) }}</strong>
                </div>
                <div class="metric-cell">
                  <span>高光率</span>
                  <strong>{{ formatPercent(item.summary.high_roll_prob) }}</strong>
                </div>
                <div class="metric-cell">
                  <span>期望评分</span>
                  <strong>{{ formatScore(item.summary.expected_score) }}</strong>
                </div>
                <div class="metric-cell">
                  <span>额外调谐器</span>
                  <strong>{{ formatInteger(item.summary.expected_tuner_cost) }}</strong>
                </div>
              </div>
              <ul class="bucket-list">
                <li v-for="bucket in item.summary.result_buckets" :key="`${item.strategy}-${bucket.label}`">
                  <span>{{ bucket.label }}</span>
                  <strong>{{ formatPercent(bucket.rate) }}</strong>
                </li>
              </ul>
            </article>
          </div>
          <p v-else class="empty-state">提交表单后显示三种策略对比。</p>
        </article>

        <article v-if="futureSummary" class="panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Continue To End</p>
              <h2>继续到底摘要</h2>
            </div>
          </div>
          <div class="metric-grid metric-grid-wide">
            <div class="metric-cell">
              <span>达标率</span>
              <strong>{{ formatPercent(futureSummary.hit_prob) }}</strong>
            </div>
            <div class="metric-cell">
              <span>高光率</span>
              <strong>{{ formatPercent(futureSummary.high_roll_prob) }}</strong>
            </div>
            <div class="metric-cell">
              <span>期望评分</span>
              <strong>{{ formatScore(futureSummary.expected_score) }}</strong>
            </div>
            <div class="metric-cell">
              <span>额外经验</span>
              <strong>{{ formatInteger(futureSummary.expected_exp_cost) }}</strong>
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { API_BASE_URL } from '@/stores/constants'
import {
  COST_OPTIONS,
  GOAL_OPTIONS,
  RESONATOR_OPTIONS,
  TARGET_PRESETS,
  WINDOW_OPTIONS,
  applyQueryToDecisionForm,
  buildEchoPayload,
  createDefaultDecisionForm,
  describeTargetBits,
  formatInteger,
  formatPercent,
  formatScore,
  substatOptions,
  tierOptionsForSubstat,
} from './decisionSupport'

type Bucket = {
  label: string
  rate: number
}

type Summary = {
  trials: number
  hit_prob: number
  high_roll_prob: number
  expected_score: number
  expected_tuner_cost: number
  expected_exp_cost: number
  result_buckets: Bucket[]
}

type StrategyItem = {
  strategy: string
  summary: Summary
}

const form = reactive(createDefaultDecisionForm())
form.trials = 8000

const loading = ref(false)
const errorMessage = ref('')
const strategies = ref<StrategyItem[]>([])
const futureSummary = ref<Summary | null>(null)
const route = useRoute()

const strategyLabel = (value: string) => {
  switch (value) {
    case 'stop_now':
      return '立刻止损'
    case 'continue_once':
      return '继续一手'
    case 'continue_to_end':
      return '继续到底'
    default:
      return value
  }
}

const payload = computed(() => ({
  echo: {
    ...buildEchoPayload(form.slots),
    user_id: Number(form.userId || 0),
  },
  user_id: Number(form.userId || 0),
  resonator: form.resonator,
  cost: form.cost,
  goal: form.goal,
  target_bits: form.targetBits,
  window: form.window,
  trials: Number(form.trials || 0),
}))

const runSimulator = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const [compareResp, futureResp] = await Promise.all([
      axios.post(`${API_BASE_URL}/simulator/echo-compare`, payload.value),
      axios.post(`${API_BASE_URL}/simulator/echo-future`, payload.value),
    ])
    strategies.value = compareResp.data.data?.strategies ?? []
    futureSummary.value = futureResp.data.data ?? null
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.message || 'Simulator 请求失败'
  } finally {
    loading.value = false
  }
}

watch(
  () => route.query,
  async (query) => {
    const shouldAutorun = applyQueryToDecisionForm(form, query as Record<string, unknown>)
    if (shouldAutorun) {
      await runSimulator()
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.sim-page {
  display: grid;
  gap: 20px;
}

.hero {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 28px;
  border-radius: 28px;
  background:
    radial-gradient(circle at top right, rgba(113, 201, 206, 0.42), transparent 30%),
    linear-gradient(135deg, #2d3047 0%, #1b4965 44%, #0f6e7f 100%);
  color: #fbfbfa;
}

.eyebrow {
  margin: 0 0 10px;
  font-size: 12px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.72);
}

h1 {
  margin: 0;
  font-size: clamp(28px, 4vw, 42px);
}

.hero-copy {
  max-width: 720px;
  margin: 12px 0 0;
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.6;
}

.hero-link {
  align-self: flex-start;
  padding: 12px 16px;
  border-radius: 999px;
  color: #10293b;
  background: #c3f0ca;
  font-weight: 700;
  text-decoration: none;
}

.sim-layout {
  display: grid;
  grid-template-columns: minmax(320px, 0.95fr) minmax(320px, 1.2fr);
  gap: 20px;
}

.panel {
  padding: 22px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(27, 73, 101, 0.09);
  box-shadow: 0 18px 36px rgba(20, 33, 61, 0.08);
}

.form-panel,
.result-stack {
  display: grid;
  gap: 18px;
}

.panel-head,
.strategy-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel-kicker {
  margin: 0 0 6px;
  font-size: 12px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: #7d8da1;
}

.panel-head h2,
.strategy-head h3 {
  margin: 0;
}

button {
  border: 0;
  border-radius: 999px;
  padding: 12px 18px;
  background: #0f4c5c;
  color: #fff;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.7;
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.field {
  display: grid;
  gap: 8px;
}

.field-wide {
  grid-column: 1 / -1;
}

.field span,
.slot-title {
  font-size: 13px;
  font-weight: 700;
  color: #36506c;
}

input,
select {
  width: 100%;
  border: 1px solid rgba(15, 76, 92, 0.16);
  border-radius: 14px;
  padding: 12px 14px;
  background: #fcfdfd;
  color: #12263a;
}

.slots {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.slot-card {
  display: grid;
  gap: 10px;
  padding: 16px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(195, 240, 202, 0.4), rgba(255, 255, 255, 0.96));
}

.overview-grid,
.metric-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.overview-item,
.metric-cell {
  display: grid;
  gap: 6px;
  padding: 16px;
  border-radius: 18px;
  background: #f5f8fa;
}

.overview-item span,
.metric-cell span {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #6c7f92;
}

.overview-item strong,
.metric-cell strong {
  font-size: 24px;
  color: #12263a;
}

.strategy-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.strategy-card {
  display: grid;
  gap: 14px;
  padding: 16px;
  border-radius: 20px;
  background: linear-gradient(180deg, rgba(15, 76, 92, 0.06), rgba(255, 255, 255, 0.98));
}

.strategy-pill {
  display: inline-flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  color: #0f4c5c;
  background: rgba(15, 76, 92, 0.1);
}

.bucket-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 8px;
}

.bucket-list li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 14px;
  color: #314355;
}

.metric-grid-wide {
  margin-top: 8px;
}

.empty-state,
.error-banner {
  margin: 0;
  padding: 14px 16px;
  border-radius: 16px;
}

.empty-state {
  background: #f4f6f8;
  color: #546476;
}

.error-banner {
  background: #ffe0dc;
  color: #8f2d23;
}

@media (max-width: 1100px) {
  .sim-layout,
  .strategy-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .hero,
  .panel-head,
  .strategy-head {
    flex-direction: column;
  }

  .field-grid,
  .slots,
  .overview-grid,
  .metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
