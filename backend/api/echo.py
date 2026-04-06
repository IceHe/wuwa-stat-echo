import datetime
import traceback
import shared
from math import ceil
from typing import Annotated
from datetime import datetime, time

from fastapi import APIRouter, Depends, Request
from sqlalchemy import String, cast, or_
from sqlmodel import func, Session, select, update, and_, not_

from auth import require_edit_permission, require_view_permission, get_operator_id, can_manage
from consts import TUNER_RECYCLED_PER_SUBSTAT, EXP, EXP_GOLD, EXP_RETURN, RESONATOR_TEMPLATES
from custom_types import SubstatItem, EchoTuneRequest, EchoFindRequest
from db import get_session
from model import EchoLog, SubstatLog
from response import Success, Error, Page
from util import bit_count, bit_pos
from ws import manager

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]
SUBSTAT_TYPE_MASK = (1 << 13) - 1
ECHO_MUTABLE_FIELDS = (
    "substat1",
    "substat2",
    "substat3",
    "substat4",
    "substat5",
    "substat_all",
    "s1_desc",
    "s2_desc",
    "s3_desc",
    "s4_desc",
    "s5_desc",
    "clazz",
    "user_id",
)


def build_substat_match(column, substat_bits: int):
    if substat_bits == 0:
        return None

    # Only the low 13 type bits are set: match any tier for the selected substat.
    if substat_bits & ~SUBSTAT_TYPE_MASK == 0:
        return column.op('&')(substat_bits) == substat_bits

    return column == substat_bits


def apply_echo_changes(target: EchoLog, payload) -> None:
    for field_name in ECHO_MUTABLE_FIELDS:
        value = getattr(payload, field_name, None)
        if value is not None:
            setattr(target, field_name, value)

    target.updated_at = datetime.now()


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

        # 推送创建消息到 WebSocket
        await manager.send_to_operator({
            "type": "create_echo_log",
            "data": echo_log.dict()
        }, echo_log.operator_id)

        return Success(echo_log.dict(), "create echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to create echo log")


@router.patch("/echo_log", dependencies=[Depends(require_edit_permission)])
async def update_echo_log(
        session: SessionDep,
        request: Request,
        echo_log: EchoLog,
):
    try:
        operator_id = await get_operator_id(request)
        if operator_id is None:
            return Error("operator not found", 401)

        existing_echo_log = session.get(EchoLog, echo_log.id)
        if existing_echo_log is None:
            return Error("echo log not found", 404)

        if existing_echo_log.operator_id != operator_id and not await can_manage(request):
            return Error("not authorized to update this echo log", 403)

        apply_echo_changes(existing_echo_log, echo_log)
        session.commit()
        session.refresh(existing_echo_log)

        # 推送更新消息到 WebSocket
        await manager.send_to_operator({
            "type": "update_echo_log",
            "data": existing_echo_log.dict()
        }, existing_echo_log.operator_id)

        return Success(existing_echo_log.dict(), "update echo log")
    except Exception as e:
        print(e)
        # 打印异常堆栈
        traceback.print_exc()
        session.rollback()
        return Error("failed to update echo log")


