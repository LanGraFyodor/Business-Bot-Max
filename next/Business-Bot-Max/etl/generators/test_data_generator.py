"""
Генератор воспроизводимых эталонных и граничных тестовых данных (Synthetic Test Data Generator).
Обеспечивает 100% воспроизводимость сценариев для жюри хакатона при фиксированном seed.
"""
import os
import sys
import json
import uuid
import random
import argparse
from pathlib import Path
from decimal import Decimal

# Базовые пути
GENERATOR_DIR = Path(__file__).resolve().parent
ETL_DIR = GENERATOR_DIR.parent
PROJECT_ROOT = ETL_DIR.parent
DATA_DIR = PROJECT_ROOT / "data"
SEEDS_DIR = DATA_DIR / "seeds"
MIGRATIONS_DIR = DATA_DIR / "migrations"

SEEDS_DIR.mkdir(parents=True, exist_ok=True)
MIGRATIONS_DIR.mkdir(parents=True, exist_ok=True)

# Имена и фамилии для реалистичных данных
FIRST_NAMES_MALE = ["Михаил", "Артур", "Дмитрий", "Александр", "Ильдар", "Роман", "Денис", "Тимур"]
LAST_NAMES_MALE = ["Северов", "Каримов", "Кузнецов", "Смирнов", "Хасанов", "Морозов", "Васильев", "Галиев"]
FIRST_NAMES_FEMALE = ["Анна", "Елена", "Дарья", "Алина", "Мария", "Лейла", "Виктория", "Полина"]
LAST_NAMES_FEMALE = ["Сладкова", "Орлова", "Соколова", "Зарипова", "Иванова", "Белова", "Волкова", "Мусина"]

CITIES_BY_REGION = {
    "77": ("Москва", "Центральный"),
    "16": ("Казань", "Приволжский"),
    "78": ("Санкт-Петербург", "Северо-Западный"),
    "66": ("Екатеринбург", "Уральский"),
    "23": ("Краснодар", "Южный"),
    "54": ("Новосибирск", "Сибирский")
}


def generate_deterministic_uuid(namespace: str, identifier: str) -> str:
    """Генерирует детерминированный UUID на основе строки для идемпотентности."""
    return str(uuid.uuid5(uuid.NAMESPACE_DNS, f"{namespace}:{identifier}"))


