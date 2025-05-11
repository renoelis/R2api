from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel, EmailStr, Field
from typing import Optional

from ..utils.token_service import TokenService

router = APIRouter(prefix="/R2api", tags=["token"])


class TokenRequest(BaseModel):
    username: str = Field(..., min_length=3, max_length=50, description="用户名")
    email: EmailStr = Field(..., description="邮箱地址")
    expires_in_days: Optional[int] = Field(30, description="令牌有效天数，-99表示永久有效")


class TokenResponse(BaseModel):
    status: str
    token: str
    expires_at: Optional[str]
    is_permanent: bool


class RenewTokenRequest(BaseModel):
    token: str = Field(..., description="要续期的令牌")
    extend_days: int = Field(30, description="延长的天数，-99表示设为永久有效")


class RenewTokenResponse(BaseModel):
    status: str
    message: str
    token: str
    old_status: str  # 原有状态描述
    is_permanent: bool
    new_expires_at: Optional[str] = None  # 如果设置了有限期限，则有值
    extended_days: Optional[int] = None  # 如果设置了有限期限，则有值


@router.post("/register", response_model=TokenResponse)
async def register_token(request: TokenRequest):
    """
    注册新的API令牌
    """
    try:
        token_service = TokenService()
        token_data = await token_service.create_token(
            username=request.username,
            email=request.email,
            expires_in_days=request.expires_in_days
        )
        
        return {
            "status": "success",
            "token": token_data["token"],
            "expires_at": token_data["expires_at"],
            "is_permanent": token_data["is_permanent"]
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"创建令牌失败: {str(e)}")


@router.post("/renew", response_model=RenewTokenResponse)
async def renew_token(request: RenewTokenRequest):
    """
    续期现有的API令牌
    """
    try:
        token_service = TokenService()
        result = await token_service.renew_token(
            token=request.token,
            extend_days=request.extend_days
        )
        
        return result
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))