@router.post("/echo_log/tune", dependencies=[Depends(require_edit_permission)])
async def tune_echo_log(
        session: SessionDep,
        request: Request,
        payload: EchoTuneRequest,
):
    try:
        operator_id = await get_operator_id(request)
        if operator_id is None:
            return Error("operator not found", 401)

        is_manager = await can_manage(request)
        echo_log = None

        if payload.id > 0:
            echo_log = session.get(EchoLog, payload.id)
            if echo_log is None:
                return Error("echo log not found", 404)
            if echo_log.operator_id != operator_id and not is_manager:
                return Error("not authorized to tune this echo log", 403)
        else:
            if not payload.user_id:
                return Error("user_id is required", 400)
            if not payload.clazz:
                return Error("clazz is required", 400)

            echo_log = EchoLog(
                user_id=payload.user_id,
                clazz=payload.clazz,
                tuned_at=datetime.now(),
                created_at=datetime.now(),
                updated_at=datetime.now(),
                operator_id=operator_id,
            )
            session.add(echo_log)
            session.flush()

        apply_echo_changes(echo_log, payload)

        tune_log = SubstatLog(
            user_id=echo_log.user_id,
            echo_id=echo_log.id,
            position=payload.position,
            substat=payload.substat,
            value=payload.value,
            operator_id=operator_id,
        )
        session.add(tune_log)
        session.commit()
        session.refresh(echo_log)
        session.refresh(tune_log)

        # 推送调谐消息到 WebSocket
        await manager.send_to_operator({
            "type": "tune_echo_log",
            "data": {
                "echo_log": echo_log.dict(),
                "tune_log": tune_log.dict()
            }
        }, operator_id)

        return Success({
            "echo_log": echo_log.dict(),
            "tune_log": tune_log.dict(),
        }, "tune echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        session.rollback()
        return Error("failed to tune echo log")


@router.delete("/echo_log/{id}", dependencies=[Depends(require_edit_permission)])
async def delete_echo_log(
        session: SessionDep,
        request: Request,
        id: int,
):
    try:
        operator_id = await get_operator_id(request)
        if operator_id is None:
            return Error("operator not found", 401)
        is_manager = await can_manage(request)

        echo_log = session.get(EchoLog, id)
        if echo_log is None:
            return Error("echo log not found")
        if echo_log.operator_id != operator_id and not is_manager:
            return Error("not authorized to delete this echo log", 403)

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

        # 推送删除消息到 WebSocket
        await manager.send_to_operator({
            "type": "delete_echo_log",
            "data": {"id": id, "deleted": "hard" if empty_echo else "soft"}
        }, echo_log.operator_id)

        return Success(result, "delete echo log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to delete echo log")


@router.post("/echo_log/{id}/recover", dependencies=[Depends(require_edit_permission)])
async def recover_echo_log(
        session: SessionDep,
        request: Request,
        id: int,
):
    try:
        operator_id = await get_operator_id(request)
        if operator_id is None:
            return Error("operator not found", 401)
        is_manager = await can_manage(request)

        echo_log = session.get(EchoLog, id)
        if echo_log is None:
            return Error("echo log not found")
        if echo_log.operator_id != operator_id and not is_manager:
            return Error("not authorized to recover this echo log", 403)

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
        request: Request,
        id: int,
):
    try:
        operator_id = await get_operator_id(request)
        if operator_id is None:
            return Error("operator not found", 401)

        stmt = select(EchoLog)
        if id > 0:
            stmt = stmt.where(EchoLog.id == id)
        else:
            stmt = stmt.where(EchoLog.deleted == 0) \
                .where(EchoLog.operator_id == operator_id) \
                .order_by(EchoLog.updated_at.desc()) \
                .limit(1)
        echo_log = session.exec(stmt).first()
        if echo_log is None:
            return Error("echo log not found", 404)

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
        request: Request,
        echo_log: EchoFindRequest,
        page_size: int = 20,
):
    has_substat_filter = (
        echo_log.substat1 | echo_log.substat2 | echo_log.substat3 | echo_log.substat4 | echo_log.substat5
    ) != 0
    keyword = (echo_log.keyword or "").strip()
    if (
            not has_substat_filter and
            int(echo_log.id or 0) <= 0 and
            int(echo_log.user_id or 0) <= 0 and
            echo_log.clazz == '' and
            keyword == ''
    ):
        return Success([], "no search condition specified, return empty list")

    try:
        echo_id = int(echo_log.id or 0)
        user_id = int(echo_log.user_id or 0)

        stmt = select(EchoLog).where(EchoLog.deleted == 0)

        if echo_id > 0:
            stmt = stmt.where(EchoLog.id == echo_id)

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

        if user_id > 0:
            stmt = stmt.where(EchoLog.user_id == user_id)
        if echo_log.clazz != '':
            stmt = stmt.where(EchoLog.clazz == echo_log.clazz)
        if keyword:
            keyword_pattern = f"%{keyword}%"
            keyword_filters = [
                EchoLog.clazz.ilike(keyword_pattern),
                EchoLog.s1_desc.ilike(keyword_pattern),
                EchoLog.s2_desc.ilike(keyword_pattern),
                EchoLog.s3_desc.ilike(keyword_pattern),
                EchoLog.s4_desc.ilike(keyword_pattern),
                EchoLog.s5_desc.ilike(keyword_pattern),
                cast(EchoLog.user_id, String).ilike(keyword_pattern),
                cast(EchoLog.id, String).ilike(keyword_pattern),
            ]
            stmt = stmt.where(or_(*keyword_filters))

        normalized_page_size = max(1, min(page_size, 100))
        stmt = stmt.order_by(EchoLog.updated_at.desc()).limit(normalized_page_size)

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
        request: Request,
        echoId: int,
        pos: int,
):
    try:
        operator_id = await get_operator_id(request)
        if operator_id is None:
            return Error("operator not found", 401)
        is_manager = await can_manage(request)

        echo_log = session.get(EchoLog, echoId)
        if echo_log is None:
            return Error("echo log not found", 404)
        if echo_log.operator_id != operator_id and not is_manager:
            return Error("not authorized to delete substats for this echo log", 403)

        stmt = update(SubstatLog) \
            .where(SubstatLog.echo_id == echoId) \
            .where(SubstatLog.position == pos)
        if not is_manager:
            stmt = stmt.where(SubstatLog.operator_id == operator_id)
        stmt = stmt.values(deleted=1)
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
