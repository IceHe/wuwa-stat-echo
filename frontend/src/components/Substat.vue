<!-- 用多种按钮来记录不同的调谐记录 -->
<template>
  <div style="min-width: 600px">
    <div>
      <h1>调谐</h1>
      <span class="name">当前孔位: </span>
      <button
        class="button"
        v-for="i in 5"
        :key="i"
        @click="pos = i - 1"
        :style="pos + 1 === i ? 'background-color: yellow; font-color: red' : ''"
      >
        {{ i }}
      </button>
    </div>
    <div v-for="substat in SUBSTAT" :key="substat">
      <span class="name">{{ substat.name }} {{recentTuneStats.substat_dict?.[substat.num]?.total}}</span>
      <button
        class="button"
        v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
        :key="value"
        @click="addTuneLog(value.substat_number, value.value_number, pos); "
        :style="`color: ${substat.font_color}`"
      >
        {{ value.desc }}<br/>{{recentTuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]?.total}}
      </button>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { API_SERV, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants.ts'
import {onMounted, ref} from 'vue'
import emitter from '@/stores/eventBus.js'

export default {
  name: 'Tune',
  props: {
    position: {
      type: Number,
      required: true,
    },
  },
  setup(props) {
    const pos = ref(props.position)

    // 展示最近各词条出现的数量
    const recentTuneStats = ref({
      data_total: 0,
      substat_dict: {},
    })
    const refreshRecentTuneStats = () => {
      const size = 39;
      axios
        .get(`http://${API_SERV}/tune_stats?size=${size}`)
        .then((response) => {
          console.log('recent tune stats: ', response.data) // DEBUG
          recentTuneStats.value = response.data.data
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }
    onMounted(refreshRecentTuneStats)

    const addTuneLog = (substat, value, position) => {
      console.log('substat:', substat, 'value:', value, 'pos:', position)
      axios
        .post(`http://${API_SERV}/tune_log`, {
          substat: substat,
          value: value,
          position: position,
          echo_id: 0,
        })
        .then((response) => {
          console.log("add tune log: ", response.data) // DEBUG
          const code = response.data.code
          if (code === 200) {
            // alert('添加调谐记录成功');
            if (pos.value < 4) {
              pos.value++
            } else {
              pos.value = 0
            }
            emitter.emit('refreshSubstatLogs')
            refreshRecentTuneStats()
          } else {
            alert('添加调谐记录失败')
          }
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }
    return {
      addTuneLog,
      pos,
      refreshRecentTuneStats,
      recentTuneStats,
      SUBSTAT,
      SUBSTAT_VALUE_MAP,
    }
  },
}
</script>

<style scoped>
.name {
  display: inline-block;
  min-width: 100px;
  width: 15%;
}
.button {
  display: inline-block;
  min-width: 30px;
  width: 10%;
  height: 40px;
}
</style>