def build_scenario_dataset(seed: int = 42) -> dict:
    """Формирует полный набор эталонных профилей, витрин, товаров, расчетов и оценок устойчивости."""
    random.seed(seed)

    scenarios = [
        # 1. Самозанятый кондитер
        {
            "key": "bento_cake",
            "first_name": "Анна",
            "last_name": "Сладкова",
            "gender": "f",
            "legal_form": "self_employed",
            "business_name": "Кондитерская «Сладкий Лис»",
            "okved_code": "10.71",
            "region_code": "77",
            "tax_regime_code": "npd",
            "has_employees": False,
            "employee_count": 0,
            "monthly_revenue_est": 95000.0,
            "monthly_expenses_est": 35000.0,
            "store_slug": "sweet_fox_bakery",
            "store_title": "Кондитерская «Сладкий Лис»",
            "store_desc": "Авторские бенто-торты и десерты ручной работы. Только натуральные ягоды и крем-чиз.",
            "categories": ["Бенто-торты", "Капкейки"],
            "products": [
                {"name": "Бенто-торт «Ягодный пломбир»", "price": 1400.0, "unit": "шт", "marking": False, "stock": 10},
                {"name": "Бенто-торт «Шоколадный трюфель»", "price": 1500.0, "unit": "шт", "marking": False, "stock": 8},
                {"name": "Набор капкейков (6 шт)", "price": 1200.0, "unit": "набор", "marking": False, "stock": 15}
            ]
        },
        # 2. Самозанятый IT-разработчик
        {
            "key": "freelance_it",
            "first_name": "Тимур",
            "last_name": "Хасанов",
            "gender": "m",
            "legal_form": "self_employed",
            "business_name": "Веб-разработка «KazanCode»",
            "okved_code": "62.01",
            "region_code": "16",
            "tax_regime_code": "npd",
            "has_employees": False,
            "employee_count": 0,
            "monthly_revenue_est": 180000.0,
            "monthly_expenses_est": 25000.0,
            "store_slug": "kazan_code_dev",
            "store_title": "Разработка чат-ботов и сайтов KazanCode",
            "store_desc": "Создание мини-приложений для мессенджера MAX, Telegram-ботов и посадочных страниц.",
            "categories": ["Разработка ботов", "Поддержка"],
            "products": [
                {"name": "Разработка бота для MAX под ключ", "price": 45000.0, "unit": "проект", "marking": False, "stock": 3},
                {"name": "Аудит и рефакторинг API", "price": 20000.0, "unit": "услуга", "marking": False, "stock": 5},
                {"name": "Месячная техподдержка бота", "price": 15000.0, "unit": "месяц", "marking": False, "stock": 10}
            ]
        },
        # 3. Самозанятый репетитор
        {
            "key": "tutor_english",
            "first_name": "Мария",
            "last_name": "Белова",
            "gender": "f",
            "legal_form": "self_employed",
            "business_name": "Языковая студия «Easy English»",
            "okved_code": "85.41",
            "region_code": "78",
            "tax_regime_code": "npd",
            "has_employees": False,
            "employee_count": 0,
            "monthly_revenue_est": 110000.0,
            "monthly_expenses_est": 15000.0,
            "store_slug": "easy_english_spb",
            "store_title": "Английский онлайн с Марией Беловой",
            "store_desc": "Индивидуальные и парные уроки английского языка для взрослых и школьников. Подготовка к собеседованиям.",
            "categories": ["Индивидуальные уроки", "Абонементы"],
            "products": [
                {"name": "Пробный урок 45 минут", "price": 800.0, "unit": "урок", "marking": False, "stock": 20},
                {"name": "Абонемент на 8 уроков (1 месяц)", "price": 14000.0, "unit": "абонемент", "marking": False, "stock": 10},
                {"name": "Экспресс-подготовка к интервью", "price": 5000.0, "unit": "курс", "marking": False, "stock": 5}
            ]
        },
        # 4. ИП Магазин одежды (УСН 15% Татарстан - льгота 7%, маркировка)
        {
            "key": "fashion_brand",
            "first_name": "Михаил",
            "last_name": "Северов",
            "gender": "m",
            "legal_form": "ip",
            "business_name": "Бренд базовой одежды «Север»",
            "inn": "165501234567",
            "okved_code": "47.71",
            "region_code": "16",
            "tax_regime_code": "usn_income_expense",
            "has_employees": True,
            "employee_count": 2,
            "monthly_revenue_est": 650000.0,
            "monthly_expenses_est": 480000.0,
            "store_slug": "sever_clothing",
            "store_title": "Одежда «Север» | Базовый гардероб",
            "store_desc": "Плотные худи, оверсайз футболки и джоггеры. Вся одежда маркирована в Честный ЗНАК.",
            "categories": ["Худи и свитшоты", "Футболки"],
            "products": [
                {"name": "Худи оверсайз «Север» Графит", "price": 4500.0, "unit": "шт", "marking": True, "stock": 25},
                {"name": "Футболка Heavy Cotton Белая", "price": 2200.0, "unit": "шт", "marking": True, "stock": 50},
                {"name": "Брюки джоггеры Черные", "price": 3800.0, "unit": "шт", "marking": True, "stock": 20}
            ]
        },
        # 5. ИП розница обуви (Екатеринбург, УСН 15% ставка 7%)
        {
            "key": "shoes_shop",
            "first_name": "Роман",
            "last_name": "Морозов",
            "gender": "m",
            "legal_form": "ip",
            "business_name": "Магазин кроссовок «StepForward»",
            "inn": "667109876543",
            "okved_code": "47.72",
            "region_code": "66",
            "tax_regime_code": "usn_income_expense",
            "has_employees": True,
            "employee_count": 3,
            "monthly_revenue_est": 920000.0,
            "monthly_expenses_est": 710000.0,
            "store_slug": "step_forward_ekb",
            "store_title": "Кроссовки и кеды StepForward",
            "store_desc": "Оригинальная спортивная обувь ведущих брендов. 100% Data Matrix маркировка.",
            "categories": ["Беговые кроссовки", "Кеды casual"],
            "products": [
                {"name": "Кроссовки Urban Runner Pro", "price": 7900.0, "unit": "пара", "marking": True, "stock": 15},
                {"name": "Кеды Classic Canvas White", "price": 4500.0, "unit": "пара", "marking": True, "stock": 30}
            ]
        },
        # 6. ИП агентство маркетинга (Москва, УСН 6%)
        {
            "key": "agency_marketing",
            "first_name": "Дарья",
            "last_name": "Соколова",
            "gender": "f",
            "legal_form": "ip",
            "business_name": "Агентство трафика «Focus Media»",
            "inn": "770412345678",
            "okved_code": "73.11",
            "region_code": "77",
            "tax_regime_code": "usn_income",
            "has_employees": True,
            "employee_count": 4,
            "monthly_revenue_est": 850000.0,
            "monthly_expenses_est": 390000.0,
            "store_slug": "focus_media_agency",
            "store_title": "Таргетированная реклама и SMM | Focus Media",
            "store_desc": "Настройка рекламы в мессенджерах, поисковиках и соцсетях. Ведение аккаунтов бизнеса в MAX.",
            "categories": ["Таргет и контекст", "Комплексный SMM"],
            "products": [
                {"name": "Настройка таргетированной рекламы", "price": 35000.0, "unit": "проект", "marking": False, "stock": 10},
                {"name": "Комплексное ведение канала в MAX (месяц)", "price": 50000.0, "unit": "месяц", "marking": False, "stock": 6}
            ]
        },
        # 7. ООО Кофейня (Казань, УСН 7%, вывеска) - умеренный риск (AT_RISK)
        {
            "key": "coffee_ooo",
            "first_name": "Артур",
            "last_name": "Каримов",
            "gender": "m",
            "legal_form": "ooo",
            "business_name": "Кофейня «Зерно & Толк»",
            "inn": "1655987654",
            "okved_code": "56.10",
            "region_code": "16",
            "tax_regime_code": "usn_income_expense",
            "has_employees": True,
            "employee_count": 5,
            "monthly_revenue_est": 1350000.0,
            "monthly_expenses_est": 1180000.0,
            "store_slug": "coffee_zerno_tolk",
            "store_title": "Спешелти кофейня «Зерно & Толк»",
            "store_desc": "Свежеобжаренный спешелти кофе, сытные круассаны и полезные боулы в центре Казани.",
            "categories": ["Кофе и напитки", "Выпечка и боулы"],
            "products": [
                {"name": "Капучино на овсяном молоке 350мл", "price": 280.0, "unit": "порция", "marking": False, "stock": 100},
                {"name": "Фильтр-кофе Эфиопия Иргачефф 250мл", "price": 200.0, "unit": "порция", "marking": False, "stock": 80},
                {"name": "Круассан с миндальным кремом", "price": 240.0, "unit": "шт", "marking": False, "stock": 30}
            ]
        },
        # 8. ООО Производство / Пекарня (Москва, УСН 15%)
        {
            "key": "bakery_factory",
            "first_name": "Дмитрий",
            "last_name": "Смирнов",
            "gender": "m",
            "legal_form": "ooo",
            "business_name": "Ремесленная пекарня «Хлебная Мануфактура»",
            "inn": "7714556677",
            "okved_code": "10.71",
            "region_code": "77",
            "tax_regime_code": "usn_income_expense",
            "has_employees": True,
            "employee_count": 8,
            "monthly_revenue_est": 2200000.0,
            "monthly_expenses_est": 1750000.0,
            "store_slug": "hleb_manufaktura",
            "store_title": "Пекарня «Хлебная Мануфактура» Опт и Розница",
            "store_desc": "Бездрожжевой тартин, подовый ржаной хлеб, поставки в рестораны и доставка частным клиентам.",
            "categories": ["Подовый хлеб", "Ремесленная сдоба"],
            "products": [
                {"name": "Хлеб Тартин на закваске 600г", "price": 220.0, "unit": "шт", "marking": False, "stock": 50},
                {"name": "Багет французский традиционный", "price": 150.0, "unit": "шт", "marking": False, "stock": 40}
            ]
        },
        # 9. [Edge Case 1]: Самозанятый на пороге лимита (доход 195k/мес = 2.34 млн/год при лимите 2.4 млн)
        {
            "key": "edge_npd_threshold",
            "first_name": "Алина",
            "last_name": "Зарипова",
            "gender": "f",
            "legal_form": "self_employed",
            "business_name": "Дизайн интерьеров «Zaripova Design»",
            "okved_code": "74.10",
            "region_code": "16",
            "tax_regime_code": "npd",
            "has_employees": False,
            "employee_count": 0,
            "monthly_revenue_est": 195000.0,
            "monthly_expenses_est": 30000.0,
            "store_slug": "zaripova_interior_design",
            "store_title": "Студия дизайна Алины Зариповой",
            "store_desc": "Дизайн-проекты квартир и коммерческих помещений. 3D-визуализации и авторский надзор.",
            "categories": ["Дизайн-проекты", "Консультации"],
            "products": [
                {"name": "Полный дизайн-проект квартиры (до 60 кв.м)", "price": 120000.0, "unit": "проект", "marking": False, "stock": 2},
                {"name": "Планировочное решение с расстановкой мебели", "price": 30000.0, "unit": "проект", "marking": False, "stock": 4}
            ]
        },
        # 10. [Edge Case 2]: Убыточный бизнес ИП (расходы 240k > выручка 180k -> CRITICAL)
        {
            "key": "edge_loss_business",
            "first_name": "Денис",
            "last_name": "Васильев",
            "gender": "m",
            "legal_form": "ip",
            "business_name": "Сувенирная лавка «Подарки Казани»",
            "inn": "165599887766",
            "okved_code": "47.78",
            "region_code": "16",
            "tax_regime_code": "usn_income_expense",
            "has_employees": True,
            "employee_count": 1,
            "monthly_revenue_est": 180000.0,
            "monthly_expenses_est": 240000.0,
            "store_slug": "kazan_souvenirs_loss",
            "store_title": "Сувениры и подарки Казани",
            "store_desc": "Национальные тюбетейки, чак-чак в подарочной упаковке и керамика ручной работы.",
            "categories": ["Национальные сувениры", "Чак-чак"],
            "products": [
                {"name": "Тюбетейка бархатная с золотой вышивкой", "price": 1800.0, "unit": "шт", "marking": False, "stock": 10},
                {"name": "Подарочный набор Чак-чак 500г", "price": 650.0, "unit": "набор", "marking": False, "stock": 25}
            ]
        }
    ]

    users = []
    business_profiles = []
    stores = []
    product_categories = []
    products = []
    orders = []
    order_items = []
    tax_calculations = []
    financial_health_assessments = []
    checklist_progress = []

    cat_counter = 1
    order_counter = 1
    order_item_counter = 1
    progress_counter = 1

    for idx, sc in enumerate(scenarios, start=1):
        uid = generate_deterministic_uuid("user", sc["key"])
        prof_id = generate_deterministic_uuid("profile", sc["key"])
        store_id = generate_deterministic_uuid("store", sc["key"])
        calc_id = generate_deterministic_uuid("tax_calc", sc["key"])
        health_id = generate_deterministic_uuid("health", sc["key"])

        city, _ = CITIES_BY_REGION.get(sc["region_code"], ("Москва", "Центральный"))

        # User
        users.append({
            "id": uid,
            "max_user_id": f"max_user_{sc['key']}_{idx:02d}",
            "username": f"{sc['store_slug']}",
            "first_name": sc["first_name"],
            "last_name": sc["last_name"]
        })

        # Business Profile
        business_profiles.append({
            "id": prof_id,
            "user_id": uid,
            "legal_form": sc["legal_form"],
            "business_name": sc["business_name"],
            "inn": sc.get("inn"),
            "okved_code": sc["okved_code"],
            "region_code": sc["region_code"],
            "city": city,
            "tax_regime_code": sc["tax_regime_code"],
            "has_employees": sc["has_employees"],
            "employee_count": sc["employee_count"],
            "monthly_revenue_est": sc["monthly_revenue_est"],
            "monthly_expenses_est": sc["monthly_expenses_est"],
            "status": "active"
        })

        # Storefront
        stores.append({
            "id": store_id,
            "business_profile_id": prof_id,
            "slug": sc["store_slug"],
            "title": sc["store_title"],
            "description": sc["store_desc"],
            "contact_phone": f"+7 (9{idx:02d}) 123-45-67",
            "pickup_address": f"г. {city}, ул. Центральная, д. {idx * 5}",
            "delivery_enabled": True,
            "delivery_terms": f"Курьерская доставка по {city} от 250 руб.",
            "delivery_min_amount": 1000.0,
            "delivery_fee": 250.0,
            "is_published": True
        })

        # Categories & Products
        created_store_products = []
        for cat_name in sc["categories"]:
            cat_id = cat_counter
            cat_counter += 1
            product_categories.append({
                "id": cat_id,
                "store_id": store_id,
                "name": cat_name,
                "sort_order": cat_id % 10
            })

            # Assign products to this category
            for pr in sc["products"]:
                p_id = generate_deterministic_uuid("product", f"{sc['key']}:{pr['name']}")
                prod_obj = {
                    "id": p_id,
                    "store_id": store_id,
                    "category_id": cat_id,
                    "type": "good" if "услуг" not in pr.get("unit", "") else "service",
                    "name": pr["name"],
                    "description": f"Качественная позиция '{pr['name']}' от проверенного поставщика.",
                    "price": pr["price"],
                    "currency": "RUB",
                    "unit": pr["unit"],
                    "is_in_stock": pr["stock"] > 0,
                    "stock_quantity": pr["stock"],
                    "requires_marking": pr["marking"],
                    "status": "active"
                }
                products.append(prod_obj)
                created_store_products.append(prod_obj)

        # Orders for store
        if created_store_products:
            ord_id = generate_deterministic_uuid("order", f"{sc['key']}:01")
            sel_prod = created_store_products[0]
            qty = random.randint(1, 2)
            total = sel_prod["price"] * qty

            orders.append({
                "id": ord_id,
                "store_id": store_id,
                "order_number": f"ORD-2026-{order_counter:04d}",
                "customer_max_user_id": f"max_buyer_{order_counter:03d}",
                "customer_name": f"Покупатель {sc['first_name']}",
                "customer_phone": f"+7 (911) {order_counter:03d}-00-00",
                "delivery_type": "delivery",
                "delivery_address": f"г. {city}, ул. Проспектная, д. {order_counter}",
                "total_amount": float(total),
                "currency": "RUB",
                "status": "confirmed",
                "customer_comment": "Доставить в рабочее время"
            })
            order_items.append({
                "id": order_item_counter,
                "order_id": ord_id,
                "product_id": sel_prod["id"],
                "product_name_snapshot": sel_prod["name"],
                "unit_price_snapshot": float(sel_prod["price"]),
                "quantity": qty,
                "total_price": float(total)
            })
            order_counter += 1
            order_item_counter += 1

        # Deterministic Tax Calculations & Comparisons
        annual_rev = sc["monthly_revenue_est"] * 12
        annual_exp = sc["monthly_expenses_est"] * 12
        payroll_annual = (sc["employee_count"] * 45000.0 * 12) if sc["has_employees"] else 0.0

        # Сравнение налоговых режимов
        npd_applicable = (sc["legal_form"] == "self_employed" and annual_rev <= 2400000.0 and not sc["has_employees"])
        npd_tax = annual_rev * 0.04 if npd_applicable else None

        # УСН 6%
        usn6_tax_raw = annual_rev * 0.06
        insurance_fixed = 53658.0 if sc["legal_form"] == "ip" else 0.0
        insurance_deduction = insurance_fixed if not sc["has_employees"] else min(insurance_fixed, usn6_tax_raw * 0.5)
        usn6_tax_payable = max(0.0, usn6_tax_raw - insurance_deduction)

        # УСН 15% (с учетом ставки региона, например 7% в Татарстане)
        applied_rate = 7.0 if sc["region_code"] == "16" else (10.0 if sc["region_code"] in ("77", "54") else 15.0)
        taxable_base = max(0.0, annual_rev - annual_exp)
        usn15_tax = taxable_base * (applied_rate / 100.0)

        # Подбор оптимального режима
        recommended = sc["tax_regime_code"]
        est_tax = usn15_tax if sc["tax_regime_code"] == "usn_income_expense" else (npd_tax if sc["tax_regime_code"] == "npd" else usn6_tax_payable)
        est_insurance = insurance_fixed

        tax_calculations.append({
            "id": calc_id,
            "business_profile_id": prof_id,
            "input_parameters": {
                "annual_revenue": annual_rev,
                "annual_expenses": annual_exp,
                "has_employees": sc["has_employees"],
                "employee_count": sc["employee_count"],
                "region_code": sc["region_code"],
                "legal_form": sc["legal_form"]
            },
            "results_comparison": {
                "npd": {"applicable": npd_applicable, "annual_tax": npd_tax, "notes": "Чеки в 'Мой налог', касса 54-ФЗ не нужна"},
                "usn_income": {"applicable": True, "annual_tax": round(usn6_tax_payable, 2), "deduction": insurance_deduction},
                "usn_income_expense": {"applicable": True, "annual_tax": round(usn15_tax, 2), "rate_percent": applied_rate}
            },
            "recommended_regime_code": recommended,
            "estimated_annual_tax": round(float(est_tax), 2),
            "estimated_annual_insurance": round(float(est_insurance), 2),
            "assumptions": [
                "Расчет является оценочным и носит справочный характер",
                f"Применена региональная ставка УСН {applied_rate}% для региона {sc['region_code']}",
                "Фиксированные страховые взносы ИП рассчитаны по ставке 2026 года (53 658 руб.)"
            ]
        })

        # Оценка финансовой устойчивости (Unit-экономика и выживаемость)
        rev = sc["monthly_revenue_est"]
        exp = sc["monthly_expenses_est"]
        fixed_costs = exp * 0.45
        variable_costs = exp * 0.55
        monthly_tax = est_tax / 12.0
        net_profit = rev - exp - monthly_tax

        margin = (net_profit / rev * 100.0) if rev > 0 else 0.0
        contribution_margin_ratio = (rev - variable_costs) / rev if rev > 0 else 0.1
        break_even = (fixed_costs / contribution_margin_ratio) if contribution_margin_ratio > 0 else fixed_costs * 2
        safety_margin = ((rev - break_even) / rev * 100.0) if rev > 0 else -100.0

        if net_profit <= 0 or safety_margin < 0:
            risk_level = "HIGH"
            verdict = "CRITICAL"
            recs = [
                "Бизнес генерирует операционный убыток. Требуется срочный пересмотр постоянных расходов.",
                "Снизить аренду или фонд оплаты труда на 25%.",
                "Рассмотреть повышение отпускных цен в каталоге магазина в MAX."
            ]
        elif safety_margin < 20.0:
            risk_level = "MEDIUM"
            verdict = "AT_RISK"
            recs = [
                "Запас финансовой прочности менее 20%. Риск кассового разрыва при задержках оплат.",
                "Сформировать подушку ликвидности в размере 2-месячных постоянных затрат.",
                "Провести оптимизацию налоговой нагрузки."
            ]
        else:
            risk_level = "LOW"
            verdict = "SUSTAINABLE"
            recs = [
                "Бизнес устойчив. Запас прочности превышает безопасный порог 25%.",
                "Рекомендуется масштабирование: подключение доставки курьерами и рекламы в MAX.",
                "Рассмотреть подачу заявки на региональный грант до 500 тыс. руб."
            ]

        # Для краевого случая 9 (порог НПД)
        if sc["key"] == "edge_npd_threshold":
            recs.insert(0, "ВНИМАНИЕ: Годовой доход приближается к 2.4 млн руб. Подготовьте документы для перехода на ИП на УСН 6%.")

        financial_health_assessments.append({
            "id": health_id,
            "business_profile_id": prof_id,
            "tax_calculation_id": calc_id,
            "monthly_revenue": round(rev, 2),
            "monthly_fixed_costs": round(fixed_costs, 2),
            "monthly_variable_costs": round(variable_costs, 2),
            "net_monthly_profit": round(net_profit, 2),
            "margin_percent": round(margin, 2),
            "break_even_point": round(break_even, 2),
            "safety_margin_percent": round(safety_margin, 2),
            "cash_gap_risk_level": risk_level,
            "verdict": verdict,
            "recommendations": recs
        })

        # Прогресс чек-листов (привязка задач)
        chk_id = 1 if sc["legal_form"] == "self_employed" else (2 if sc["legal_form"] == "ip" else 3)
        for item_offset in range(1, 4):
            checklist_progress.append({
                "id": progress_counter,
                "business_profile_id": prof_id,
                "checklist_item_id": (chk_id - 1) * 4 + item_offset,
                "is_completed": (item_offset <= 2)
            })
            progress_counter += 1

    return {
        "users": users,
        "business_profiles": business_profiles,
        "stores": stores,
        "product_categories": product_categories,
        "products": products,
        "orders": orders,
        "order_items": order_items,
        "tax_calculations": tax_calculations,
        "financial_health_assessments": financial_health_assessments,
        "user_checklist_progress": checklist_progress
    }


