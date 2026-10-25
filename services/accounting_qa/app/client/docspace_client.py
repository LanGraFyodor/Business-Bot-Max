import httpx
from typing import Any, Dict, Optional, Tuple

class DocspaceApiClient:
    """Асинхронный клиент для взаимодействия с Go REST API системы Docspace."""

    def __init__(self, base_url: str, timeout: float = 15.0):
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    async def get_health(self) -> Tuple[int, Dict[str, Any], float]:
        """Проверка доступности сервера /api/health."""
        url = f"{self.base_url}/api/health"
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            resp = await client.get(url)
            try:
                data = resp.json()
            except Exception:
                data = {"raw": resp.text}
            return resp.status_code, data, resp.elapsed.total_seconds() * 1000

    async def get_channels_unauthorized(self) -> Tuple[int, Dict[str, Any], float]:
        """Тест безопасности: запрос каналов без авторизационного токена (ожидается 401)."""
        url = f"{self.base_url}/api/integrations/channels"
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            resp = await client.get(url)
            try:
                data = resp.json()
            except Exception:
                data = {"raw": resp.text}
            return resp.status_code, data, resp.elapsed.total_seconds() * 1000

    async def get_channels(self, token: str) -> Tuple[int, Any, float]:
        """Получение списка доступных каналов команды по токену интеграции."""
        url = f"{self.base_url}/api/integrations/channels"
        headers = {"Authorization": f"Bearer {token}"}
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            resp = await client.get(url, headers=headers)
            try:
                data = resp.json()
            except Exception:
                data = {"raw": resp.text}
            return resp.status_code, data, resp.elapsed.total_seconds() * 1000

    async def upload_document(
        self,
        token: str,
        channel_id: str,
        title: str,
        filename: str,
        external_id: str,
        file_bytes: bytes,
        review: bool = True,
        due: Optional[str] = None,
        description: Optional[str] = None,
    ) -> Tuple[int, Dict[str, Any], float]:
        """
        Импорт бинарного отчета 1С через POST /api/integrations/documents.
        Параметр external_id гарантирует идемпотентность.
        """
        url = f"{self.base_url}/api/integrations/documents"
        params: Dict[str, str] = {
            "channelId": channel_id,
            "title": title,
            "filename": filename,
            "externalId": external_id,
            "review": "1" if review else "0",
        }
        if due:
            params["due"] = due
        if description:
            params["description"] = description

        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/octet-stream",
        }

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            resp = await client.post(url, params=params, headers=headers, content=file_bytes)
            try:
                data = resp.json()
            except Exception:
                data = {"raw": resp.text}
            return resp.status_code, data, resp.elapsed.total_seconds() * 1000

    async def get_document_status(self, token: str, external_id: str) -> Tuple[int, Dict[str, Any], float]:
        """Чтение статуса и решения по документу по externalId."""
        url = f"{self.base_url}/api/integrations/documents"
        params = {"externalId": external_id}
        headers = {"Authorization": f"Bearer {token}"}
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            resp = await client.get(url, params=params, headers=headers)
            try:
                data = resp.json()
            except Exception:
                data = {"raw": resp.text}
            return resp.status_code, data, resp.elapsed.total_seconds() * 1000
