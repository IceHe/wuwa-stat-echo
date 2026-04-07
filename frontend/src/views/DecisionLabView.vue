<template>
  <section class="lab-page">
    <header class="hero">
      <div>
        <p class="eyebrow">Decision Lab</p>
        <h1>继续一手，还是立刻止损</h1>
        <p class="hero-copy">
          输入当前声骸状态，系统会结合历史样本分位、下一手命中概率和继续到底的达标率给出建议。
        </p>
      </div>
      <RouterLink class="hero-link" to="/simulator">切到 Simulator</RouterLink>
    </header>

    <div class="lab-layout">
      <form class="panel form-panel" @submit.prevent="runDecision">
        <div class="panel-head">
          <div>
            <p class="panel-kicker">Input</p>
            <h2>当前声骸</h2>
          </div>
          <button type="submit" :disabled="loading">
            {{ loading ? '计算中...' : '生成建议' }}
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
            <span>样本窗口</span>
            <select v-model="form.window">
              <option v-for="item in WINDOW_OPTIONS" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>
          <label class="field">
            <span>目标词条</span>
            <select v-model.number="form.targetBits">
              <option v-for="item in TARGET_PRESETS" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>
        </div>

        <div class="slots">
          <div v-for="(slot, index) in form.slots" :key="`slot-${index}`" class="slot-card">
            <p class="slot-title">当前第 {{ index + 1 }} 词条</p>
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

      <section class="result-column">
        <article class="panel summary-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Recommendation</p>
              <h2>建议结论</h2>
            </div>
            <span class="pill" :class="recommendationClass">{{ recommendationLabel }}</span>
          </div>
          <div v-if="result" class="summary-grid">
            <div class="summary-card">
              <span>当前评分</span>
              <strong>{{ formatScore(result.current_score) }}</strong>
            </div>
            <div class="summary-card">
              <span>同阶段分位</span>
              <strong>{{ formatPercent(result.percentile) }}</strong>
            </div>
            <div class="summary-card">
              <span>继续一手命中率</span>
              <strong>{{ formatPercent(result.continue_to_next_prob) }}</strong>
            </div>
            <div class="summary-card">
              <span>继续到底达标率</span>
              <strong>{{ formatPercent(result.continue_to_finish_prob) }}</strong>
            </div>
          </div>
          <p v-else class="empty-state">填写当前词条后生成建议。</p>
        </article>

        <article class="panel metrics-panel">
          <div class="panel-head">
            <div>
              <p class="panel-kicker">Decision Basis</p>
              <h2>判断依据</h2>
            </div>
          </div>
          <div v-if="result" class="metric-list">
            <div class="metric-item">
              <span>有效词条数</span>
              <strong>{{ formatInteger(result.effective_substat_count) }}</strong>
            </div>
            <div class="metric-item">
              <span>锁定价值</span>
              <strong>{{ formatPercent(result.locked_value) }}</strong>
            </div>
            <div class="metric-item">
              <span>额外调谐器</span>
              <strong>{{ formatInteger(result.expected_extra_tuner) }}</strong>
            </div>
            <div class="metric-item">
              <span>额外经验</span>
              <strong>{{ formatInteger(result.expected_extra_exp) }}</strong>
            </div>
            <div class="metric-item wide">
              <span>目标词条</span>
              <strong>{{ describeTargetBits(form.targetBits) }}</strong>
            </div>
          </div>
          <ul v-if="result?.reasons?.length" class="reason-list">
            <li v-for="reason in result.reasons" :key="reason">{{ reason }}</li>
          </ul>
        </article>
      </section>
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

type DecisionResult = {
  current_score: number
  percentile: number
  effective_substat_count: number
  locked_value: number
  continue_to_next_prob: number
  continue_to_finish_prob: number
  expected_extra_tuner: number
  expected_extra_exp: number
  recommendation: string
  reasons: string[]
}

const form = reactive(createDefaultDecisionForm())
const loading = ref(false)
const errorMessage = ref('')
const result = ref<DecisionResult | null>(null)
const route = useRoute()

