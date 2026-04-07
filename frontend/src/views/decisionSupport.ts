import { ECHO_COST, RESONATORS, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants'

export const WINDOW_OPTIONS = [
  { value: 'all', label: '全量样本' },
  { value: 'last_100', label: '最近 100' },
  { value: 'last_500', label: '最近 500' },
  { value: 'last_1000', label: '最近 1000' },
  { value: 'day_7', label: '最近 7 天' },
  { value: 'day_30', label: '最近 30 天' },
]

export const GOAL_OPTIONS = ['保底', '小毕业', '毕业', '神品']
export const COST_OPTIONS = ECHO_COST
export const RESONATOR_OPTIONS = RESONATORS

export const TARGET_PRESETS = [
  { value: 3, label: '双暴' },
  { value: 7, label: '双暴 + 攻击' },
  { value: 259, label: '双暴 + 共效' },
  { value: 2051, label: '双暴 + 共解' },
  { value: 11, label: '双暴 + 重击' },
  { value: 515, label: '双暴 + 普攻' },
]

export type EchoSlotForm = {
  substat: string
  tier: string
}

export type DecisionFormState = {
  userId: string
  resonator: string
  cost: string
  goal: string
  trials: number
  window: string
  targetBits: number
  slots: EchoSlotForm[]
}

export const createDefaultDecisionForm = (): DecisionFormState => ({
  userId: '',
  resonator: '弗洛洛',
  cost: '4C',
  goal: '毕业',
  trials: 5000,
  window: 'all',
  targetBits: 3,
  slots: Array.from({ length: 5 }, () => ({
    substat: '',
    tier: '0',
  })),
})

export const buildEchoPayload = (slots: EchoSlotForm[]) => {
  const payload: Record<string, number> = {
    substat1: 0,
    substat2: 0,
    substat3: 0,
    substat4: 0,
    substat5: 0,
    substat_all: 0,
  }
  slots.forEach((slot, index) => {
    const encoded = encodeSubstat(slot.substat, slot.tier)
    payload[`substat${index + 1}`] = encoded
    payload.substat_all |= encoded & ((1 << 13) - 1)
  })
  return payload
}

export const decodeSubstat = (encoded: number): EchoSlotForm => {
  if (!encoded) {
    return { substat: '', tier: '0' }
  }
  const mask = encoded & ((1 << 13) - 1)
  let substatNum = -1
  for (let i = 0; i < 13; i++) {
    if (((mask >> i) & 1) === 1) {
      substatNum = i
      break
    }
  }
  let tierNum = 0
  const tierBits = encoded >> 13
  for (let i = 0; i < 8; i++) {
    if (((tierBits >> i) & 1) === 1) {
      tierNum = i
      break
    }
  }
  return {
    substat: substatNum >= 0 ? String(substatNum) : '',
    tier: String(tierNum),
  }
}

export const applyQueryToDecisionForm = (
  form: DecisionFormState,
  query: Record<string, unknown>,
) => {
  const read = (key: string) => {
    const value = query[key]
    return Array.isArray(value) ? String(value[0] ?? '') : String(value ?? '')
  }

  const userId = read('user_id')
  const resonator = read('resonator')
  const cost = read('cost')
  const goal = read('goal')
  const window = read('window')
  const targetBits = read('target_bits')
  const trials = read('trials')

  if (userId) form.userId = userId
  if (resonator) form.resonator = resonator
  if (cost) form.cost = cost
  if (goal) form.goal = goal
  if (window) form.window = window
  if (targetBits) form.targetBits = Number(targetBits)
  if (trials) form.trials = Number(trials)

  form.slots.forEach((slot, index) => {
    const raw = Number(read(`s${index + 1}`) || 0)
    const decoded = decodeSubstat(raw)
    slot.substat = decoded.substat
    slot.tier = decoded.tier
  })

  return read('autorun') === '1'
}

export const buildDecisionQueryFromEncoded = (input: {
  userId?: number | string
  resonator?: string
  cost?: string
  goal?: string
  window?: string
  targetBits?: number
  trials?: number
  substats: number[]
  autorun?: boolean
}) => {
  const query: Record<string, string> = {}
  if (input.userId) query.user_id = String(input.userId)
  if (input.resonator) query.resonator = input.resonator
  if (input.cost) query.cost = input.cost
  if (input.goal) query.goal = input.goal
  if (input.window) query.window = input.window
  if (input.targetBits) query.target_bits = String(input.targetBits)
  if (input.trials) query.trials = String(input.trials)
  input.substats.forEach((value, index) => {
    if (value) {
      query[`s${index + 1}`] = String(value)
    }
  })
  if (input.autorun) query.autorun = '1'
  return query
}

export const encodeSubstat = (substatValue: string, tierValue: string) => {
  if (!substatValue) {
    return 0
  }
  const substatNum = Number(substatValue)
  const tierNum = Number(tierValue || 0)
  if (!Number.isFinite(substatNum) || substatNum < 0) {
    return 0
  }
  return (1 << substatNum) | ((1 << tierNum) << 13)
}

export const formatPercent = (value: number | null | undefined, digits = 2) => {
  if (value == null || Number.isNaN(value)) {
    return '--'
  }
  return `${(value * 100).toFixed(digits)}%`
}

export const formatScore = (value: number | null | undefined) => {
  if (value == null || Number.isNaN(value)) {
    return '--'
  }
  return value.toFixed(2)
}

export const formatInteger = (value: number | null | undefined) => {
  if (value == null || Number.isNaN(value)) {
    return '--'
  }
  return Math.round(value).toString()
}

export const describeTargetBits = (bits: number) => {
  const names = SUBSTAT.filter((item) => (bits & (1 << item.num)) !== 0).map((item) => item.name)
  return names.length ? names.join(' / ') : '自动'
}

export const substatOptions = SUBSTAT.map((item) => ({
  value: String(item.num),
  label: item.name,
}))

export const tierOptionsForSubstat = (substatValue: string) => {
  const substatNum = Number(substatValue)
  if (!Number.isFinite(substatNum) || !(substatNum in SUBSTAT_VALUE_MAP)) {
    return []
  }
  return SUBSTAT_VALUE_MAP[substatNum].map((item) => ({
    value: String(item.value_number),
    label: item.desc_full,
  }))
}
