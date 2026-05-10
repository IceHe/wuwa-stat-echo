<template>
  <h1>{{ title }}</h1>
  <div class="find-panel">
    <div class="find-toolbar-row">
      <span class="name">玩家ID</span>
      <input
        class="button user-id-input"
        type="text"
        v-model="echoLog.user_id"
        placeholder="当前玩家ID"
        @change="setUserId(echoLog.user_id)" />
      <button class="button clear-button" @click="resetEchoLog()">清空</button>
      <input
        class="button keyword-input"
        type="text"
        v-model.trim="echoLog.keyword"
        placeholder="玩家ID / 声骸ID / 套装 / 词条"
        @keyup.enter="findEchoLog(false)" />
      <span class="clazz-chip" :style="`color: ${CLASS_COLORS[echoLog.clazz]};`">
        {{ echoLog.clazz.substring(0, 4) }}
      </span>
      <button class="button search-button" @click="findEchoLog(false)">搜索</button>
    </div>
    <div class="find-toolbar-row">
      <span class="name">搜索方式</span>
      <div class="mode-switch">
        <button
          class="button mode-button"
          @click="setSearchMode(SEARCH_MODE.POSITIONAL)"
          :class="{ 'mode-active': echoLog.search_mode === SEARCH_MODE.POSITIONAL }"
        >
          孔位搜索
        </button>
        <button
          class="button mode-button"
          @click="setSearchMode(SEARCH_MODE.SUBSTAT_SET)"
          :class="{ 'mode-active': echoLog.search_mode === SEARCH_MODE.SUBSTAT_SET }"
        >
          副词条搜索
        </button>
      </div>
    </div>
    <div v-if="echoLog.search_mode === SEARCH_MODE.SUBSTAT_SET" class="find-mode-hint">
      忽略孔位和档位，仅按包含哪些副词条搜索
    </div>
    <div class="suite-row">
      <span class="name">声骸套装</span>
      <div class="suite-scroll">
        <button
          class="button suite-button"
          @click="setClazz('')"
          :style="echoLog.clazz === '' ? 'background-color: yellow;' : ''"
        >
          不限
        </button>
        <button
          class="button suite-button"
          v-for="clazz in CLASSES"
          :key="clazz"
          @click="setClazz(clazz)"
          :style="echoLog.clazz === clazz ? 'background-color: yellow;' : ''"
        >
          <span :style="`color: ${CLASS_COLORS[clazz]};`"> {{ clazz.substring(0, 4) }}</span>
        </button>
      </div>
    </div>
    <template v-if="echoLog.search_mode === SEARCH_MODE.POSITIONAL">
      <div class="find-toolbar-row">
        <span class="name">当前孔位 </span>
        <div class="find-position-row">
          <button class="substat"
                  key="1"
                  @click="echoLog.pos = 0"
                  :style="echoLog.pos === 0 ? 'background-color: yellow; font-color: red' : ''"
          >
            {{ echoLog.s1_desc ? echoLog.s1_desc : "1" }}
          </button>
          <button class="substat"
                  key="2"
                  @click="echoLog.pos = 1"
                  :style="echoLog.pos === 1 ? 'background-color: yellow; font-color: red' : ''"
          >
            {{ echoLog.s2_desc ? echoLog.s2_desc : "2" }}
          </button>
          <button class="substat"
                  key="3"
                  @click="echoLog.pos = 2"
                  :style="echoLog.pos === 2 ? 'background-color: yellow; font-color: red' : ''"
          >
            {{ echoLog.s3_desc ? echoLog.s3_desc : "3" }}
          </button>
          <button class="substat"
                  key="4"
                  @click="echoLog.pos = 3"
                  :style="echoLog.pos === 3 ? 'background-color: yellow; font-color: red' : ''"
          >
            {{ echoLog.s4_desc ? echoLog.s4_desc : "4" }}
          </button>
          <button class="substat"
                  key="5"
                  @click="echoLog.pos = 4"
                  :style="echoLog.pos === 4 ? 'background-color: yellow; font-color: red' : ''"
          >
            {{ echoLog.s5_desc ? echoLog.s5_desc : "5" }}
          </button>
        </div>
      </div>
      <div v-for="substat in SUBSTAT" :key="substat.num" class="find-substat-row">
        <span class="name" :style="`color: ${substat.font_color}; font-weight: bolder;`">{{ substat.name.substring(0, 4) }}</span>
        <div class="find-substat-buttons">
          <button
            class="button compact-button"
            @click="addAnyTuneToFind(substat.num)"
            :style="`color: ${substat.font_color}`"
          >
            不限
          </button>
          <button
            class="button compact-button"
            v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
            :key="value.value_number"
            @click="addTuneToFind(value.substat_number, value.value_number)"
            :style="`color: ${substat.font_color}`"
          >
            {{ value.desc }}
          </button>
        </div>
      </div>
    </template>
    <template v-else>
      <div class="find-substat-set-row">
        <span class="name">副词条</span>
        <div class="find-set-buttons">
          <button
            v-for="substat in SUBSTAT"
            :key="substat.num"
            class="button set-substat-button"
            @click="toggleSubstatSet(substat.num)"
            :style="getSubstatSetButtonStyle(substat)"
          >
            {{ substat.name }}
          </button>
        </div>
      </div>
      <div class="find-toolbar-row">
        <span class="name">已选</span>
        <div class="selected-substat-summary">{{ selectedSubstatSummary() }}</div>
      </div>
    </template>
  </div>
  <div class="find-results">
    <button @click="findEchoLog()">{{ refreshButtonLabel }}</button>
    <table class="my-table">
      <thead>
      <tr style="text-align: left;">
        <th>
          <div>玩家/声骸</div>
          <div style="font-size: 10px; color: #888; font-weight: normal;">自己的声骸或管理员可修改</div>
        </th>
        <th>套装</th>
        <th>词条1</th>
        <th>词条2</th>
        <th>词条3</th>
        <th>词条4</th>
        <th>词条5</th>
        <th>记录于</th>
      </tr>
      </thead>
      <tbody>
      <EchoLogRow
        v-for="echoLog in echoLogs"
        :key="echoLog.id + echoLog.updated_at + echoLog.deleted"
        :echoLog="echoLog"
        :operatorId="operatorId"
        :canManage="canManage"
        :showActions="false"
        :showScore="false"
      />
      </tbody>
    </table>
  </div>
