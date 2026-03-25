import hashlib
import logging
import os
from time import monotonic
from typing import Optional

import httpx
from fastapi import HTTPException, Request
from env import load_env_file


AUTH_INVALID_DETAIL = "Token 无效或已过期"
AUTH_FORBIDDEN_DETAIL = "权限不足"
AUTH_UNAVAILABLE_DETAIL = "鉴权服务不可用"

_TOKEN_CACHE_PREFIX = "_auth_token_cache_"

load_env_file()

AUTH_SERVICE_URL = os.getenv("AUTH_SERVICE_URL", "http://127.0.0.1:8080")
AUTH_SERVICE_TIMEOUT_SECONDS = float(os.getenv("AUTH_SERVICE_TIMEOUT_SECONDS", "3"))

logger = logging.getLogger(__name__)


def _token_fingerprint(token: str) -> str:
    return hashlib.sha256(token.encode("utf-8")).hexdigest()[:12]


def _has_permission(permissions: list[str], required_permission: str) -> bool:
    return "manage" in permissions or required_permission in permissions


def extract_token_from_request(request: Request) -> Optional[str]:
    authorization = request.headers.get("Authorization")
    if authorization:
        scheme, _, token = authorization.partition(" ")
        if scheme.lower() == "bearer" and token.strip():
            return token.strip()

    x_token = request.headers.get("X-Token")
    if x_token and x_token.strip():
        return x_token.strip()

    token = request.query_params.get("token", "").strip()
    if token:
        return token

    return None


async def _request_auth_api(
        path: str,
        *,
        json: Optional[dict] = None,
        headers: Optional[dict[str, str]] = None,
        method: str = "POST",
) -> httpx.Response:
    url = f"{AUTH_SERVICE_URL.rstrip('/')}{path}"
    started = monotonic()
    token = ""
    if json and "token" in json:
        token = str(json["token"])
    elif headers:
        authorization = headers.get("Authorization", "")
        if authorization.lower().startswith("bearer "):
            token = authorization[7:].strip()
        else:
            token = headers.get("X-Token", "").strip()
    token_fp = _token_fingerprint(token) if token else "-"

    try:
        async with httpx.AsyncClient(timeout=AUTH_SERVICE_TIMEOUT_SECONDS) as client:
            response = await client.request(method, url, json=json, headers=headers)
    except httpx.RequestError as exc:
        elapsed_ms = int((monotonic() - started) * 1000)
        logger.warning(
            "Auth service request failed: path=%s, token_fp=%s, timeout_s=%s, elapsed_ms=%s, error=%s",
            path,
            token_fp,
            AUTH_SERVICE_TIMEOUT_SECONDS,
            elapsed_ms,
            exc,
        )
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL) from exc

    if response.status_code >= 500:
        logger.warning(
            "Auth service upstream error: path=%s, token_fp=%s, status_code=%s, body=%s",
            path,
            token_fp,
            response.status_code,
            response.text,
        )
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL)

    return response


async def validate_token(token: str, required_permission: Optional[str] = None) -> dict:
    payload: dict[str, str] = {"token": token}
    if required_permission:
        payload["permission"] = required_permission

    response = await _request_auth_api("/api/validate", json=payload)
    if response.status_code != 200:
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL)

    try:
        result = response.json()
    except ValueError as exc:
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL) from exc

    permissions = result.get("permissions")
    if not isinstance(permissions, list):
        permissions = []

    reason = str(result.get("reason", "")).lower()
    if not result.get("valid"):
        if reason == "forbidden":
            raise HTTPException(status_code=403, detail=AUTH_FORBIDDEN_DETAIL)
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)

    if required_permission and not _has_permission(permissions, required_permission):
        raise HTTPException(status_code=403, detail=AUTH_FORBIDDEN_DETAIL)

    operator_id = result.get("id")
    if operator_id is not None:
        operator_id = int(operator_id)

    return {"permissions": permissions, "operator_id": operator_id}


async def require_view_permission(request: Request) -> list[str]:
    token = extract_token_from_request(request)
    if not token:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)
    cache_key = f"{_TOKEN_CACHE_PREFIX}{_token_fingerprint(token)}"
    if hasattr(request.state, cache_key):
        cached = getattr(request.state, cache_key)
        return cached["permissions"]
    result = await validate_token(token, "view")
    setattr(request.state, cache_key, result)
    return result["permissions"]


async def require_edit_permission(request: Request) -> list[str]:
    token = extract_token_from_request(request)
    if not token:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)
    cache_key = f"{_TOKEN_CACHE_PREFIX}{_token_fingerprint(token)}"
    if hasattr(request.state, cache_key):
        cached = getattr(request.state, cache_key)
        return cached["permissions"]
    result = await validate_token(token, "edit")
    setattr(request.state, cache_key, result)
    return result["permissions"]


async def get_operator_id(request: Request) -> Optional[int]:
    """Extract operator_id from the request's token, using cache if available."""
    token = extract_token_from_request(request)
    if not token:
        return None
    cache_key = f"{_TOKEN_CACHE_PREFIX}{_token_fingerprint(token)}"
    if hasattr(request.state, cache_key):
        return getattr(request.state, cache_key).get("operator_id")
    return None


async def require_manage_permission(request: Request) -> list[str]:
    token = extract_token_from_request(request)
    if not token:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)
    cache_key = f"{_TOKEN_CACHE_PREFIX}{_token_fingerprint(token)}"
    if hasattr(request.state, cache_key):
        cached = getattr(request.state, cache_key)
        return cached["permissions"]
    result = await validate_token(token, "manage")
    setattr(request.state, cache_key, result)
    return result["permissions"]


async def proxy_login(token: str) -> dict:
    response = await _request_auth_api("/api/login", json={"token": token})
    if response.status_code == 400:
        raise HTTPException(status_code=400, detail="token is required")
    if response.status_code == 401:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)
    if response.status_code != 200:
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL)
    return response.json()


async def proxy_me(token: str) -> dict:
    response = await _request_auth_api(
        "/api/me",
        method="GET",
        headers={"Authorization": f"Bearer {token}"},
    )
    if response.status_code == 401:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)
    if response.status_code == 403:
        raise HTTPException(status_code=403, detail=AUTH_FORBIDDEN_DETAIL)
    if response.status_code != 200:
        raise HTTPException(status_code=503, detail=AUTH_UNAVAILABLE_DETAIL)
    return response.json()
