<template>
  <div style="min-width: 900px; font-size: 12px">
    <button @click="fetchEchoDcritCounts">统计数据 - 刷新</button>
    &nbsp;
    <span>声骸数量：{{ echo_dcrit_count?.echo_count }}</span>，
    <span>双暴数量：{{ echo_dcrit_count?.dcrit_total }}</span>
    <table class="my-table">
      <thead>
      <tr>
        <th>暴击 \ 暴伤</th>
        <th v-for="y in critDmgValues" :key="y.value_number">{{ y.desc }}</th>
      </tr>
      </thead>
      <tbody>
      <template v-for="x in critRateValues.length" :key="x">
        <tr style="text-align: center">
          <td>{{ critRateValues[x - 1]?.desc }}</td>
          <td v-for="y in critDmgValues" :key="(x - 1) * 100 + y.value_number">
            {{ echo_dcrit_count.counts?.[x - 1]?.[y.value_number] ?? "-0" }}
          </td>
        </tr>
      </template>
      </tbody>
    </table>
    <div style="min-height: 40px;"></div>
  </div>
</template>

<script lang="ts">
import {API_SERV, SUBSTAT, SUBSTAT_VALUE_MAP} from '@/stores/constants.ts'
import {onMounted, ref} from 'vue'
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
          .get(`http://${API_SERV}/counts/echo_dcrit?after_id=${after_id}&before_id=${before_id}`)
          .then((response) => {
            console.log(response.data) // DEBUG
            echo_dcrit_count.value = response.data.data
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
    }
    onMounted(fetchEchoDcritCounts)

    // 返回模板需要的数据和方法
    return {
      fetchEchoDcritCounts,
      echo_dcrit_count,
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

.force-bold-shadow {
  text-shadow: 0.5px 0 0 currentColor, /* 右阴影 */ -0.5px 0 0 currentColor; /* 左阴影 */
}
</style>
