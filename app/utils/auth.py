from fastapi import Request, HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from typing import Optional, Dict, Any

from .token_service import TokenService

# 定义安全模型
security = HTTPBearer()


async def get_current_token(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> Dict[str, Any]:
    """
    依赖项函数，用于验证请求中的Token
    """
    token_service = TokenService()
    token = credentials.credentials
    
    is_valid, token_data, _ = await token_service.validate_token(token)
    
    if not is_valid or not token_data:
        raise HTTPException(
            status_code=401,
            detail="无效的认证凭据",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    return token_data