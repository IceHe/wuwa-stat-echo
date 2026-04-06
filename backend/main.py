from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from starlette.middleware.cors import CORSMiddleware

from api.auth import router as auth_router
from api.echo import router as echo_router
from api.substat import router as substat_router
from api.analysis import router as analysis_router
from api.db_data import router as db_data_router
from api.predict import router as predict_router
from api.counts import router as test_router
from shared import init_tune_stats
from ws import manager
import asyncio


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("🚀 FastAPI 启动中...")
    await asyncio.sleep(1)
    init_tune_stats()
    print("✅ FastAPI 启动完成")
    yield


app = FastAPI(lifespan=lifespan)

# 配置 CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # 允许所有来源，生产环境中应限制为特定域名
    allow_credentials=True,  # 允许凭据
    allow_methods=["*"],  # 允许所有方法
    allow_headers=["*"],  # 允许所有 headers
)

app.include_router(echo_router)
app.include_router(substat_router)
app.include_router(analysis_router)
app.include_router(db_data_router)
app.include_router(predict_router)
app.include_router(test_router)
app.include_router(auth_router)


@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket, operator_id: str):
    """WebSocket 端点，接收 operator_id 查询参数"""
    await manager.connect(websocket, operator_id)
    try:
        while True:
            # 保持连接，接收心跳消息
            data = await websocket.receive_text()
            # 可以处理客户端发送的消息，暂时只保持连接
    except WebSocketDisconnect:
        manager.disconnect(websocket, operator_id)


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8888)
