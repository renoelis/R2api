import os
from dotenv import load_dotenv

# 加载环境变量
load_dotenv()

# 青流平台API配置
QINGFLOW_API_BASE_URL = "https://api.qingflow.com"
QINGFLOW_APP_ID = "aqddbt0obk02"
QINGFLOW_ACCESS_TOKEN = os.getenv("QINGFLOW_ACCESS_TOKEN", "72e12c93-debd-4def-a6ff-708c671425c9")

# 青流平台字段ID
FIELD_ID_MAP = {
    "id": "360860723",
    "active": "360860724",
    "username": "360860725",
    "email": "360860726",
    "token": "360860727",
    "created_at": "360860728",
    "expires_at": "360860729",
    "is_permanent": "360860730"
}

# 应用配置
MAX_FILE_SIZE = 200 * 1024 * 1024  # 200MB
API_VERSION = "v1"
SERVICE_NAME = "r2-uploader"

# 日志配置
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")