</template>

<script>
import axios from 'axios'
import {API_BASE_URL, CLASS_COLORS, CLASSES, SUBSTAT, SUBSTAT_VALUE_MAP} from '@/stores/constants.ts'
import {computed, onMounted, ref} from 'vue'
import {useRoute} from 'vue-router';
import EchoLogRow from "@/components/EchoLogRow.vue";
import emitter from "@/stores/eventBus.js";
import {authState} from '@/auth'

const SEARCH_MODE = {
  POSITIONAL: 'positional',
  SUBSTAT_SET: 'substat_set',
}

export default {
  name: 'Find Echo',
  props: {
    title: {
      type: String,
      required: false,
      default: '搜索声骸',
    },
    refreshButtonLabel: {
      type: String,
      required: false,
      default: '声骸搜索列表 - 刷新',
    },
    defaultOperatorScope: {
      type: String,
      required: false,
      default: 'all',
    },
    allowEmptyQuery: {
      type: Boolean,
      required: false,
      default: false,
    },
    autoLoad: {
      type: Boolean,
      required: false,
      default: false,
    },
    syncFromEditor: {
      type: Boolean,
      required: false,
      default: true,
    },
    refreshEventName: {
      type: String,
      required: false,
      default: '',
    },
  },
  computed: {
    CLASS_COLORS() {
      return CLASS_COLORS
    }
  },
  components: {EchoLogRow},
  created() {
    if (this.syncFromEditor) {
      this.handleSetUserId = (id) => {
        this.setUserId(id)
      }
      this.handleSetClazz = (clazz) => {
        this.setClazz(clazz)
      }
      emitter.on('setUserId', this.handleSetUserId)
      emitter.on('setClazz', this.handleSetClazz)
    }
    if (this.refreshEventName) {
      this.handleRefreshEvent = () => {
        this.findEchoLog(false)
      }
      emitter.on(this.refreshEventName, this.handleRefreshEvent)
    }
  },
  beforeUnmount() {
    if (this.handleSetUserId) {
      emitter.off('setUserId', this.handleSetUserId)
    }
    if (this.handleSetClazz) {
      emitter.off('setClazz', this.handleSetClazz)
    }
    if (this.handleRefreshEvent && this.refreshEventName) {
      emitter.off(this.refreshEventName, this.handleRefreshEvent)
    }
  },
  setup: function (props) {
    const route = useRoute();
    const normalizeUserId = (userId) => {
      if (userId === '' || userId === null || userId === undefined) {
        return 0
      }
      const normalized = Number(userId)
      return Number.isNaN(normalized) ? 0 : normalized
    }

    const buildInitialTemplate = () => ({
      clazz: props.syncFromEditor ? String(route.query.clazz || '') : '',
      user_id: props.syncFromEditor ? normalizeUserId(route.query.user_id) : 0,
    })

    const template = ref(buildInitialTemplate())

    const newEchoLog = () => ({
      clazz: template.value.clazz,
      user_id: template.value.user_id,
      id: 0,
      keyword: "",
      search_mode: SEARCH_MODE.POSITIONAL,
      substat_all_mask: 0,
      pos: 0, // 当前孔位
      substat1: 0,
      substat2: 0,
      substat3: 0,
      substat4: 0,
      substat5: 0,
      s1_desc: "",
      s2_desc: "",
      s3_desc: "",
      s4_desc: "",
      s5_desc: "",
      // tuned_at: new Date().toISOString(), // FIXME
    })

    const echoLog = ref(newEchoLog())
    const resetEchoLog = () => {
      template.value = buildInitialTemplate()
      echoLog.value = newEchoLog()
      if (props.allowEmptyQuery) {
        findEchoLog(false)
        return
      }
      echoLogs.value = []
    }

    const setUserId = (userId) => {
      const normalizedUserId = normalizeUserId(userId)
      template.value.user_id = normalizedUserId
      echoLog.value.user_id = normalizedUserId
      findEchoLog(false)
    }
    const setClazz = (clazz) => {
      template.value.clazz = clazz
      echoLog.value.clazz = clazz
      findEchoLog(false)
    }

    const setSearchMode = (mode) => {
      if (echoLog.value.search_mode === mode) {
        return
      }
      echoLog.value.search_mode = mode
      findEchoLog(false)
    }

    const buildFindPayload = () => {
      const searchMode = echoLog.value.search_mode === SEARCH_MODE.SUBSTAT_SET
        ? SEARCH_MODE.SUBSTAT_SET
        : SEARCH_MODE.POSITIONAL

      return {
        id: Number(echoLog.value.id || 0),
        user_id: normalizeUserId(echoLog.value.user_id),
        clazz: echoLog.value.clazz,
        keyword: echoLog.value.keyword?.trim() || '',
        operator_scope: props.defaultOperatorScope,
        allow_empty_query: props.allowEmptyQuery,
        search_mode: searchMode,
        substat_all_mask: searchMode === SEARCH_MODE.SUBSTAT_SET ? echoLog.value.substat_all_mask : 0,
        substat1: searchMode === SEARCH_MODE.POSITIONAL ? echoLog.value.substat1 : 0,
        substat2: searchMode === SEARCH_MODE.POSITIONAL ? echoLog.value.substat2 : 0,
        substat3: searchMode === SEARCH_MODE.POSITIONAL ? echoLog.value.substat3 : 0,
        substat4: searchMode === SEARCH_MODE.POSITIONAL ? echoLog.value.substat4 : 0,
        substat5: searchMode === SEARCH_MODE.POSITIONAL ? echoLog.value.substat5 : 0,
      }
    }

    const findEchoLog = async (nextPos = true) => {
      try {
        const response = await axios.post(`${API_BASE_URL}/echo_log/find?page_size=20`, buildFindPayload())
        console.log('listEchoLog: ', response.data)
        if (response.data.code === 200) {
          echoLogs.value = response.data.data
          if (nextPos && echoLog.value.search_mode === SEARCH_MODE.POSITIONAL && echoLog.value.pos < 4) {
            echoLog.value.pos++
          }
        } else {
          console.error('查询声骸 失败:', response.data)
          alert('查询声骸 失败')
        }
      } catch (error) {
        console.error('查询声骸 请求失败:', error)
      }
    }

    const echoLogs = ref([])
    onMounted(() => {
      if (props.autoLoad) {
        findEchoLog(false)
      }
    })
    const operatorId = computed(() => authState.user?.id ?? null)
    const canManage = computed(() => authState.user?.permissions?.includes('manage') ?? false)
    const hasDuplicatedSubstat = (substat) => (
      (1 << substat) & (
        (echoLog.value.pos !== 0 ? echoLog.value.substat1 : 0) |
        (echoLog.value.pos !== 1 ? echoLog.value.substat2 : 0) |
        (echoLog.value.pos !== 2 ? echoLog.value.substat3 : 0) |
        (echoLog.value.pos !== 3 ? echoLog.value.substat4 : 0) |
        (echoLog.value.pos !== 4 ? echoLog.value.substat5 : 0)
      )
    )

    const isSubstatSetSelected = (substat) => ((echoLog.value.substat_all_mask >> substat) & 1) === 1

    const toggleSubstatSet = (substat) => {
      echoLog.value.substat_all_mask ^= 1 << substat
      findEchoLog(false)
    }

    const selectedSubstatSummary = () => {
      const names = SUBSTAT
        .filter((substat) => isSubstatSetSelected(substat.num))
        .map((substat) => substat.name)
      return names.length > 0 ? names.join(' / ') : '不限'
    }

    const getSubstatSetButtonStyle = (substat) => (
      isSubstatSetSelected(substat.num)
        ? `background-color: #fff4a3; color: ${substat.font_color}; border-color: ${substat.font_color}; font-weight: 700;`
        : `color: ${substat.font_color};`
    )

    const setSubstatToCurrentPos = (substatBits, substatDesc) => {
      switch (echoLog.value.pos) {
        case 0:
          echoLog.value.substat1 = substatBits
          echoLog.value.s1_desc = substatDesc
          break
        case 1:
          echoLog.value.substat2 = substatBits
          echoLog.value.s2_desc = substatDesc
          break
        case 2:
          echoLog.value.substat3 = substatBits
          echoLog.value.s3_desc = substatDesc
          break
        case 3:
          echoLog.value.substat4 = substatBits
          echoLog.value.s4_desc = substatDesc
          break
        case 4:
          echoLog.value.substat5 = substatBits
          echoLog.value.s5_desc = substatDesc
          break
        default:
          alert('请先选择孔位')
          return false
      }
      return true
    }

    const addAnyTuneToFind = (substat) => {
      if (hasDuplicatedSubstat(substat)) {
        alert('已存在相同词条，请检查')
        return
      }

      if (setSubstatToCurrentPos(1 << substat, `${SUBSTAT[substat].name} 不限`)) {
        findEchoLog()
      }
    }

    const addTuneToFind = (substat, value) => {
      if (hasDuplicatedSubstat(substat)) {
        alert('已存在相同词条，请检查')
        return
      }

      // DEBUG
      console.log('add tune to find, echo_id:', echoLog.value.id, ', pos:', echoLog.value.pos, ', substat:', substat, ', value:', value)
      const substatDesc = SUBSTAT_VALUE_MAP[substat][value].desc_full
      if (!setSubstatToCurrentPos(1 << substat | 1 << (value + 13), substatDesc)) {
        return
      }
      // axios.post(`${API_BASE_URL}/echo_log/find`, {
      //   user_id: echoLog.value.user_id,
      //   clazz: echoLog.value.clazz,
      //   substat1: echoLog.value.substat1,
      //   substat2: echoLog.value.substat2,
      //   substat3: echoLog.value.substat3,
      //   substat4: echoLog.value.substat4,
      //   substat5: echoLog.value.substat5,
      // })
      //   .then((response) => {
      //     console.log('find echo log:', response.data) // DEBUG
      //     if (response.data.code === 200) {
      //       echoLogs.value = response.data.data
      //       if (echoLog.value.pos < 4) {
      //         echoLog.value.pos++
      //       }
      //     } else {
      //       alert('查询声骸失败')
      //     }
      //   })
      //   .catch((error) => {
      //     console.error('获取声骸 请求失败:', error)
      //     alert('查询声骸 请求失败')
      //   })
      findEchoLog()
    }

    return {
      echoLog,
      echoLogs,
      setClazz,
      setUserId,
      setSearchMode,
      newEchoLog,
      resetEchoLog,
      addAnyTuneToFind,
      addTuneToFind,
      findEchoLog,
      toggleSubstatSet,
      selectedSubstatSummary,
      getSubstatSetButtonStyle,
      operatorId,
      canManage,
      CLASSES,
      SEARCH_MODE,
      SUBSTAT,
      SUBSTAT_VALUE_MAP,
    }
  },
}
</script>

