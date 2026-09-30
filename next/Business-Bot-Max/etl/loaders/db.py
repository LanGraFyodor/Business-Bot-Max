"""
Модуль управления подключением к базе данных PostgreSQL.
Поддерживает повторные попытки (retry) при старте в Docker Compose.
"""
import time
import psycopg2
from psycopg2.extras import RealDictCursor
from etl.config import DATABASE_URL, DB_CONNECT_RETRIES, DB_CONNECT_DELAY_SECONDS
from etl.logger import log_info, log_error, log_warning, log_success

def get_connection():
    """
    Возвращает активное соединение с базой данных PostgreSQL.
    Пытается переподключиться при ожидании готовности контейнера базы.
    """
    last_err = None
    for attempt in range(1, DB_CONNECT_RETRIES + 1):
        try:
            conn = psycopg2.connect(DATABASE_URL)
            conn.autocommit = False
            return conn
        except psycopg2.OperationalError as e:
            last_err = e
            if attempt < DB_CONNECT_RETRIES:
                log_warning(f"Ожидание готовности PostgreSQL ({attempt}/{DB_CONNECT_RETRIES})... Повтор через {DB_CONNECT_DELAY_SECONDS}с")
                time.sleep(DB_CONNECT_DELAY_SECONDS)

    log_error(f"Не удалось подключиться к базе данных после {DB_CONNECT_RETRIES} попыток: {last_err}")
    raise last_err

def test_db_connection() -> bool:
    try:
        conn = get_connection()
        with conn.cursor() as cur:
            cur.execute("SELECT version();")
            ver = cur.fetchone()[0]
            log_success(f"Подключение к БД успешно установлено: {ver.split(',')[0]}")
        conn.close()
        return True
    except Exception as e:
        log_error(f"Ошибка проверки связи с БД: {e}")
        return False
