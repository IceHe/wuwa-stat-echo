<!-- src/components/TuneLogRow.vue -->
<template>
  <tr :style="echoLog?.deleted === 1 ? 'text-decoration: line-through;' : ''">
    <td>
      <button
        v-if="canModify"
        class="echo-btn"
        @click="setEchoId(echoLog.id)"
      >
        {{ echoLog.user_id }}<br/>{{ echoLog.id }}
      </button>
      <span v-else style="color: #999;">{{ echoLog.user_id }}<br/>{{ echoLog.id }}</span>
    </td>
    <td :style="`color: ${CLASS_COLORS[echoLog.clazz]}; width: 50px;`">{{ echoLog.clazz.substring(0, 4) }}</td>
    <td :style="`color: ${getSubstatColor(echoLog.substat1)};`">{{ echoLog.s1_desc }}</td>
    <td :style="`color: ${getSubstatColor(echoLog.substat2)};`">{{ echoLog.s2_desc }}</td>
    <td :style="`color: ${getSubstatColor(echoLog.substat3)};`">{{ echoLog.s3_desc }}</td>
    <td :style="`color: ${getSubstatColor(echoLog.substat3)};`">{{ echoLog.s4_desc }}</td>
    <td :style="`color: ${getSubstatColor(echoLog.substat3)};`">{{ echoLog.s5_desc }}</td>
    <td style="max-width: 77px;">{{ created_at }}</td>
    <td v-if="showActions" style="max-width: 48px;">
      <button v-if="canDelete" @click="del" style="color: darkred">删除</button>
      <button v-else-if="canRecover" @click="recover">恢复</button>
    </td>
  </tr>
</template>

<style scoped>
.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 8px;
}
.echo-btn {
  background: #4a90d9;
  color: #fff;
  border: 1px solid #4a90d9;
  border-radius: 4px;
  padding: 2px 6px;
  cursor: pointer;
  font-size: 12px;
  line-height: 1.4;
  text-align: center;
  width: 100%;
}
.echo-btn:hover {
  background: #357abd;
  border-color: #357abd;
}
.echo-btn:disabled {
  background: #999;
  border-color: #999;
  cursor: not-allowed;
}
</style>

<script>
import moment from 'moment'
import axios from 'axios'
import {API_BASE_URL, CLASS_COLORS, getSubstatColor} from '@/stores/constants.ts'
import echoLogs from "@/components/EchoLogs.vue";
import emitter from "@/stores/eventBus.js";
import echo from "@/components/Echo.vue";

export default {
  name: 'EchoLogRow',
  props: {
    echoLog: {
      type: Object,
      required: true,
    },
    operatorId: {
      type: Number,
      required: false,
      default: null,
    },
    canManage: {
      type: Boolean,
      required: false,
      default: false,
    },
    showActions: {
      type: Boolean,
      required: false,
      default: true,
    },
  },
  computed: {
    canModify() {
      return this.canManage || (this.operatorId != null && this.echoLog?.operator_id === this.operatorId)
    },
    canDelete() {
      return this.canModify && this.echoLog?.deleted === 0
    },
    canRecover() {
      return this.canModify && this.echoLog?.deleted === 1
    },
    echo() {
      return echo
    },
    CLASS_COLORS() {
      return CLASS_COLORS
    },
    echoLogs() {
      return echoLogs
    }
  },
  setup(props) {
    const recoverTuneLog = (id) => {
      axios
        .post(`${API_BASE_URL}/echo_log/${id}/recover`, {
          id: props.echoLog.id,
        })
        .then((response) => {
          console.log(response.data) // DEBUG
          const code = response.data.code
          if (code === 200) {
            props.echoLog.deleted = 0
          } else {
            alert('恢复失败')
          }
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }

    const deleteEchoLog = (id) => {
      axios
        .delete(`${API_BASE_URL}/echo_log/${id}`, {
          id: props.echoLog.id,
        })
        .then((response) => {
          console.log("delete echo log: ", response.data) // DEBUG
          const code = response.data.code
          if (code === 200) {
            props.echoLog.deleted = 1
          } else {
            alert('删除失败')
          }
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }

    const created_at = moment(new Date(props.echoLog.created_at)).format('MM/DD HH:mm:ss')

    const setEchoId = (echoId) => {
      emitter.emit('getEchoLog', echoId)
    }

    return {
      echoLog: props.echoLog,
      created_at,
      recover: () => {
        recoverTuneLog(props.echoLog.id)
        setTimeout(() => {
          emitter.emit('refreshEchoLogs')
          emitter.emit('refreshSubstatLogs')
        }, 1000)
      },
      del: () => {
        deleteEchoLog(props.echoLog.id)
        setTimeout(() => {
          emitter.emit('refreshEchoLogs')
          emitter.emit('refreshSubstatLogs')
        }, 1000)
      },
      setEchoId,
      getSubstatColor,
    }
  },
}
</script>
