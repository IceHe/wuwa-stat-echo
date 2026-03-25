import datetime
import traceback
import shared
from math import ceil
from typing import Annotated
from datetime import datetime, time

from fastapi import APIRouter, Depends, Request
from sqlmodel import func, Session, select, update, and_, not_

from auth import require_edit_permission, require_view_permission, get_operator_id
from consts import TUNER_RECYCLED_PER_SUBSTAT, EXP, EXP_GOLD, EXP_RETURN, RESONATOR_TEMPLATES
from custom_types import SubstatItem
from db import get_session
from model import EchoLog, SubstatLog
from response import Success, Error, Page
from util import bit_count, bit_pos

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]
SUBSTAT_TYPE_MASK = (1 << 13) - 1


def build_substat_match(column, substat_bits: int):
    if substat_bits == 0:
        return None

    # Only the low 13 type bits are set: match any tier for the selected substat.
    if substat_bits & ~SUBSTAT_TYPE_MASK == 0:
        return column.op('&')(substat_bits) == substat_bits

    return column == substat_bits


@router.get("/echo_logs", dependencies=[Depends(require_view_permission)])
async def list_echo_log(
        session: SessionDep,
        page: int = 1,
        page_size: int = 20,
):
    try:
        stmt = select(EchoLog) \
            .order_by(EchoLog.updated_at.desc()) \
            .offset((page - 1) * page_size) \
            .limit(page_size)
        data = session.exec(stmt).all()
        data_total = session.exec(select(func.count(EchoLog.id)) \
                                  .where(EchoLog.deleted == 0)).one()
        return Page("echo logs", data, data_total, page, page_size)
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get echo logs")


@router.post("/echo_log", dependencies=[Depends(require_edit_permission)])
async def create_echo_log(
        session: SessionDep,
        request: Request,
        echo_log: EchoLog,
):
    try:
        if echo_log.tuned_at is None:
            echo_log.tuned_at = datetime.now()
        elif isinstance(echo_log.tuned_at, str):
            echo_log.tuned_at = datetime.fromisoformat(echo_log.tuned_at.split('Z')[0])
        else:
            raise Error(f"tuned_at is not a datetime or str: {echo_log.tuned_at}")
        echo_log.updated_at = datetime.now()
        echo_log.created_at = datetime.now()
        echo_log.operator_id = await get_operator_id(request)

        session.add(echo_log)
        session.commit()
        session.refresh(echo_log)
        return Success(echo_log.dict(), "create echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to create echo log")


@router.patch("/echo_log", dependencies=[Depends(require_edit_permission)])
async def update_echo_log(
        session: SessionDep,
        echo_log: EchoLog,
):
    try:
        echo_log.updated_at = datetime.now()
        log_dict = echo_log.dict(exclude={
            'id', 'deleted', 'tuned_at', 'created_at',
        })

        stmt = update(EchoLog) \
            .where(EchoLog.id == echo_log.id) \
            .values(**log_dict)
        session.exec(stmt)
        session.commit()

        return Success(log_dict, "update echo log")
    except Exception as e:
        print(e)
        # 打印异常堆栈
        traceback.print_exc()
        return Error("failed to update echo log")


@router.delete("/echo_log/{id}", dependencies=[Depends(require_edit_permission)])
async def delete_echo_log(
        session: SessionDep,
        id: int,
):
    try:
        echo_log = session.get(EchoLog, id)
        if echo_log is None:
            return Error("echo log not found")

        empty_echo = (
            echo_log.substat1 == 0 and
            echo_log.substat2 == 0 and
            echo_log.substat3 == 0 and
            echo_log.substat4 == 0 and
            echo_log.substat5 == 0
        )

        if empty_echo:
            substat_logs = session.exec(
                select(SubstatLog).where(SubstatLog.echo_id == id)
            ).all()
            for substat_log in substat_logs:
                session.delete(substat_log)
            session.delete(echo_log)
            result = {"deleted": "hard", "id": id}
        else:
            stmt = update(EchoLog) \
                .where(EchoLog.id == id) \
                .values(deleted=1)
            result = session.exec(stmt)

            stmt2 = update(SubstatLog) \
                .where(SubstatLog.echo_id == id) \
                .values(deleted=1)
            session.exec(stmt2)

        session.commit()
        return Success(result, "delete echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to delete echo log")


@router.post("/echo_log/{id}/recover", dependencies=[Depends(require_edit_permission)])
async def recover_echo_log(
        session: SessionDep,
        id: int,
):
    try:
        echo_log = session.get(EchoLog, id)
        if echo_log is None:
            return Error("echo log not found")

        stmt = update(EchoLog) \
            .where(EchoLog.id == id) \
            .values(deleted=0)
        result = session.exec(stmt)

        stmt2 = update(SubstatLog) \
            .where(SubstatLog.echo_id == id) \
            .values(deleted=0)
        session.exec(stmt2)

        session.commit()
        return Success(result, "recover echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to recover echo log")


@router.get("/echo_log/{id}", dependencies=[Depends(require_view_permission)])
async def get_echo_log(
        session: SessionDep,
        id: int,
):
    try:
        stmt = select(EchoLog)
        if id > 0:
            stmt = stmt.where(EchoLog.id == id)
        else:
            stmt = stmt.where(EchoLog.deleted == 0) \
                .order_by(EchoLog.updated_at.desc()) \
                .limit(1)
        echo_log = session.exec(stmt).first()
        if echo_log is None:
            return Error("echo log not found")

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

        return Success({
            **echo_log.dict(),
            "pos_total": pos_total,
        }, "echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get echo log")


