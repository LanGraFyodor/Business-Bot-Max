"""
Загрузчик нормативно-правовых источников и бизнес-требований.
Гарантирует идемпотентность через ON CONFLICT DO UPDATE.
"""
import json
from psycopg2.extras import Json
from etl.config import REFERENCE_DIR
from etl.logger import log_info, log_success
from etl.validator import validate_legal_sources, validate_business_requirements

def load_legal_data(conn) -> tuple:
    """
    Загружает legal_sources и business_requirements в PostgreSQL.
    Возвращает статистику (read, valid, inserted, errors).
    """
    stats = []

    # 1. legal_sources
    sources_path = REFERENCE_DIR / "legal_sources.json"
    with open(sources_path, "r", encoding="utf-8") as f:
        raw_sources = json.load(f)

    valid_sources, source_errors = validate_legal_sources(raw_sources)
    known_source_ids = {s['id'] for s in valid_sources}

    inserted_sources = 0
    with conn.cursor() as cur:
        for s in valid_sources:
            cur.execute("""
                INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    name = EXCLUDED.name,
                    official_url = EXCLUDED.official_url,
                    publisher = EXCLUDED.publisher,
                    source_type = EXCLUDED.source_type,
                    last_verified_at = EXCLUDED.last_verified_at,
                    status = EXCLUDED.status;
            """, (
                s['id'], s['name'], s['official_url'], s['publisher'],
                s['source_type'], s['last_verified_at'], s.get('status', 'active')
            ))
            inserted_sources += 1
    conn.commit()

    stats.append({
        'entity': 'legal_sources',
        'read_count': len(raw_sources),
        'valid_count': len(valid_sources),
        'inserted_count': inserted_sources,
        'updated_count': inserted_sources,
        'error_count': len(source_errors),
        'status': 'OK' if not source_errors else 'WARN'
    })
    log_success(f"legal_sources: обработано {inserted_sources} записей (ошибок: {len(source_errors)})")

    # 2. business_requirements
    reqs_path = REFERENCE_DIR / "business_requirements.json"
    with open(reqs_path, "r", encoding="utf-8") as f:
        raw_reqs = json.load(f)

    valid_reqs, req_errors = validate_business_requirements(raw_reqs, known_source_ids)

    inserted_reqs = 0
    with conn.cursor() as cur:
        for r in valid_reqs:
            cur.execute("""
                INSERT INTO business_requirements (
                    id, source_id, category, title, summary, legal_citation,
                    applies_to_legal_forms, applies_to_okved_prefixes, region_code,
                    is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    source_id = EXCLUDED.source_id,
                    category = EXCLUDED.category,
                    title = EXCLUDED.title,
                    summary = EXCLUDED.summary,
                    legal_citation = EXCLUDED.legal_citation,
                    applies_to_legal_forms = EXCLUDED.applies_to_legal_forms,
                    applies_to_okved_prefixes = EXCLUDED.applies_to_okved_prefixes,
                    region_code = EXCLUDED.region_code,
                    is_mandatory = EXCLUDED.is_mandatory,
                    penalty_risk_description = EXCLUDED.penalty_risk_description,
                    action_plan = EXCLUDED.action_plan,
                    is_simulated_in_mvp = EXCLUDED.is_simulated_in_mvp;
            """, (
                r['id'], r['source_id'], r['category'], r['title'], r['summary'],
                r['legal_citation'], r.get('applies_to_legal_forms', []),
                r.get('applies_to_okved_prefixes'), r.get('region_code'),
                r.get('is_mandatory', True), r.get('penalty_risk_description'),
                Json(r.get('action_plan', [])), r.get('is_simulated_in_mvp', False)
            ))
            inserted_reqs += 1

        # Обновляем автоинкремент ID
        cur.execute("SELECT setval('business_requirements_id_seq', COALESCE((SELECT MAX(id) FROM business_requirements), 1));")
    conn.commit()

    stats.append({
        'entity': 'business_requirements',
        'read_count': len(raw_reqs),
        'valid_count': len(valid_reqs),
        'inserted_count': inserted_reqs,
        'updated_count': inserted_reqs,
        'error_count': len(req_errors),
        'status': 'OK' if not req_errors else 'WARN'
    })
    log_success(f"business_requirements: обработано {inserted_reqs} требований (ошибок: {len(req_errors)})")

    return stats
