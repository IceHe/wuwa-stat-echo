import traceback
from typing import Annotated

from fastapi import APIRouter, Depends
from sqlmodel import Session, select
from auth import require_manage_permission
from db import get_session
from model import SubstatLog, EchoLog
from response import Success, Error

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]


@router.get("/db/echo_logs/write_substat_all", dependencies=[Depends(require_manage_permission)])
async def echo_log_write_substat_all(session: SessionDep):
    try:
        stmt_query = select(EchoLog).where(EchoLog.deleted == 0)
        logs = session.exec(stmt_query).all()
        print(f"total logs: {len(logs)}")

        successTotal = 0
        for log in logs:
            if log.substat_all == 0:
                print(f"echo_id: {log.id}")
                substat_all = (log.substat1 |
                               log.substat2 |
                               log.substat3 |
                               log.substat4 |
                               log.substat5) & 0b1111111111111
                log.substat_all = substat_all
                session.add(log)
                successTotal += 1
        session.commit()

        return Success({
            "success_total": successTotal,
            "total": len(logs),
        }, "write substat all")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to write substat all")


@router.get("/db/substat_logs/write_user_id", dependencies=[Depends(require_manage_permission)])
async def substat_logs_write_user_id(session: SessionDep):
    try:
        # count logs
        stmt_query = select(SubstatLog).where(SubstatLog.deleted == 0).order_by(SubstatLog.id.desc())
        logs = session.exec(stmt_query).all()
        print(f"total logs: {len(logs)}")

        successTotal = 0
        for log in logs:
            echo_id = log.echo_id
            user_id = log.user_id
            if echo_id > 0 and user_id == 0:
                print(f"echo_id: {echo_id}")
                stmt = select(EchoLog).where(EchoLog.id == echo_id)
                echo_log = session.exec(stmt).one()
                if echo_log is not None and echo_log.user_id > 0:
                    log.user_id = echo_log.user_id
                    session.add(log)
                    successTotal += 1
        session.commit()

        return Success({
            "success_total": successTotal,
            "total": len(logs),
        }, "write id")
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get stats")
