import os

from sqlmodel import create_engine, Session
from sqlmodel.pool import StaticPool
from env import load_env_file

load_env_file()

postgres_url = os.getenv("DATABASE_URL")
if not postgres_url:
    raise RuntimeError(
        "DATABASE_URL is not set. Configure backend/.env or export the variable before starting the backend."
    )

if postgres_url.startswith("postgresql://") and "+psycopg" not in postgres_url:
    postgres_url = postgres_url.replace("postgresql://", "postgresql+psycopg://", 1)

engine = create_engine(
    postgres_url,
    connect_args={},
    poolclass=StaticPool,
)


def get_session():
    with Session(engine) as session:
        yield session
