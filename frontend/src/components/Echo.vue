<!-- 用多种按钮来记录不同的调谐记录 -->
<template>
  <div style="min-width: 750px;">
    <!--<h1>调谐声骸</h1>-->

    <div class="template-row">
      <!--<span class="name">-->
      <!--  评分模板-->
      <!--  <br/>-->
      <!--  &nbsp;&nbsp;{{ scoreTemplate.resonator ? scoreTemplate.resonator : '无' }}-->
      <!--  <br/>-->
      <!--  &nbsp;&nbsp;{{ scoreTemplate.cost ? scoreTemplate.cost : '无' }}-->
      <!--</span>-->
      <button class="template-label-button">
        <!--共鸣者-->
        评分模板
      </button>
      <div class="template-scroll">
        <button
            v-for="resonator in RESONATORS"
            class="button template-button"
            :style="scoreTemplate.resonator === resonator ? 'background-color: yellow;' : ''"
            @click="setResonator(resonator)"
        >
          {{ resonator }}
        </button>
      </div>
      <div class="template-cost-group">
        <button class="template-cost-label">Cost主词条</button>
        <button
            v-for="cost in ECHO_COST"
            class="button template-cost-button"
            :style="scoreTemplate.cost === cost ? 'background-color: yellow;' : ''"
            @click="setCost(cost)"
        >
          {{ cost }}
        </button>
      </div>
    </div>

    <div class="player-info-row">
      <span class="name substat-name-cell">玩家ID</span>
      <input
          class="button player-info-input"
          type="text"
          v-model="echoLog.user_id"
          placeholder="当前玩家ID"
          @change="setUserId(echoLog.user_id)" />
      <span class="button player-info-label">声骸ID</span>
      <input
          class="button player-info-input echo-id-input"
          type="text"
          v-model="echoLog.id"
          placeholder="当前声骸ID"
          @change="setEchoId(echoLog.id)"
      />
      <!--      <span class="button">调谐日期</span>-->
      <!--      <input-->
      <!--        class="button"-->
      <!--        style="text-align: center; font-weight: bolder;"-->
      <!--        type="text"-->
      <!--        v-model="echoLog.id"-->
      <!--        placeholder="调谐时间"-->
      <!--        @change="setTunedAt(echoLog.id)"-->
      <!--      />-->
      <span class="player-info-chip" :style="`color: ${CLASS_COLORS[echoLog.clazz]};`">
        {{ echoLog.clazz.substring(0, 4) }}
      </span>
      <span class="player-stats-text">
        <!--当前玩家 -->
        距上次双暴
        <span style="color: red;">{{ currentUser.target_echo_distance }}</span> 声骸
        <span style="color: orange;">{{ currentUser.target_substat_distance }}</span> 词条
      </span>
      <span class="inline-double-crit-rate player-info-chip">
        双暴成功率 {{ echoAnalysis.two_crit_percent }}%
      </span>
      <!--<span>-->
      <!--  &nbsp;-->
      <!--  所有玩家 距离上次双暴-->
      <!--  <span style="color: red;">{{allUsers.target_substat_distance}}</span> 个声骸-->
      <!--  <span style="color: orange;">{{allUsers.target_echo_distance}}</span> 个词条-->
      <!--</span>-->
    </div>
    <div class="substat-row">
      <span class="name substat-name-cell">
        <span style="font-weight: bolder; color: red;">
          声骸副词条
        </span>
        <!--<span style="font-size: small;">ww面板</span>-->
      </span>
      <div class="top-summary-track">
        <div class="substat-summary-buttons">
          <button class="substat"
                  key="1"
                  @click="setPos(0)"
                  :style="(echoLog.pos === 0 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat1)};`"
          >
            {{ echoLog.s1_desc ? echoLog.s1_desc : "1" }}
            <br />
            <span v-if="echoLog.substat1" style="font-weight: bolder;">
              {{ echoAnalysis.score.substat1 }} 分
            </span>
          </button>
          <button class="substat"
                  key="2"
                  @click="setPos(1)"
                  :style="(echoLog.pos === 1 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat2)};`"
          >
            {{ echoLog.s2_desc ? echoLog.s2_desc : "2" }}
            <br />
            <span v-if="echoLog.substat2" style="font-weight: bolder;">
              {{ echoAnalysis.score.substat2 }} 分
            </span>
          </button>
          <button class="substat"
                  key="3"
                  @click="setPos(2)"
                  :style="(echoLog.pos === 2 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat3)};`"
          >
            {{ echoLog.s3_desc ? echoLog.s3_desc : "3" }}
            <br />
            <span v-if="echoLog.substat3" style="font-weight: bolder;">
              {{ echoAnalysis.score.substat3 }} 分
            </span>
          </button>
          <button class="substat"
                  key="4"
                  @click="setPos(3)"
                  :style="(echoLog.pos === 3 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat4)};`"
          >
            {{ echoLog.s4_desc ? echoLog.s4_desc : "4" }}
            <br />
            <span v-if="echoLog.substat4" style="font-weight: bolder;">
              {{ echoAnalysis.score.substat4 }} 分
            </span>
          </button>
          <button class="substat"
                  key="5"
                  @click="setPos(4)"
                  :style="(echoLog.pos === 4 ? 'background-color: yellow;' : '') + `color: ${getSubstatColor(echoLog.substat5)};`"
          >
            {{ echoLog.s5_desc ? echoLog.s5_desc : "5" }}
            <br />
            <span v-if="echoLog.substat5" style="font-weight: bolder;">
              {{ echoAnalysis.score.substat5 }} 分
            </span>
          </button>
          <button class="substat"
                  key="end"
                  @click="setPos(5)"
                  :style="echoLog.pos === 5 ? 'background-color: yellow; color: red;' : 'color: red;'"
                  :disabled="!canCreate && echoLog.pos !== 5 && !!echoLog.id"
          >
            {{ echoLog.pos === 5 ? '下一个' : 'END' }}
            <br />
            <span style="font-weight: bolder;">
              {{ echoLog.pos === 5 ? '创建声骸' : '结束当前' }}
            </span>
          </button>
        </div>
        <div class="score-panel-slot">
          <span v-if="echoAnalysis.score.substat_all" class="score-panel current-score-panel">
            当前评分<br />
            <span class="suite-score-panel-value">{{ echoAnalysis.score.substat_all }}</span>
          </span>
        </div>
      </div>
    </div>
    <div class="suite-row">
      <span class="name substat-name-cell">
        声骸套装<br />
        <span v-for="_ in 6">&#9472;</span><br />
        <!--词条x次|间隔<br />-->
        <span class="suite-note">(柱长表示频次)</span><br />
        词条 | 间隔
      </span>
      <div class="top-summary-track">
        <div class="suite-scroll">
          <button
              class="button suite-button"
              v-for="clazz in CLASSES"
              :key="clazz"
              @click="setClazz(clazz)"
              :style="echoLog.clazz === clazz ? 'background-color: yellow;' : ''"
              :disabled="!canCreate && echoLog.pos !== 5 && !!echoLog.id"
          >
            <span :style="`color: ${CLASS_COLORS[clazz]};`"> {{ clazz.substring(0, 4) }}</span>
          </button>
        </div>
        <div class="score-panel-slot suite-score-panels">
          <span v-if="showPotentialMaxScore()" class="score-panel suite-score-panel">
            理论最高<br />
            <span class="suite-score-panel-value">{{ getPotentialMaxScore() }}</span>
          </span>
        </div>
      </div>
    </div>
    <div v-for="substat in SUBSTAT" :key="substat" :style="'display: flex;'">
      <span class="name substat-name-cell">
        <div class="substat-summary-text">
          <span class="substat-name-label" :style="`color: ${substat.font_color};`">
            {{ substat.name.substring(0, 4) }}
          </span>
          <span :style="
            recentTuneStats.substat_dict?.[substat.num]?.total > 4
              ? 'color: green;'
              : (recentTuneStats.substat_dict?.[substat.num]?.total < 2 ? 'color: red;' : '')
          ">
            x{{ recentTuneStats.substat_dict?.[substat.num]?.total }}
          </span>
          <span class="substat-distance-text" :style="getRecentDistanceColor(substat.num)">
            |
            {{ getRecentDistanceDisplay(substat.num) }}
            <sup
                v-if="isRecentDistanceOverflow(substat.num)"
                class="distance-overflow-mark"
                :title="getRecentDistanceTitle(substat.num)"
            >+</sup>
          </span>
        </div>
        <div class="button-mini-bar-shell summary-bar-shell name-mini-bar-shell">
          <div
              class="button-mini-bar-fill"
              :style="getRecentSubstatTotalBarStyle(substat.num, substat.font_color)"
          />
        </div>
      </span>
      <button
          class="button stat-button"
          v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
          :key="value"
          @click="doTune(value.substat_number, value.value_number); "
          :style="`color: ${(substat.bitmap & echoLog.substat_all) === 0 ? substat.font_color : '#808080'};`"
          :disabled="echoLog.pos === 5 || (substat.bitmap & echoLog.substat_all) !== 0"
      >
        <span class="stat-button-label">
          {{ value.desc }}
        </span>
        <span class="stat-button-count">
          {{ getRecentValueCount(substat.num, value.value_number) || '' }}
        </span>
        <div class="button-mini-bar-shell">
          <div
              class="button-mini-bar-fill"
              :style="getRecentValueBarStyle(substat.num, value.value_number, substat.font_color)"
          />
        </div>
      </button>
      <span
          class="substat-current-position-rate"
          :style="`color: ${substat.font_color};`"
          :title="`${substat.name} 条件出率（当前孔位）：${echoAnalysis.substat_dict?.[substat.num]?.cur_pos_percent || '暂无数据'}。基于历史记录并排除已出现副词条。`"
      >
        {{ echoAnalysis.substat_dict?.[substat.num]?.cur_pos_percent || '' }}
      </span>

    </div>
  </div>
  <div class="target-row">
    <span class="name substat-name-cell">目标词条</span>
    <div class="target-controls">
      <button
          class="target_button"
          v-for="substat in SUBSTAT"
          :key="substat"
          @click="toggleTargetSubstat(substat.bitmap)"
          :style="(targetSubstatBitmap & substat.bitmap ? 'background-color: yellow;' : '') + `color: ${substat.font_color}`"
      >
        {{ substat.name.substring(0, 4) }}
      </button>
      <input
          class="button target-input"
          type="date"
          v-model="template.substat_since_date"
          @input="setSubstatSinceDate(template.substat_since_date)"
      />
      <span
          class="target-hint"
          :title="'仅影响“当前玩家”统计；统计范围为所选日期 04:00 之后的词条记录。'">
        ?
      </span>
    </div>
  </div>
  <div class="target-stats-table-wrap">
    <table class="my-table" style="margin-bottom: 5px;">
      <thead>
      <tr>
        <th rowspan="2">玩家</th>
        <th colspan="3">
          目标
          <span style="color: gray;">(默认未出货时回收调谐器&密音筒)</span>
        </th>
        <th colspan="2">调谐器</th>
        <th colspan="2">金密音筒</th>
      </tr>
      <tr>
        <th rowspan="2">总数</th>
        <th>平均间隔声骸</th>
        <th>平均间隔词条</th>
        <th rowspan="2">总消耗</th>
        <th rowspan="2">平均消耗</th>
        <th rowspan="2">总消耗</th>
        <th rowspan="2">平均消耗</th>
      </tr>
      </thead>
      <tbody>
      <tr>
        <td>
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
  </div>
  <span style="margin-bottom: 5px;"></span>
