import hashlib
import random
from typing import Dict, Any, Tuple
from app.models.schemas import ReportType, MockReportConfig, MockReportResponse

DEPARTMENTS = [
    "Бухгалтерия", "Казначейство", "Финансовый контроль", 
    "Логистика и фулфилмент Ozon", "IT-инфраструктура", 
    "Разработка учетных систем", "Коммерческий департамент"
]

CASH_FLOW_ITEMS = [
    ("Поступления от покупателей (Ozon Marketplace)", "Приход", 500000, 2500000),
    ("Оплата поставщикам и фулфилмент-партнерам", "Расход", 200000, 1200000),
    ("Аренда сортировочных центров", "Расход", 150000, 600000),
    ("Фонд оплаты труда (ФОТ)", "Расход", 300000, 1800000),
    ("Налоговые платежи (НДС, Налог на прибыль)", "Расход", 100000, 800000),
    ("Прочие доходы от эквайринга", "Приход", 50000, 300000),
]

EMPLOYEES = [
    ("001042", "Иванов Алексей Петрович", "Ведущий разработчик 1С", 220000),
    ("001043", "Смирнова Елена Сергеевна", "Главный бухгалтер", 250000),
    ("001044", "Кузнецов Дмитрий Игоревич", "QA Automation Engineer (Python)", 180000),
    ("001045", "Васильева Ольга Николаевна", "Финансовый аналитик", 160000),
    ("001046", "Попов Михаил Андреевич", "Инженер сопровождения систем учета", 150000),
    ("001047", "Федорова Екатерина Павловна", "Специалист казначейства", 140000),
]

class MockAccountingReportGenerator:
    """Генератор синтетических отчетов 1С для тестирования учетных систем."""

    @staticmethod
    def generate(config: MockReportConfig) -> MockReportResponse:
        if config.report_type == ReportType.PLAN_FACT:
            filename, content, totals = MockAccountingReportGenerator._generate_plan_fact(config)
        elif config.report_type == ReportType.CASH_FLOW:
            filename, content, totals = MockAccountingReportGenerator._generate_cash_flow(config)
        elif config.report_type == ReportType.PAYROLL:
            filename, content, totals = MockAccountingReportGenerator._generate_payroll(config)
        elif config.report_type == ReportType.RECONCILIATION:
            filename, content, totals = MockAccountingReportGenerator._generate_reconciliation(config)
        else:
            filename, content, totals = MockAccountingReportGenerator._generate_plan_fact(config)

        checksum = hashlib.sha256(content.encode("utf-8")).hexdigest()

        return MockReportResponse(
            report_type=config.report_type,
            filename=filename,
            csv_content=content,
            row_count=config.row_count,
            checksum_sha256=checksum,
            summary_totals=totals,
        )

    @staticmethod
    def _generate_plan_fact(config: MockReportConfig) -> Tuple[str, str, Dict[str, Any]]:
        filename = f"1c_plan_fact_{config.period.lower()}_{random.randint(100, 999)}.csv"
        lines = ["Отдел;План;Факт;Отклонение;Процент_выполнения"]
        
        total_plan = 0
        total_fact = 0
        
        depts = DEPARTMENTS.copy()
        random.shuffle(depts)
        selected_depts = (depts * ((config.row_count // len(depts)) + 1))[:config.row_count]

        for i, dept in enumerate(selected_depts):
            plan = random.randint(10, 50) * 10000
            
            # Если внедряем аномалию на последней строке: критический перерасход
            if config.inject_anomaly and i == config.row_count - 1:
                fact = int(plan * 1.85)  # +85% перерасход
            else:
                variance_ratio = random.uniform(0.92, 1.08)
                fact = int(plan * variance_ratio)

            deviation = fact - plan
            percent = round((fact / plan) * 100, 2) if plan > 0 else 0.0

            total_plan += plan
            total_fact += fact
            lines.append(f"{dept};{plan};{fact};{deviation};{percent}%")

        totals = {
            "total_plan_rub": total_plan,
            "total_fact_rub": total_fact,
            "total_deviation_rub": total_fact - total_plan,
            "total_execution_percent": round((total_fact / total_plan) * 100, 2) if total_plan > 0 else 0,
            "has_critical_overrun": config.inject_anomaly
        }

        return filename, "\n".join(lines) + "\n", totals

    @staticmethod
    def _generate_cash_flow(config: MockReportConfig) -> Tuple[str, str, Dict[str, Any]]:
        filename = f"1c_dds_cashflow_{config.period.lower()}_{random.randint(100, 999)}.csv"
        lines = ["Статья_ДДС;Направление;Сумма_руб;Корр_счет;Статус"]
        
        total_inflow = 0
        total_outflow = 0

        for i in range(config.row_count):
            item_tmpl = CASH_FLOW_ITEMS[i % len(CASH_FLOW_ITEMS)]
            title, direction, min_amt, max_amt = item_tmpl
            amt = random.randint(min_amt // 1000, max_amt // 1000) * 1000

            if config.inject_anomaly and i == 0:
                amt = -abs(amt)  # Аномалия: отрицательная сумма в проводке ДДС

            if direction == "Приход":
                total_inflow += amt
            else:
                total_outflow += amt

            corr_acc = f"51.0{random.randint(1, 3)}"
            lines.append(f"{title};{direction};{amt};{corr_acc};Проведено")

        totals = {
            "total_inflow_rub": total_inflow,
            "total_outflow_rub": total_outflow,
            "net_cash_flow_rub": total_inflow - total_outflow,
            "has_negative_amount": config.inject_anomaly
        }

        return filename, "\n".join(lines) + "\n", totals

    @staticmethod
    def _generate_payroll(config: MockReportConfig) -> Tuple[str, str, Dict[str, Any]]:
        filename = f"1c_payroll_sheet_{config.period.lower()}_{random.randint(100, 999)}.csv"
        lines = ["Табельный_номер;Сотрудник;Должность;Оклад;НДФЛ_13;К_выплате"]
        
        total_gross = 0
        total_net = 0

        for i in range(config.row_count):
            tab, name, pos, salary = EMPLOYEES[i % len(EMPLOYEES)]
            ndfl = int(salary * 0.13)
            net = salary - ndfl

            total_gross += salary
            total_net += net

            lines.append(f"{tab};{name};{pos};{salary};{ndfl};{net}")

        totals = {
            "total_gross_rub": total_gross,
            "total_ndfl_rub": total_gross - total_net,
            "total_net_rub": total_net,
            "headcount": config.row_count
        }

        return filename, "\n".join(lines) + "\n", totals

    @staticmethod
    def _generate_reconciliation(config: MockReportConfig) -> Tuple[str, str, Dict[str, Any]]:
        filename = f"1c_reconciliation_act_{config.period.lower()}_{random.randint(100, 999)}.csv"
        lines = ["Дата;Документ_основание;Дебет_руб;Кредит_руб;Сальдо_руб"]
        
        running_balance = 0
        for i in range(config.row_count):
            day = min(i + 1, 28)
            date_str = f"2026-09-{day:02d}"
            doc = f"Акт выполненных работ № {1000 + i}"
            debit = random.randint(10, 80) * 1000 if i % 2 == 0 else 0
            credit = random.randint(10, 80) * 1000 if i % 2 == 1 else 0
            
            running_balance += (debit - credit)
            lines.append(f"{date_str};{doc};{debit};{credit};{running_balance}")

        totals = {
            "final_balance_rub": running_balance,
            "records_count": config.row_count
        }

        return filename, "\n".join(lines) + "\n", totals
