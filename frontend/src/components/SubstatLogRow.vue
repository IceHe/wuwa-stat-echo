<!-- src/components/TuneLogRow.vue -->
<template>
  <tr :style="`color: ${fontColor}`">
    <td>{{tuneLog.echo_id}} / {{ position }} / {{ tuneLog.user_id }}</td>
    <td>{{ substat }}</td>
    <td>{{ substatValue }}</td>
    <td>{{ timestamp }}</td>
    <td>
      <button v-if="canEdit" @click="del" style="color: darkred">删除</button>
    </td>
  </tr>
</template>

<style scoped>
.my-table td,
.my-table th {
  border: 1px solid #ddd; /* 统一设置单元格边框 */
  padding: 8px;
}
</style>

<script>
import moment from 'moment'
import axios from 'axios'
import { API_SERV, SUBSTAT, SUBSTAT_VALUE_MAP } from '@/stores/constants.ts'

export default {
  name: 'TuneLogRow',
  props: {
    tuneLog: {
      type: Object,
      required: true,
    },
    refreshSubstatLogs: {
      type: Function,
      required: true,
    },
    operatorId: {
      type: Number,
      required: false,
      default: null,
    },
  },
  computed: {
    canEdit() {
      return this.operatorId != null && this.tuneLog?.operator_id === this.operatorId
    },
  },
  setup(props) {
    const deleteTuneLog = (id) => {
      axios
        .post(`http://${API_SERV}/tune_log/${id}/delete`, {
          id: props.tuneLog.id,
        })
        .then((response) => {
          console.log(response.data) // DEBUG
          const code = response.data.code
          if (code === 200) {
            props.refreshSubstatLogs()
          } else {
            alert('删除失败')
          }
        })
        .catch((error) => {
          console.error('请求失败:', error)
        })
    }

    const del = () => {
      deleteTuneLog(props.tuneLog.id)
    }

    const substat = SUBSTAT[props.tuneLog.substat].name
    const substatValue = SUBSTAT_VALUE_MAP[props.tuneLog.substat][props.tuneLog.value].desc
    const position = props.tuneLog.position + 1
    const timestamp = moment(new Date(props.tuneLog.timestamp)).format('MM-DD HH:mm:ss')
    const fontColor = SUBSTAT[props.tuneLog.substat].font_color

    return {
      substat,
      substatValue,
      position,
      timestamp,
      del,
      fontColor,
    }
  },
}
</script>
