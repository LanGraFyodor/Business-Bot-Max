import asyncio
from datetime import datetime, timezone
import statistics
import time
import uuid
from typing import List, Optional

from app.client.docspace_client import DocspaceApiClient
from app.config import settings
from app.generator.mock_reports import MockAccountingReportGenerator
from app.models.schemas import (
    MockReportConfig,
    ReportType,
    TestCaseResult,
    TestStatus,
    TestSuiteExecution,
    TestSuiteRunRequest,
    MetricsSummary,
)

class AccountingQARunner:
    """Движок автоматизированного тестирования учетных интеграций."""

    def __init__(self):
        self.history: List[TestSuiteExecution] = []

    async def execute_suite(self, request: TestSuiteRunRequest) -> TestSuiteExecution:
        target_url = (request.target_api_url or settings.docspace_api_url).rstrip("/")
        token = request.integration_token or settings.integration_token
        client = DocspaceApiClient(base_url=target_url, timeout=settings.default_timeout_sec)

        suite_id = f"suite-{uuid.uuid4().hex[:8]}"
        start_time = time.perf_counter()
        results: List[TestCaseResult] = []

        # ---------------------------------------------------------
        # TC-01: Health check availability
        # ---------------------------------------------------------
        tc01_start = time.perf_counter()
        try:
            status_code, data, latency = await client.get_health()
            if status_code == 200 and data.get("ok") is True:
                results.append(TestCaseResult(
                    test_id="TC-01-HEALTH",
                    name="Проверка доступности Healthcheck Go API",
                    description="GET /api/health должен возвращать статус 200 и поле ok=true",
                    status=TestStatus.PASSED,
                    duration_ms=round((time.perf_counter() - tc01_start) * 1000, 2),
                    http_status=status_code,
                    details=data
                ))
            else:
                results.append(TestCaseResult(
                    test_id="TC-01-HEALTH",
                    name="Проверка доступности Healthcheck Go API",
                    description="GET /api/health должен возвращать статус 200 и поле ok=true",
                    status=TestStatus.FAILED,
                    duration_ms=round((time.perf_counter() - tc01_start) * 1000, 2),
                    http_status=status_code,
                    error_message=f"Неожиданный ответ: {data}",
                    details=data
                ))
        except Exception as e:
            results.append(TestCaseResult(
                test_id="TC-01-HEALTH",
                name="Проверка доступности Healthcheck Go API",
                description="GET /api/health должен возвращать статус 200 и поле ok=true",
                status=TestStatus.ERROR,
                duration_ms=round((time.perf_counter() - tc01_start) * 1000, 2),
                error_message=f"Сетевая ошибка: {str(e)}"
            ))

        # ---------------------------------------------------------
        # TC-02: Security verification (Missing Token)
        # ---------------------------------------------------------
        tc02_start = time.perf_counter()
        try:
            status_code, data, latency = await client.get_channels_unauthorized()
            if status_code == 401:
                results.append(TestCaseResult(
                    test_id="TC-02-SECURITY",
                    name="Защита API от неавторизованного доступа",
                    description="GET /api/integrations/channels без Bearer токена должен отклоняться 401",
                    status=TestStatus.PASSED,
                    duration_ms=round((time.perf_counter() - tc02_start) * 1000, 2),
                    http_status=status_code,
                    details=data
                ))
            else:
                results.append(TestCaseResult(
                    test_id="TC-02-SECURITY",
                    name="Защита API от неавторизованного доступа",
                    description="GET /api/integrations/channels без Bearer токена должен отклоняться 401",
                    status=TestStatus.FAILED,
                    duration_ms=round((time.perf_counter() - tc02_start) * 1000, 2),
                    http_status=status_code,
                    error_message=f"Ожидался 401, получен {status_code}",
                    details=data
                ))
        except Exception as e:
            results.append(TestCaseResult(
                test_id="TC-02-SECURITY",
                name="Защита API от неавторизованного доступа",
                description="GET /api/integrations/channels без Bearer токена должен отклоняться 401",
                status=TestStatus.ERROR,
                duration_ms=round((time.perf_counter() - tc02_start) * 1000, 2),
                error_message=f"Ошибка соединения: {str(e)}"
            ))

        # ---------------------------------------------------------
        # TC-03: Channel Discovery
        # ---------------------------------------------------------
        tc03_start = time.perf_counter()
        resolved_channel_id = request.channel_id
        if not token:
            results.append(TestCaseResult(
                test_id="TC-03-DISCOVERY",
                name="Обнаружение учетных каналов интеграции",
                description="GET /api/integrations/channels с Bearer токеном возвращает список каналов",
                status=TestStatus.SKIPPED,
                duration_ms=0,
                error_message="Пропущен: не передан DOCSPACE_TEST_INTEGRATION_TOKEN"
            ))
        else:
            try:
                status_code, data, latency = await client.get_channels(token)
                if status_code == 200 and isinstance(data, list) and len(data) > 0:
                    if not resolved_channel_id:
                        resolved_channel_id = data[0].get("id")
                    results.append(TestCaseResult(
                        test_id="TC-03-DISCOVERY",
                        name="Обнаружение учетных каналов интеграции",
                        description="GET /api/integrations/channels возвращает доступные каналы",
                        status=TestStatus.PASSED,
                        duration_ms=round((time.perf_counter() - tc03_start) * 1000, 2),
                        http_status=status_code,
                        details={"channels_count": len(data), "selected_channel_id": resolved_channel_id}
                    ))
                else:
                    results.append(TestCaseResult(
                        test_id="TC-03-DISCOVERY",
                        name="Обнаружение учетных каналов интеграции",
                        description="GET /api/integrations/channels возвращает доступные каналы",
                        status=TestStatus.FAILED,
                        duration_ms=round((time.perf_counter() - tc03_start) * 1000, 2),
                        http_status=status_code,
                        error_message=f"Не удалось получить каналы: статус {status_code}, тело {data}",
                        details={"response": data}
                    ))
            except Exception as e:
                results.append(TestCaseResult(
                    test_id="TC-03-DISCOVERY",
                    name="Обнаружение учетных каналов интеграции",
                    description="GET /api/integrations/channels возвращает доступные каналы",
                    status=TestStatus.ERROR,
                    duration_ms=round((time.perf_counter() - tc03_start) * 1000, 2),
                    error_message=f"Ошибка запроса: {str(e)}"
                ))

        # ---------------------------------------------------------
        # TC-04: 1C Report Ingestion
        # ---------------------------------------------------------
        tc04_start = time.perf_counter()
        report_external_id = f"ozon-qa-{uuid.uuid4().hex[:6]}"
        mock_report = MockAccountingReportGenerator.generate(MockReportConfig(
            report_type=ReportType.PLAN_FACT,
            row_count=5,
            period="2026-Q3"
        ))
        
        if not token or not resolved_channel_id:
            results.append(TestCaseResult(
                test_id="TC-04-INGESTION",
                name="Загрузка отчета 1С через бинарный POST стрим",
                description="POST /api/integrations/documents загружает файл отчета в очередь согласования",
                status=TestStatus.SKIPPED,
                duration_ms=0,
                error_message="Пропущен: отсутствует токен или валидный channel_id"
            ))
        else:
            try:
                status_code, data, latency = await client.upload_document(
                    token=token,
                    channel_id=resolved_channel_id,
                    title="Тестовый план-факт отчет 1С Ozon",
                    filename=mock_report.filename,
                    external_id=report_external_id,
                    file_bytes=mock_report.csv_content.encode("utf-8"),
                    review=True,
                    description="Автоматизированный прогон учетной системы Ozon QA"
                )
                if status_code in (201, 200) and data.get("externalId") == report_external_id:
                    results.append(TestCaseResult(
                        test_id="TC-04-INGESTION",
                        name="Загрузка отчета 1С через бинарный POST стрим",
                        description="POST /api/integrations/documents загружает файл отчета в очередь согласования",
                        status=TestStatus.PASSED,
                        duration_ms=round((time.perf_counter() - tc04_start) * 1000, 2),
                        http_status=status_code,
                        details={"document_id": data.get("id"), "status": data.get("status"), "external_id": report_external_id}
                    ))
                else:
                    results.append(TestCaseResult(
                        test_id="TC-04-INGESTION",
                        name="Загрузка отчета 1С через бинарный POST стрим",
                        description="POST /api/integrations/documents загружает файл отчета в очередь согласования",
                        status=TestStatus.FAILED,
                        duration_ms=round((time.perf_counter() - tc04_start) * 1000, 2),
                        http_status=status_code,
                        error_message=f"Ошибка загрузки: статус {status_code}, тело {data}",
                        details={"response": data}
                    ))
            except Exception as e:
                results.append(TestCaseResult(
                    test_id="TC-04-INGESTION",
                    name="Загрузка отчета 1С через бинарный POST стрим",
                    description="POST /api/integrations/documents загружает файл отчета в очередь согласования",
                    status=TestStatus.ERROR,
                    duration_ms=round((time.perf_counter() - tc04_start) * 1000, 2),
                    error_message=f"Ошибка выполнения: {str(e)}"
                ))

        # ---------------------------------------------------------
        # TC-05: Idempotency by external_id
        # ---------------------------------------------------------
        tc05_start = time.perf_counter()
        if not token or not resolved_channel_id:
            results.append(TestCaseResult(
                test_id="TC-05-IDEMPOTENCY",
                name="Проверка идемпотентности по externalId",
                description="Повторная отправка того же externalId возвращает 200 OK без создания дубликата",
                status=TestStatus.SKIPPED,
                duration_ms=0,
                error_message="Пропущен: отсутствует токен или канал"
            ))
        else:
            try:
                # Повторный запрос с тем же external_id
                status_code, data, latency = await client.upload_document(
                    token=token,
                    channel_id=resolved_channel_id,
                    title="Повторный запрос того же отчета",
                    filename=mock_report.filename,
                    external_id=report_external_id,
                    file_bytes=mock_report.csv_content.encode("utf-8"),
                    review=True
                )
                if status_code == 200 and data.get("externalId") == report_external_id:
                    results.append(TestCaseResult(
                        test_id="TC-05-IDEMPOTENCY",
                        name="Проверка идемпотентности по externalId",
                        description="Повторная отправка того же externalId возвращает 200 OK без создания дубликата",
                        status=TestStatus.PASSED,
                        duration_ms=round((time.perf_counter() - tc05_start) * 1000, 2),
                        http_status=status_code,
                        details={"status": "idempotent_duplicate_prevented", "document_id": data.get("id")}
                    ))
                else:
                    results.append(TestCaseResult(
                        test_id="TC-05-IDEMPOTENCY",
                        name="Проверка идемпотентности по externalId",
                        description="Повторная отправка того же externalId возвращает 200 OK без создания дубликата",
                        status=TestStatus.FAILED,
                        duration_ms=round((time.perf_counter() - tc05_start) * 1000, 2),
                        http_status=status_code,
                        error_message=f"Ожидался статус 200 на дубль, получен {status_code}",
                        details={"response": data}
                    ))
            except Exception as e:
                results.append(TestCaseResult(
                    test_id="TC-05-IDEMPOTENCY",
                    name="Проверка идемпотентности по externalId",
                    description="Повторная отправка того же externalId возвращает 200 OK без создания дубликата",
                    status=TestStatus.ERROR,
                    duration_ms=round((time.perf_counter() - tc05_start) * 1000, 2),
                    error_message=f"Ошибка проверки идемпотентности: {str(e)}"
                ))

        # ---------------------------------------------------------
        # TC-06: Report Status Tracking
        # ---------------------------------------------------------
        tc06_start = time.perf_counter()
        if not token:
            results.append(TestCaseResult(
                test_id="TC-06-TRACKING",
                name="Снятие статуса согласования отчета внешней системой",
                description="GET /api/integrations/documents?externalId=... возвращает решение и ревизию",
                status=TestStatus.SKIPPED,
                duration_ms=0,
                error_message="Пропущен: не передан токен"
            ))
        else:
            try:
                status_code, data, latency = await client.get_document_status(token, report_external_id)
                if status_code == 200 and data.get("externalId") == report_external_id:
                    results.append(TestCaseResult(
                        test_id="TC-06-TRACKING",
                        name="Снятие статуса согласования отчета внешней системой",
                        description="GET /api/integrations/documents?externalId=... возвращает решение и ревизию",
                        status=TestStatus.PASSED,
                        duration_ms=round((time.perf_counter() - tc06_start) * 1000, 2),
                        http_status=status_code,
                        details={"status": data.get("status"), "revision": data.get("revision")}
                    ))
                else:
                    results.append(TestCaseResult(
                        test_id="TC-06-TRACKING",
                        name="Снятие статуса согласования отчета внешней системой",
                        description="GET /api/integrations/documents?externalId=... возвращает решение и ревизию",
                        status=TestStatus.FAILED,
                        duration_ms=round((time.perf_counter() - tc06_start) * 1000, 2),
                        http_status=status_code,
                        error_message=f"Не удалось получить карточку документа: {data}",
                        details={"response": data}
                    ))
            except Exception as e:
                results.append(TestCaseResult(
                    test_id="TC-06-TRACKING",
                    name="Снятие статуса согласования отчета внешней системой",
                    description="GET /api/integrations/documents?externalId=... возвращает решение и ревизию",
                    status=TestStatus.ERROR,
                    duration_ms=round((time.perf_counter() - tc06_start) * 1000, 2),
                    error_message=f"Ошибка проверки статуса: {str(e)}"
                ))

        # ---------------------------------------------------------
        # TC-07: Flaky Benchmark & Latency Stability
        # ---------------------------------------------------------
        is_flaky = False
        flaky_summary = None

        if request.include_flaky_benchmark:
            tc07_start = time.perf_counter()
            latencies = []
            failures = 0
            iterations = request.benchmark_iterations

            for it in range(iterations):
                it_start = time.perf_counter()
                try:
                    code, _, lat = await client.get_health()
                    if code == 200:
                        latencies.append(round((time.perf_counter() - it_start) * 1000, 2))
                    else:
                        failures += 1
                except Exception:
                    failures += 1
                await asyncio.sleep(0.05)

            duration_total = round((time.perf_counter() - tc07_start) * 1000, 2)
            if latencies and len(latencies) >= 2:
                avg_lat = statistics.mean(latencies)
                max_lat = max(latencies)
                min_lat = min(latencies)
                jitter = max_lat - min_lat

                # Если разброс более 350мс или были единичные сбои — помечаем как flaky
                if failures > 0 and failures < iterations:
                    is_flaky = True
                    flaky_summary = f"Обнаружено нестабильное поведение (flaky): {failures}/{iterations} сбоев запросов"
                elif jitter > 350.0:
                    is_flaky = True
                    flaky_summary = f"Высокий джиттер задержки: min={min_lat}ms, max={max_lat}ms (разброс {round(jitter, 2)}ms)"

                test_status = TestStatus.FAILED if is_flaky else TestStatus.PASSED
                results.append(TestCaseResult(
                    test_id="TC-07-FLAKY-BENCHMARK",
                    name="Бенчмарк стабильности и детекция flaky-тестов",
                    description=f"Серия из {iterations} запросов для замера джиттера и выявления нестабильных ответов",
                    status=test_status,
                    duration_ms=duration_total,
                    details={
                        "iterations": iterations,
                        "success_iterations": len(latencies),
                        "failed_iterations": failures,
                        "latencies_ms": latencies,
                        "avg_latency_ms": round(avg_lat, 2),
                        "jitter_ms": round(jitter, 2),
                        "flaky_flag": is_flaky
                    }
                ))
            else:
                results.append(TestCaseResult(
                    test_id="TC-07-FLAKY-BENCHMARK",
                    name="Бенчмарк стабильности и детекция flaky-тестов",
                    description="Серия запросов замера стабильности",
                    status=TestStatus.FAILED if failures > 0 else TestStatus.PASSED,
                    duration_ms=duration_total,
                    error_message="Недостаточно данных для расчета джиттера" if failures > 0 else None
                ))

        total_duration = round((time.perf_counter() - start_time) * 1000, 2)
        passed_count = sum(1 for r in results if r.status == TestStatus.PASSED)
        failed_count = sum(1 for r in results if r.status in (TestStatus.FAILED, TestStatus.ERROR))

        execution = TestSuiteExecution(
            suite_id=suite_id,
            timestamp=datetime.now(timezone.utc),
            target_api_url=target_url,
            total_tests=len(results),
            passed_tests=passed_count,
            failed_tests=failed_count,
            total_duration_ms=total_duration,
            is_flaky_detected=is_flaky,
            flaky_summary=flaky_summary,
            results=results
        )

        self.history.insert(0, execution)
        if len(self.history) > settings.history_limit:
            self.history = self.history[:settings.history_limit]

        return execution

    def get_history(self, limit: int = 10) -> List[TestSuiteExecution]:
        return self.history[:limit]

    def get_metrics(self) -> MetricsSummary:
        if not self.history:
            return MetricsSummary(
                total_suites_executed=0,
                passed_suites=0,
                failed_suites=0,
                success_rate_percent=100.0,
                avg_duration_ms=0.0,
                flaky_events_detected=0
            )

        total = len(self.history)
        passed = sum(1 for s in self.history if s.failed_tests == 0)
        failed = total - passed
        durations = [s.total_duration_ms for s in self.history]
        flaky_count = sum(1 for s in self.history if s.is_flaky_detected)

        return MetricsSummary(
            total_suites_executed=total,
            passed_suites=passed,
            failed_suites=failed,
            success_rate_percent=round((passed / total) * 100, 2),
            avg_duration_ms=round(statistics.mean(durations), 2) if durations else 0.0,
            flaky_events_detected=flaky_count
        )

# Глобальный синглтон раннера
qa_runner = AccountingQARunner()
