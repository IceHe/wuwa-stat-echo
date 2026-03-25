import traceback
from collections import defaultdict

from consts import SUBSTAT_DICT
from custom_types import SubstatItem, SubstatValueStat
from db import engine
from model import SubstatLog
from sqlmodel import func, Session, select, and_

tune_stats = {
    "position_total": [0] * 5,
}


def init_tune_stats():
    try:
        with Session(engine) as session:
            # init substat dict
            substat_dict: dict[str, SubstatItem] = defaultdict(None)
            for i, substat in SUBSTAT_DICT.items():
                substat_dict[i] = SubstatItem(substat.number, substat.name, substat.name_cn)
                for j, value in substat.value_dict.items():
                    substat_dict[i].value_dict[j] = SubstatValueStat(
                        value.value_number,
                        value.value_desc,
                        value.value_desc_full,
                    )

            # count logs
            stmt_count = select(func.count(SubstatLog.id)).where(SubstatLog.deleted == 0)
            stmt_query = select(SubstatLog).where(SubstatLog.deleted == 0).order_by(SubstatLog.id.desc())

            logs_total = session.exec(stmt_count).one()
            logs = session.exec(stmt_query).all()

            # 其它统计：一种词条多久没出现
            index = -1
            distances = [-1] * 13

            position_total = [0] * 5
            substat_pos_total = [[0] * 5 for _ in range(13)]
            for tune_log in logs:
                index += 1
                if distances[tune_log.substat] == -1:
                    distances[tune_log.substat] = index

                substat = substat_dict[str(tune_log.substat)]
                substat.total += 1
                substat_value = substat.value_dict[str(tune_log.value)]
                substat_value.total += 1
                substat_value.position_dict[str(tune_log.position)].total += 1
                position_total[tune_log.position] += 1
                substat_pos_total[tune_log.substat][tune_log.position] += 1

                substat_all = substat.value_dict['all']
                substat_all.total += 1
                substat_all.position_dict[str(tune_log.position)].total += 1

            # calculate percent
            for _, substat in substat_dict.items():
                substat_all = substat.value_dict['all']

                if logs_total > 0:
                    substat.percent = round(substat.total / logs_total * 100, 2)
                substat_total = substat.total
                for x, value in substat.value_dict.items():
                    substat_all_total = substat_all.total
                    if substat_total > 0:
                        value.percent_substat = round(value.total / substat_total * 100, 2)
                    if x == 'all':
                        if logs_total > 0:
                            value.percent = round(value.total / logs_total * 100, 2)
                    elif substat_all_total > 0:
                        value.percent = round(value.total / substat_all_total * 100, 2)

                    for k, position in value.position_dict.items():
                        substat_all_value_total = substat_all.position_dict[str(k)].total
                        if position.total > 0 and substat_all_value_total > 0:
                            position.percent = round(position.total / substat_all_value_total * 100, 2)
                        if position.total > 0 and position_total[int(k)] > 0:
                            position.percent_all = round(position.total / position_total[int(k)] * 100, 1)

                for i, position in substat_all.position_dict.items():
                    if position.total > 0 and sum(position_total) > 0:
                        position.percent = round(position.total / position_total[int(i)] * 100, 2)

            tune_stats.update({
                "data_total": logs_total,
                "substat_dict": substat_dict,
                "substat_distance": distances,
                "substat_pos_total": substat_pos_total,
                "position_total": position_total,
            })
    except Exception as e:
        print(e)
        traceback.print_exc()
