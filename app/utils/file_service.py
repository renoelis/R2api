import httpx
import asyncio
import boto3
import uuid
import tempfile
import os
from typing import Dict, Any, Tuple, BinaryIO, IO
from botocore.exceptions import ClientError
from fastapi import UploadFile

from .config import MAX_FILE_SIZE


class FileService:
    def __init__(self):
        self.max_file_size = MAX_FILE_SIZE

    async def download_file(self, file_url: str) -> Tuple[BinaryIO, str, int]:
        """
        异步下载文件并返回临时文件对象、内容类型和文件大小
        """
        temp_file = tempfile.NamedTemporaryFile(delete=False)
        content_type = None
        file_size = 0
        
        try:
            async with httpx.AsyncClient(timeout=60.0, follow_redirects=True) as client:
                # 发送HEAD请求获取文件大小和内容类型
                head_response = await client.head(file_url)
                head_response.raise_for_status()
                
                content_type = head_response.headers.get("content-type", "application/octet-stream")
                
                # 检查Content-Length头，如果存在则验证文件大小
                content_length = head_response.headers.get("content-length")
                if content_length and int(content_length) > self.max_file_size:
                    raise ValueError(f"文件大小超过限制：{int(content_length)} > {self.max_file_size}")
                
                # 使用流式下载以支持大文件
                async with client.stream("GET", file_url) as response:
                    response.raise_for_status()
                    
                    async for chunk in response.aiter_bytes(chunk_size=8192):
                        file_size += len(chunk)
                        
                        # 检查文件大小是否超过限制
                        if file_size > self.max_file_size:
                            temp_file.close()
                            os.unlink(temp_file.name)
                            raise ValueError(f"文件大小超过限制：{file_size} > {self.max_file_size}")
                        
                        temp_file.write(chunk)
            
            temp_file.flush()
            temp_file.seek(0)
            return temp_file, content_type, file_size
            
        except httpx.RequestError as e:
            temp_file.close()
            os.unlink(temp_file.name)
            raise Exception(f"下载文件时出错: {str(e)}")
        except ValueError as e:
            # 重新抛出ValueError，用于文件大小验证
            raise
        except Exception as e:
            temp_file.close()
            os.unlink(temp_file.name)
            raise Exception(f"处理文件时出错: {str(e)}")

    async def upload_to_r2(
        self, 
        file: BinaryIO, 
        content_type: str,
        bucket_name: str, 
        object_key: str,
        endpoint: str,
        access_key_id: str,
        secret_access_key: str,
        custom_domain: str
    ) -> Dict[str, Any]:
        """
        将文件上传到R2存储桶并返回公共URL
        """
        try:
            # 创建S3客户端连接R2
            s3_client = boto3.client(
                service_name='s3',
                endpoint_url=endpoint,
                aws_access_key_id=access_key_id,
                aws_secret_access_key=secret_access_key
            )
            
            # 确保获取文件大小
            file.seek(0, os.SEEK_END)
            file_size = file.tell()
            file.seek(0)
            
            # 上传文件
            s3_client.upload_fileobj(
                file,
                bucket_name,
                object_key,
                ExtraArgs={
                    'ContentType': content_type
                }
            )
            
            # 构建公共URL
            if custom_domain:
                # 使用自定义域名
                public_url = f"{custom_domain.rstrip('/')}/{object_key}"
            else:
                # 使用默认R2 URL
                public_url = f"{endpoint.rstrip('/')}/{bucket_name}/{object_key}"
            
            return {
                "public_url": public_url,
                "size": file_size,
                "content_type": content_type
            }
            
        except ClientError as e:
            raise Exception(f"上传到R2时出错: {str(e)}")
        except Exception as e:
            raise Exception(f"处理R2上传时出错: {str(e)}")
        finally:
            # 关闭并删除临时文件
            file.close()
            if hasattr(file, 'name') and os.path.exists(file.name):
                os.unlink(file.name)

    async def upload_file_directly(
        self, 
        upload_file: UploadFile,
        bucket_name: str, 
        object_key: str,
        endpoint: str,
        access_key_id: str,
        secret_access_key: str,
        custom_domain: str
    ) -> Dict[str, Any]:
        """
        直接上传文件到R2存储桶并返回公共URL
        """
        temp_file = tempfile.NamedTemporaryFile(delete=False)
        file_size = 0
        
        try:
            # 将上传的文件保存到临时文件
            content = await upload_file.read()
            file_size = len(content)
            
            # 检查文件大小是否超过限制
            if file_size > self.max_file_size:
                raise ValueError(f"文件大小超过限制：{file_size} > {self.max_file_size}")
            
            # 写入临时文件
            temp_file.write(content)
            temp_file.flush()
            temp_file.seek(0)
            
            # 获取content_type
            content_type = upload_file.content_type or "application/octet-stream"
            
            # 创建S3客户端连接R2
            s3_client = boto3.client(
                service_name='s3',
                endpoint_url=endpoint,
                aws_access_key_id=access_key_id,
                aws_secret_access_key=secret_access_key
            )
            
            # 上传文件
            s3_client.upload_fileobj(
                temp_file,
                bucket_name,
                object_key,
                ExtraArgs={
                    'ContentType': content_type
                }
            )
            
            # 构建公共URL
            if custom_domain:
                # 使用自定义域名
                public_url = f"{custom_domain.rstrip('/')}/{object_key}"
            else:
                # 使用默认R2 URL
                public_url = f"{endpoint.rstrip('/')}/{bucket_name}/{object_key}"
            
            return {
                "public_url": public_url,
                "size": file_size,
                "content_type": content_type,
                "file_name": upload_file.filename
            }
            
        except ClientError as e:
            raise Exception(f"上传到R2时出错: {str(e)}")
        except ValueError as e:
            # 重新抛出ValueError，用于文件大小验证
            raise
        except Exception as e:
            raise Exception(f"处理文件上传时出错: {str(e)}")
        finally:
            # 关闭并删除临时文件
            temp_file.close()
            if os.path.exists(temp_file.name):
                os.unlink(temp_file.name)