def export_dataset_to_json(dataset: dict, output_path: Path):
    """Экспорт датасета в JSON-файл."""
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(dataset, f, ensure_ascii=False, indent=2)
    print(f"[OK] Сгенерирован JSON-датасет: {output_path}")


def export_dataset_to_sql(dataset: dict, output_path: Path):
    """Экспорт датасета в чистый идемпотентный SQL-скрипт миграции PostgreSQL."""
    sql = []
    sql.append("-- ============================================================================")
    sql.append("-- Миграция 04_seed_test_dataset.sql")
    sql.append("-- Наполнение эталонными и граничными тестовыми данными для жюри")
    sql.append("-- Идемпотентность: повторный запуск безопасен (ON CONFLICT DO UPDATE)")
    sql.append("-- ============================================================================\n")

    # 1. users
    sql.append("-- 1. Пользователи MAX")
    for u in dataset["users"]:
        sql.append(f"INSERT INTO users (id, max_user_id, username, first_name, last_name) "
                   f"VALUES ('{u['id']}', '{u['max_user_id']}', '{u['username']}', '{u['first_name']}', '{u['last_name']}') "
                   f"ON CONFLICT (max_user_id) DO UPDATE SET username = EXCLUDED.username;")
    sql.append("")

    # 2. business_profiles
    sql.append("-- 2. Профили бизнеса")
    for p in dataset["business_profiles"]:
        inn_val = f"'{p['inn']}'" if p["inn"] else "NULL"
        has_emp = "TRUE" if p["has_employees"] else "FALSE"
        sql.append(f"INSERT INTO business_profiles (id, user_id, legal_form, business_name, inn, okved_code, region_code, city, tax_regime_code, has_employees, employee_count, monthly_revenue_est, monthly_expenses_est, status) "
                   f"VALUES ('{p['id']}', '{p['user_id']}', '{p['legal_form']}', '{p['business_name']}', {inn_val}, '{p['okved_code']}', '{p['region_code']}', '{p['city']}', '{p['tax_regime_code']}', {has_emp}, {p['employee_count']}, {p['monthly_revenue_est']}, {p['monthly_expenses_est']}, '{p['status']}') "
                   f"ON CONFLICT (id) DO UPDATE SET monthly_revenue_est = EXCLUDED.monthly_revenue_est, monthly_expenses_est = EXCLUDED.monthly_expenses_est;")
    sql.append("")

    # 3. stores
    sql.append("-- 3. Интернет-магазины в MAX")
    for s in dataset["stores"]:
        deliv = "TRUE" if s["delivery_enabled"] else "FALSE"
        pub = "TRUE" if s["is_published"] else "FALSE"
        desc_esc = s['description'].replace("'", "''")
        terms_esc = s['delivery_terms'].replace("'", "''")
        sql.append(f"INSERT INTO stores (id, business_profile_id, slug, title, description, contact_phone, pickup_address, delivery_enabled, delivery_terms, delivery_min_amount, delivery_fee, is_published) "
                   f"VALUES ('{s['id']}', '{s['business_profile_id']}', '{s['slug']}', '{s['title']}', '{desc_esc}', '{s['contact_phone']}', '{s['pickup_address']}', {deliv}, '{terms_esc}', {s['delivery_min_amount']}, {s['delivery_fee']}, {pub}) "
                   f"ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description;")
    sql.append("")

    # 4. product_categories
    sql.append("-- 4. Категории товаров")
    for c in dataset["product_categories"]:
        sql.append(f"INSERT INTO product_categories (id, store_id, name, sort_order) "
                   f"VALUES ({c['id']}, '{c['store_id']}', '{c['name']}', {c['sort_order']}) "
                   f"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;")
    sql.append("SELECT setval('product_categories_id_seq', (SELECT MAX(id) FROM product_categories));\n")

    # 5. products
    sql.append("-- 5. Товары и услуги")
    for pr in dataset["products"]:
        stock = "TRUE" if pr["is_in_stock"] else "FALSE"
        mark = "TRUE" if pr["requires_marking"] else "FALSE"
        desc_esc = pr['description'].replace("'", "''")
        name_esc = pr['name'].replace("'", "''")
        sql.append(f"INSERT INTO products (id, store_id, category_id, type, name, description, price, currency, unit, is_in_stock, stock_quantity, requires_marking, status) "
                   f"VALUES ('{pr['id']}', '{pr['store_id']}', {pr['category_id']}, '{pr['type']}', '{name_esc}', '{desc_esc}', {pr['price']}, '{pr['currency']}', '{pr['unit']}', {stock}, {pr['stock_quantity']}, {mark}, '{pr['status']}') "
                   f"ON CONFLICT (id) DO UPDATE SET price = EXCLUDED.price, is_in_stock = EXCLUDED.is_in_stock;")
    sql.append("")

    # 6. orders & items
    sql.append("-- 6. Клиентские заказы")
    for o in dataset["orders"]:
        sql.append(f"INSERT INTO orders (id, store_id, order_number, customer_max_user_id, customer_name, customer_phone, delivery_type, delivery_address, total_amount, currency, status, customer_comment) "
                   f"VALUES ('{o['id']}', '{o['store_id']}', '{o['order_number']}', '{o['customer_max_user_id']}', '{o['customer_name']}', '{o['customer_phone']}', '{o['delivery_type']}', '{o['delivery_address']}', {o['total_amount']}, '{o['currency']}', '{o['status']}', '{o['customer_comment']}') "
                   f"ON CONFLICT (id) DO UPDATE SET total_amount = EXCLUDED.total_amount, status = EXCLUDED.status;")
    for oi in dataset["order_items"]:
        snap_esc = oi['product_name_snapshot'].replace("'", "''")
        sql.append(f"INSERT INTO order_items (id, order_id, product_id, product_name_snapshot, unit_price_snapshot, quantity, total_price) "
                   f"VALUES ({oi['id']}, '{oi['order_id']}', '{oi['product_id']}', '{snap_esc}', {oi['unit_price_snapshot']}, {oi['quantity']}, {oi['total_price']}) "
                   f"ON CONFLICT (id) DO UPDATE SET total_price = EXCLUDED.total_price;")
    sql.append("SELECT setval('order_items_id_seq', (SELECT MAX(id) FROM order_items));\n")

    # 7. tax_calculations
    sql.append("-- 7. Расчеты налогового калькулятора")
    for tc in dataset["tax_calculations"]:
        inp_json = json.dumps(tc["input_parameters"], ensure_ascii=False).replace("'", "''")
        res_json = json.dumps(tc["results_comparison"], ensure_ascii=False).replace("'", "''")
        assump_array = "ARRAY[" + ", ".join([f"'{a.replace(\"'\", \"''\")}'" for a in tc["assumptions"]]) + "]::text[]"
        sql.append(f"INSERT INTO tax_calculations (id, business_profile_id, input_parameters, results_comparison, recommended_regime_code, estimated_annual_tax, estimated_annual_insurance, assumptions) "
                   f"VALUES ('{tc['id']}', '{tc['business_profile_id']}', '{inp_json}'::jsonb, '{res_json}'::jsonb, '{tc['recommended_regime_code']}', {tc['estimated_annual_tax']}, {tc['estimated_annual_insurance']}, {assump_array}) "
                   f"ON CONFLICT (id) DO UPDATE SET estimated_annual_tax = EXCLUDED.estimated_annual_tax;")
    sql.append("")

    # 8. financial_health_assessments
    sql.append("-- 8. Оценки финансовой устойчивости бизнеса")
    for fh in dataset["financial_health_assessments"]:
        recs_json = json.dumps(fh["recommendations"], ensure_ascii=False).replace("'", "''")
        sql.append(f"INSERT INTO financial_health_assessments (id, business_profile_id, tax_calculation_id, monthly_revenue, monthly_fixed_costs, monthly_variable_costs, net_monthly_profit, margin_percent, break_even_point, safety_margin_percent, cash_gap_risk_level, verdict, recommendations) "
                   f"VALUES ('{fh['id']}', '{fh['business_profile_id']}', '{fh['tax_calculation_id']}', {fh['monthly_revenue']}, {fh['monthly_fixed_costs']}, {fh['monthly_variable_costs']}, {fh['net_monthly_profit']}, {fh['margin_percent']}, {fh['break_even_point']}, {fh['safety_margin_percent']}, '{fh['cash_gap_risk_level']}', '{fh['verdict']}', '{recs_json}'::jsonb) "
                   f"ON CONFLICT (id) DO UPDATE SET net_monthly_profit = EXCLUDED.net_monthly_profit, margin_percent = EXCLUDED.margin_percent, verdict = EXCLUDED.verdict;")
    sql.append("")

    # 9. user_checklist_progress
    sql.append("-- 9. Прогресс чек-листов")
    for cp in dataset["user_checklist_progress"]:
        comp_val = "TRUE" if cp["is_completed"] else "FALSE"
        sql.append(f"INSERT INTO user_checklist_progress (id, business_profile_id, checklist_item_id, is_completed) "
                   f"VALUES ({cp['id']}, '{cp['business_profile_id']}', {cp['checklist_item_id']}, {comp_val}) "
                   f"ON CONFLICT (business_profile_id, checklist_item_id) DO UPDATE SET is_completed = EXCLUDED.is_completed;")
    sql.append("SELECT setval('user_checklist_progress_id_seq', (SELECT MAX(id) FROM user_checklist_progress));\n")

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql))
    print(f"[OK] Сгенерирован SQL-скрипт сидов: {output_path}")


