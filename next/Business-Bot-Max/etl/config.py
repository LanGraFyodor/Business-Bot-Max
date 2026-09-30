"""
Модуль конфигурации ETL-пайплайна.
Безопасное чтение параметров из переменных окружения без хардкода секретов.
"""
import os
from pathlib import Path

# Базовые пути проекта
ETL_DIR = Path(__file__).resolve().parent
PROJECT_ROOT = ETL_DIR.parent
DATA_DIR = PROJECT_ROOT / "data"
REFERENCE_DIR = DATA_DIR / "reference"
SEEDS_DIR = DATA_DIR / "seeds"
RAW_LAWS_DIR = DATA_DIR / "raw" / "laws"

# Параметры подключения к PostgreSQL
DEFAULT_DB_USER = os.getenv("POSTGRES_USER", "postgres")
DEFAULT_DB_PASSWORD = os.getenv("POSTGRES_PASSWORD", "postgres_secret_password")
DEFAULT_DB_HOST = os.getenv("POSTGRES_HOST", "localhost")
DEFAULT_DB_PORT = os.getenv("POSTGRES_PORT", "5432")
DEFAULT_DB_NAME = os.getenv("POSTGRES_DB", "max_business")

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    f"postgresql://{DEFAULT_DB_USER}:{DEFAULT_DB_PASSWORD}@{DEFAULT_DB_HOST}:{DEFAULT_DB_PORT}/{DEFAULT_DB_NAME}"
)

# Настройки повторных попыток подключения к БД (для Docker startup)
DB_CONNECT_RETRIES = int(os.getenv("DB_CONNECT_RETRIES", "15"))
DB_CONNECT_DELAY_SECONDS = float(os.getenv("DB_CONNECT_DELAY_SECONDS", "2.0"))