const recommendationLabelMap: Record<string, string> = {
  stop: '建议止损',
  continue_once: '建议继续一手',
  continue_to_end: '建议继续到底',
  high_risk: '高风险继续',
}

const recommendationLabel = computed(() =>
  result.value ? recommendationLabelMap[result.value.recommendation] ?? result.value.recommendation : '等待计算',
)

const recommendationClass = computed(() => {
  const code = result.value?.recommendation
  return code ? `pill-${code}` : 'pill-pending'
})

const runDecision = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const payload = {
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
      trials: 3000,
    }
    const response = await axios.post(`${API_BASE_URL}/decision/echo-next-step`, payload)
    result.value = response.data.data
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.message || 'Decision Lab 请求失败'
  } finally {
    loading.value = false
  }
}

watch(
  () => route.query,
  async (query) => {
    const shouldAutorun = applyQueryToDecisionForm(form, query as Record<string, unknown>)
    if (shouldAutorun) {
      await runDecision()
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.lab-page {
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
    radial-gradient(circle at top left, rgba(255, 203, 119, 0.55), transparent 36%),
    linear-gradient(135deg, #14213d 0%, #213555 52%, #274c77 100%);
  color: #fefcf7;
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
  color: rgba(255, 255, 255, 0.82);
  line-height: 1.6;
}

.hero-link {
  align-self: flex-start;
  padding: 12px 16px;
  border-radius: 999px;
  color: #14213d;
  background: #f6bd60;
  font-weight: 700;
  text-decoration: none;
}

.lab-layout {
  display: grid;
  grid-template-columns: minmax(320px, 1.15fr) minmax(320px, 1fr);
  gap: 20px;
}

.panel {
  padding: 22px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(20, 33, 61, 0.08);
  box-shadow: 0 18px 36px rgba(31, 41, 55, 0.08);
}

.form-panel {
  display: grid;
  gap: 18px;
}

.panel-head {
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
  color: #8d99ae;
}

.panel-head h2 {
  margin: 0;
  font-size: 24px;
}

button {
  border: 0;
  border-radius: 999px;
  padding: 12px 18px;
  background: #14213d;
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

.field span,
.slot-title {
  font-size: 13px;
  font-weight: 700;
  color: #415a77;
}

input,
select {
  width: 100%;
  border: 1px solid rgba(20, 33, 61, 0.15);
  border-radius: 14px;
  padding: 12px 14px;
  background: #fffdf8;
  color: #132238;
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
  background: linear-gradient(180deg, rgba(246, 189, 96, 0.14), rgba(255, 255, 255, 0.92));
}

.result-column {
  display: grid;
  gap: 20px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.summary-card,
.metric-item {
  display: grid;
  gap: 6px;
  padding: 16px;
  border-radius: 18px;
  background: #f7f8fa;
}

.summary-card span,
.metric-item span {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #7a8aa0;
}

.summary-card strong,
.metric-item strong {
  font-size: 24px;
  color: #132238;
}

.metric-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.wide {
  grid-column: 1 / -1;
}

.reason-list {
  margin: 14px 0 0;
  padding-left: 18px;
  color: #324154;
  line-height: 1.7;
}

.pill {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 700;
}

.pill-stop {
  background: #ffe0dc;
  color: #9f2a1d;
}

.pill-continue_once {
  background: #ffe7b8;
  color: #8a5600;
}

.pill-continue_to_end {
  background: #d8f3dc;
  color: #1b6b39;
}

.pill-high_risk,
.pill-pending {
  background: #e6eef8;
  color: #36506c;
}

.empty-state,
.error-banner {
  margin: 0;
  padding: 14px 16px;
  border-radius: 16px;
}

.empty-state {
  background: #f4f5f7;
  color: #5b687a;
}

.error-banner {
  background: #ffe0dc;
  color: #912f24;
}

@media (max-width: 980px) {
  .lab-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .hero {
    padding: 22px;
  }

  .hero,
  .panel-head {
    flex-direction: column;
  }

  .field-grid,
  .slots,
  .summary-grid,
  .metric-list {
    grid-template-columns: 1fr;
  }
}
</style>
