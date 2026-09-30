"""
Загрузчик нормативно-справочной информации:
tax_regimes, tax_regional_rates, checklists, checklist_items, signboard_regulations.
"""
import json
from etl.config import REFERENCE_DIR
from etl.logger import log_success
from etl.validator import validate_tax_regimes

def load_reference_data(conn) -> list:
    stats = []

    # 1. tax_regimes
    with open(REFERENCE_DIR / "tax_regimes.json", "r", encoding="utf-8-sig") as f:
        raw_regimes = json.load(f)
    valid_regimes, regime_errs = validate_tax_regimes(raw_regimes)

    inserted_regimes = 0
    with conn.cursor() as cur:
        for r in valid_regimes:
            cur.execute("""
                INSERT INTO tax_regimes (
                    code, name, short_name, description, allowed_legal_forms,
                    max_annual_revenue, max_employees, base_tax_rate_percent,
                    has_regional_discounts, fixed_insurance_required,
                    fixed_insurance_amount_annual, reporting_frequency,
                    kkt_required_by_default, legal_source_id, effective_year
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (code) DO UPDATE SET
                    name = EXCLUDED.name,
                    short_name = EXCLUDED.short_name,
                    description = EXCLUDED.description,
                    allowed_legal_forms = EXCLUDED.allowed_legal_forms,
                    max_annual_revenue = EXCLUDED.max_annual_revenue,
                    max_employees = EXCLUDED.max_employees,
                    base_tax_rate_percent = EXCLUDED.base_tax_rate_percent,
                    fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;
            """, (
                r['code'], r['name'], r['short_name'], r.get('description'),
                r['allowed_legal_forms'], r['max_annual_revenue'], r['max_employees'],
                r['base_tax_rate_percent'], r.get('has_regional_discounts', False),
                r.get('fixed_insurance_required', False), r.get('fixed_insurance_amount_annual', 0.0),
                r.get('reporting_frequency', 'annual'), r.get('kkt_required_by_default', True),
                r.get('legal_source_id'), r.get('effective_year', 2026)
            ))
            inserted_regimes += 1
    conn.commit()

    stats.append({
        'entity': 'tax_regimes',
        'read_count': len(raw_regimes),
        'valid_count': len(valid_regimes),
        'inserted_count': inserted_regimes,
        'updated_count': inserted_regimes,
        'error_count': len(regime_errs),
        'status': 'OK' if not regime_errs else 'WARN'
    })
    log_success(f"tax_regimes: загружено {inserted_regimes} режимов")

    # 2. tax_regional_rates
    with open(REFERENCE_DIR / "tax_regional_rates.json", "r", encoding="utf-8-sig") as f:
        raw_rates = json.load(f)

    inserted_rates = 0
    with conn.cursor() as cur:
        for rr in raw_rates:
            cur.execute("""
                INSERT INTO tax_regional_rates (
                    id, tax_regime_code, region_code, rate_percent,
                    conditions_description, legal_act_title, effective_from, effective_to
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    rate_percent = EXCLUDED.rate_percent,
                    conditions_description = EXCLUDED.conditions_description,
                    legal_act_title = EXCLUDED.legal_act_title;
            """, (
                rr['id'], rr['tax_regime_code'], rr['region_code'], rr['rate_percent'],
                rr.get('conditions_description'), rr.get('legal_act_title'),
                rr['effective_from'], rr.get('effective_to')
            ))
            inserted_rates += 1
        cur.execute("SELECT setval('tax_regional_rates_id_seq', COALESCE((SELECT MAX(id) FROM tax_regional_rates), 1));")
    conn.commit()

    stats.append({
        'entity': 'tax_regional_rates',
        'read_count': len(raw_rates),
        'valid_count': len(raw_rates),
        'inserted_count': inserted_rates,
        'updated_count': inserted_rates,
        'error_count': 0,
        'status': 'OK'
    })
    log_success(f"tax_regional_rates: загружено {inserted_rates} региональных ставок")

    # 3. signboard_regulations
    with open(REFERENCE_DIR / "signboard_regulations.json", "r", encoding="utf-8-sig") as f:
        raw_signs = json.load(f)

    inserted_signs = 0
    with conn.cursor() as cur:
        for sb in raw_signs:
            cur.execute("""
                INSERT INTO signboard_regulations (
                    id, region_code, city, authority_name, legal_act_title,
                    max_sign_height_meters, max_sign_width_percent_of_facade,
                    requires_special_permit, required_documents, processing_time_days,
                    official_portal_url, is_mocked_data
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    authority_name = EXCLUDED.authority_name,
                    max_sign_height_meters = EXCLUDED.max_sign_height_meters,
                    max_sign_width_percent_of_facade = EXCLUDED.max_sign_width_percent_of_facade;
            """, (
                sb['id'], sb['region_code'], sb['city'], sb['authority_name'],
                sb['legal_act_title'], sb['max_sign_height_meters'],
                sb['max_sign_width_percent_of_facade'], sb['requires_special_permit'],
                sb.get('required_documents', []), sb['processing_time_days'],
                sb.get('official_portal_url'), sb.get('is_mocked_data', True)
            ))
            inserted_signs += 1
        cur.execute("SELECT setval('signboard_regulations_id_seq', COALESCE((SELECT MAX(id) FROM signboard_regulations), 1));")
    conn.commit()

    stats.append({
        'entity': 'signboard_regulations',
        'read_count': len(raw_signs),
        'valid_count': len(raw_signs),
        'inserted_count': inserted_signs,
        'updated_count': inserted_signs,
        'error_count': 0,
        'status': 'OK'
    })
    log_success(f"signboard_regulations: загружено {inserted_signs} регламентов вывесок")

    # 4. checklists & checklist_items
    with open(REFERENCE_DIR / "checklists.json", "r", encoding="utf-8-sig") as f:
        raw_chks = json.load(f)
    with open(REFERENCE_DIR / "checklist_items.json", "r", encoding="utf-8-sig") as f:
        raw_items = json.load(f)

    inserted_chks = 0
    with conn.cursor() as cur:
        for chk in raw_chks:
            cur.execute("""
                INSERT INTO checklists (id, code, title, description, target_legal_form, target_okved_prefix)
                VALUES (%s, %s, %s, %s, %s, %s)
                ON CONFLICT (code) DO UPDATE SET
                    title = EXCLUDED.title,
                    description = EXCLUDED.description;
            """, (
                chk['id'], chk['code'], chk['title'], chk.get('description'),
                chk.get('target_legal_form', 'all'), chk.get('target_okved_prefix')
            ))
            inserted_chks += 1
        cur.execute("SELECT setval('checklists_id_seq', COALESCE((SELECT MAX(id) FROM checklists), 1));")
    conn.commit()

    inserted_items = 0
    with conn.cursor() as cur:
        for ci in raw_items:
            cur.execute("""
                INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    title = EXCLUDED.title,
                    description = EXCLUDED.description,
                    sort_order = EXCLUDED.sort_order;
            """, (
                ci['id'], ci['checklist_id'], ci.get('requirement_id'),
                ci['title'], ci.get('description'), ci.get('sort_order', 0),
                ci.get('is_mandatory', True)
            ))
            inserted_items += 1
        cur.execute("SELECT setval('checklist_items_id_seq', COALESCE((SELECT MAX(id) FROM checklist_items), 1));")
    conn.commit()

    stats.append({
        'entity': 'checklists',
        'read_count': len(raw_chks),
        'valid_count': len(raw_chks),
        'inserted_count': inserted_chks,
        'updated_count': inserted_chks,
        'error_count': 0,
        'status': 'OK'
    })
    stats.append({
        'entity': 'checklist_items',
        'read_count': len(raw_items),
        'valid_count': len(raw_items),
        'inserted_count': inserted_items,
        'updated_count': inserted_items,
        'error_count': 0,
        'status': 'OK'
    })
    log_success(f"checklists: загружено {inserted_chks} шаблонов ({inserted_items} пунктов задач)")

    return stats

