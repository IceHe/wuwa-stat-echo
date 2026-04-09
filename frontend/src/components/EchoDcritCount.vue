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
        <th v-for="y in critDmgValues" :key="y.value_number">{{ y.desc }}</th>
        <th class="total-cell">合计</th>
      </tr>
      </thead>
      <tbody>
      <template v-for="x in critRateValues.length" :key="x">
        <tr style="text-align: center">
          <td class="axis-cell">{{ critRateValues[x - 1]?.desc }}</td>
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
        <td v-for="y in critDmgValues" :key="`total-${y.value_number}`" class="total-cell">
          {{ getColumnTotal(y.value_number) }}
        </td>
        <td class="total-cell">{{ grandTotal }}</td>
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
      const max = maxCellCount.value
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

    // 返回模板需要的数据和方法
    return {
      fetchEchoDcritCounts,
      echo_dcrit_count,
      getCellCount,
      getRowTotal,
      getColumnTotal,
      grandTotal,
      getDataCellStyle,
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

.total-row td,
.total-row th,
.total-cell {
  font-weight: 700;
  background: #f5efe2;
}

.force-bold-shadow {
  text-shadow: 0.5px 0 0 currentColor, /* 右阴影 */ -0.5px 0 0 currentColor; /* 左阴影 */
}
</style>
