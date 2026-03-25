import traceback
from collections import defaultdict
from typing import Annotated

from fastapi import APIRouter, Depends
from sqlmodel import func, Session, select, and_

from auth import require_view_permission
from consts import SUBSTAT_DICT, RESONATOR_TEMPLATES
from custom_types import SubstatItem, SubstatValueStat
from db import get_session
from model import SubstatLog, EchoLog
from response import Success, Error

import shared
from util import bit_pos, bit_count

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]


@router.get("/tune_stats", dependencies=[Depends(require_view_permission)])
async def substat_statistics(
        session: SessionDep,
        size: int = 0,
        user_id: int = 0,
        after_id: int = 0,
        before_id: int = 0,
):
    try:
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
        if size > 0:
            stmt_query = stmt_query.limit(size)
        if user_id > 0:
            print(f"filter by user_id: {user_id}")
            stmt_query = stmt_query.where(and_(SubstatLog.user_id == user_id))
            stmt_count = stmt_count.where(and_(SubstatLog.user_id == user_id))
        if after_id > 0:
            print(f"filter by after_id: {after_id}")
            stmt_query = stmt_query.where(and_(SubstatLog.id > after_id))
            stmt_count = stmt_count.where(and_(SubstatLog.id > after_id))
        if before_id > 0:
            print(f"filter by before_id: {before_id}")
            stmt_query = stmt_query.where(and_(SubstatLog.id < before_id))
            stmt_count = stmt_count.where(and_(SubstatLog.id < before_id))

        print('stmt_query:', stmt_query)
        logs_total = session.exec(stmt_count).one()
        logs = session.exec(stmt_query).all()
        if size > 0:
            logs_total = len(logs)

        # 其它统计：一种词条多久没出现
        index = -1
        distances = [-1] * 13

        position_total = [0] * 5
        substat_pos_total = [[0] * 5 for _ in range(13)]
        for tune_log in logs:
            index += 1
            # print(tune_log.substat)
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
                if position.total > 0 and position_total[int(i)] > 0:
                    position.percent = round(position.total / position_total[int(i)] * 100, 2)

        shared.tune_stats = {
            "data_total": logs_total,
            "substat_dict": substat_dict,
            "substat_distance": distances,
            "substat_pos_total": substat_pos_total,
            "position_total": position_total,
        }
        return Success(shared.tune_stats, "tune stats")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get stats")


@router.get("/substat_distance_analysis", dependencies=[Depends(require_view_permission)])
async def substat_statistics(
        session: SessionDep,
        size: int = 0,
):
    try:
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
        if size > 0:
            stmt_query = stmt_query.limit(size)

        logs_total = session.exec(stmt_count).one()
        logs = session.exec(stmt_query).all()
        if size > 0:
            logs_total = len(logs)

        # 其它统计：一种词条多久没出现
        index = -1
        distances = [-1] * 13

        position_total = [0] * 5
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
                    # if position.total > 0 and position_total[int(k)] > 0:
                    #     position.percent_all = round(position.total / position_total[int(k)] * 100, 0)

            for i, position in substat_all.position_dict.items():
                if position.total > 0 and sum(position_total) > 0:
                    position.percent = round(position.total / position_total[int(i)] * 100, 2)

        return Success({
            "data_total": logs_total,
            "substat_dict": substat_dict,
            "substat_distance": distances,
        }, "tune stats")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get stats")


PERCENT = [
    # [0暴, 1暴 2暴]
    [12.82, 12.82, 100.0] * 3,  # pos 1
    [9.09, 33.33, 100.0],  # pos 2
    [5.45, 27.27, 100.0],  # 3
    [2.22, 20.0, 100.0],  # 4
    [0.0, 11.11, 100.0],  # 5
]


@router.post("/analyze_echo", dependencies=[Depends(require_view_permission)])
async def analyze_echo_log(
        echo_log: EchoLog,
        resonator: str = '',
        cost: str = '',
):
    if cost == '':
        cost = '1C'
    print('resonator:', resonator)
    print('cost:', cost)

    try:
        pos = 0
        if echo_log.substat1:
            pos = 1
        if echo_log.substat2:
            pos = 2
        if echo_log.substat3:
            pos = 3
        if echo_log.substat4:
            pos = 4

        tune_stats = shared.tune_stats
        pos_total = tune_stats["position_total"][pos]
        if echo_log.substat1:
            pos_total -= tune_stats["substat_pos_total"][bit_pos(echo_log.substat1)][pos]
        if echo_log.substat2:
            pos_total -= tune_stats["substat_pos_total"][bit_pos(echo_log.substat2)][pos]
        if echo_log.substat3:
            pos_total -= tune_stats["substat_pos_total"][bit_pos(echo_log.substat3)][pos]
        if echo_log.substat4:
            pos_total -= tune_stats["substat_pos_total"][bit_pos(echo_log.substat4)][pos]
        if echo_log.substat5:
            pos_total -= tune_stats["substat_pos_total"][bit_pos(echo_log.substat5)][pos]

        if pos_total > 0:
            substat_dict: dict[str, SubstatItem] = tune_stats['substat_dict']
            for _, substat in substat_dict.items():
                show = not ((echo_log.substat_all >> substat.number) & 1)
                substat.cur_pos_percent = (str(round(
                    tune_stats['substat_pos_total'][substat.number][pos] * 100 / pos_total, 1
                )) + '%') if show else ''
                for _, value in substat.value_dict.items():
                    pos_stat = value.position_dict[str(pos)]
                    pos_stat.percent = (str(round(
                        pos_stat.total * 100 / pos_total, 1
                    )) + '%') if show and pos_stat.total > 0 else ''

        resonator_template = RESONATOR_TEMPLATES[resonator]
        tune_stats["score"] = resonator_template.echo_score(echo_log, cost)
        tune_stats["score"].resonator = resonator_template.name
        tune_stats["resonator_template"] = resonator_template

        print('resonator_template.name:', resonator_template.name)

        # 双暴概率
        crit_count = bit_count(echo_log.substat_all & 0b11)
        tune_stats["two_crit_percent"] = PERCENT[pos][crit_count]

        return Success(tune_stats, "echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get echo log")
