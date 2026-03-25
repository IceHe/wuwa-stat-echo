from datetime import datetime
from typing import Optional

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
    timestamp: Optional[datetime]
    deleted: Optional[int]


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
