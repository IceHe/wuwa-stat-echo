<template>
  <div style="min-width: 900px; font-size: 12px">
    <button @click="fetchEchoDcritCounts">统计数据 - 刷新</button>
    &nbsp;
    <span>声骸数量：{{ echo_dcrit_count?.echo_count }}</span>，
    <span>双暴数量：{{ echo_dcrit_count?.dcrit_total }}</span>
    <table class="my-table">
      <thead>
      <tr class="axis-row">
        <th class="axis-cell">暴击 \ 暴伤</th>
        <th class="rate-head-cell"></th>
        <th v-for="y in critDmgValues" :key="y.value_number">{{ y.desc }}</th>
        <th class="total-cell">合计</th>
      </tr>
      </thead>
      <tbody>
      <tr class="rate-row" style="text-align: center">
        <td class="axis-cell"></td>
        <td class="rate-cell rate-axis-cell">统计出率</td>
        <td v-for="y in critDmgValues" :key="`rate-${y.value_number}`" class="rate-cell">
          {{ getColumnRate(y.value_number) }}
        </td>
        <td class="total-cell">-</td>
      </tr>
      <template v-for="x in critRateValues.length" :key="x">
        <tr style="text-align: center">
          <td class="axis-cell">{{ critRateValues[x - 1]?.desc }}</td>
          <td class="rate-cell">{{ getRowRate(x - 1) }}</td>
          <td
            v-for="y in critDmgValues"
            :key="(x - 1) * 100 + y.value_number"
            class="data-cell"
            :style="getDataCellStyle(x - 1, y.value_number)"
          >
            <span class="data-cell-value">{{ getCellCount(x - 1, y.value_number) }}</span>
          </td>
          <td class="total-cell">{{ getRowTotal(x - 1) }}</td>
        </tr>
      </template>
      <tr class="total-row" style="text-align: center">
        <td class="total-cell">合计</td>
        <td class="total-cell"></td>
        <td v-for="y in critDmgValues" :key="`total-${y.value_number}`" class="total-cell">
          {{ getColumnTotal(y.value_number) }}
        </td>
        <td class="total-cell">{{ grandTotal }}</td>
      </tr>
      </tbody>
    </table>
    <div style="height: 20px;"></div>
    <div class="section-title">4 格合并统计</div>
    <table class="my-table aggregate-table">
      <thead>
      <tr class="axis-row">
        <th class="axis-cell">暴击 \ 暴伤</th>
        <th v-for="group in fourCellColGroups" :key="`four-col-${group.start}`">{{ formatGroupLabel(critDmgValues, group) }}</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="rowGroup in fourCellRowGroups" :key="`four-row-${rowGroup.start}`" style="text-align: center">
        <td class="axis-cell">{{ formatGroupLabel(critRateValues, rowGroup) }}</td>
        <td
          v-for="colGroup in fourCellColGroups"
          :key="`four-cell-${rowGroup.start}-${colGroup.start}`"
          class="data-cell"
          :style="getAggregatedCellStyle(rowGroup, colGroup, maxFourCellCount)"
        >
          <span class="data-cell-value">{{ getAggregatedCellCount(rowGroup, colGroup) }}</span>
        </td>
      </tr>
      </tbody>
    </table>
    <div style="height: 20px;"></div>
    <div class="section-title">9 格合并统计</div>
    <table class="my-table aggregate-table">
      <thead>
      <tr class="axis-row">
        <th class="axis-cell">暴击 \ 暴伤</th>
        <th v-for="group in nineCellColGroups" :key="`nine-col-${group.start}`">{{ formatGroupLabel(critDmgValues, group) }}</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="rowGroup in nineCellRowGroups" :key="`nine-row-${rowGroup.start}`" style="text-align: center">
        <td class="axis-cell">{{ formatGroupLabel(critRateValues, rowGroup) }}</td>
        <td
          v-for="colGroup in nineCellColGroups"
          :key="`nine-cell-${rowGroup.start}-${colGroup.start}`"
          class="data-cell"
          :style="getAggregatedCellStyle(rowGroup, colGroup, maxNineCellCount)"
        >
          <span class="data-cell-value">{{ getAggregatedCellCount(rowGroup, colGroup) }}</span>
        </td>
      </tr>
      </tbody>
    </table>
    <div style="min-height: 40px;"></div>
  </div>
