<!-- 用多种按钮来记录不同的调谐记录 -->
<template>
  <div style="min-width: 750px;">
    <div style="display: flex;">
      <span class="name">目标词条</span>
      <button
          class="target_button"
          v-for="substat in SUBSTAT"
          :key="substat"
          @click="toggleTargetSubstat(substat.bitmap)"
          :style="(targetSubstatBitmap & substat.bitmap ? 'background-color: yellow;' : '') + `color: ${substat.font_color}`"
      >
        {{ substat.name.substring(0, 4) }}
      </button>
    </div>
    <table class="my-table" style="margin-bottom: 5px;">
      <thead>
      <tr>
        <th rowspan="2" style="width: 12%;">玩家</th>
        <th colspan="3">
          目标
          <span style="color: gray;">(默认未出货时回收调谐器&密音筒)</span>
        </th>
        <th colspan="2">调谐器</th>
        <th colspan="2">金密音筒</th>
      </tr>
      <tr>
        <th rowspan="2" style="width: 8%;">总数</th>
        <th style="width: 15%;">平均间隔声骸</th>
        <th style="width: 15%;">平均间隔词条</th>
        <th rowspan="2" style="width: 12%;">总消耗</th>
        <th rowspan="2" style="width: 12%;">平均消耗</th>
        <th rowspan="2" style="width: 12%;">总消耗</th>
        <th rowspan="2" style="width: 12%;">平均消耗</th>
      </tr>
      </thead>
      <tbody>
      <tr>
        <td>
          <!--当前玩家-->
          {{ echoLog.user_id }}
        </td>
        <td>
          <span style="color: red">{{ currentUser.target }}</span>
        </td>
        <td>
          <span style="color: orange">{{ currentUser.target_avg_echo }}</span>
        </td>
        <td>
          <span style="color: orange">{{ currentUser.target_avg_substat }}</span>
        </td>
        <td>
          <span style="color: red">{{ currentUser.tuner_consumed }}</span>
        </td>
        <td>
          <span style="color: orange">{{ currentUser.tuner_consumed_avg }}</span>
        </td>
        <td>
          <span style="color: red">{{ currentUser.exp_consumed }}</span>
        </td>
        <td>
          <span style="color: orange">{{ currentUser.exp_consumed_avg }}</span>
        </td>
      </tr>
      <tr>
        <td>
          所有玩家
        </td>
        <td>
          <span style="color: red">{{ allUsers.target }}</span>
        </td>
        <td>
          <span style="color: orange">{{ allUsers.target_avg_echo }}</span>
        </td>
        <td>
          <span style="color: orange">{{ allUsers.target_avg_substat }}</span>
        </td>
        <td>
          <span style="color: red">{{ allUsers.tuner_consumed }}</span>
        </td>
        <td>
          <span style="color: orange">{{ allUsers.tuner_consumed_avg }}</span>
        </td>
        <td>
          <span style="color: red">{{ allUsers.exp_consumed }}</span>
        </td>
        <td>
          <span style="color: orange">{{ allUsers.exp_consumed_avg }}</span>
        </td>
      </tr>
      </tbody>
    </table>

    <!--<div>-->
    <!--  <span class="name">-->
    <!--    词条 x 数量 - 间隔-->
    <!--  </span>-->
    <!--</div>-->
    <div style="display: flex;">
      <span class="name">
        当前声骸:
        <span style="font-weight: bolder;">
          {{ echoLog.id }}
        </span>
        <br />
        套装:
        <span :style="`color: ${CLASS_COLORS[echoLog.clazz]};`">
          {{ echoLog.clazz.substring(0, 4) }}
        </span>
      </span>
      <button class="substat"
              key="1"
              :style="(echoLog.pos === 0 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat1)};`"
      >
        {{ echoLog.s1_desc ? echoLog.s1_desc : "1" }}
      </button>
      <button class="substat"
              key="2"
              :style="(echoLog.pos === 1 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat2)};`"
      >
        {{ echoLog.s2_desc ? echoLog.s2_desc : "2" }}
      </button>
      <button class="substat"
              key="3"
              :style="(echoLog.pos === 2 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat3)};`"
      >
        {{ echoLog.s3_desc ? echoLog.s3_desc : "3" }}
      </button>
      <button class="substat"
              key="4"
              :style="(echoLog.pos === 3 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat4)};`"
      >
        {{ echoLog.s4_desc ? echoLog.s4_desc : "4" }}
      </button>
      <button class="substat"
              key="5"
              :style="(echoLog.pos === 4 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat5)};`"
      >
        {{ echoLog.s5_desc ? echoLog.s5_desc : "5" }}
      </button>
    </div>
    <div v-for="substat in SUBSTAT" :key="substat" style="display: flex;">
      <span class="name">
        {{ substat.name.substring(0, 4) }}
        <span :style="recentTuneStats.substat_dict?.[substat.num]?.total > 6 ? 'color: green;' : (
            recentTuneStats.substat_dict?.[substat.num]?.total < 3 ? 'color: red;' : ''
          )">
          x {{ recentTuneStats.substat_dict?.[substat.num]?.total }}
        </span>
        <span :style="
          recentTuneStats.substat_distance?.[substat.num] >= 39 ? 'color: red;' : (
            recentTuneStats.substat_distance?.[substat.num] >= 20 ? 'color: orange;' : ''
          )
        ">
          - {{ recentTuneStats.substat_distance?.[substat.num] }}
        </span>
      </span>
      <button
          class="button"
          v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
          :key="value"
          :style="`color: ${substat.font_color}; display: ${
            recentTuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]?.total > 0
            ? 'hidden' : ''
          }`"
      >
        <span v-if="recentTuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]?.total"
              style="font-weight: bold;"
        >
          {{ value.desc }}
          x {{ recentTuneStats.substat_dict?.[substat.num]?.value_dict[value.value_number]?.total }}
        </span>
      </button>
      <span v-if="substat.num === 5">
        &nbsp;
        当前玩家 距离上次双暴
        <span style="color: red;">{{ currentUser.target_echo_distance }}</span> 个声骸
        <span style="color: orange;">{{ currentUser.target_substat_distance }}</span> 个词条
      </span>
      <!--<span v-if="substat.num === 6">-->
      <!--  &nbsp;-->
      <!--  所有玩家 距离上次双暴-->
      <!--  <span style="color: red;">{{allUsers.target_substat_distance}}</span> 个声骸-->
      <!--  <span style="color: orange;">{{allUsers.target_echo_distance}}</span> 个词条-->
      <!--</span>-->
    </div>
  </div>
  <span style="margin-bottom: 5px;"></span>
</template>

<script>
import axios from 'axios'
import {
  API_SERV,
  CLASS_COLORS,
  CLASSES,
  getSubstatColor,
  SUBSTAT,
  SUBSTAT_VALUE_MAP
} from '@/stores/constants.ts'
import {onMounted, ref} from 'vue'
import {useRoute, useRouter} from 'vue-router';
import {template} from "@antfu/utils";

export default {
  name: 'Echo',
  methods: {template},
  computed: {
    CLASS_COLORS() {
      return CLASS_COLORS
    }
  },
  props: {},
  setup: function (props) {
    const route = useRoute();
    const echoLog = ref({
      clazz: '',
      user_id: 0,
      id: route.query.echo_id ?? 0,
      pos: 0, // 当前孔位
      substat1: 0,
      substat2: 0,
      substat3: 0,
      substat4: 0,
      substat5: 0,
      substat_all: 0,
      s1_desc: "",
      s2_desc: "",
      s3_desc: "",
      s4_desc: "",
      s5_desc: "",
      // tuned_at: new Date().toISOString(), // FIXME
    })

    const getEchoLog = async (echoId = 0) => {
      const targetEchoId = Number(echoId ? echoId : echoLog.value.id)
      try {
        const response = await axios.get(`http://${API_SERV}/echo_log/${targetEchoId}`)
        console.log("get echo log:", response.data) // DEBUG
        if (response.data.code === 200) {
          echoLog.value = response.data.data
          echoLog.value.pos = 0
          if (echoLog.value.substat1 > 0) {
            echoLog.value.pos = 1
          }
          if (echoLog.value.substat2 > 0) echoLog.value.pos = 2
          if (echoLog.value.substat3 > 0) echoLog.value.pos = 3
          if (echoLog.value.substat4 > 0) echoLog.value.pos = 4
          return true
        } else {
          if (targetEchoId <= 0 && response.data.message === 'echo log not found') {
            echoLog.value = {
              ...echoLog.value,
              clazz: '',
              user_id: 0,
              id: 0,
              pos: 0,
              substat1: 0,
              substat2: 0,
              substat3: 0,
              substat4: 0,
              substat5: 0,
              substat_all: 0,
              s1_desc: "",
              s2_desc: "",
              s3_desc: "",
              s4_desc: "",
              s5_desc: "",
            }
            return false
          }
          alert('获取声骸失败')
          return false
        }
      } catch (error) {
        if (targetEchoId <= 0 && error?.response?.data?.message === 'echo log not found') {
          echoLog.value = {
            ...echoLog.value,
            clazz: '',
            user_id: 0,
            id: 0,
            pos: 0,
            substat1: 0,
            substat2: 0,
            substat3: 0,
            substat4: 0,
            substat5: 0,
            substat_all: 0,
            s1_desc: "",
            s2_desc: "",
            s3_desc: "",
            s4_desc: "",
            s5_desc: "",
          }
          return false
        }
        console.error('获取声骸 请求失败:', error)
        alert('获取声骸 请求失败')
        return false
      }
    }

    const currentUser = ref({
      target_echo_distance: -1,
      target_substat_distance: -1,
      target: 0,
      target_avg_echo: 0.0,
      target_avg_substat: 0.0,
      tuner_consumed: 0,
      tuner_consumed_avg: 0.0,
      exp_consumed: 0,
      exp_consumed_avg: 0.0,
    })
    const allUsers = ref({
      target_echo_distance: -1,
      target_substat_distance: -1,
      target: 0,
      target_avg_echo: 0.0,
      target_avg_substat: 0.0,
      tuner_consumed: 0,
      tuner_consumed_avg: 0.0,
      exp_consumed: 0,
      exp_consumed_avg: 0.0,
    })
    const refreshEchoLogsAnalysis = async (size = 0) => {
      try {
        const response = await axios.get(`http://${API_SERV}/echo_logs/analysis?size=${size}&user_id=${echoLog.value.user_id}&target_bits=${targetSubstatBitmap.value}`)
        console.log('current user: ', response.data) // DEBUG
        currentUser.value = response.data.data
      } catch (error) {
        console.error('请求失败:', error)
      }
      try {
        const response = await axios.get(`http://${API_SERV}/echo_logs/analysis?page_size=${size}&target_bits=${targetSubstatBitmap.value}`)
        console.log('all users: ', response.data) // DEBUG
        allUsers.value = response.data.data
      } catch (error) {
        console.error('请求失败:', error)
      }
    }

    const targetSubstatBitmap = ref(0b11)
    const toggleTargetSubstat = (bitmap) => {
      targetSubstatBitmap.value ^= bitmap
      refreshEchoLogsAnalysis()
    }

    // 展示最近各词条出现的数量
    const recentTuneStats = ref({
      data_total: 0,
      substat_dict: {},
      substat_distance: [],
    })
    const refreshRecentTuneStats = async (size = 52) => {
      try {
        const response = await axios.get(`http://${API_SERV}/tune_stats?size=${size}&user_id=${echoLog.value.user_id}`)
        console.log("refreshRecentTuneStats: ", response.data) // DEBUG
        recentTuneStats.value = response.data.data
      } catch (error) {
        console.error('请求失败:', error)
      }
    }
    onMounted(async () => {
      await getEchoLog()
      await refreshEchoLogsAnalysis()
      await refreshRecentTuneStats()
    })

    return {
      targetSubstatBitmap,
      echoLog,
      recentTuneStats,
      currentUser,
      allUsers,
      getSubstatColor,
      toggleTargetSubstat,
      CLASSES,
      SUBSTAT,
      SUBSTAT_VALUE_MAP,
    }
  },
}
</script>

<style scoped>
.name {
  display: inline-block;
  text-align-all: center;
  min-width: 120px;
  width: 12%;
}

.button {
  display: inline-block;
  width: 10.5%;
  height: 35px;
  text-align: center;
}

.target_button {
  display: inline-block;
  width: 6.1%;
  height: 50px;
  text-align: center;
}

.substat {
  display: inline-block;
  min-width: 90px;
  max-width: 120px;
  width: 19.5%;
  height: 60px;
  text-align: center;
}

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
