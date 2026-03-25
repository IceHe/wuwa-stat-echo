from datetime import datetime
from typing import Optional

from sqlalchemy import Column, DateTime, Integer, text
from sqlmodel import Field, SQLModel


class SubstatLog(SQLModel, table=True):
    __tablename__ = "wuwa_tune_log"
    id: int = Field(primary_key=True)
    substat: int
    value: int
    position: int
    echo_id: int
    user_id: int
    operator_id: Optional[int] = None
    timestamp: Optional[datetime] = Field(
        default=None,
        sa_column=Column(DateTime(timezone=True), nullable=False, server_default=text("CURRENT_TIMESTAMP")),
    )
    deleted: Optional[int] = Field(
        default=None,
        sa_column=Column(Integer, nullable=False, server_default=text("0")),
    )


class EchoLog(SQLModel, table=True):
    __tablename__ = "wuwa_echo_log"
    id: int = Field(primary_key=True)
    substat1: int = 0
    substat2: int = 0
    substat3: int = 0
    substat4: int = 0
    substat5: int = 0
    substat_all: int = 0
    s1_desc: str = ""
    s2_desc: str = ""
    s3_desc: str = ""
    s4_desc: str = ""
    s5_desc: str = ""
    clazz: str = ""
    user_id: int = 0
    operator_id: Optional[int] = None
    deleted: int = 0
    tuned_at: Optional[datetime]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]
