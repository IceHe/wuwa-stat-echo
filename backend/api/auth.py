from fastapi import APIRouter, Body, HTTPException, Request

from auth import AUTH_INVALID_DETAIL, extract_token_from_request, proxy_login, proxy_me


router = APIRouter(prefix="/auth", tags=["auth"])


@router.post("/login")
async def login(payload: dict = Body(...)):
    token = str(payload.get("token", "")).strip()
    return await proxy_login(token)


@router.get("/me")
async def me(request: Request):
    token = extract_token_from_request(request)
    if not token:
        raise HTTPException(status_code=401, detail=AUTH_INVALID_DETAIL)
    return await proxy_me(token)
