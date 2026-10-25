from datetime import datetime, timezone
from typing import List, Optional

from fastapi import FastAPI, Query, Response
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import RedirectResponse

from app.config import settings
from app.generator.mock_reports import MockAccountingReportGenerator
from app.models.schemas import (
    HealthResponse,
    MetricsSummary,
    MockReportConfig,
    MockReportResponse,
    ReportType,
    TestSuiteExecution,
    TestSuiteRunRequest,
)
from app.runner.test_runner import qa_runner

app = FastAPI(
    title="Ozon Tech - Accounting Systems QA & Mock Runner",
    description=(
        "Микросервис автоматизированного тестирования, сбора метрик качества "
        "и генерации синтетических отчетов 1С для учетных систем (Ozon Tech)."
    ),
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

# Разрешаем CORS для локальной разработки и тестирования
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/", include_in_schema=False)
async def root_redirect():
    return RedirectResponse(url="/docs")


@app.get("/api/v1/health", response_model=HealthResponse, tags=["Health & System"])
async def get_health():
    """Проверка доступности самого сервиса тестирования учетных систем."""
    return HealthResponse(
        status="ok",
        timestamp=datetime.now(timezone.utc),
        service="accounting-qa-runner",
        version="1.0.0",
        target_api_url=settings.docspace_api_url,
    )


# ---------------------------------------------------------------------------
# Endpoints: Генерация учетных данных 1С (Mock Generator)
# ---------------------------------------------------------------------------
@app.post(
    "/api/v1/generator/report",
    response_model=MockReportResponse,
    tags=["1C Data Generator"],
    summary="Сгенерировать синтетический отчет 1С",
)
async def generate_mock_report(config: MockReportConfig):
    """
    Генерирует реалистичный отчет 1С (План-факт, ДДС/Cashflow, Расчетная ведомость, Акт сверки).
    Поддерживает инъекцию аномалий для тестирования валидаторов и отказоустойчивости.
    """
    return MockAccountingReportGenerator.generate(config)


@app.get(
    "/api/v1/generator/sample-csv",
    tags=["1C Data Generator"],
    summary="Скачать готовый образец отчета 1С (text/csv)",
)
async def download_sample_csv(
    report_type: ReportType = Query(default=ReportType.PLAN_FACT, description="Тип отчета 1С"),
    rows: int = Query(default=5, ge=1, le=50, description="Количество строк"),
    inject_anomaly: bool = Query(default=False, description="Внедрить аномалию"),
):
    """Возвращает сырой CSV-файл для скачивания и отправки в учетную систему."""
    report = MockAccountingReportGenerator.generate(
        MockReportConfig(report_type=report_type, row_count=rows, inject_anomaly=inject_anomaly)
    )
    return Response(
        content=report.csv_content,
        media_type="text/csv; charset=utf-8",
        headers={"Content-Disposition": f'attachment; filename="{report.filename}"'},
    )


# ---------------------------------------------------------------------------
# Endpoints: Запуск автотестов и метрики (QA & Test Automation)
# ---------------------------------------------------------------------------
@app.post(
    "/api/v1/qa/run-suite",
    response_model=TestSuiteExecution,
    tags=["QA & Test Automation"],
    summary="Запустить комплексный сьют автотестов учетной интеграции",
)
async def run_qa_suite(request: Optional[TestSuiteRunRequest] = None):
    """
    Выполняет набор интеграционных тестов против Go REST API:
    - TC-01: Healthcheck API
    - TC-02: Защита от неавторизованного доступа (401 на отсутствие Bearer)
    - TC-03: Обнаружение каналов интеграции
    - TC-04: Бинарная загрузка отчета 1С с флагом review=1
    - TC-05: Контроль идемпотентности по externalId
    - TC-06: Получение статуса согласования отчета
    - TC-07: Бенчмарк стабильности и детекция flaky-поведения (джиттер)
    """
    req = request or TestSuiteRunRequest()
    return await qa_runner.execute_suite(req)


@app.get(
    "/api/v1/qa/history",
    response_model=List[TestSuiteExecution],
    tags=["QA & Test Automation"],
    summary="История прогонов автоматических тестов",
)
async def get_test_history(limit: int = Query(default=10, ge=1, le=50)):
    """Возвращает историю запусков с деталями по каждому тесту."""
    return qa_runner.get_history(limit=limit)


@app.get(
    "/api/v1/qa/metrics",
    response_model=MetricsSummary,
    tags=["QA & Test Automation"],
    summary="Сводные метрики стабильности и успешности интеграции",
)
async def get_qa_metrics():
    """
    Метрики качества интеграции:
    - Общий процент успешных прогонов (success_rate_percent)
    - Средняя длительность сьюта (avg_duration_ms)
    - Количество зафиксированных событий нестабильности (flaky_events_detected)
    """
    return qa_runner.get_metrics()
