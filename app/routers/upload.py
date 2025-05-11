from fastapi import APIRouter, HTTPException, Depends, UploadFile, Form, File
from pydantic import BaseModel, HttpUrl, Field, validator
from typing import Dict, Any, Optional

from ..utils.auth import get_current_token
from ..utils.file_service import FileService
from ..utils.config import MAX_FILE_SIZE

router = APIRouter(prefix="/R2api", tags=["upload"])


class UploadRequest(BaseModel):
    fileUrl: HttpUrl = Field(..., description="要下载的文件URL")
    bucketName: str = Field(..., min_length=1, description="R2存储桶名称")
    objectKey: str = Field(..., min_length=1, description="对象键名(可包含路径，如'images/photo.jpg')")
    endpoint: HttpUrl = Field(..., description="R2存储桶端点URL")
    accessKeyId: str = Field(..., min_length=1, description="访问密钥ID")
    secretAccessKey: str = Field(..., min_length=1, description="访问密钥")
    customdomain: Optional[HttpUrl] = Field(None, description="自定义域名(可选)")
    
    @validator('objectKey')
    def validate_object_key(cls, v):
        # 验证 objectKey 不以 / 开头
        if v.startswith('/'):
            raise ValueError("objectKey 不能以 '/' 开头")
        return v


class UploadResponse(BaseModel):
    status: str
    message: str
    data: Dict[str, Any]


@router.post("/upload", response_model=UploadResponse)
async def upload_file(
    request: UploadRequest,
    token_data: Dict[str, Any] = Depends(get_current_token)
):
    """
    从URL下载文件并上传到R2存储桶
    """
    file_service = FileService()
    
    try:
        # 下载文件
        file, content_type, file_size = await file_service.download_file(str(request.fileUrl))
        
        # 验证文件大小
        if file_size > MAX_FILE_SIZE:
            raise HTTPException(
                status_code=413,
                detail=f"文件大小超过限制: {file_size} > {MAX_FILE_SIZE} 字节"
            )
        
        # 上传到R2
        result = await file_service.upload_to_r2(
            file=file,
            content_type=content_type,
            bucket_name=request.bucketName,
            object_key=request.objectKey,
            endpoint=str(request.endpoint),
            access_key_id=request.accessKeyId,
            secret_access_key=request.secretAccessKey,
            custom_domain=str(request.customdomain) if request.customdomain else None
        )
        
        return {
            "status": "success",
            "message": "文件上传成功",
            "data": result
        }
        
    except ValueError as e:
        # 处理文件大小或其他验证错误
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        # 处理其他错误
        raise HTTPException(status_code=500, detail=f"上传失败: {str(e)}")


@router.post("/upload-direct", response_model=UploadResponse)
async def upload_file_directly(
    bucket_name: str = Form(..., description="R2存储桶名称"),
    object_key: Optional[str] = Form(None, description="对象键名(可包含路径，如'images/photo.jpg')，为空则使用原文件名"),
    endpoint: str = Form(..., description="R2存储桶端点URL"),
    access_key_id: str = Form(..., description="访问密钥ID"),
    secret_access_key: str = Form(..., description="访问密钥"),
    custom_domain: Optional[str] = Form(None, description="自定义域名(可选)"),
    file: UploadFile = File(..., description="要上传的文件"),
    token_data: Dict[str, Any] = Depends(get_current_token)
):
    """
    通过form-data直接上传文件到R2存储桶
    """
    file_service = FileService()
    
    try:
        # 如果未提供object_key，则使用原文件名
        if not object_key or object_key.strip() == "":
            object_key = file.filename
        
        # 验证object_key
        if object_key.startswith('/'):
            raise HTTPException(
                status_code=400,
                detail="objectKey 不能以 '/' 开头"
            )
        
        # 上传到R2
        result = await file_service.upload_file_directly(
            upload_file=file,
            bucket_name=bucket_name,
            object_key=object_key,
            endpoint=endpoint,
            access_key_id=access_key_id,
            secret_access_key=secret_access_key,
            custom_domain=custom_domain
        )
        
        return {
            "status": "success",
            "message": "文件上传成功",
            "data": result
        }
        
    except ValueError as e:
        # 处理文件大小或其他验证错误
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        # 处理其他错误
        raise HTTPException(status_code=500, detail=f"上传失败: {str(e)}")