import pytest
from fastapi.testclient import TestClient

from app.generator.mock_reports import MockAccountingReportGenerator
from app.main import app
from app.models.schemas import MockReportConfig, ReportType, TestSuiteRunRequest
from app.runner.test_runner import qa_runner

client = TestClient(app)


def test_health_endpoint():
    response = client.get("/api/v1/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["service"] == "accounting-qa-runner"
    assert "timestamp" in data


def test_generator_plan_fact():
    config = MockReportConfig(report_type=ReportType.PLAN_FACT, row_count=5)
    report = MockAccountingReportGenerator.generate(config)
    assert report.report_type == ReportType.PLAN_FACT
    assert report.filename.startswith("1c_plan_fact_")
    assert len(report.checksum_sha256) == 64
    assert report.row_count == 5
    lines = [line for line in report.csv_content.strip().split("\n") if line]
    # 1 header + 5 rows
    assert len(lines) == 6
    assert "Отдел;План;Факт;Отклонение;Процент_выполнения" in lines[0]


def test_generator_cash_flow():
    config = MockReportConfig(report_type=ReportType.CASH_FLOW, row_count=6)
    report = MockAccountingReportGenerator.generate(config)
    assert report.report_type == ReportType.CASH_FLOW
    assert report.filename.startswith("1c_dds_cashflow_")
    assert "total_inflow_rub" in report.summary_totals
    assert "total_outflow_rub" in report.summary_totals


def test_generator_payroll():
    config = MockReportConfig(report_type=ReportType.PAYROLL, row_count=4)
    report = MockAccountingReportGenerator.generate(config)
    assert report.report_type == ReportType.PAYROLL
    assert report.filename.startswith("1c_payroll_sheet_")
    assert report.summary_totals["headcount"] == 4
    assert report.summary_totals["total_net_rub"] > 0


def test_generator_anomaly_injection():
    config = MockReportConfig(report_type=ReportType.PLAN_FACT, row_count=3, inject_anomaly=True)
    report = MockAccountingReportGenerator.generate(config)
    assert report.summary_totals["has_critical_overrun"] is True


def test_generator_sample_csv_endpoint():
    response = client.get("/api/v1/generator/sample-csv?report_type=plan_fact&rows=3")
    assert response.status_code == 200
    assert "text/csv" in response.headers["content-type"]
    assert "attachment" in response.headers["content-disposition"]
    content = response.text
    assert "Отдел;План;Факт" in content


def test_qa_metrics_initial():
    response = client.get("/api/v1/qa/metrics")
    assert response.status_code == 200
    data = response.json()
    assert "total_suites_executed" in data
    assert "success_rate_percent" in data


def test_qa_runner_handles_unreachable_target():
    # Тест устойчивости раннера при недоступном сетевом адресе
    import asyncio
    request = TestSuiteRunRequest(
        target_api_url="http://127.0.0.1:59999",
        include_flaky_benchmark=False
    )
    result = asyncio.run(qa_runner.execute_suite(request))
    assert result.total_tests >= 2
    assert result.failed_tests >= 1
    assert result.suite_id.startswith("suite-")
