"""
Модуль структурированного цветного логирования для ETL-пайплайна.
Обеспечивает информативный вывод: что загружено, сколько строк, ошибки и статус.
"""
import sys
import time
from datetime import datetime

class Colors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
    RESET = '\033[0m'

def _timestamp() -> str:
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")

def log_info(msg: str):
    print(f"[{_timestamp()}] {Colors.CYAN}[INFO]{Colors.RESET} {msg}", flush=True)

def log_success(msg: str):
    print(f"[{_timestamp()}] {Colors.GREEN}{Colors.BOLD}[SUCCESS]{Colors.RESET} {msg}", flush=True)

def log_warning(msg: str):
    print(f"[{_timestamp()}] {Colors.YELLOW}{Colors.BOLD}[WARNING]{Colors.RESET} {msg}", flush=True)

def log_error(msg: str):
    print(f"[{_timestamp()}] {Colors.RED}{Colors.BOLD}[ERROR]{Colors.RESET} {msg}", file=sys.stderr, flush=True)

def log_section(title: str):
    border = "=" * 70
    print(f"\n{Colors.BLUE}{border}{Colors.RESET}")
    print(f"{Colors.BOLD}{Colors.BLUE}  {title}{Colors.RESET}")
    print(f"{Colors.BLUE}{border}{Colors.RESET}", flush=True)

def print_summary_table(results: list):
    """
    Выводит итоговую таблицу статистики ETL-процесса.
    results: список словарей с полями:
      entity, read_count, valid_count, inserted_count, updated_count, error_count, duration_sec, status
    """
    header = f"{'Сущность / Таблица':<28} | {'Считано':<8} | {'Валидно':<8} | {'Записано':<8} | {'Ошибки':<8} | {'Статус':<10}"
    separator = "-" * len(header)
    
    print("\n" + "=" * len(header))
    print(f"{Colors.BOLD}ИТОГОВЫЙ ОТЧЕТ ETL-ПРОЦЕССА (DATA PIPELINE SUMMARY){Colors.RESET}")
    print("=" * len(header))
    print(header)
    print(separator)
    
    total_read = 0
    total_written = 0
    total_errors = 0
    
    for r in results:
        total_read += r['read_count']
        total_written += r['inserted_count']
        total_errors += r['error_count']
        
        status_color = Colors.GREEN if r['status'] == 'OK' else (Colors.YELLOW if r['status'] == 'WARN' else Colors.RED)
        status_str = f"{status_color}{r['status']:<10}{Colors.RESET}"
        
        print(f"{r['entity']:<28} | {r['read_count']:<8} | {r['valid_count']:<8} | {r['inserted_count']:<8} | {r['error_count']:<8} | {status_str}")
    
    print(separator)
    print(f"{'ИТОГО':<28} | {total_read:<8} | {'-':<8} | {total_written:<8} | {total_errors:<8} | {'ВЫПОЛНЕНО':<10}")
    print("=" * len(header) + "\n")
