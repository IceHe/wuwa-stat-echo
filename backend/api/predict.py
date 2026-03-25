import traceback
from typing import Annotated

from fastapi import APIRouter, Depends
from sqlmodel import Session, select, and_

from auth import require_view_permission
from db import get_session
from model import EchoLog
from response import Success, Error
from util import bit_pos

MASK = 0b1111111111111

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]


@router.post("/predict/echo_substat", dependencies=[Depends(require_view_permission)])
async def predict_echo_substat(
        session: SessionDep,
        echo_log: EchoLog,
):
    if echo_log is None:
        return Error("echo_log cannot be None")

    substat_target = echo_log.substat_all & MASK
    if substat_target < 0:
        return Error("substat_target must >= 0")

    if bit_pos(substat_target) >= 5:
        return Success({
            "count_total": 0,
            "count": [0] * 14,
            "percent": [0.0] * 14,
        }, "predict echo substat")

    try:
        stmt = select(EchoLog).where(and_(
            EchoLog.deleted == 0,
            EchoLog.substat_all.op('&')(substat_target) == substat_target
        ))
        logs = session.exec(stmt).all()
        cnt = [0] * 15
        for log in logs:
            if echo_log.substat1 == 0:
                cnt[bit_pos(log.substat1 & MASK)] += 1
                continue
            if echo_log.substat1 & MASK != log.substat1 & MASK:
                continue

            if echo_log.substat2 == 0:
                cnt[bit_pos(log.substat2 & MASK)] += 1
                continue
            if echo_log.substat2 & MASK != log.substat2 & MASK:
                continue

            if echo_log.substat3 == 0:
                cnt[bit_pos(log.substat3 & MASK)] += 1
                continue
            if echo_log.substat3 & MASK != log.substat3 & MASK:
                continue

            if echo_log.substat4 == 0:
                cnt[bit_pos(log.substat4 & MASK)] += 1
                continue
            if echo_log.substat4 & MASK != log.substat4 & MASK:
                continue

            if echo_log.substat5 == 0:
                cnt[bit_pos(log.substat5 & MASK)] += 1
                continue

        # return Page("predict", logs, len(logs), -1, -1)
        cnt = cnt[:13]
        total = sum(cnt)
        return Success({
            "count_total": total,
            "count": cnt,
            "percent": [round(x / total * 100, 1) if total > 0 else 0.0 for x in cnt],
        }, "predict echo substat")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get echo logs")
