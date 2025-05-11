import uuid
import httpx
import json
from datetime import datetime, timedelta
import secrets
from typing import Optional, Dict, Any, Tuple
import asyncio

from .config import QINGFLOW_API_BASE_URL, QINGFLOW_APP_ID, QINGFLOW_ACCESS_TOKEN, FIELD_ID_MAP


class TokenService:
    def __init__(self):
        self.api_base_url = QINGFLOW_API_BASE_URL
        self.app_id = QINGFLOW_APP_ID
        self.access_token = QINGFLOW_ACCESS_TOKEN
        self.field_id_map = FIELD_ID_MAP
        self.headers = {
            "Content-Type": "application/json",
            "accessToken": self.access_token
        }

    def _format_datetime(self, dt: datetime) -> str:
        """格式化日期时间为青流平台接受的格式"""
        return dt.strftime("%Y-%m-%d %H:%M:%S")

    async def create_token(self, username: str, email: str, expires_in_days: int = 30) -> Dict[str, Any]:
        """创建新的API Token"""
        # 生成唯一ID和Token
        token_id = str(uuid.uuid4())
        token_value = secrets.token_urlsafe(32)
        
        # 计算过期时间
        created_at = datetime.now()
        is_permanent = expires_in_days == -99
        expires_at = None if is_permanent else created_at + timedelta(days=expires_in_days)
        
        # 准备请求体
        answers = []
        
        # ID字段
        answers.append({
            "queId": int(self.field_id_map["id"]),
            "queTitle": "id",
            "values": [{"value": token_id}]
        })
        
        # 激活状态
        answers.append({
            "queId": int(self.field_id_map["active"]),
            "queTitle": "active",
            "values": [{"value": "true"}]
        })
        
        # 用户名
        answers.append({
            "queId": int(self.field_id_map["username"]),
            "queTitle": "username",
            "values": [{"value": username}]
        })
        
        # 邮箱
        answers.append({
            "queId": int(self.field_id_map["email"]),
            "queTitle": "email",
            "values": [{"value": email}]
        })
        
        # Token值
        answers.append({
            "queId": int(self.field_id_map["token"]),
            "queTitle": "token",
            "values": [{"value": token_value}]
        })
        
        # 创建时间
        answers.append({
            "queId": int(self.field_id_map["created_at"]),
            "queTitle": "created_at",
            "values": [{"value": self._format_datetime(created_at)}]
        })
        
        # 过期时间
        if not is_permanent:
            answers.append({
                "queId": int(self.field_id_map["expires_at"]),
                "queTitle": "expires_at",
                "values": [{"value": self._format_datetime(expires_at)}]
            })
        
        # 是否永久有效
        answers.append({
            "queId": int(self.field_id_map["is_permanent"]),
            "queTitle": "is_permanent",
            "values": [{"value": str(is_permanent).lower()}]
        })
        
        # 发送请求创建记录
        max_retries = 3
        retry_count = 0
        
        while retry_count < max_retries:
            try:
                url = f"{self.api_base_url}/app/{self.app_id}/apply"
                payload = {"answers": answers}
                
                async with httpx.AsyncClient(timeout=30.0) as client:
                    response = await client.post(url, headers=self.headers, json=payload)
                    response.raise_for_status()
                    
                    return {
                        "id": token_id,
                        "token": token_value,
                        "created_at": created_at.isoformat(),
                        "expires_at": expires_at.isoformat() if expires_at else None,
                        "is_permanent": is_permanent
                    }
                    
            except httpx.ConnectTimeout:
                # 连接超时，尝试重试
                retry_count += 1
                if retry_count >= max_retries:
                    raise Exception(f"创建Token连接超时，已重试{retry_count}次")
                # 等待短暂时间后重试
                await asyncio.sleep(1)
            except httpx.HTTPError as e:
                raise Exception(f"Failed to create token: {str(e)}")

    async def validate_token(self, token: str) -> Tuple[bool, Optional[Dict[str, Any]], Optional[str]]:
        """验证Token是否有效，并返回Token数据和数据ID"""
        max_retries = 3
        retry_count = 0
        
        while retry_count < max_retries:
            try:
                # 查询Token
                url = f"{self.api_base_url}/app/{self.app_id}/apply/filter"
                payload = {
                    "pageSize": 1,
                    "pageNum": 1,
                    "queries": [
                        {
                            "queId": int(self.field_id_map["token"]),
                            "queTitle": "token",
                            "searchKey": token
                        }
                    ]
                }
                
                async with httpx.AsyncClient(timeout=30.0) as client:
                    response = await client.post(url, headers=self.headers, json=payload)
                    response.raise_for_status()
                    data = response.json()
                    
                    # 检查是否有结果
                    results = data.get("result", {}).get("result", [])
                    if not results or len(results) == 0:
                        return False, None, None
                    
                    # 获取数据ID
                    apply_id = results[0].get("applyId")
                    
                    # 解析Token数据
                    token_data = {}
                    for answer in results[0].get("answers", []):
                        que_id = str(answer.get("queId"))
                        
                        # 查找字段名称
                        field_name = None
                        for key, value in self.field_id_map.items():
                            if value == que_id:
                                field_name = key
                                break
                        
                        if field_name and answer.get("values") and len(answer.get("values")) > 0:
                            token_data[field_name] = answer.get("values")[0].get("value")
                    
                    # 验证Token是否有效
                    if not token_data:
                        return False, None, None
                    
                    # 检查是否激活
                    if token_data.get("active", "").lower() != "true":
                        return False, None, None
                    
                    # 检查是否永久有效或未过期
                    is_permanent = token_data.get("is_permanent", "").lower() == "true"
                    if not is_permanent:
                        expires_at = token_data.get("expires_at")
                        if expires_at:
                            expires_date = datetime.strptime(expires_at, "%Y-%m-%d %H:%M:%S")
                            if datetime.now() > expires_date:
                                return False, None, None
                    
                    return True, token_data, apply_id
                    
            except httpx.ConnectTimeout:
                # 连接超时，尝试重试
                retry_count += 1
                if retry_count >= max_retries:
                    raise Exception(f"Token验证连接超时，已重试{retry_count}次")
                # 等待短暂时间后重试
                await asyncio.sleep(1)
            except httpx.HTTPError as e:
                raise Exception(f"Token validation failed: {str(e)}")
            except Exception as e:
                raise Exception(f"Token validation error: {str(e)}")

    async def renew_token(self, token: str, extend_days: int = 30) -> Dict[str, Any]:
        """
        续用或修改Token有效期
        extend_days参数：
        - 正数：设置或延长指定天数
        - -99：设为永久有效
        """
        try:
            # 先验证Token并获取数据ID
            is_valid, token_data, apply_id = await self.validate_token(token)
            
            if not is_valid or not apply_id:
                raise Exception("无效的Token或Token已过期")
                
            # 检查是否为永久Token
            is_permanent = token_data.get("is_permanent", "").lower() == "true"
            
            # 准备更新请求
            answers = []
            old_expires_at = token_data.get("expires_at")
            
            # 情况1: 将永久有效的token修改为有限期限
            if is_permanent and extend_days != -99:
                # 计算新的过期时间（从当前时间起）
                current_time = datetime.now()
                new_expires_at = current_time + timedelta(days=extend_days)
                
                # 更新is_permanent字段为false
                answers.append({
                    "queId": int(self.field_id_map["is_permanent"]),
                    "queTitle": "is_permanent",
                    "values": [{"value": "false"}]
                })
                
                # 设置新的过期时间
                answers.append({
                    "queId": int(self.field_id_map["expires_at"]),
                    "queTitle": "expires_at",
                    "values": [{"value": self._format_datetime(new_expires_at)}]
                })
                
                message = "Token已从永久有效修改为有限期限"
            
            # 情况2: 当前已是永久有效，请求再次设为永久有效
            elif is_permanent and extend_days == -99:
                return {
                    "status": "success",
                    "message": "Token已是永久有效状态",
                    "token": token,
                    "is_permanent": True
                }
            
            # 情况3: 将有期限的token设为永久有效
            elif not is_permanent and extend_days == -99:
                # 更新is_permanent字段
                answers.append({
                    "queId": int(self.field_id_map["is_permanent"]),
                    "queTitle": "is_permanent",
                    "values": [{"value": "true"}]
                })
                
                # 清空expires_at字段
                answers.append({
                    "queId": int(self.field_id_map["expires_at"]),
                    "queTitle": "expires_at",
                    "values": [{"value": ""}]
                })
                
                new_expires_at = None
                message = "Token已设为永久有效"
            
            # 情况4: 常规延期
            else:
                # 计算新的过期时间
                current_expires_at = datetime.strptime(token_data.get("expires_at"), "%Y-%m-%d %H:%M:%S")
                new_expires_at = current_expires_at + timedelta(days=extend_days)
                
                # 更新过期时间
                answers.append({
                    "queId": int(self.field_id_map["expires_at"]),
                    "queTitle": "expires_at",
                    "values": [{"value": self._format_datetime(new_expires_at)}]
                })
                
                message = "Token有效期已延长"
            
            # 发送更新请求
            url = f"{self.api_base_url.replace('/app', '')}/{self.app_id}/apply/{apply_id}"
            payload = {"answers": answers}
            
            max_retries = 3
            retry_count = 0
            
            while retry_count < max_retries:
                try:
                    async with httpx.AsyncClient(timeout=30.0) as client:
                        response = await client.post(url, headers=self.headers, json=payload)
                        response.raise_for_status()
                        break
                except httpx.ConnectTimeout:
                    # 连接超时，尝试重试
                    retry_count += 1
                    if retry_count >= max_retries:
                        raise Exception(f"续期Token连接超时，已重试{retry_count}次")
                    # 等待短暂时间后重试
                    await asyncio.sleep(1)
            
            # 构建返回结果
            result = {
                "status": "success",
                "message": message,
                "token": token,
                "old_status": "永久有效" if is_permanent else f"有效期至 {old_expires_at}",
                "is_permanent": extend_days == -99
            }
            
            # 添加新的过期时间信息（如果不是永久有效）
            if extend_days != -99:
                formatted_expires = self._format_datetime(new_expires_at) if new_expires_at else None
                result["new_expires_at"] = formatted_expires
                result["extended_days"] = extend_days
            
            return result
                
        except httpx.HTTPError as e:
            raise Exception(f"续期Token失败: {str(e)}")
        except Exception as e:
            raise Exception(f"续期Token错误: {str(e)}")