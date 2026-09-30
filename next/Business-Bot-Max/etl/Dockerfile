# ==============================================================================
# Dockerfile для сервиса ETL-пайплайна
# Легковесный образ (< 100MB), сборка за 20 секунд
# ==============================================================================
FROM python:3.11-slim

# Отключение буферизации вывода для мгновенного отображения логов
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONPATH=/app

WORKDIR /app

# Установка зависимостей
COPY etl/requirements.txt /app/etl/requirements.txt
RUN pip install --no-cache-dir -r /app/etl/requirements.txt

# Копирование исходного кода и данных
COPY . /app

# Запуск единой команды ETL-пайплайна
CMD ["python", "-m", "etl.run_pipeline"]
