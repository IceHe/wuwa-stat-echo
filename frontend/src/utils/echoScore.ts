import { SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants.ts'

const MASK = 0b1111111111111
const SUBSTAT_BIT_WIDTH = 13

export type ResonatorTemplate = {
  name?: string
  echo_max_score?: Record<string, number>
  mainstat_max_score?: Record<string, number>
  substat_weight?: Record<string, number>
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

const getSubstatNumericValue = (substatNum: number, valueNum: number) => {
  const desc = SUBSTAT_VALUE_MAP[substatNum]?.[valueNum]?.desc ?? '0'
  return Number.parseFloat(String(desc).replace('%', '')) || 0
}

const getEchoMaxScoreBase = (template: ResonatorTemplate, cost: string) =>
  Number(template.echo_max_score?.[String(cost || '1C').slice(0, 1)] ?? 0)

const getMainstatBaseScore = (template: ResonatorTemplate, cost: string) =>
  Number(template.mainstat_max_score?.[cost || '1C'] ?? 0)

const getSubstatWeight = (template: ResonatorTemplate, substatNum: number) => {
  const fullName = getSubstatFullName(substatNum)
  return Number(template.substat_weight?.[fullName] ?? 0)
}

const getScaledSubstatScore = (template: ResonatorTemplate, cost: string, substatNum: number, valueNum: number) => {
  const echoMaxScoreBase = getEchoMaxScoreBase(template, cost)
  if (echoMaxScoreBase <= 0) {
    return 0
  }
  const numericValue = getSubstatNumericValue(substatNum, valueNum)
  const weight = getSubstatWeight(template, substatNum)
  return weight * numericValue / echoMaxScoreBase * 50
}

export const calculateEchoPotentialMaxScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
) => {
  if (!template) {
    return 0
  }
  const selectedSubstats = getSelectedSubstats(echoLog)
  const currentTotal = selectedSubstats.reduce((total, substatBits) => {
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

  return getMainstatBaseScore(template, cost) + currentTotal + remainingTotal
}

export const calculateEchoCurrentScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
) => {
  if (!template) {
    return 0
  }
  const selectedSubstats = getSelectedSubstats(echoLog)
  const currentTotal = selectedSubstats.reduce((total, substatBits) => {
    const substatNum = bitPos(substatBits & MASK)
    const valueNum = bitPos(substatBits >> SUBSTAT_BIT_WIDTH)
    if (substatNum < 0 || valueNum < 0) {
      return total
    }
    return total + getScaledSubstatScore(template, cost, substatNum, valueNum)
  }, 0)
  return getMainstatBaseScore(template, cost) + currentTotal
}

export const formatEchoPotentialMaxScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
) => {
  const total = calculateEchoPotentialMaxScore(echoLog, template, cost)
  return total > 0 ? total.toFixed(2) : ''
}

export const formatEchoCurrentScore = (
  echoLog: EchoSubstatCarrier,
  template: ResonatorTemplate | null | undefined,
  cost = '1C',
) => {
  const total = calculateEchoCurrentScore(echoLog, template, cost)
  return total > 0 ? total.toFixed(2) : ''
}