def main():
    parser = argparse.ArgumentParser(description="Генератор тестовых и эталонных данных")
    parser.add_argument("--seed", type=int, default=42, help="Seed для генератора псевдослучайных чисел (по умолчанию: 42)")
    parser.add_argument("--output-json", type=str, default=str(SEEDS_DIR / "test_dataset.json"), help="Путь к результирующему JSON")
    parser.add_argument("--output-sql", type=str, default=str(MIGRATIONS_DIR / "04_seed_test_dataset.sql"), help="Путь к результирующему SQL")
    args = parser.parse_args()

    print(f"=== Запуск генератора тестовых данных (Seed={args.seed}) ===")
    dataset = build_scenario_dataset(seed=args.seed)

    print(f"Сгенерировано:")
    print(f"  - Пользователей: {len(dataset['users'])}")
    print(f"  - Профилей бизнеса: {len(dataset['business_profiles'])}")
    print(f"  - Витрин магазинов: {len(dataset['stores'])}")
    print(f"  - Категорий товаров: {len(dataset['product_categories'])}")
    print(f"  - Товаров и услуг: {len(dataset['products'])}")
    print(f"  - Заказов покупателей: {len(dataset['orders'])}")
    print(f"  - Расчетов налогов: {len(dataset['tax_calculations'])}")
    print(f"  - Оценок фин. устойчивости: {len(dataset['financial_health_assessments'])}")
    print(f"  - Прогресса по чек-листам: {len(dataset['user_checklist_progress'])}")

    export_dataset_to_json(dataset, Path(args.output_json))
    export_dataset_to_sql(dataset, Path(args.output_sql))
    print("=== Генерация тестовых данных успешно завершена ===")


if __name__ == "__main__":
    main()
