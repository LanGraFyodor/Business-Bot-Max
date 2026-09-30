#!/usr/bin/env python3
"""
Единая точка входа ELT/ETL-пайплайна.
Запуск из одной строки:
  В Docker: docker compose run --rm etl
  Локально: python -m etl.run_pipeline
"""
import sys
import time
from etl.logger import log_section, log_info, log_success, log_error, print_summary_table
from etl.loaders.db import get_connection, test_db_connection
from etl.loaders.legal_loader import load_legal_data
from etl.loaders.reference_loader import load_reference_data
from etl.loaders.seed_loader import load_seed_demo_data

def run_smoke_verification(conn):
    """Проверяет фактическое наполнение таблиц в БД."""
    log_section("ЭТАП 4: SMOKE-ПРОВЕРКА ЦЕЛОСТНОСТИ ДАННЫХ В БД")
    tables = [
        "legal_sources", "business_requirements", "tax_regimes",
        "tax_regional_rates", "signboard_regulations", "checklists",
        "checklist_items", "users", "business_profiles", "stores",
        "product_categories", "products", "orders", "order_items",
        "tax_calculations", "financial_health_assessments", "user_checklist_progress"
    ]
    with conn.cursor() as cur:
        for t in tables:
            cur.execute(f"SELECT COUNT(*) FROM {t};")
            count = cur.fetchone()[0]
            log_info(f"Таблица '{t}': {count} записей в БД")
    log_success("Все ключевые таблицы содержат валидные данные!")

def main():
    start_time = time.time()
    log_section("ЗАПУСК ИДЕМПОТЕНТНОГО ETL-ПАЙПЛАЙНА (ХАКАТОН MAX)")
    
    # 0. Проверка доступности БД
    log_info("Шаг 0: Проверка соединения с PostgreSQL...")
    if not test_db_connection():
        log_error("Критическая ошибка: База данных недоступна. Пайплайн остановлен.")
        sys.exit(1)

    all_stats = []
    conn = get_connection()

    try:
        # 1. Загрузка правовых данных (Extract & Load Legal Data)
        log_section("ЭТАП 1: ОБРАБОТКА И ЗАГРУЗКА ПРАВОВОЙ БАЗЫ И ТРЕБОВАНИЙ")
        legal_stats = load_legal_data(conn)
        all_stats.extend(legal_stats)

        # 2. Загрузка нормативно-справочной информации (Reference Data)
        log_section("ЭТАП 2: ЗАГРУЗКА СИСТЕМНЫХ СПРАВОЧНИКОВ И НАЛОГОВЫХ РЕЖИМОВ")
        ref_stats = load_reference_data(conn)
        all_stats.extend(ref_stats)

        # 3. Загрузка эталонных тестовых сценариев (Demo Seeds)
        log_section("ЭТАП 3: НАПОЛНЕНИЕ ЭТАЛОННЫХ ДЕМО-ДАННЫХ ДЛЯ ЖЮРИ")
        seed_stats = load_seed_demo_data(conn)
        all_stats.extend(seed_stats)

        # 4. Проверка целостности
        run_smoke_verification(conn)

        # 5. Итоговый отчет
        duration = round(time.time() - start_time, 2)
        print_summary_table(all_stats)
        log_success(f"ETL-пайплайн успешно завершен за {duration} сек! Данные консистентны и готовы к работе.")

    except Exception as e:
        log_error(f"Непредвиденный сбой во время выполнения ETL: {e}")
        conn.rollback()
        sys.exit(1)
    finally:
        conn.close()

if __name__ == "__main__":
    main()