</template>

<script>
import axios from 'axios'
import { API_BASE_URL,
  CLASS_COLORS,
  CLASSES, ECHO_COST,
  getSubstatColor, RESONATORS,
  SUBSTAT,
  SUBSTAT_VALUE_MAP
} from '@/stores/constants.ts'
import {onMounted, ref, computed, watch} from 'vue'
import emitter from '@/stores/eventBus.js'
import {useRoute, useRouter} from 'vue-router';
import {authState} from '@/auth'

const MASK = 0b1111111111111;
const SUBSTAT_BIT_WIDTH = 13
export default {
  name: 'Echo',
  computed: {
    CLASS_COLORS() {
      return CLASS_COLORS
    }
  },
  props: {},
  created() {
    emitter.on('getEchoLog', (echoId) => {
      this.getEchoLog(echoId)
    })
  },
  setup: function (props) {
    const router = useRouter();
    const route = useRoute();
    const currentOperatorId = computed(() => authState.user?.id ?? null)
    const canManage = computed(() => authState.user?.permissions?.includes('manage') ?? false)
    const canModify = computed(() =>
      canManage.value || (currentOperatorId.value != null && echoLog.value?.operator_id === currentOperatorId.value)
    )
    // 允许创建/结束的条件：
    // - 没有绑定声骸 id（新建）
    // - 当前用户是创建者（canModify）
    // - 当前声骸已结束（pos === 5）
    // - 记录没有 operator_id（旧数据或未记操作者）
    const canCreate = computed(() =>
      !echoLog.value.id ||
      canModify.value ||
      echoLog.value.pos === 5 ||
      !echoLog.value.operator_id
    )
    const updateQueryParam = (key, value) => {
      router.replace({
        query: {
          ...route.query,
          [key]: value,
        },
      });
    }

    const scoreTemplate = ref({
      resonator: route.query.resonator || '',
      cost: route.query.cost || '',
    })
    const setResonator = (resonator) => {
      updateQueryParam('resonator', resonator)
      scoreTemplate.value.resonator = resonator
      fetchEchoAnalysis()
    }
    const setCost = (cost) => {
      updateQueryParam('cost', cost)
      scoreTemplate.value.cost = cost
      fetchEchoAnalysis()
    }

    const getEchoLog = async (echoId = 0, options = {}) => {
      const { silent = false } = options
      const targetEchoId = Number(echoId ? echoId : echoLog.value.id)
      try {
        const response = await axios.get(`${API_BASE_URL}/echo_log/${targetEchoId}?resonator=${route.query.resonator}&cost=${route.query.cost}`)
        console.log("get echo log:", response.data) // DEBUG
        if (response.data.code === 200) {
          if (!canManage.value && currentOperatorId.value != null && response.data.data?.operator_id !== currentOperatorId.value) {
            clearEchoEditor()
            return false
          }
          echoLog.value = response.data.data
          template.value.user_id = echoLog.value.user_id
          template.value.clazz = echoLog.value.clazz
          updateQueryParam('user_id', echoLog.value.user_id)
          updateQueryParam('clazz', echoLog.value.clazz)
          echoLog.value.pos = 0
          if (echoLog.value.substat1 > 0) {
          echoLog.value.pos = 1
        }
        if (echoLog.value.substat2 > 0) echoLog.value.pos = 2
        if (echoLog.value.substat3 > 0) echoLog.value.pos = 3
        if (echoLog.value.substat4 > 0) echoLog.value.pos = 4
        if (echoLog.value.substat5 > 0) echoLog.value.pos = 5
        updateQueryParam('echo_id', echoLog.value.id)
        refreshEchoLogsAnalysis()
        fetchEchoAnalysis()
        return true
        } else {
          if (silent || (targetEchoId <= 0 && response.data.message === 'echo log not found')) {
            clearEchoEditor()
            return false
          }
          alert('获取声骸失败')
          return false
        }
      } catch (error) {
        if (silent || (targetEchoId <= 0 && error?.response?.data?.message === 'echo log not found')) {
          clearEchoEditor()
          return false
        }
        console.error('获取声骸 请求失败:', error)
        alert('获取声骸 请求失败')
        return false
      }
    }

    const template = ref({
      clazz: route.query.clazz || '',
      user_id: route.query.user_id || 0,
      substat_since_date: route.query.substat_since_date || '',
    })
    const getActiveUserId = () => Number(
      template.value.user_id || echoLog.value.user_id || 0
    )
    const setSubstatSinceDate = (date) => {
      template.value.substat_since_date = date
      updateQueryParam('substat_since_date', date)
      refreshEchoLogsAnalysis()
    }

    const buildEmptyEchoLog = () => {
      return {
        clazz: template.value.clazz,
        user_id: template.value.user_id,
        id: 0,
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
        pos_total: 0,
      }
    }
    const newEchoLog = () => buildEmptyEchoLog()
    const echoLog = ref(newEchoLog())
    const clearEchoEditor = () => {
      echoLog.value = buildEmptyEchoLog()
      updateQueryParam('echo_id', 0)
    }
    const emitSyncEchoLog = () => {
      if (!echoLog.value.id) {
        return
      }
      emitter.emit('syncEchoLog', {
        ...echoLog.value,
      })
    }

    const addEchoLog = async (doIfSuccess = () => {
    }) => {
      echoLog.value = newEchoLog()
      if (!echoLog.value.user_id) {
        alert('请先输入玩家ID')
        return false
      }
      if (!echoLog.value.clazz) {
        alert('请先选择套装')
        return false
      }
      try {
        const response = await axios
            .post(`${API_BASE_URL}/echo_log`, {
              user_id: echoLog.value.user_id,
              clazz: echoLog.value.clazz,
              // tuned_at: echoLog.value.tuned_at, // FIXME
            })
        console.log("create echo log: ", response.data) // DEBUG
        if (response.data.code === 200) {
          echoLog.value = {
            ...echoLog.value,
            ...response.data.data,
            pos: echoLog.value.pos,
          }
          updateQueryParam('echo_id', echoLog.value.id)
          emitSyncEchoLog()
          updateQueryParam('echo_id', echoLog.value.id)
          doIfSuccess()
          return true
        } else {
          alert('创建声骸失败')
          return false
        }
      } catch (error) {
        console.error('创建声骸 请求失败:', error)
        alert('创建声骸 请求失败')
        return false
      }
    }

    const updateEchoLog = async (doIfSuccess = () => {
    }) => {
      if (!echoLog.value.id) {
        return false
      }
      try {
        const response = await axios
            .patch(`${API_BASE_URL}/echo_log`, {
              id: echoLog.value.id,
              substat1: echoLog.value.substat1,
              substat2: echoLog.value.substat2,
              substat3: echoLog.value.substat3,
              substat4: echoLog.value.substat4,
              substat5: echoLog.value.substat5,
              substat_all: echoLog.value.substat_all,
              s1_desc: echoLog.value.s1_desc,
              s2_desc: echoLog.value.s2_desc,
              s3_desc: echoLog.value.s3_desc,
              s4_desc: echoLog.value.s4_desc,
              s5_desc: echoLog.value.s5_desc,
              clazz: echoLog.value.clazz,
              user_id: echoLog.value.user_id,
              // tuned_at: echoLog.value.tuned_at, // FIXME
            })
        console.log("update echo log:", response.data) // DEBUG
        if (response.data.code === 200) {
          doIfSuccess()
          return true
        } else {
          alert('更新声骸记录失败')
          return false
        }
      } catch (error) {
        console.error('更新声骸记录 请求失败:', error)
        alert('更新声骸记录 请求失败')
        return false
      }
    }

    const delSubstatLog = (echoId, pos) => {
      if (echoId <= 0 || pos < 0 || pos > 4) {
        return
      }
      axios
          .delete(`${API_BASE_URL}/echo_log/${echoId}/substat_pos/${pos}`)
          .then((response) => {
            console.log('delSubstatLog:', response.data) // DEBUG
            const code = response.data.code
            if (code === 200) {
              // alert('删除成功');
            } else {
              alert('跟据echoId和pos删除词条失败')
            }
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
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
    const refreshEchoLogsAnalysis = (size = 0) => {
      const activeUserId = getActiveUserId()
      if (activeUserId > 0) {
        axios
            .get(`${API_BASE_URL}/echo_logs/analysis?size=${size}&user_id=${activeUserId}&target_bits=${targetSubstatBitmap.value}&substat_since_date=${template.value.substat_since_date}`)
            .then((response) => {
              console.log('current user: ', response.data) // DEBUG
              currentUser.value = response.data.data
            })
            .catch((error) => {
              console.error('请求失败:', error)
            })
      } else {
        currentUser.value = {
          target_echo_distance: -1,
          target_substat_distance: -1,
          target: 0,
          target_avg_echo: 0.0,
          target_avg_substat: 0.0,
          tuner_consumed: 0,
          tuner_consumed_avg: 0.0,
          exp_consumed: 0,
          exp_consumed_avg: 0.0,
        }
      }
      axios
          .get(`${API_BASE_URL}/echo_logs/analysis?size=${size}&target_bits=${targetSubstatBitmap.value}`)
          .then((response) => {
            console.log('all users: ', response.data) // DEBUG
            allUsers.value = response.data.data
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
    }
    onMounted(refreshEchoLogsAnalysis)

    const targetSubstatBitmap = ref(0b11)
    const toggleTargetSubstat = (bitmap) => {
      targetSubstatBitmap.value ^= bitmap
      refreshEchoLogsAnalysis()
    }

    const setUserId = (userId) => {
      updateQueryParam('user_id', userId)
      template.value.user_id = userId
      echoLog.value.user_id = userId
      refreshEchoLogsAnalysis()
      refreshRecentTuneStats()
      emitter.emit("setUserId", userId)
      if (echoLog.value.id > 0 && canModify.value) {
        emitSyncEchoLog()
        updateEchoLog(() => emitter.emit('refreshEchoLogs'))
      }
    }
    const setClazz = (clazz) => {
      updateQueryParam('clazz', clazz)
      template.value.clazz = clazz
      echoLog.value.clazz = clazz
      emitter.emit("setClazz", clazz)
      if (echoLog.value.id > 0 && canModify.value) {
        emitSyncEchoLog()
        updateEchoLog(() => emitter.emit('refreshEchoLogs'))
      }
    }
    const setEchoId = (id) => {
      updateQueryParam('echo_id', id)
      echoLog.value.id = id
      getEchoLog()
    }
    const hasInitializedEchoEditor = ref(false)
    const initEchoEditor = () => {
      if (hasInitializedEchoEditor.value || currentOperatorId.value == null) {
        return
      }
      hasInitializedEchoEditor.value = true
      const initialEchoId = Number(route.query.echo_id || 0)
      getEchoLog(initialEchoId > 0 ? initialEchoId : 0, { silent: true })
    }
    onMounted(initEchoEditor)
    watch(currentOperatorId, (nextOperatorId, previousOperatorId) => {
      if (nextOperatorId !== previousOperatorId) {
        hasInitializedEchoEditor.value = false
      }
      if (
        !canManage.value &&
        currentOperatorId.value != null &&
        echoLog.value.id > 0 &&
        echoLog.value.operator_id !== currentOperatorId.value
      ) {
        clearEchoEditor()
      }
      initEchoEditor()
    })
    const setPos = (pos) => {
      if (!canCreate.value) {
        return
      }
      if (echoLog.value.id < 0) {
        return
      }
      if (pos === 5) {
        if (echoLog.value.pos === 5) {
          addEchoLog()
          return
        }
        echoLog.value.pos = 5
        return
      }
      switch (pos) {
        case 0:
          if (echoLog.value.substat1 > 0) {
            delSubstatLog(echoLog.value.id, 0)
            echoLog.value.substat1 = 0
            echoLog.value.s1_desc = ''
          }
          break
        case 1:
          if (echoLog.value.substat2 > 0) {
            delSubstatLog(echoLog.value.id, 1)
            echoLog.value.substat2 = 0
            echoLog.value.s2_desc = ''
          }
          break
        case 2:
          if (echoLog.value.substat3 > 0) {
            delSubstatLog(echoLog.value.id, 2)
            echoLog.value.substat3 = 0
            echoLog.value.s3_desc = ''
          }
          break
        case 3:
          if (echoLog.value.substat4 > 0) {
            delSubstatLog(echoLog.value.id, 3)
            echoLog.value.substat4 = 0
            echoLog.value.s4_desc = ''
          }
          break
        case 4:
          if (echoLog.value.substat5 > 0) {
            delSubstatLog(echoLog.value.id, 4)
            echoLog.value.substat5 = 0
            echoLog.value.s5_desc = ''
          }
          break
        default:
          alert('请先选择孔位')
          return
      }
      echoLog.value.substat_all = (
          echoLog.value.substat1 |
          echoLog.value.substat2 |
          echoLog.value.substat3 |
          echoLog.value.substat4 |
          echoLog.value.substat5
      ) & MASK

      echoLog.value.pos = pos
      emitSyncEchoLog()
      updateEchoLog(() => {
        emitter.emit('refreshEchoLogs')
        emitter.emit('refreshSubstatLogs')
      })
    }

    const addSubstatLog = async (substat, value, position) => {
      if (!echoLog.value.id) {
        alert('请先创建声骸')
        return false
      }
      try {
        const response = await axios
            .post(`${API_BASE_URL}/tune_log`, {
              substat: substat,
              value: value,
              position: position,
              echo_id: echoLog.value.id,
              user_id: echoLog.value.user_id,
            })
        console.log(response.data) // DEBUG
        const code = response.data.code
        if (code === 200) {
          return true
        } else {
          alert('添加调谐记录失败')
          return false
        }
      } catch (error) {
        console.error('添加调谐记录 请求失败:', error)
        alert('添加调谐记录 请求失败')
        return false
      }
    }

    const echoAnalysis = ref({
      data_total: 0,
      substat_dict: {},
      position_total: [0, 0, 0, 0, 0],
      substat_pos_total: [[0, 0, 0, 0, 0],],
      score: {
        substat1: 0,
        substat2: 0,
        substat3: 0,
        substat4: 0,
        substat5: 0,
        substat_all: 0,
      },
      two_crit_percent: 0.0,
    })
    const fetchEchoAnalysis = () => {
      axios
          .post(`${API_BASE_URL}/analyze_echo?resonator=${scoreTemplate.value.resonator}&cost=${scoreTemplate.value.cost}`, {
            ...echoLog.value
          })
          .then((response) => {
            console.log("analyze_echo:", response.data) // DEBUG
            echoAnalysis.value = response.data.data
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
    }
    onMounted(fetchEchoAnalysis)

    const doTune = async (substat, value) => {
      if (!echoLog.value.id) {
        if (!echoLog.value.user_id) {
          alert('请先输入玩家ID')
          return
        }
        if (!echoLog.value.clazz) {
          alert('请先选择套装')
          return
        }
      }
      if (echoLog.value.pos === 5) {
        alert('当前声骸已结束，请先创建新声骸，或手动选择要修改的孔位')
        return
      }
      if ((1 << substat) & (
          (echoLog.value.pos !== 0 ? echoLog.value.substat1 : 0) |
          (echoLog.value.pos !== 1 ? echoLog.value.substat2 : 0) |
          (echoLog.value.pos !== 2 ? echoLog.value.substat3 : 0) |
          (echoLog.value.pos !== 3 ? echoLog.value.substat4 : 0) |
          (echoLog.value.pos !== 4 ? echoLog.value.substat5 : 0)
      )) {
        alert('已存在相同词条，请检查')
        return
      }

      const tunePos = echoLog.value.pos
      console.log('add tune log, echo_id:', echoLog.value.id, ', pos:', tunePos, ', substat:', substat, ', value:', value)
      const substatDesc = SUBSTAT_VALUE_MAP[substat][value].desc_full
      const nextEchoLog = {
        ...echoLog.value,
      }
      switch (tunePos) {
        case 0:
          nextEchoLog.substat1 = 1 << substat | 1 << (value + 13)
          nextEchoLog.s1_desc = substatDesc
          break
        case 1:
          nextEchoLog.substat2 = 1 << substat | 1 << (value + 13)
          nextEchoLog.s2_desc = substatDesc
          break
        case 2:
          nextEchoLog.substat3 = 1 << substat | 1 << (value + 13)
          nextEchoLog.s3_desc = substatDesc
          break
        case 3:
          nextEchoLog.substat4 = 1 << substat | 1 << (value + 13)
          nextEchoLog.s4_desc = substatDesc
          break
        case 4:
          nextEchoLog.substat5 = 1 << substat | 1 << (value + 13)
          nextEchoLog.s5_desc = substatDesc
          break
        default:
          alert('请先选择孔位')
          return
      }

      nextEchoLog.substat_all = (
          nextEchoLog.substat1 |
          nextEchoLog.substat2 |
          nextEchoLog.substat3 |
          nextEchoLog.substat4 |
          nextEchoLog.substat5
      ) & MASK
      try {
        const response = await axios.post(`${API_BASE_URL}/echo_log/tune`, {
          id: nextEchoLog.id,
          user_id: nextEchoLog.user_id,
          clazz: nextEchoLog.clazz,
          substat1: nextEchoLog.substat1,
          substat2: nextEchoLog.substat2,
          substat3: nextEchoLog.substat3,
          substat4: nextEchoLog.substat4,
          substat5: nextEchoLog.substat5,
          substat_all: nextEchoLog.substat_all,
          s1_desc: nextEchoLog.s1_desc,
          s2_desc: nextEchoLog.s2_desc,
          s3_desc: nextEchoLog.s3_desc,
          s4_desc: nextEchoLog.s4_desc,
          s5_desc: nextEchoLog.s5_desc,
          position: tunePos,
          substat,
          value,
        })
        console.log('tune echo log:', response.data) // DEBUG
        if (response.data.code !== 200) {
          alert('添加调谐记录失败')
          return
        }

        const savedEchoLog = response.data.data.echo_log
        savedEchoLog.pos = 0
        if (savedEchoLog.substat1 > 0) savedEchoLog.pos = 1
        if (savedEchoLog.substat2 > 0) savedEchoLog.pos = 2
        if (savedEchoLog.substat3 > 0) savedEchoLog.pos = 3
        if (savedEchoLog.substat4 > 0) savedEchoLog.pos = 4
        if (savedEchoLog.substat5 > 0) savedEchoLog.pos = 5
        echoLog.value = savedEchoLog
        updateQueryParam('echo_id', echoLog.value.id)
        emitSyncEchoLog()
        fetchEchoAnalysis()
        emitter.emit('refreshEchoLogs')
        emitter.emit('refreshSubstatLogs')
        refreshRecentTuneStats()
        refreshEchoLogsAnalysis()
      } catch (error) {
        console.error('调谐声骸 请求失败:', error)
        alert('添加调谐记录失败')
      }
    }

    // 展示最近各词条出现的数量
    const recentTuneStats = ref({
      data_total: 0,
      substat_dict: {},
      substat_distance: [],
    })
    const refreshRecentTuneStats = (size = 39) => {
      axios
          .get(`${API_BASE_URL}/tune_stats?size=${size}&user_id=${template.value.user_id}`)
          .then((response) => {
            console.log(response.data) // DEBUG
            recentTuneStats.value = response.data.data
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
    }
    onMounted(refreshRecentTuneStats)

    const getRecentValueCount = (substatNum, valueNum) =>
      recentTuneStats.value.substat_dict?.[substatNum]?.value_dict?.[valueNum]?.total ?? 0

    const bitPos = (value) => {
      if (!value) {
        return -1
      }
      let pos = 0
      while (((value >> pos) & 1) === 0) {
        pos += 1
      }
      return pos
    }

    const getSelectedSubstats = () =>
      [echoLog.value.substat1, echoLog.value.substat2, echoLog.value.substat3, echoLog.value.substat4, echoLog.value.substat5]
        .filter((substat) => substat > 0)

    const getSubstatFullName = (substatNum) =>
      SUBSTAT_VALUE_MAP[substatNum]?.[0]?.desc_full?.replace(/\s.*$/, '') ?? ''

    const getSubstatNumericValue = (substatNum, valueNum) => {
      const desc = SUBSTAT_VALUE_MAP[substatNum]?.[valueNum]?.desc ?? '0'
      return Number.parseFloat(String(desc).replace('%', '')) || 0
    }

    const getResonatorTemplate = () => echoAnalysis.value.resonator_template ?? {}

    const getCurrentCost = () => scoreTemplate.value.cost || '1C'

    const getEchoMaxScoreBase = () => {
      const template = getResonatorTemplate()
      const cost = getCurrentCost()
      return Number(template.echo_max_score?.[String(cost).slice(0, 1)] ?? 0)
    }

    const getMainstatBaseScore = () => {
      const template = getResonatorTemplate()
      const cost = getCurrentCost()
      return Number(template.mainstat_max_score?.[cost] ?? 0)
    }

    const getSubstatWeight = (substatNum) => {
      const template = getResonatorTemplate()
      const fullName = getSubstatFullName(substatNum)
      return Number(template.substat_weight?.[fullName] ?? 0)
    }

    const getScaledSubstatScore = (substatNum, valueNum) => {
      const echoMaxScoreBase = getEchoMaxScoreBase()
      if (echoMaxScoreBase <= 0) {
        return 0
      }
      const numericValue = getSubstatNumericValue(substatNum, valueNum)
      const weight = getSubstatWeight(substatNum)
      return weight * numericValue / echoMaxScoreBase * 50
    }

    const getCurrentSubstatScoreTotal = () =>
      getSelectedSubstats().reduce((total, substatBits) => {
        const substatNum = bitPos(substatBits & MASK)
        const valueNum = bitPos(substatBits >> SUBSTAT_BIT_WIDTH)
        if (substatNum < 0 || valueNum < 0) {
          return total
        }
        return total + getScaledSubstatScore(substatNum, valueNum)
      }, 0)

    const getRemainingPotentialScoreTotal = () => {
      const selectedNums = new Set(
        getSelectedSubstats()
          .map((substatBits) => bitPos(substatBits & MASK))
          .filter((substatNum) => substatNum >= 0)
      )
      const remainingSlots = Math.max(0, 5 - selectedNums.size)
      if (remainingSlots <= 0) {
        return 0
      }
      const candidates = SUBSTAT
        .filter((substat) => !selectedNums.has(substat.num))
        .map((substat) => {
          const values = SUBSTAT_VALUE_MAP[substat.num] ?? []
          const maxValueNum = values.length - 1
          return maxValueNum >= 0 ? getScaledSubstatScore(substat.num, maxValueNum) : 0
        })
        .sort((a, b) => b - a)
      return candidates.slice(0, remainingSlots).reduce((sum, score) => sum + score, 0)
    }

    const getPotentialMaxScore = () => {
      const mainstatBaseScore = getMainstatBaseScore()
      const total = mainstatBaseScore + getCurrentSubstatScoreTotal() + getRemainingPotentialScoreTotal()
      return total > 0 ? total.toFixed(2) : ''
    }

    const showPotentialMaxScore = () => !!getPotentialMaxScore()

    const getRecentSubstatTotal = (substatNum) =>
      recentTuneStats.value.substat_dict?.[substatNum]?.total ?? 0

    const getRecentRawDistance = (substatNum) =>
      recentTuneStats.value.substat_distance?.[substatNum] ?? -1

    const getRecentDistanceLimit = () =>
      Math.max(recentTuneStats.value.data_total ?? 0, 0)

    const isRecentDistanceOverflow = (substatNum) =>
      getRecentRawDistance(substatNum) < 0

    const getRecentDistanceDisplay = (substatNum) => {
      const distance = getRecentRawDistance(substatNum)
      if (distance >= 0) {
        return distance
      }
      return getRecentDistanceLimit()
    }

    const getRecentDistanceColor = (substatNum) => {
      const distance = getRecentDistanceDisplay(substatNum)
      if (isRecentDistanceOverflow(substatNum) || distance > 26) {
        return 'color: red;'
      }
      if (distance > 13) {
        return 'color: orange;'
      }
      return ''
    }

    const getRecentDistanceTitle = (substatNum) => {
      const limit = getRecentDistanceLimit()
      if (!isRecentDistanceOverflow(substatNum)) {
        return `最近 ${limit} 个词条内，距离上次出现为 ${getRecentDistanceDisplay(substatNum)}`
      }
      return `最近 ${limit} 个词条统计范围内未出现，实际间隔大于 ${limit}`
    }

    const getRecentGlobalMaxValueCount = () => {
      const counts = SUBSTAT.flatMap((substat) =>
        SUBSTAT_VALUE_MAP[substat.num].map((value) => getRecentValueCount(substat.num, value.value_number))
      )
      const maxCount = Math.max(...counts, 0)
      return maxCount > 0 ? maxCount : 1
    }

    const getRecentValueBarStyle = (substatNum, valueNum, color) => {
      const count = getRecentValueCount(substatNum, valueNum)
      const maxCount = getRecentGlobalMaxValueCount()
      const width = count > 0 ? Math.max(10, Math.round(count * 100 / maxCount)) : 0
      return {
        width: `${width}%`,
        backgroundColor: color,
        opacity: count > 0 ? 0.85 : 0,
      }
    }

    const getRecentSubstatTotalBarStyle = (substatNum, color) => {
      const total = getRecentSubstatTotal(substatNum)
      const maxTotal = Math.max(...SUBSTAT.map((substat) => getRecentSubstatTotal(substat.num)), 0) || 1
      const width = total > 0 ? Math.max(10, Math.round(total * 100 / maxTotal)) : 0
      return {
        width: `${width}%`,
        backgroundColor: color,
        opacity: total > 0 ? 0.55 : 0,
      }
    }

    const tuneStats = ref({
      data_total: 0,
      substat_dict: {},
      position_total: [0, 0, 0, 0, 0],
      substat_pos_total: [[0, 0, 0, 0, 0],],
    })
    const fetchTuneStats = () => {
      axios
          .get(`${API_BASE_URL}/tune_stats?`)
          .then((response) => {
            console.log("tune_stats:", response.data) // DEBUG
            tuneStats.value = response.data.data
          })
          .catch((error) => {
            console.error('请求失败:', error)
          })
    }
    onMounted(fetchTuneStats)

    return {
      targetSubstatBitmap,
      echoLog,
      canModify,
      canCreate,
      recentTuneStats,
      currentUser,
      allUsers,
      tuneStats,
      echoAnalysis,
      scoreTemplate,
      template,
      setResonator,
      setCost,
      setPos,
      setClazz,
      setUserId,
      setEchoId,
      getEchoLog,
      addEchoLog,
      doTune,
      getRecentValueCount,
      getPotentialMaxScore,
      getRecentValueBarStyle,
      getRecentSubstatTotal,
      getRecentDistanceColor,
      getRecentDistanceDisplay,
      getRecentDistanceTitle,
      getRecentSubstatTotalBarStyle,
      getRecentRawDistance,
      getRecentDistanceLimit,
      isRecentDistanceOverflow,
      showPotentialMaxScore,
      getSubstatColor,
      toggleTargetSubstat,
      setSubstatSinceDate,
      CLASSES,
      SUBSTAT,
      SUBSTAT_VALUE_MAP,
      RESONATORS,
      ECHO_COST,
    }
  },
}
</script>

<style scoped>
.name {
  font-size: medium;
  display: inline-block;
  text-align-all: center;
  width: 12.5%;
}

.substat-name-cell {
  width: 9.5%;
  min-width: 84px;
}

.suite-note {
  font-size: 12px;
}

.button {
  font-size: medium;
  display: inline-block;
  min-width: 30px;
  max-width: 120px;
  width: 10%;
  height: 35px;
  text-align: center;
}

.template-row {
  display: flex;
  align-items: flex-start;
  gap: 0;
}

.player-info-row {
  display: flex;
  align-items: center;
  gap: 0;
  margin-bottom: 4px;
}

.player-info-input {
  min-width: 90px;
  text-align: center;
  font-weight: 700;
}

.echo-id-input {
  min-width: 80px;
}

.player-info-label {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 80px;
  font-weight: 700;
}

.player-info-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 35px;
  padding: 0 10px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.04);
  font-weight: 700;
  white-space: nowrap;
}

.player-stats-text {
  font-size: medium;
  white-space: nowrap;
}

.template-label-button {
  flex: 0 0 44px;
  width: 44px;
  min-width: 44px;
  max-width: 44px;
  height: 100px;
  color: red;
  font-size: medium;
}

.template-scroll {
  display: flex;
  flex: 1;
  gap: 0;
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 560px;
  padding-bottom: 6px;
}

.template-button {
  flex: 0 0 44px;
  width: 44px;
  min-width: 44px;
  max-width: 44px;
  height: 100px;
}

.template-cost-group {
  display: flex;
  flex: 0 0 auto;
  gap: 0;
}

.template-cost-label {
  width: 45px;
  min-width: 45px;
  max-width: 45px;
  height: 100px;
  color: red;
  font-size: medium;
}

.template-cost-button {
  flex: 0 0 40px;
  width: 40px;
  min-width: 40px;
  max-width: 40px;
  height: 100px;
}

.substat-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.suite-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.top-summary-track {
  display: flex;
  flex: 1 1 auto;
  width: auto;
  min-width: 0;
  gap: 10px;
}

.substat-summary-buttons {
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
}

.suite-scroll {
  display: flex;
  flex: 0 1 616px;
  width: 616px;
  gap: 0;
  overflow-x: auto;
  overflow-y: hidden;
  max-width: calc(100% - 104px);
  padding-bottom: 6px;
}

.suite-button {
  flex: 0 0 44px;
  width: 44px;
  min-width: 44px;
  max-width: 44px;
  height: 100px;
}

.score-panel-slot {
  display: flex;
  flex: 0 0 96px;
  justify-content: flex-end;
}

.suite-score-panels {
  gap: 8px;
}

.score-panel {
  display: inline-flex;
  flex: 0 0 auto;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 88px;
  height: 100px;
  padding: 0 8px;
  border-radius: 10px;
  background: rgba(0, 0, 0, 0.04);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.3;
  text-align: center;
  color: #555;
}

.suite-score-panel,
.current-score-panel {
}

.current-score-panel {
  height: 60px;
}

.suite-score-panel-value {
  color: #c62828;
  font-size: 18px;
  font-weight: 800;
}

.target-row {
  display: flex;
  align-items: center;
}

.target-controls {
  display: flex;
  flex: 1;
  min-width: 0;
}

.target-stats-table-wrap {
  width: 89.5%;
  max-width: 89.5%;
}

.target_button {
  display: inline-block;
  flex: 1 1 0;
  min-width: 0;
  max-width: 48px;
  height: 40px;
  text-align: center;
}

.target-input {
  flex: 0 0 140px;
  width: 140px;
  min-width: 140px;
  max-width: 140px;
  height: 40px;
  text-align: center;
  font-weight: 700;
}

.stat-button,
.stat-summary-button {
  position: relative;
  width: 10%;
  min-width: 52px;
  max-width: 120px;
  height: 40px;
  overflow: hidden;
}

.stat-button-label {
  position: relative;
  z-index: 2;
  font-weight: 700;
}

.stat-button-count {
  position: absolute;
  right: 4px;
  top: 3px;
  z-index: 2;
  min-width: 18px;
  padding: 1px 4px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0, 0, 0, 0.12);
  font-size: 12px;
  font-weight: 800;
  line-height: 1.1;
  color: #222;
  text-align: center;
}

.button-mini-bar-shell {
  position: relative;
  width: calc(100% - 8px);
  height: 6px;
  margin: 4px auto 0;
  border-radius: 999px;
  background: #ececec;
  overflow: hidden;
}

.summary-bar-shell {
  background: #e2e2e2;
}

.name-mini-bar-shell {
  width: calc(100% - 12px);
  margin-top: 3px;
}

.substat-summary-text {
  line-height: 1.15;
  white-space: nowrap;
}

.substat-name-label {
  font-weight: 800;
}

.substat-distance-text {
  font-size: 12px;
}

.distance-overflow-mark {
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
}

.substat-current-position-rate {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 52px;
  margin-left: 6px;
  font-size: 12px;
  font-weight: 700;
  cursor: help;
}

.button-mini-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 160ms ease;
}

.substat {
  display: inline-block;
  flex: 1 1 0;
  min-width: 84px;
  max-width: 88px;
  width: auto;
  height: 60px;
  text-align: center;
}

.end-double-crit-rate {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 90px;
  max-width: 120px;
  width: 16%;
  height: 60px;
  font-size: 13px;
  font-weight: 700;
  color: red;
  line-height: 1.2;
  text-align: center;
}

.inline-double-crit-rate {
  margin-left: 4px;
  font-size: medium;
  font-weight: 700;
  color: red;
}

.my-table {
  box-sizing: border-box;
  width: auto;
  max-width: 100%;
  table-layout: auto;
  border-collapse: collapse; /* 关键：合并边框 */
  border: 1px solid #e0e0e0; /* 表格边框 */
  font-weight: bolder;
//font-size: medium;
}

.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 8px;
}
.target-hint {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 35px;
  height: 35px;
  margin-left: 6px;
  border-radius: 50%;
  border: 1px solid #ccc;
  font-weight: 700;
  cursor: help;
  color: #555;
  background: #fafafa;
  user-select: none;
}
</style>
