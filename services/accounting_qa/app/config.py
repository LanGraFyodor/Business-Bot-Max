import os
from dataclasses import dataclass

@dataclass
class Settings:
    docspace_api_url: str = os.getenv("DOCSPACE_API_URL", "http://127.0.0.1:8080").rstrip("/")
    integration_token: str = os.getenv("DOCSPACE_TEST_INTEGRATION_TOKEN", "").strip()
    service_port: int = int(os.getenv("ACCOUNTING_QA_PORT", "8081"))
    default_timeout_sec: float = float(os.getenv("DEFAULT_TIMEOUT_SEC", "15.0"))
    history_limit: int = int(os.getenv("HISTORY_LIMIT", "50"))

settings = Settings()