</template>

<script lang="ts">
import {API_BASE_URL, SUBSTAT, SUBSTAT_VALUE_MAP} from '@/stores/constants.ts'
import {computed, onMounted, ref} from 'vue'
import axios from 'axios'
import {useRoute} from "vue-router";

export default {
  name: 'TuneStats',
  setup() {
    type ValueGroup = { start: number, size: number }
    const critRateValues = SUBSTAT_VALUE_MAP[0] ?? []
    const critDmgValues = SUBSTAT_VALUE_MAP[1] ?? []

    const echo_dcrit_count = ref({
      echo_count: 0,
      dcrit_total: 0,
      counts: {}
    })

    const route = useRoute()

    // cell-bg
    const cellBgColor = ['#646464', '#4B4B4B', '#646464', '#4B4B4B', '#646464']

    const fetchEchoDcritCounts = () => {
      const after_id = route.query.after_id || 0
      const before_id = route.query.before_id || 0
      axios
          .get(`${API_BASE_URL}/counts/echo_dcrit?after_id=${after_id}&before_id=${before_id}`)
          .then((response) => {
            console.log(response.data) // DEBUG
            echo_dcrit_count.value = response.data.data
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
    }
    onMounted(fetchEchoDcritCounts)

    const getCellCount = (critRateValueNumber: number, critDmgValueNumber: number) =>
      Number(echo_dcrit_count.value.counts?.[critRateValueNumber]?.[critDmgValueNumber] ?? 0)

    const getRowTotal = (critRateValueNumber: number) =>
      critDmgValues.reduce((sum, critDmgValue) => sum + getCellCount(critRateValueNumber, critDmgValue.value_number), 0)

    const getColumnTotal = (critDmgValueNumber: number) =>
      critRateValues.reduce((sum, critRateValue) => sum + getCellCount(critRateValue.value_number, critDmgValueNumber), 0)

    const grandTotal = computed(() =>
      critRateValues.reduce((sum, critRateValue) => sum + getRowTotal(critRateValue.value_number), 0)
    )

    const formatRate = (count: number) => {
      const total = grandTotal.value
      if (total <= 0) {
        return '0%'
      }
      return `${(count / total * 100).toFixed(1)}%`
    }

    const getRowRate = (critRateValueNumber: number) => formatRate(getRowTotal(critRateValueNumber))

    const getColumnRate = (critDmgValueNumber: number) => formatRate(getColumnTotal(critDmgValueNumber))

    const maxCellCount = computed(() => {
      let max = 0
      for (const critRateValue of critRateValues) {
        for (const critDmgValue of critDmgValues) {
          max = Math.max(max, getCellCount(critRateValue.value_number, critDmgValue.value_number))
        }
      }
      return max
    })

    const getDataCellStyle = (critRateValueNumber: number, critDmgValueNumber: number) => {
      const count = getCellCount(critRateValueNumber, critDmgValueNumber)
      return buildHeatmapStyle(count, maxCellCount.value)
    }

    const buildHeatmapStyle = (count: number, max: number) => {
      const ratio = max > 0 ? count / max : 0
      const barWidth = Math.max(0, Math.min(100, ratio * 100))
      const hue = 120 * (1 - ratio)
      const saturation = 75
      const lightness = 92 - ratio * 44
      const barColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`
      const fadeColor = `hsl(${hue}, ${Math.max(18, saturation - 25)}%, 98%)`
      const emptyColor = '#ffffff'

      return {
        background: `linear-gradient(90deg, ${barColor} 0%, ${fadeColor} ${barWidth}%, ${emptyColor} ${barWidth}%, ${emptyColor} 100%)`,
      }
    }

    const buildGroups = (length: number, preferredSize: number): ValueGroup[] => {
      const groups: ValueGroup[] = []
      let index = 0
      while (index < length) {
        const remaining = length - index
        let size = preferredSize
        if (preferredSize === 3 && remaining < 3) {
          size = Math.min(2, remaining)
        } else if (preferredSize === 3 && remaining === 4) {
          size = 2
        } else if (remaining < preferredSize) {
          size = remaining
        }
        groups.push({start: index, size})
        index += size
      }
      return groups
    }

    const fourCellRowGroups = computed(() => buildGroups(critRateValues.length, 2))
    const fourCellColGroups = computed(() => buildGroups(critDmgValues.length, 2))
    const nineCellRowGroups = computed(() => buildGroups(critRateValues.length, 3))
    const nineCellColGroups = computed(() => buildGroups(critDmgValues.length, 3))

    const getAggregatedCellCount = (rowGroup: ValueGroup, colGroup: ValueGroup) => {
      let total = 0
      for (let rowOffset = 0; rowOffset < rowGroup.size; rowOffset++) {
        for (let colOffset = 0; colOffset < colGroup.size; colOffset++) {
          total += getCellCount(rowGroup.start + rowOffset, colGroup.start + colOffset)
        }
      }
      return total
    }

    const buildMaxAggregatedCount = (rowGroups: ValueGroup[], colGroups: ValueGroup[]) =>
      rowGroups.reduce((rowMax, rowGroup) => {
        const groupMax = colGroups.reduce((colMax, colGroup) => Math.max(colMax, getAggregatedCellCount(rowGroup, colGroup)), 0)
        return Math.max(rowMax, groupMax)
      }, 0)

    const maxFourCellCount = computed(() => buildMaxAggregatedCount(fourCellRowGroups.value, fourCellColGroups.value))
    const maxNineCellCount = computed(() => buildMaxAggregatedCount(nineCellRowGroups.value, nineCellColGroups.value))

    const getAggregatedCellStyle = (rowGroup: ValueGroup, colGroup: ValueGroup, max: number) =>
      buildHeatmapStyle(getAggregatedCellCount(rowGroup, colGroup), max)

    const formatGroupLabel = (values: Array<{ desc: string }>, group: ValueGroup) => {
      const first = values[group.start]?.desc ?? ''
      const last = values[group.start + group.size - 1]?.desc ?? first
      return first === last ? first : `${first} - ${last}`
    }

    // 返回模板需要的数据和方法
    return {
      fetchEchoDcritCounts,
      echo_dcrit_count,
      getCellCount,
      getRowTotal,
      getColumnTotal,
      getRowRate,
      getColumnRate,
      grandTotal,
      getDataCellStyle,
      fourCellRowGroups,
      fourCellColGroups,
      nineCellRowGroups,
      nineCellColGroups,
      getAggregatedCellCount,
      getAggregatedCellStyle,
      maxFourCellCount,
      maxNineCellCount,
      formatGroupLabel,
      cellBgColor,
      critRateValues,
      critDmgValues,
      SUBSTAT,
      SUBSTAT_VALUE_MAP,
    }
  },
}
</script>

<style scoped>
.my-table {
  width: 100%;
  border-collapse: collapse; /* 关键：合并边框 */
  border: 1px solid #e0e0e0; /* 表格边框 */
}

.aggregate-table {
  width: auto;
  min-width: 720px;
}

.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 8px;
}

.data-cell {
  min-width: 72px;
  border: 1px solid #cfcfcf;
  transition: background 160ms ease;
}

.data-cell-value {
  display: inline-block;
  min-width: 2ch;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.axis-cell {
  font-weight: 700;
  background: #f5efe2;
}

.axis-row th {
  font-weight: 700;
  background: #f5efe2;
}

.rate-head-cell,
.rate-cell {
  font-weight: 700;
  color: #1f3558;
  background: #e7f0ff;
}

.rate-axis-cell {
  color: #6b7280;
}

.total-row td,
.total-row th,
.total-cell {
  font-weight: 700;
  background: #f5efe2;
}

.section-title {
  margin: 8px 0;
  font-weight: 700;
  color: #4b5563;
}

.force-bold-shadow {
  text-shadow: 0.5px 0 0 currentColor, /* 右阴影 */ -0.5px 0 0 currentColor; /* 左阴影 */
}
</style>
