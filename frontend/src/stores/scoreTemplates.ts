import axios from 'axios'
import { reactive } from 'vue'

import { API_BASE_URL } from '@/stores/constants.ts'
import type { ResonatorTemplate } from '@/utils/echoScore'

const storageKey = 'wuwa-echo-score-template-config'
const contextKey = 'wuwa-echo-score-template-context'
const builtinVersion = 'builtin-2026-04-09'

const buildTemplate = (
  name: string,
  echoMax: Record<string, number>,
  mainstat: Record<string, number>,
  weights: Record<string, number>,
): ResonatorTemplate => {
  const baseWeights: Record<string, number> = {
    暴击: 2.0,
    暴击伤害: 1.0,
    攻击: 1.1,
    攻击固定值: 0.1,
  }
  Object.entries(weights).forEach(([key, value]) => {
    baseWeights[key] = value
  })
  return {
    name,
    echo_max_score: echoMax,
    mainstat_max_score: mainstat,
    substat_weight: baseWeights,
  }
}

const builtinTemplates: Record<string, ResonatorTemplate> = {
  '': buildTemplate('通用', { 4: 80, 3: 80, 1: 80 }, { '4C': 8.86, '3C属伤': 6.78, '3C攻击': 6.73, '3C其它': 1.57, '1C': 4.76 }, { 共鸣效率: 0.3, 普攻: 0.05, 重击: 0.05, 共鸣技能: 0.05, 共鸣解放: 0.05 }),
  通用: buildTemplate('通用', { 4: 80, 3: 80, 1: 80 }, { '4C': 8.86, '3C属伤': 6.78, '3C攻击': 6.73, '3C其它': 1.57, '1C': 4.76 }, { 共鸣效率: 0.3, 普攻: 0.05, 重击: 0.05, 共鸣技能: 0.05, 共鸣解放: 0.05 }),
  暗主: buildTemplate('暗主', { 4: 82.527, 3: 78.527, 1: 74.977 }, { '4C': 8.93, '3C属伤': 6.84, '3C攻击': 6.84, '3C其它': 1.59, '1C': 4.8 }, { 共鸣效率: 0.5, 普攻: 0.275, 共鸣技能: 0.22, 共鸣解放: 0.605 }),
  椿: buildTemplate('椿', { 4: 83.8, 3: 79.8, 1: 76.25 }, { '4C': 8.79, '3C属伤': 6.8, '3C攻击': 6.8, '3C其它': 1.6, '1C': 4.72 }, { 共鸣效率: 0.15, 普攻: 0.715, 共鸣解放: 0.275 }),
  珂莱塔: buildTemplate('珂莱塔', { 4: 86.066, 3: 82.066, 1: 78.516 }, { '4C': 8.56, '3C属伤': 6.54, '3C攻击': 6.54, '3C其它': 1.52, '1C': 4.58 }, { 共鸣效率: 0.2, 共鸣技能: 0.91 }),
  今汐: buildTemplate('今汐', { 4: 83.8, 3: 79.8, 1: 76.25 }, { '4C': 8.79, '3C属伤': 6.72, '3C攻击': 6.72, '3C其它': 1.56, '1C': 4.72 }, { 共鸣效率: 0.25, 共鸣技能: 0.715, 共鸣解放: 0.33 }),
  长离: buildTemplate('长离', { 4: 83.17, 3: 79.17, 1: 75.62 }, { '4C': 8.86, '3C属伤': 6.78, '3C攻击': 6.78, '3C其它': 1.57, '1C': 4.76 }, { 共鸣效率: 0.3, 共鸣技能: 0.66, 共鸣解放: 0.44 }),
  坎特蕾拉: buildTemplate('坎特蕾拉', { 4: 83.17, 3: 79.17, 1: 75.62 }, { '4C': 8.86, '3C属伤': 6.78, '3C攻击': 6.78, '3C其它': 1.57, '1C': 4.76 }, { 共鸣效率: 0.5, 普攻: 0.66 }),
  折枝: buildTemplate('折枝', { 4: 81.89, 3: 77.89, 1: 74.34 }, { '4C': 8.99, '3C属伤': 6.89, '3C攻击': 6.89, '3C其它': 1.6, '1C': 4.84 }, { 共鸣效率: 0.2, 普攻: 0.55, 重击: 0.22, 共鸣技能: 0.22 }),
  忌炎: buildTemplate('忌炎', { 4: 83.8, 3: 79.8, 1: 76.25 }, { '4C': 8.79, '3C属伤': 6.72, '3C攻击': 6.72, '3C其它': 1.56, '1C': 4.72 }, { 共鸣效率: 0.3, 普攻: 0.165, 重击: 0.715, 共鸣技能: 0.33 }),
  相里要: buildTemplate('相里要', { 4: 83.8, 3: 79.8, 1: 76.25 }, { '4C': 8.79, '3C属伤': 6.72, '3C攻击': 6.72, '3C其它': 1.56, '1C': 4.72 }, { 共鸣效率: 0.3, 普攻: 0.165, 共鸣技能: 0.22, 共鸣解放: 0.715 }),
  洛可可: buildTemplate('洛可可', { 4: 85.25, 3: 81.25, 1: 77.7 }, { '4C': 8.64, '3C属伤': 6.6, '3C攻击': 6.6, '3C其它': 1.53, '1C': 4.63 }, { 共鸣效率: 0.3, 重击: 0.84 }),
  布兰特: buildTemplate('布兰特', { 4: 77.33, 3: 74.03, 1: 71.88 }, { '4C': 8.17, '3C属伤': 6.31, '3C攻击': 6.31, '3C其它': 6.31, '1C': 5 }, { 攻击: 0.44, 攻击固定值: 0.044, 共鸣效率: 0.8, 普攻: 0.66, 共鸣解放: 0.165 }),
  菲比: buildTemplate('菲比', { 4: 78.76, 3: 74.76, 1: 71.21 }, { '4C': 9.36, '3C属伤': 7.18, '3C攻击': 6.78, '3C其它': 1.57, '1C': 5.05 }, { 暴击: 1.58, 共鸣效率: 0.1, 普攻: 0.088, 重击: 0.66, 共鸣技能: 0.055, 共鸣解放: 0.187 }),
  赞妮: buildTemplate('赞妮', { 4: 83.8, 3: 79.8, 1: 76.25 }, { '4C': 8.79, '3C属伤': 6.72, '3C攻击': 6.72, '3C其它': 1.56, '1C': 4.72 }, { 共鸣效率: 0.3, 重击: 0.715, 共鸣解放: 0.154 }),
  夏空: buildTemplate('夏空', { 4: 82.78, 3: 78.78, 1: 75.23 }, { '4C': 8.9, '3C属伤': 6.81, '3C攻击': 6.81, '3C其它': 1.58, '1C': 4.78 }, { 共鸣效率: 0.3, 普攻: 0.506, 重击: 0.363, 共鸣解放: 0.627 }),
  卡提希娅: buildTemplate('卡提希娅', { 4: 79.726, 3: 76.871, 1: 78.986 }, { '4C': 6.89, '3C属伤': 5.46, '3C攻击': 5.46, '3C其它': 0, '1C': 6.48 }, { 攻击: 0, 攻击固定值: 0, 生命: 1.1, 生命固定值: 0.01, 共鸣效率: 0.1, 普攻: 0.704, 共鸣解放: 0.308 }),
  露帕: buildTemplate('露帕', { 4: 84.059, 3: 80.059, 1: 76.509 }, { '4C': 8.77, '3C属伤': 6.71, '3C攻击': 6.71, '3C其它': 1.56, '1C': 4.7 }, { 共鸣效率: 0.2, 普攻: 0.077, 重击: 0.055, 共鸣技能: 0.231, 共鸣解放: 0.737 }),
  弗洛洛: buildTemplate('弗洛洛', { 4: 84.059, 3: 80.059, 1: 76.509 }, { '4C': 8.77, '3C属伤': 6.71, '3C攻击': 6.71, '3C其它': 1.56, '1C': 4.7 }, { 共鸣技能: 0.737 }),
  奥古斯塔: buildTemplate('奥古斯塔', { 4: 85.161, 3: 81.161, 1: 77.611 }, { '4C': 8.65, '3C属伤': 6.62, '3C攻击': 6.62, '3C其它': 1.54, '1C': 4.63 }, { 重击: 0.832, 共鸣效率: 0.2 }),
  尤诺: buildTemplate('尤诺', { 4: 83.804, 3: 79.804, 1: 76.254 }, { '4C': 8.79, '3C属伤': 6.72, '3C攻击': 6.72, '3C其它': 1.56, '1C': 4.72 }, { 共鸣效率: 0.2, 共鸣解放: 0.715 }),
  嘉贝莉娜: buildTemplate('嘉贝莉娜', { 4: 80.358, 3: 76.358, 1: 72.808 }, { '4C': 8.79, '3C属伤': 7.03, '3C攻击': 7.03, '3C其它': 1.63, '1C': 4.94 }, { 共鸣效率: 0.2, 重击: 0.418 }),
  仇远: buildTemplate('仇远', { 4: 82.017, 3: 78.017, 1: 74.467 }, { '4C': 8.98, '3C属伤': 6.88, '3C攻击': 6.88, '3C其它': 1.6, '1C': 4.83 }, { 重击: 0.561, 共鸣效率: 0.2 }),
  琳奈: buildTemplate('琳奈', { 4: 84.117, 3: 80.117, 1: 76.567 }, { '4C': 8.75, '3C属伤': 6.7, '3C攻击': 6.7, '3C其它': 1.56, '1C': 4.7 }, { 攻击: 1.05, 普攻: 0.792, 共鸣解放: 0.253, 共鸣效率: 0.2 }),
  莫宁: buildTemplate('莫宁', { 4: 64.515, 3: 64.099, 1: 63.698 }, { '4C': 9.71, '3C属伤': 5.85, '3C攻击': 0, '3C其它': 8.73, '1C': 8.47 }, { 防御: 1.25, 防御固定值: 0.1, 共鸣效率: 1.3, 暴击: 0.1, 暴击伤害: 0.3, 共鸣解放: 0.44 }),
  陆赫斯: buildTemplate('陆赫斯', { 4: 85.915, 3: 81.915, 1: 78.365 }, { '4C': 8.58, '3C属伤': 6.55, '3C攻击': 6.55, '3C其它': 1.52, '1C': 4.59 }, { 攻击: 1.15, 普攻: 0.847, 共鸣效率: 0.15 }),
  爱弥斯: buildTemplate('爱弥斯', { 4: 85.642, 3: 81.642, 1: 78.092 }, { '4C': 8.6, '3C属伤': 6.58, '3C攻击': 6.58, '3C其它': 1.53, '1C': 4.6 }, { 攻击固定值: 0.12, 共鸣解放: 0.77, 共鸣效率: 0.2 }),
  西格莉卡: buildTemplate('通用', { 4: 80, 3: 80, 1: 80 }, { '4C': 8.86, '3C属伤': 6.78, '3C攻击': 6.73, '3C其它': 1.57, '1C': 4.76 }, { 共鸣效率: 0.3, 普攻: 0.05, 重击: 0.05, 共鸣技能: 0.05, 共鸣解放: 0.05 }),
}