@router.post("/echo_log/find", dependencies=[Depends(require_view_permission)])
async def find_echo_log(
        session: SessionDep,
        echo_log: EchoLog,
):
    if echo_log.substat1 | echo_log.substat2 | echo_log.substat3 | echo_log.substat4 | echo_log.substat5 == 0:
        return Success([], "no substat specified, return empty list")

    try:
        user_id = int(echo_log.user_id or 0)

        stmt = select(EchoLog)
        for column, substat_bits in (
                (EchoLog.substat1, echo_log.substat1),
                (EchoLog.substat2, echo_log.substat2),
                (EchoLog.substat3, echo_log.substat3),
                (EchoLog.substat4, echo_log.substat4),
                (EchoLog.substat5, echo_log.substat5),
        ):
            filter_expr = build_substat_match(column, substat_bits)
            if filter_expr is not None:
                stmt = stmt.where(and_(filter_expr))

        stmt = stmt.where(EchoLog.deleted == 0)
        if user_id > 0:
            stmt = stmt.where(EchoLog.user_id == user_id)
        if echo_log.clazz != '':
            stmt = stmt.where(EchoLog.clazz == echo_log.clazz)

        data = session.exec(stmt).all()
        return Success(data, "find echo logs")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to find echo logs")


# 跟据 echo id 和 pos 位置，删除 substat log
@router.delete("/echo_log/{echoId}/substat_pos/{pos}", dependencies=[Depends(require_edit_permission)])
async def delete_substat_log(
        session: SessionDep,
        echoId: int,
        pos: int,
):
    try:
        stmt = update(SubstatLog) \
            .where(SubstatLog.echo_id == echoId) \
            .where(SubstatLog.position == pos) \
            .values(deleted=1)
        result = session.exec(stmt)
        session.commit()
        return Success(result, "delete substat log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to delete substat log")


# 统计距离上一次目标间隔了多少个声骸和副词条，平均目标间隔等
@router.get("/echo_logs/analysis", dependencies=[Depends(require_view_permission)])
async def echo_log_analysis(
        session: SessionDep,
        user_id: int = 0,
        size: int = 0,
        target_bits: int = 0b11,
        substat_since_date: str = '',
):
    try:
        selectStmt = select(EchoLog) \
            .where(EchoLog.deleted == 0) \
            .order_by(EchoLog.updated_at.desc())
        countStmt = select(func.count(EchoLog.id)).where(EchoLog.deleted == 0)
        if user_id > 0:
            selectStmt = selectStmt.where(and_(EchoLog.user_id == user_id))
            countStmt = countStmt.where(and_(EchoLog.user_id == user_id))

            if substat_since_date:
                start_at = datetime.combine(datetime.fromisoformat(substat_since_date).date(), time(4, 0, 0))
                echo_ids = session.exec(
                    select(SubstatLog.echo_id)
                    .where(SubstatLog.deleted == 0)
                    .where(SubstatLog.user_id == user_id)
                    .where(SubstatLog.timestamp >= start_at)
                ).all()
                echo_ids = list(dict.fromkeys(echo_ids))
                if echo_ids:
                    selectStmt = selectStmt.where(EchoLog.id.in_(echo_ids))
                    countStmt = countStmt.where(EchoLog.id.in_(echo_ids))
                else:
                    selectStmt = selectStmt.where(EchoLog.id == -1)
                    countStmt = countStmt.where(EchoLog.id == -1)
        if size > 0:
            selectStmt = selectStmt.limit(size)

        # FIXME 临时逻辑
        # selectStmt = selectStmt.where(and_(EchoLog.clazz.in_(['沉日劫明', '命理崩毁之弦'])))
        # selectStmt = selectStmt.where(and_(not_(EchoLog.clazz == '凌冽决断之心')))

        data = session.exec(selectStmt).all()
        data_total = session.exec(countStmt).one()

        found = False
        idx = 0
        target = 0
        target_echo_distance = -1
        target_substat_distance = -1
        substat_total = 0
        tuner_recycled = 0
        exp_total = 0
        exp_recycled = 0
        for echo_log in data:
            substat_all = 0b1111111111111 & (
                    echo_log.substat1 |
                    echo_log.substat2 |
                    echo_log.substat3 |
                    echo_log.substat4 |
                    echo_log.substat5
            )
            # substat_count = substat_all.bit_count()
            substat_count = bit_count(substat_all)
            substat_total += substat_count
            exp_total += EXP[0][substat_count]

            if substat_all & target_bits == target_bits:
                target += 1
                if not found:
                    found = True
                    target_echo_distance = idx
                    target_substat_distance = substat_total
            else:
                tuner_recycled += substat_count * TUNER_RECYCLED_PER_SUBSTAT
                exp_recycled += EXP_RETURN[substat_count]
            idx += 1

        if not found:
            target_echo_distance = idx
            target_substat_distance = substat_total

        tuner_consumed = ceil(substat_total * 10 - tuner_recycled)
        exp_consumed = ceil((exp_total - exp_recycled) / EXP_GOLD)

        analysis = {
            "target_echo_distance": target_echo_distance,
            "target_substat_distance": target_substat_distance,
            "target": target,
            "target_avg_echo": 0.0,
            "target_avg_substat": 0.0,
            "tuner_consumed": tuner_consumed,
            "tuner_consumed_avg": 0.0,
            "exp_consumed": exp_consumed,
            "exp_consumed_avg": 0.0,
        }
        if target > 0:
            analysis["target_avg_echo"] = round(data_total / target, 1)
            analysis["target_avg_substat"] = round(substat_total / target, 1)
            analysis["tuner_consumed_avg"] = ceil(tuner_consumed / target)
            analysis["exp_consumed_avg"] = ceil(exp_consumed / target)

        return Success(analysis, "echo logs analysis")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get echo logs")
