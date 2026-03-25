<template>
  <div style="min-width: 900px; font-size: 12px">
    <button @click="fetchTuneStats">统计数据 - 刷新</button>
    &nbsp;
    <span>调谐次数：{{ tuneStats?.data_total }}</span>
    <table class="my-table" style="margin-top: 5px;">
      <thead>
        <tr>
          <th rowspan="2" style="width: 25px">词条</th>
          <th rowspan="2" style="width: 40px">档位</th>
          <template v-for="i in 5" :key="`head-${i}`">
            <th colspan="2">孔{{ i }}</th>
          </template>
          <th colspan="2">所有孔位</th>
<!--          <th colspan="2">所有档位</th>-->
        </tr>
        <tr>
          <template v-for="i in 5" :key="`head2-${i}`">
            <th style="width: 50px">次数</th>
            <th style="width: 50px">占比</th>
          </template>
          <th style="width: 50px">总次数</th>
          <th style="width: 50px">总占比</th>
        </tr>
      </thead>
      <tbody>
        <template v-for="substat in SUBSTAT" :key="substat">
          <tr
            v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
            :key="`${substat.num}-${value.value_number}`"
            :style="substat.font_color == '' ? '' : `color: ${substat.font_color}`"
          >
            <td v-if="value.value_number === 0" :rowspan="SUBSTAT_VALUE_MAP[substat.num].length + 1">
              {{ substat.name }}
            </td>
            <td>{{ value.desc }}</td>
            <template v-for="i in 5" :key="`cell-${substat.num}-${value.value_number}-${i}`">
              <td>
                {{
                  tuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]
                    ?.position_dict?.[i - 1]?.total
                }}
              </td>
              <td>
                {{
                  tuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]
                    ?.position_dict?.[i - 1]?.percent
                }}%
              </td>
            </template>
            <td>
              {{ tuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]?.total }}
            </td>
            <td>
              {{ tuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]?.percent }}%
            </td>
<!--            <td v-if="value.value_number === 0" :rowspan="SUBSTAT_VALUE_MAP[substat.num].length + 1">-->
<!--              {{ tuneStats.substat_dict?.[substat.num]?.total }}-->
<!--            </td>-->
<!--            <td v-if="value.value_number === 0" :rowspan="SUBSTAT_VALUE_MAP[substat.num].length + 1">-->
<!--              {{ tuneStats.substat_dict?.[substat.num]?.percent }}%-->
<!--            </td>-->
          </tr>
          <tr class="force-bold-shadow" :style="substat.font_color == '' ? '' : `color: ${substat.font_color}; font-weight: 700;`" >
            <td>全档位</td>
            <template v-for="i in 5" :key="`cell-all-${substat.num}-${i}`">
              <!-- <td :style="cellBgColor[i - 1] == '' ? '' : `background-color: ${cellBgColor[i - 1]}`" > -->
              <td>
                {{
                  tuneStats.substat_dict?.[substat.num]?.value_dict['all']
                    ?.position_dict?.[i - 1]?.total
                }}
              </td>
              <!-- <td :style="cellBgColor[i - 1] == '' ? '' : `background-color: ${cellBgColor[i - 1]}`" > -->
              <td>
                {{
                  tuneStats.substat_dict?.[substat.num]?.value_dict['all']
                    ?.position_dict?.[i - 1]?.percent
                }}%
              </td>
            </template>
            <td>
              {{ tuneStats.substat_dict?.[substat.num]?.value_dict['all']?.total }}
            </td>
            <td>
              {{ tuneStats.substat_dict?.[substat.num]?.value_dict['all']?.percent }}%
            </td>
          </tr>
        </template>
      </tbody>
    </table>
    <div style="min-height: 40px;"></div>
  </div>
</template>

<script lang="ts">
import { API_SERV, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants.ts'
import { onMounted, ref } from 'vue'
import axios from 'axios'
import {useRoute} from "vue-router";

export default {
  name: 'TuneStats',
  setup() {
    const tuneStats = ref({
      data_total: 0,
      substat_dict: {},
    })

    const route = useRoute()

    // cell-bg
    const cellBgColor = ['#646464', '#4B4B4B', '#646464', '#4B4B4B', '#646464']

    const fetchTuneStats = () => {
      const user_id = route.query.user_id ?? 0
      const after_id = route.query.after_id ?? ''
      const before_id = route.query.before_id ?? ''
      let url = `http://${API_SERV}/tune_stats?foo=bar`;
      if (user_id) {
        url += `&user_id=${user_id}`
      }
      if (after_id) {
        url += `&after_id=${after_id}`
      }
      if (before_id) {
        url += `&before_id=${before_id}`
      }
      axios
        .get(url)
        // .get(`http://${API_SERV}/tune_stats?`)
        .then((response) => {
          console.log(response.data) // DEBUG
          tuneStats.value = response.data.data
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }
    onMounted(fetchTuneStats)

    // 返回模板需要的数据和方法
    return {
      fetchTuneStats,
      tuneStats,
      cellBgColor,
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
  text-shadow:
    0.5px 0 0 currentColor, /* 右阴影 */
    -0.5px 0 0 currentColor; /* 左阴影 */
}
</style>
