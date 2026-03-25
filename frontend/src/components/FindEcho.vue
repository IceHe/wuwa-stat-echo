<template>
  <h1>搜索声骸</h1>
  <div style="min-width: 750px;">
    <div>
      <span class="name">玩家ID</span>
      <input
        class="button"
        style="text-align: center; font-weight: bolder; min-width: 90px;"
        type="text"
        v-model="echoLog.user_id"
        placeholder="当前玩家ID"
        @change="setUserId(echoLog.user_id)" />
      <!--      <span class="button">调谐日期</span>-->
      <!--      <input-->
      <!--        class="button"-->
      <!--        style="text-align: center; font-weight: bolder;"-->
      <!--        type="text"-->
      <!--        v-model="echoLog.id"-->
      <!--        placeholder="调谐时间"-->
      <!--        @change="setTunedAt(echoLog.id)"-->
      <!--      />-->
      &nbsp;
      <span :style="`color: ${CLASS_COLORS[echoLog.clazz]}; font-weight: bolder;`">
        {{ echoLog.clazz.substring(0, 4) }}
      </span>
      &nbsp;
      <button class="button" style="color: red; min-width: 80px;" @click="echoLog = newEchoLog()">清空</button>
    </div>
    <div>
      <span class="name">当前孔位 </span>
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
    <div v-for="substat in SUBSTAT" :key="substat">
      <span class="name">{{ substat.name.substring(0, 4) }}</span>
      <button
        class="button"
        @click="addAnyTuneToFind(substat.num)"
        :style="`color: ${substat.font_color}`"
      >
        不限
      </button>
      <button
        class="button"
        v-for="value in SUBSTAT_VALUE_MAP[substat.num]"
        :key="value"
        @click="addTuneToFind(value.substat_number, value.value_number); "
        :style="`color: ${substat.font_color}`"
      >
        {{ value.desc }}
      </button>
    </div>
  </div>
  <div style="min-width: 480px">
    <button @click="findEchoLog()">声骸搜索列表 - 刷新</button>
    <table class="my-table">
      <thead>
      <tr style="text-align: left;">
        <th>玩家/声骸</th>
        <th>套装</th>
        <th>词条1</th>
        <th>词条2</th>
        <th>词条3</th>
        <th>词条4</th>
        <th>词条5</th>
        <th>记录于</th>
        <th>操作</th>
      </tr>
      </thead>
      <tbody>
      <EchoLogRow
        v-for="echoLog in echoLogs"
        :key="echoLog.id + echoLog.updated_at + echoLog.deleted"
        :echoLog="echoLog"
      />
      </tbody>
    </table>
  </div>
</template>

<script>
import axios from 'axios'
import {API_SERV, CLASS_COLORS, CLASSES, SUBSTAT, SUBSTAT_VALUE_MAP} from '@/stores/constants.ts'
import {ref} from 'vue'
import {useRoute} from 'vue-router';
import EchoLogRow from "@/components/EchoLogRow.vue";
import emitter from "@/stores/eventBus.js";

export default {
  name: 'Find Echo',
  computed: {
    CLASS_COLORS() {
      return CLASS_COLORS
    }
  },
  components: {EchoLogRow},
  created() {
    emitter.on('setUserId', (id) => {
      this.setUserId(id)
    })
    emitter.on('setClazz', (clazz) => {
      this.setClazz(clazz)
    })
  },
  setup: function (props) {
    const route = useRoute();
    const template = ref({
      clazz: route.query.clazz || '',
      user_id: route.query.user_id || 0,
    })

    const newEchoLog = () => ({
      clazz: template.value.clazz,
      user_id: template.value.user_id,
      id: 0,
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
    const normalizeUserId = (userId) => {
      if (userId === '' || userId === null || userId === undefined) {
        return 0
      }
      const normalized = Number(userId)
      return Number.isNaN(normalized) ? 0 : normalized
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

    const findEchoLog = async (nextPos = true) => {
      try {
        const response = await axios.post(`http://${API_SERV}/echo_log/find?page_size=20`, {
          user_id: normalizeUserId(echoLog.value.user_id),
          clazz: echoLog.value.clazz,
          substat1: echoLog.value.substat1,
          substat2: echoLog.value.substat2,
          substat3: echoLog.value.substat3,
          substat4: echoLog.value.substat4,
          substat5: echoLog.value.substat5,
        })
        console.log('listEchoLog: ', response.data)
        if (response.data.code === 200) {
          echoLogs.value = response.data.data
          if (nextPos && echoLog.value.pos < 4) {
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
    const hasDuplicatedSubstat = (substat) => (
      (1 << substat) & (
        echoLog.value.pos !== 0 ? echoLog.value.substat1 : 0 |
        echoLog.value.pos !== 1 ? echoLog.value.substat2 : 0 |
        echoLog.value.pos !== 2 ? echoLog.value.substat3 : 0 |
        echoLog.value.pos !== 3 ? echoLog.value.substat4 : 0 |
        echoLog.value.pos !== 4 ? echoLog.value.substat5 : 0
      )
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
      // axios.post(`http://${API_SERV}/echo_log/find`, {
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
      newEchoLog,
      addAnyTuneToFind,
      addTuneToFind,
      findEchoLog,
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
  min-width: 70px;
  width: 7%;
}

.button {
  display: inline-block;
  min-width: 30px;
  max-width: 120px;
  width: 10%;
  height: 40px;
  text-align: center;
}

.suite-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.suite-scroll {
  display: flex;
  gap: 4px;
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 620px;
  padding-bottom: 6px;
}

.suite-button {
  flex: 0 0 44px;
  width: 44px;
  min-width: 44px;
  max-width: 44px;
  height: 90px;
}

.substat {
  display: inline-block;
  min-width: 100px;
  max-width: 120px;
  width: 19.4%;
  height: 40px;
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
</style>
