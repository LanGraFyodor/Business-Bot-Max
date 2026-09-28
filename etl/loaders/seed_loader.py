"""
Загрузчик эталонных синтетических данных (Seed Data Loader).
Загружает пользователей MAX, карточки бизнеса, витрины, категории, товары и заказы.
"""
import json
from etl.config import SEEDS_DIR
from etl.logger import log_success
from etl.validator import validate_business_profiles, validate_products

def load_seed_demo_data(conn) -> list:
    stats = []
    seeds_file = SEEDS_DIR / "test_dataset.json"
    if not seeds_file.exists():
        seeds_file = SEEDS_DIR / "seed_demo_data.json"
    if not seeds_file.exists():
        return stats

    with open(seeds_file, "r", encoding="utf-8") as f:
        demo = json.load(f)

    # 1. users
    raw_users = demo.get("users", [])
    inserted_users = 0
    with conn.cursor() as cur:
        for u in raw_users:
            cur.execute("""
                INSERT INTO users (id, max_user_id, username, first_name, last_name)
                VALUES (%s, %s, %s, %s, %s)
                ON CONFLICT (max_user_id) DO UPDATE SET
                    username = EXCLUDED.username,
                    first_name = EXCLUDED.first_name,
                    last_name = EXCLUDED.last_name;
            """, (u['id'], u['max_user_id'], u.get('username'), u.get('first_name'), u.get('last_name')))
            inserted_users += 1
    conn.commit()
    stats.append({
        'entity': 'users (demo)',
        'read_count': len(raw_users),
        'valid_count': len(raw_users),
        'inserted_count': inserted_users,
        'updated_count': inserted_users,
        'error_count': 0,
        'status': 'OK'
    })

    # 2. business_profiles
    raw_profiles = demo.get("business_profiles", [])
    valid_profiles, prof_errs = validate_business_profiles(raw_profiles)
    inserted_profs = 0
    with conn.cursor() as cur:
        for p in valid_profiles:
            cur.execute("""
                INSERT INTO business_profiles (
                    id, user_id, legal_form, business_name, inn, okved_code,
                    region_code, city, tax_regime_code, has_employees,
                    employee_count, monthly_revenue_est, monthly_expenses_est, status
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    business_name = EXCLUDED.business_name,
                    monthly_revenue_est = EXCLUDED.monthly_revenue_est,
                    monthly_expenses_est = EXCLUDED.monthly_expenses_est,
                    tax_regime_code = EXCLUDED.tax_regime_code;
            """, (
                p['id'], p['user_id'], p['legal_form'], p['business_name'],
                p.get('inn'), p['okved_code'], p['region_code'], p['city'],
                p['tax_regime_code'], p.get('has_employees', False),
                p.get('employee_count', 0), p.get('monthly_revenue_est', 0.0),
                p.get('monthly_expenses_est', 0.0), p.get('status', 'active')
            ))
            inserted_profs += 1
    conn.commit()
    stats.append({
        'entity': 'business_profiles (demo)',
        'read_count': len(raw_profiles),
        'valid_count': len(valid_profiles),
        'inserted_count': inserted_profs,
        'updated_count': inserted_profs,
        'error_count': len(prof_errs),
        'status': 'OK' if not prof_errs else 'WARN'
    })

    # 3. stores
    raw_stores = demo.get("stores", [])
    inserted_stores = 0
    with conn.cursor() as cur:
        for s in raw_stores:
            cur.execute("""
                INSERT INTO stores (
                    id, business_profile_id, slug, title, description,
                    contact_phone, pickup_address, delivery_enabled, delivery_terms,
                    delivery_min_amount, delivery_fee, is_published
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (slug) DO UPDATE SET
                    title = EXCLUDED.title,
                    description = EXCLUDED.description,
                    delivery_enabled = EXCLUDED.delivery_enabled,
                    is_published = EXCLUDED.is_published;
            """, (
                s['id'], s['business_profile_id'], s['slug'], s['title'],
                s.get('description'), s.get('contact_phone'), s.get('pickup_address'),
                s.get('delivery_enabled', False), s.get('delivery_terms'),
                s.get('delivery_min_amount', 0.0), s.get('delivery_fee', 0.0),
                s.get('is_published', False)
            ))
            inserted_stores += 1
    conn.commit()
    stats.append({
        'entity': 'stores (demo)',
        'read_count': len(raw_stores),
        'valid_count': len(raw_stores),
        'inserted_count': inserted_stores,
        'updated_count': inserted_stores,
        'error_count': 0,
        'status': 'OK'
    })

    # 4. product_categories
    raw_cats = demo.get("product_categories", [])
    inserted_cats = 0
    with conn.cursor() as cur:
        for c in raw_cats:
            cur.execute("""
                INSERT INTO product_categories (id, store_id, name, sort_order)
                VALUES (%s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, sort_order = EXCLUDED.sort_order;
            """, (c['id'], c['store_id'], c['name'], c.get('sort_order', 0)))
            inserted_cats += 1
        cur.execute("SELECT setval('product_categories_id_seq', COALESCE((SELECT MAX(id) FROM product_categories), 1));")
    conn.commit()
    stats.append({
        'entity': 'product_categories (demo)',
        'read_count': len(raw_cats),
        'valid_count': len(raw_cats),
        'inserted_count': inserted_cats,
        'updated_count': inserted_cats,
        'error_count': 0,
        'status': 'OK'
    })

    # 5. products
    raw_prods = demo.get("products", [])
    valid_prods, prod_errs = validate_products(raw_prods)
    inserted_prods = 0
    with conn.cursor() as cur:
        for pr in valid_prods:
            cur.execute("""
                INSERT INTO products (
                    id, store_id, category_id, type, name, description,
                    price, currency, unit, is_in_stock, stock_quantity,
                    requires_marking, status
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    name = EXCLUDED.name,
                    price = EXCLUDED.price,
                    description = EXCLUDED.description,
                    is_in_stock = EXCLUDED.is_in_stock,
                    requires_marking = EXCLUDED.requires_marking;
            """, (
                pr['id'], pr['store_id'], pr.get('category_id'), pr.get('type', 'good'),
                pr['name'], pr.get('description'), pr['price'], pr.get('currency', 'RUB'),
                pr.get('unit', 'шт'), pr.get('is_in_stock', True),
                pr.get('stock_quantity', 0), pr.get('requires_marking', False),
                pr.get('status', 'active')
            ))
            inserted_prods += 1
    conn.commit()
    stats.append({
        'entity': 'products (demo)',
        'read_count': len(raw_prods),
        'valid_count': len(valid_prods),
        'inserted_count': inserted_prods,
        'updated_count': inserted_prods,
        'error_count': len(prod_errs),
        'status': 'OK' if not prod_errs else 'WARN'
    })

    # 6. orders & order_items
    raw_orders = demo.get("orders", [])
    raw_items = demo.get("order_items", [])
    inserted_orders = 0
    with conn.cursor() as cur:
        for ord_item in raw_orders:
            cur.execute("""
                INSERT INTO orders (
                    id, store_id, order_number, customer_max_user_id,
                    customer_name, customer_phone, delivery_type, delivery_address,
                    total_amount, currency, status, customer_comment
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    status = EXCLUDED.status,
                    total_amount = EXCLUDED.total_amount;
            """, (
                ord_item['id'], ord_item['store_id'], ord_item['order_number'],
                ord_item['customer_max_user_id'], ord_item['customer_name'],
                ord_item['customer_phone'], ord_item['delivery_type'],
                ord_item.get('delivery_address'), ord_item['total_amount'],
                ord_item.get('currency', 'RUB'), ord_item.get('status', 'new'),
                ord_item.get('customer_comment')
            ))
            inserted_orders += 1

        for oi in raw_items:
            cur.execute("""
                INSERT INTO order_items (id, order_id, product_id, product_name_snapshot, unit_price_snapshot, quantity, total_price)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET total_price = EXCLUDED.total_price;
            """, (
                oi['id'], oi['order_id'], oi.get('product_id'),
                oi['product_name_snapshot'], oi['unit_price_snapshot'],
                oi['quantity'], oi['total_price']
            ))
        cur.execute("SELECT setval('order_items_id_seq', COALESCE((SELECT MAX(id) FROM order_items), 1));")
    conn.commit()

    stats.append({
        'entity': 'orders & items (demo)',
        'read_count': len(raw_orders) + len(raw_items),
        'valid_count': len(raw_orders) + len(raw_items),
        'inserted_count': inserted_orders + len(raw_items),
        'updated_count': inserted_orders + len(raw_items),
        'error_count': 0,
        'status': 'OK'
    })

    # 7. tax_calculations
    raw_calcs = demo.get("tax_calculations", [])
    inserted_calcs = 0
    with conn.cursor() as cur:
        for tc in raw_calcs:
            cur.execute("""
                INSERT INTO tax_calculations (
                    id, business_profile_id, input_parameters, results_comparison,
                    recommended_regime_code, estimated_annual_tax, estimated_annual_insurance, assumptions
                )
                VALUES (%s, %s, %s::jsonb, %s::jsonb, %s, %s, %s, %s)
                ON CONFLICT (id) DO UPDATE SET
                    estimated_annual_tax = EXCLUDED.estimated_annual_tax,
                    estimated_annual_insurance = EXCLUDED.estimated_annual_insurance;
            """, (
                tc['id'], tc['business_profile_id'], json.dumps(tc['input_parameters'], ensure_ascii=False),
                json.dumps(tc['results_comparison'], ensure_ascii=False), tc['recommended_regime_code'],
                tc['estimated_annual_tax'], tc['estimated_annual_insurance'], tc.get('assumptions', [])
            ))
            inserted_calcs += 1
    conn.commit()
    if raw_calcs:
        stats.append({
            'entity': 'tax_calculations',
            'read_count': len(raw_calcs),
            'valid_count': len(raw_calcs),
            'inserted_count': inserted_calcs,
            'updated_count': inserted_calcs,
            'error_count': 0,
            'status': 'OK'
        })

    # 8. financial_health_assessments
    raw_health = demo.get("financial_health_assessments", [])
    inserted_health = 0
    with conn.cursor() as cur:
        for fh in raw_health:
            cur.execute("""
                INSERT INTO financial_health_assessments (
                    id, business_profile_id, tax_calculation_id, monthly_revenue,
                    monthly_fixed_costs, monthly_variable_costs, net_monthly_profit,
                    margin_percent, break_even_point, safety_margin_percent,
                    cash_gap_risk_level, verdict, recommendations
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s::jsonb)
                ON CONFLICT (id) DO UPDATE SET
                    net_monthly_profit = EXCLUDED.net_monthly_profit,
                    margin_percent = EXCLUDED.margin_percent,
                    verdict = EXCLUDED.verdict;
            """, (
                fh['id'], fh['business_profile_id'], fh.get('tax_calculation_id'),
                fh['monthly_revenue'], fh['monthly_fixed_costs'], fh['monthly_variable_costs'],
                fh['net_monthly_profit'], fh['margin_percent'], fh['break_even_point'],
                fh['safety_margin_percent'], fh['cash_gap_risk_level'], fh['verdict'],
                json.dumps(fh.get('recommendations', []), ensure_ascii=False)
            ))
            inserted_health += 1
    conn.commit()
    if raw_health:
        stats.append({
            'entity': 'financial_health',
            'read_count': len(raw_health),
            'valid_count': len(raw_health),
            'inserted_count': inserted_health,
            'updated_count': inserted_health,
            'error_count': 0,
            'status': 'OK'
        })

    # 9. user_checklist_progress
    raw_chk = demo.get("user_checklist_progress", [])
    inserted_chk = 0
    with conn.cursor() as cur:
        for cp in raw_chk:
            cur.execute("""
                INSERT INTO user_checklist_progress (id, business_profile_id, checklist_item_id, is_completed)
                VALUES (%s, %s, %s, %s)
                ON CONFLICT (business_profile_id, checklist_item_id) DO UPDATE SET
                    is_completed = EXCLUDED.is_completed;
            """, (cp['id'], cp['business_profile_id'], cp['checklist_item_id'], cp.get('is_completed', False)))
            inserted_chk += 1
        cur.execute("SELECT setval('user_checklist_progress_id_seq', COALESCE((SELECT MAX(id) FROM user_checklist_progress), 1));")
    conn.commit()
    if raw_chk:
        stats.append({
            'entity': 'checklist_progress',
            'read_count': len(raw_chk),
            'valid_count': len(raw_chk),
            'inserted_count': inserted_chk,
            'updated_count': inserted_chk,
            'error_count': 0,
            'status': 'OK'
        })

    log_success(f"seed_loader: успешно загружен тестовый датасет (профили, витрины, расчеты, оценки, чек-листы)")

    return stats

