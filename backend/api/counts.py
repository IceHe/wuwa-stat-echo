import traceback
from collections import defaultdict
from typing import Annotated

from fastapi import APIRouter, Depends
from sqlmodel import func, Session, select

from auth import require_view_permission
from db import get_session
from model import EchoLog
from response import Success, Error
from util import bit_pos

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]


@router.get("/counts/echo_dcrit", dependencies=[Depends(require_view_permission)])
async def substat_statistics(
        session: SessionDep,
        size: int = 0,
        before_id: int = 0,
        after_id: int = 0,
):
    try:
        # count logs
        stmt_count = select(func.count(EchoLog.id)).where(EchoLog.deleted == 0)
        stmt_query = select(EchoLog).where(EchoLog.deleted == 0)
        if size > 0:
            stmt_query = stmt_query.limit(size)
            stmt_count = stmt_count.limit(size)
        if after_id > 0:
            stmt_query = stmt_query.where(EchoLog.id > after_id)
            stmt_count = stmt_count.where(EchoLog.id > after_id)
            print(f'after_id={after_id}')
        if before_id > 0:
            stmt_query = stmt_query.where(EchoLog.id < before_id)
            stmt_count = stmt_count.where(EchoLog.id < before_id)
            print(f'before_id={before_id}')

        echo_count = session.exec(stmt_count).one()
        echoes = session.exec(stmt_query).all()
        if size > 0:
            echo_count = len(echoes)

        dcrit_total = 0
        counts = defaultdict(lambda: defaultdict(int))
        for echo in echoes:
            substat_all = echo.substat1 | echo.substat2 | echo.substat3 | echo.substat4 | echo.substat5
            if (substat_all & 0b1111111111111) != (substat_all & 0b1111111111111):
                print(f"Inconsistent substat_all: {substat_all} for echo {echo.id}")
            if substat_all & 0b11 == 0b11:
                dcrit_total += 1
                crit_rate_num = 0
                crit_dmg_num = 0
                if echo.substat1 & 0b01:
                    crit_rate_num = echo.substat1 >> 13
                elif echo.substat2 & 0b01:
                    crit_rate_num = echo.substat2 >> 13
                elif echo.substat3 & 0b01:
                    crit_rate_num = echo.substat3 >> 13
                elif echo.substat4 & 0b01:
                    crit_rate_num = echo.substat4 >> 13
                elif echo.substat5 & 0b01:
                    crit_rate_num = echo.substat5 >> 13

                if echo.substat1 & 0b10:
                    crit_dmg_num = echo.substat1 >> 13
                elif echo.substat2 & 0b10:
                    crit_dmg_num = echo.substat2 >> 13
                elif echo.substat3 & 0b10:
                    crit_dmg_num = echo.substat3 >> 13
                elif echo.substat4 & 0b10:
                    crit_dmg_num = echo.substat4 >> 13
                elif echo.substat5 & 0b10:
                    crit_dmg_num = echo.substat5 >> 13

                counts[str(bit_pos(crit_rate_num))][str(bit_pos(crit_dmg_num))] += 1

        return Success({
            "echo_count": echo_count,
            "dcrit_total": dcrit_total,
            "counts": counts,
        }, "test")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to test")


@router.get("/test/0", dependencies=[Depends(require_view_permission)])
async def substat_statistics(
        session: SessionDep,
        size: int = 0,
):
    try:
        # count logs
        stmt_count = select(func.count(EchoLog.id)).where(EchoLog.deleted == 0)
        stmt_query = select(EchoLog).where(EchoLog.deleted == 0)
        if size > 0:
            stmt_query = stmt_query.limit(size)

        echo_count = session.exec(stmt_count).one()
        echoes = session.exec(stmt_query).all()
        if size > 0:
            echo_count = len(echoes)

        dcrit_total = 0
        dcrit2_total = 0
        dcrit3_total = 0
        dcrit4_total = 0
        for echo in echoes:
            substat_all = echo.substat1 | echo.substat2 | echo.substat3 | echo.substat4 | echo.substat5
            if (substat_all & 0b1111111111111) != (substat_all & 0b1111111111111):
                print(f"Inconsistent substat_all: {substat_all} for echo {echo.id}")
            if substat_all & 0b11 == 0b11:
                dcrit_total += 1
                if (echo.substat1 | echo.substat2) & 0b11 == 0b11:
                    dcrit2_total += 1
                if (echo.substat1 | echo.substat2 | echo.substat3) & 0b11 == 0b11:
                    dcrit3_total += 1
                if (echo.substat1 | echo.substat2 | echo.substat3 | echo.substat4) & 0b11 == 0b11:
                    dcrit4_total += 1

        return Success({
            "echo_count": echo_count,
            "dcrit_total": dcrit_total,
            "dcrit2_total": dcrit2_total,
            "dcrit3_total": dcrit3_total,
            "dcrit4_total": dcrit4_total,
            "dcrit2_rate": str(
                dcrit2_total / echo_count * 100 if dcrit_total > 0 else 0
            ) + "%",
            "dcrit3_rate": str(
                dcrit3_total / echo_count * 100 if dcrit_total > 0 else 0
            ) + "%",
            "dcrit4_rate": str(
                dcrit4_total / echo_count * 100 if dcrit_total > 0 else 0
            ) + "%",
            "dcrit2_per_echoes": str(
                1.0 / (dcrit2_total / echo_count) if dcrit_total > 0 else 0
            ),
            "dcrit3_per_echoes": str(
                1.0 / (dcrit3_total / echo_count) if dcrit_total > 0 else 0
            ),
            "dcrit4_per_echoes": str(
                1.0 / (dcrit4_total / echo_count) if dcrit_total > 0 else 0
            ),
        }, "test")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to test")
