from typing import Dict, List
from fastapi import WebSocket
from fastapi.encoders import jsonable_encoder
import json


class ConnectionManager:
    def __init__(self):
        # 按 operator_id 存储连接列表: {operator_id: [websocket1, websocket2, ...]}
        self.active_connections: Dict[str, List[WebSocket]] = {}

    @staticmethod
    def _normalize_operator_id(operator_id) -> str:
        return str(operator_id)

    async def connect(self, websocket: WebSocket, operator_id: str):
        """接受连接并添加到对应 operator_id 的连接列表"""
        operator_key = self._normalize_operator_id(operator_id)
        await websocket.accept()
        if operator_key not in self.active_connections:
            self.active_connections[operator_key] = []
        self.active_connections[operator_key].append(websocket)

    def disconnect(self, websocket: WebSocket, operator_id: str):
        """移除断开的连接"""
        operator_key = self._normalize_operator_id(operator_id)
        if operator_key in self.active_connections:
            if websocket in self.active_connections[operator_key]:
                self.active_connections[operator_key].remove(websocket)
                # 如果该 operator_id 没有连接了，删除键
                if not self.active_connections[operator_key]:
                    del self.active_connections[operator_key]

    async def send_to_operator(self, message: dict, operator_id):
        """向指定 operator_id 的所有连接发送消息"""
        operator_key = self._normalize_operator_id(operator_id)
        encoded_message = jsonable_encoder(message)
        if operator_key in self.active_connections:
            for connection in self.active_connections[operator_key]:
                await connection.send_text(json.dumps(encoded_message, ensure_ascii=False))

    async def broadcast(self, message: dict):
        """向所有连接广播消息"""
        encoded_message = jsonable_encoder(message)
        for operator_id in self.active_connections:
            for connection in self.active_connections[operator_id]:
                await connection.send_text(json.dumps(encoded_message, ensure_ascii=False))


# 全局连接管理器实例
manager = ConnectionManager()