type RemoteTemplatePayload = {
  version?: string
  updated_at?: string
  resonator_templates?: Record<string, ResonatorTemplate>
}

const readStorage = () => {
  if (typeof window === 'undefined') {
    return null
  }
  const raw = window.localStorage.getItem(storageKey)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as RemoteTemplatePayload
  } catch {
    return null
  }
}

const writeStorage = (payload: RemoteTemplatePayload) => {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.setItem(storageKey, JSON.stringify(payload))
}

const readContext = () => {
  if (typeof window === 'undefined') {
    return null
  }
  const raw = window.localStorage.getItem(contextKey)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as { resonator?: string; cost?: string }
  } catch {
    return null
  }
}

const writeContext = (payload: { resonator: string; cost: string }) => {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.setItem(contextKey, JSON.stringify(payload))
}

export const scoreTemplateState = reactive({
  templates: builtinTemplates as Record<string, ResonatorTemplate>,
  version: builtinVersion,
  updatedAt: '',
  source: 'builtin',
  loading: false,
  initialized: false,
  error: '',
})

export const scoreTemplateContext = reactive({
  resonator: '',
  cost: '',
})

const applyTemplatePayload = (payload: RemoteTemplatePayload | null, source: string) => {
  const templates = payload?.resonator_templates
  if (!templates || Object.keys(templates).length === 0) {
    return false
  }
  scoreTemplateState.templates = templates
  scoreTemplateState.version = payload?.version || builtinVersion
  scoreTemplateState.updatedAt = payload?.updated_at || ''
  scoreTemplateState.source = source
  scoreTemplateState.error = ''
  scoreTemplateState.initialized = true
  return true
}

