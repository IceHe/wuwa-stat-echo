import traceback
from typing import Annotated

from fastapi import APIRouter, Depends, Request
from sqlmodel import func, Session, insert, select, delete

from auth import require_edit_permission, require_view_permission, get_operator_id
from db import get_session
from model import SubstatLog
from response import Success, Error, Page

router = APIRouter()
SessionDep = Annotated[Session, Depends(get_session)]


@router.get("/", dependencies=[Depends(require_view_permission)])
async def root():
    return {"message": "Hello, Wuwa!"}


@router.get("/substat_logs", dependencies=[Depends(require_view_permission)])
async def list_tune_log(
        session: SessionDep,
        page: int = 1,
        page_size: int = 20,
):
    try:
        stmt = select(SubstatLog) \
            .order_by(SubstatLog.id.desc()) \
            .offset((page - 1) * page_size) \
            .limit(page_size)
        data = session.exec(stmt).all()
        data_total = session.exec(select(func.count(SubstatLog.id))).one()
        return Page("tune logs", data, data_total, page, page_size)
    except Exception as e:
        print(e)
        traceback.print_exc()
        return Error("failed to get tune logs")


@router.post("/tune_log/{id}/delete", dependencies=[Depends(require_edit_permission)])
async def delete_substat_log(
        session: SessionDep,
        request: Request,
        id: int,
):
    try:
        operator_id = await get_operator_id(request)
        tune_log = session.get(SubstatLog, id)
        if tune_log is None:
            return Error("tune log not found")
        if tune_log.operator_id != operator_id:
            return Error("not authorized to delete this tune log")

        stmt = delete(SubstatLog).where(SubstatLog.id == id)
        result = session.exec(stmt)
        session.commit()
        return Success({"row_deleted": result.rowcount}, f"delete tune log {id}")
    except Exception as e:
        print(e)
        traceback.print_exc()
        session.rollback()
        return Error(f"failed to delete tune log {id}")


@router.post("/tune_log", dependencies=[Depends(require_edit_permission)])
# params: 词条 word，档位 value，孔位 position，时间戳 timestamp
async def add_substat_log(
        session: SessionDep,
        request: Request,
        tuneLog: SubstatLog,
):
    try:
        operator_id = await get_operator_id(request)
        stmt = insert(SubstatLog).values({
            "user_id": tuneLog.user_id,
            "echo_id": tuneLog.echo_id,
            "position": tuneLog.position,
            "substat": tuneLog.substat,
            "value": tuneLog.value,
            "operator_id": operator_id,
        })
        session.exec(stmt)
        session.commit()
        return Success({}, "add tune log")
    except Exception as e:
        print(e)
        traceback.print_exc()
        session.rollback()
        return Error("failed to add tune log")
