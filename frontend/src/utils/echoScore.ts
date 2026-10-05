import { getMainstatOptions, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants.ts'

const MASK = 0b1111111111111
const SUBSTAT_BIT_WIDTH = 13

export type ResonatorTemplate = {
  id?: string
  variant?: string
  name?: string
  source_name?: string
  echo_max_score?: Record<string, number>
  mainstat_max_score?: Record<string, number>
  substat_weight?: Record<string, number>
  main_props?: Record<string, Record<string, number>>
  skill_weight?: number[]
  props_grade?: number[][]
  total_grade?: number[]
}

export type EchoSubstatCarrier = {
  substat1?: number
  substat2?: number
  substat3?: number
  substat4?: number
  substat5?: number
}

const bitPos = (value: number) => {
  if (!value) {
    return -1
  }
  let pos = 0
  while (((value >> pos) & 1) === 0) {
    pos += 1
  }
  return pos
}

const getSelectedSubstats = (echoLog: EchoSubstatCarrier) =>
  [
    Number(echoLog.substat1 || 0),
    Number(echoLog.substat2 || 0),
    Number(echoLog.substat3 || 0),
    Number(echoLog.substat4 || 0),
    Number(echoLog.substat5 || 0),
  ].filter((substat) => substat > 0)

const getSubstatFullName = (substatNum: number) =>
  SUBSTAT_VALUE_MAP[substatNum]?.[0]?.desc_full?.replace(/\s.*$/, '') ?? ''

const skillWeightIndex: Record<string, number> = {
  普攻: 0,
  重击: 1,
  共鸣技能: 2,
  共鸣解放: 3,
}

const truncateScore = (value: number) => Math.trunc((value + 1e-12) * 100) / 100

const getXwuidPropertyName = (name: string) =>
  name.endsWith('伤害加成') ? '属性伤害加成' : name

const getSubstatNumericValue = (substatNum: number, valueNum: number) => {
  const desc = SUBSTAT_VALUE_MAP[substatNum]?.[valueNum]?.desc ?? '0'
  return Number.parseFloat(String(desc).replace('%', '')) || 0
}

const getEchoMaxScoreBase = (template: ResonatorTemplate, cost: string) =>
  Number(template.echo_max_score?.[String(cost || '1C').slice(0, 1)] ?? 0)

const getMainstatBaseScore = (template: ResonatorTemplate, cost: string) =>
  Number(template.mainstat_max_score?.[cost || '1C'] ?? 0)

const MAINSTAT_VALUES: Record<string, number> = {
  '1C:生命%': 22.8,
  '1C:攻击%': 18,
  '1C:防御%': 22.8,
  '3C:共鸣效率': 32,
  '3C:生命%': 30,
  '3C:攻击%': 30,
  '3C:防御%': 38,
  '3C:冷凝伤害加成': 30,
  '3C:热熔伤害加成': 30,
  '3C:导电伤害加成': 30,
  '3C:气动伤害加成': 30,
  '3C:衍射伤害加成': 30,
  '3C:湮灭伤害加成': 30,
  '4C:暴击': 22,
  '4C:暴击伤害': 44,
  '4C:治疗效果加成': 26,
}

const getMainstatScore = (template: ResonatorTemplate, cost: string, mainstat: string) => {
  const normalizedCost = String(cost || '1C').startsWith('3C') ? '3C' : String(cost || '1C')
  if (!getMainstatOptions(normalizedCost).includes(mainstat)) return 0
  const value = MAINSTAT_VALUES[`${normalizedCost}:${mainstat}`] || 0
  const propName = mainstat.endsWith('伤害加成') ? '属性伤害加成' : mainstat
  const weight = Number(template.main_props?.[normalizedCost.slice(0, 1)]?.[propName] ?? 0)
  return truncateScore(value * weight / Number(template.echo_max_score?.[normalizedCost.slice(0, 1)] || 0) * 50)
}

const getSubstatWeight = (template: ResonatorTemplate, substatNum: number) => {
  const fullName = getSubstatFullName(substatNum)
  if (skillWeightIndex[fullName] !== undefined) {
    const genericWeight = Number(template.substat_weight?.['技能伤害加成'] ?? 0)
    const skillWeight = Number(template.skill_weight?.[skillWeightIndex[fullName]] ?? 0)
    return genericWeight * skillWeight
  }
  return Number(template.substat_weight?.[fullName] ?? 0)
}

const getScaledSubstatScore = (template: ResonatorTemplate, cost: string, substatNum: number, valueNum: number) => {
  const echoMaxScoreBase = getEchoMaxScoreBase(template, cost)
  if (echoMaxScoreBase <= 0) {
    return 0
  }
  const numericValue = getSubstatNumericValue(substatNum, valueNum)
  const weight = getSubstatWeight(template, substatNum)
  return truncateScore(weight * numericValue / echoMaxScoreBase * 50)
}

export const calculateEchoPotentialMaxScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
  mainstat = '',
) => {
  if (!template) {
    return 0
  }
  const selectedSubstats = getSelectedSubstats(echoLog)
  const currentTotal = getMainstatScore(template, cost, mainstat) + selectedSubstats.reduce((total, substatBits) => {
    const substatNum = bitPos(substatBits & MASK)
    const valueNum = bitPos(substatBits >> SUBSTAT_BIT_WIDTH)
    if (substatNum < 0 || valueNum < 0) {
      return total
    }
    return total + getScaledSubstatScore(template, cost, substatNum, valueNum)
  }, 0)
  const selectedNums = new Set(
    selectedSubstats
      .map((substatBits) => bitPos(substatBits & MASK))
      .filter((substatNum) => substatNum >= 0),
  )
  const remainingSlots = Math.max(0, 5 - selectedNums.size)
  const remainingTotal = remainingSlots <= 0
    ? 0
    : SUBSTAT
      .filter((substat) => !selectedNums.has(substat.num))
      .map((substat) => {
        const values = SUBSTAT_VALUE_MAP[substat.num] ?? []
        const maxValueNum = values.length - 1
        return maxValueNum >= 0 ? getScaledSubstatScore(template, cost, substat.num, maxValueNum) : 0
      })
      .sort((a, b) => b - a)
      .slice(0, remainingSlots)
      .reduce((sum, score) => sum + score, 0)

  return Number((currentTotal + remainingTotal).toFixed(2))
}

export const calculateEchoCurrentScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
  mainstat = '',
) => {
  if (!template) {
    return 0
  }
  const selectedSubstats = getSelectedSubstats(echoLog)
  const currentTotal = getMainstatScore(template, cost, mainstat) + selectedSubstats.reduce((total, substatBits) => {
    const substatNum = bitPos(substatBits & MASK)
    const valueNum = bitPos(substatBits >> SUBSTAT_BIT_WIDTH)
    if (substatNum < 0 || valueNum < 0) {
      return total
    }
    return total + getScaledSubstatScore(template, cost, substatNum, valueNum)
  }, 0)
  return Number(currentTotal.toFixed(2))
}

export const formatEchoPotentialMaxScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
  mainstat = '',
) => {
  const total = calculateEchoPotentialMaxScore(echoLog, template, cost, mainstat)
  return total > 0 ? total.toFixed(2) : ''
}

export const formatEchoCurrentScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
  mainstat = '',
) => {
  const total = calculateEchoCurrentScore(echoLog, template, cost, mainstat)
  return total > 0 ? total.toFixed(2) : ''
}