<style scoped>
.find-panel,
.find-results {
  width: 100%;
  max-width: 620px;
}

.find-panel {
  margin-bottom: 16px;
}

.find-toolbar-row,
.find-substat-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.find-substat-set-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.name {
  flex: 0 0 64px;
  min-width: 64px;
  padding-top: 10px;
}

.button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
  padding: 0 10px;
  height: 40px;
  text-align: center;
}

.user-id-input {
  width: 120px;
  min-width: 120px;
  font-weight: bolder;
}

.keyword-input {
  flex: 1 1 180px;
  min-width: 140px;
  justify-content: flex-start;
}

.clazz-chip {
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  font-weight: bolder;
}

.clear-button {
  min-width: 64px;
  color: red;
}

.search-button {
  min-width: 64px;
}

.mode-switch {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mode-button {
  min-width: 96px;
}

.mode-active {
  background-color: #fff4a3;
  font-weight: 700;
}

.find-mode-hint {
  margin: -2px 0 8px 72px;
  color: #64748b;
  font-size: 12px;
}

.find-position-row,
.find-substat-buttons,
.find-set-buttons {
  min-width: 0;
}

.find-position-row {
  display: flex;
  flex-wrap: nowrap;
  gap: 0;
}

.find-substat-buttons {
  display: flex;
  flex-wrap: nowrap;
  gap: 0;
  overflow-x: visible;
}

.find-set-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.suite-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
}

.suite-scroll {
  display: flex;
  gap: 0;
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 100%;
  padding-bottom: 6px;
}

.suite-button {
  flex: 0 0 42px;
  width: 42px;
  min-width: 42px;
  max-width: 42px;
  height: 84px;
}

.substat {
  width: 88px;
  min-width: 88px;
  max-width: 88px;
  height: 40px;
  text-align: center;
}

.compact-button {
  width: 54px;
  min-width: 54px;
  max-width: 54px;
  padding: 0 6px;
}

.set-substat-button {
  width: auto;
  min-width: 72px;
  max-width: none;
}

.selected-substat-summary {
  padding-top: 10px;
  color: #334155;
}

.my-table {
  width: 100%;
  border-collapse: collapse; /* 关键：合并边框 */
  border: 1px solid #e0e0e0; /* 表格边框 */
  table-layout: fixed;
  font-size: 12px;
}

.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 6px;
  word-break: break-word;
}
</style>