const hydrate = () => {
  const stored = readStorage()
  if (!applyTemplatePayload(stored, 'cache')) {
    scoreTemplateState.initialized = true
  }
  const storedContext = readContext()
  if (storedContext) {
    scoreTemplateContext.resonator = storedContext.resonator || ''
    scoreTemplateContext.cost = storedContext.cost || ''
  }
}

hydrate()

export const setScoreTemplateContext = (payload: { resonator?: string; cost?: string }) => {
  const next = {
    resonator: typeof payload.resonator === 'string' ? payload.resonator : scoreTemplateContext.resonator,
    cost: typeof payload.cost === 'string' ? payload.cost : scoreTemplateContext.cost,
  }
  scoreTemplateContext.resonator = next.resonator
  scoreTemplateContext.cost = next.cost
  writeContext(next)
}

export const getResonatorTemplate = (resonator: string) =>
  scoreTemplateState.templates[resonator] ||
  scoreTemplateState.templates['通用'] ||
  scoreTemplateState.templates[''] ||
  null

export const refreshScoreTemplates = async (force = false) => {
  if (scoreTemplateState.loading) {
    return
  }
  scoreTemplateState.loading = true
  scoreTemplateState.error = ''
  try {
    const response = await axios.get(`${API_BASE_URL}/score_templates${force ? '?force=1' : ''}`)
    const payload = response.data?.data as RemoteTemplatePayload | undefined
    if (!applyTemplatePayload(payload || null, 'remote')) {
      throw new Error('invalid score template payload')
    }
    writeStorage(payload || {})
  } catch (error) {
    scoreTemplateState.error = error instanceof Error ? error.message : 'failed to load score templates'
  } finally {
    scoreTemplateState.loading = false
    scoreTemplateState.initialized = true
  }
}

export const ensureScoreTemplatesLoaded = async () => {
  if (scoreTemplateState.loading) {
    return
  }
  await refreshScoreTemplates(false)
}
