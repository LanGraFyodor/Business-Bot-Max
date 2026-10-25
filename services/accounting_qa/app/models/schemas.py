from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field

class ReportType(str, Enum):
    PLAN_FACT = "plan_fact"
    CASH_FLOW = "cash_flow"
    PAYROLL = "payroll"
    RECONCILIATION = "reconciliation"

class TestStatus(str, Enum):
    PASSED = "PASSED"
    FAILED = "FAILED"
    ERROR = "ERROR"
    SKIPPED = "SKIPPED"

class MockReportConfig(BaseModel):
    report_type: ReportType = Field(default=ReportType.PLAN_FACT, description="Тип отчета 1С")
    row_count: int = Field(default=5, ge=1, le=100, description="Количество строк в отчете")
    inject_anomaly: bool = Field(default=False, description="Внедрить ли аномалию (превышение лимита, отриц. значения) для проверки валидации")
    company_name: str = Field(default="ООО Озон Маркетплейс - Учетные системы", description="Название организации")
    period: str = Field(default="2026-Q3", description="Отчетный период")

class MockReportResponse(BaseModel):
    report_type: ReportType
    filename: str
    csv_content: str
    row_count: int
    checksum_sha256: str
    summary_totals: Dict[str, Any]

class TestCaseResult(BaseModel):
    __test__ = False
    test_id: str
    name: str
    description: str
    status: TestStatus
    duration_ms: float
    http_status: Optional[int] = None
    error_message: Optional[str] = None
    details: Dict[str, Any] = Field(default_factory=dict)

class TestSuiteRunRequest(BaseModel):
    __test__ = False
    target_api_url: Optional[str] = Field(default=None, description="URL Go Docspace API (по умолчанию берется из настроек)")
    integration_token: Optional[str] = Field(default=None, description="Токен интеграции (Bearer). Если не указан, берется из окружения")
    channel_id: Optional[str] = Field(default=None, description="Канал для загрузки отчета (если не указан, берется первый доступный)")
    include_flaky_benchmark: bool = Field(default=True, description="Выполнять ли тест на детекцию нестабильности (flaky)")
    benchmark_iterations: int = Field(default=3, ge=1, le=10, description="Число итераций для бенчмарка стабильности")

class TestSuiteExecution(BaseModel):
    __test__ = False
    suite_id: str
    timestamp: datetime
    target_api_url: str
    total_tests: int
    passed_tests: int
    failed_tests: int
    total_duration_ms: float
    is_flaky_detected: bool
    flaky_summary: Optional[str] = None
    results: List[TestCaseResult]

class MetricsSummary(BaseModel):
    total_suites_executed: int
    passed_suites: int
    failed_suites: int
    success_rate_percent: float
    avg_duration_ms: float
    flaky_events_detected: int

class HealthResponse(BaseModel):
    status: str
    timestamp: datetime
    service: str
    version: str
    target_api_url: str
