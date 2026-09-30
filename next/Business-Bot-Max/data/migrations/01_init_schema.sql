-- ============================================================================
-- Проект: Конструктор бизнеса и экосистема агентов-помощников в MAX
-- Трек: «Эффективный бизнес» (Хакатон MAX 2026)
-- Скрипт: 01_init_schema.sql
-- Назначение: Инициализация полной структуры реляционной БД PostgreSQL 15+
-- Идемпотентность: повторный запуск безопасен (IF NOT EXISTS)
-- ============================================================================

-- 0. Расширения
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ----------------------------------------------------------------------------
-- Вспомогательная функция для автоматического обновления updated_at
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 1. ДОМЕН: ПОЛЬЗОВАТЕЛИ И ПРОФИЛИ БИЗНЕСА (PROFILE DOMAIN)
-- ============================================================================

-- Таблица пользователей экосистемы MAX
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    max_user_id VARCHAR(128) UNIQUE NOT NULL,      -- Обезличенный ID пользователя в MAX
    username VARCHAR(64),                          -- Никнейм в MAX (при наличии)
    first_name VARCHAR(128),
    last_name VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_max_id ON users (max_user_id);

-- Таблица карточек/профилей бизнеса
CREATE TABLE IF NOT EXISTS business_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    legal_form VARCHAR(20) NOT NULL,
    business_name VARCHAR(255) NOT NULL,
    inn VARCHAR(12),                               -- Опционально для МСП / ИП
    okved_code VARCHAR(16) NOT NULL,               -- Основной код деятельности ОКВЭД-2
    region_code VARCHAR(10) NOT NULL,              -- Код субъекта РФ ('77', '16' и т.д.)
    city VARCHAR(128) NOT NULL,
    tax_regime_code VARCHAR(32) NOT NULL,          -- Код выбранного налогового режима
    has_employees BOOLEAN NOT NULL DEFAULT FALSE,  -- Есть ли наемные работники
    employee_count INTEGER NOT NULL DEFAULT 0,     -- Число сотрудников
    monthly_revenue_est NUMERIC(14,2) NOT NULL DEFAULT 0.00,  -- Оценочная месячная выручка
    monthly_expenses_est NUMERIC(14,2) NOT NULL DEFAULT 0.00, -- Оценочные месячные расходы
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ограничения целостности (Data Quality Constraints):
    CONSTRAINT chk_legal_form CHECK (legal_form IN ('self_employed', 'ip', 'ooo')),
    CONSTRAINT chk_profile_status CHECK (status IN ('draft', 'active', 'archived')),
    CONSTRAINT chk_non_negative_finances CHECK (
        monthly_revenue_est >= 0 AND 
        monthly_expenses_est >= 0 AND 
        employee_count >= 0
    ),
    -- Бизнес-правило DQ-02: Самозанятый не может иметь сотрудников
    CONSTRAINT chk_npd_no_employees CHECK (
        legal_form <> 'self_employed' OR (has_employees = FALSE AND employee_count = 0)
    ),
    -- Бизнес-правило DQ-01: Контроль порога выручки самозанятого (2.4 млн руб. в год)
    CONSTRAINT chk_npd_revenue_limit CHECK (
        legal_form <> 'self_employed' OR (monthly_revenue_est * 12 <= 2400000.00)
    )
);

CREATE INDEX IF NOT EXISTS idx_business_profiles_user_id ON business_profiles (user_id);
CREATE INDEX IF NOT EXISTS idx_business_profiles_region ON business_profiles (region_code);
CREATE INDEX IF NOT EXISTS idx_business_profiles_okved ON business_profiles (okved_code);

-- ============================================================================
-- 2. ДОМЕН: КОНСТРУКТОР МАГАЗИНА / ВИТРИНЫ MAX (COMMERCE DOMAIN)
-- ============================================================================

-- Витрины интернет-магазинов в MAX
CREATE TABLE IF NOT EXISTS stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_profile_id UUID NOT NULL REFERENCES business_profiles(id) ON DELETE CASCADE,
    slug VARCHAR(64) UNIQUE NOT NULL,              -- Ссылка на витрину в MAX: max.ru/store?slug=torty_kzn
    title VARCHAR(128) NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    banner_url VARCHAR(512),
    contact_phone VARCHAR(32),
    contact_max_username VARCHAR(64),
    pickup_address TEXT,
    delivery_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    delivery_terms TEXT,
    delivery_min_amount NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    delivery_fee NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_store_delivery_amounts CHECK (delivery_min_amount >= 0 AND delivery_fee >= 0)
);

CREATE INDEX IF NOT EXISTS idx_stores_profile ON stores (business_profile_id);
CREATE INDEX IF NOT EXISTS idx_stores_slug ON stores (slug);

-- Категории товаров / услуг
CREATE TABLE IF NOT EXISTS product_categories (
    id BIGSERIAL PRIMARY KEY,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_store ON product_categories (store_id);

-- Каталог товаров и услуг
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES product_categories(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'good',
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    unit VARCHAR(20) NOT NULL DEFAULT 'шт',
    sku VARCHAR(64),
    is_in_stock BOOLEAN NOT NULL DEFAULT TRUE,
    stock_quantity INTEGER DEFAULT 0,
    image_url VARCHAR(512),
    requires_marking BOOLEAN NOT NULL DEFAULT FALSE, -- Признак маркировки «Честный ЗНАК»
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_product_type CHECK (type IN ('good', 'service')),
    CONSTRAINT chk_product_status CHECK (status IN ('active', 'draft', 'archived')),
    CONSTRAINT chk_product_price CHECK (price >= 0),
    CONSTRAINT chk_product_stock CHECK (stock_quantity >= 0)
);

CREATE INDEX IF NOT EXISTS idx_products_store ON products (store_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);
-- Полнотекстовый поиск по товарам на русском языке
CREATE INDEX IF NOT EXISTS idx_products_fts ON products USING gin(to_tsvector('russian', name || ' ' || COALESCE(description, '')));

-- Заказы и заявки клиентов
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_number VARCHAR(32) NOT NULL,             -- Номер вида ORD-2026-0001
    customer_max_user_id VARCHAR(128) NOT NULL,    -- Идентификатор покупателя в MAX
    customer_name VARCHAR(128) NOT NULL,
    customer_phone VARCHAR(32) NOT NULL,
    delivery_type VARCHAR(20) NOT NULL,
    delivery_address TEXT,
    total_amount NUMERIC(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status VARCHAR(32) NOT NULL DEFAULT 'new',
    customer_comment TEXT,
    merchant_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_order_delivery_type CHECK (delivery_type IN ('pickup', 'delivery')),
    CONSTRAINT chk_order_amount CHECK (total_amount >= 0),
    CONSTRAINT chk_order_status CHECK (status IN ('new', 'confirmed', 'in_progress', 'ready_for_pickup', 'delivering', 'completed', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_orders_store ON orders (store_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders (customer_max_user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

-- Позиции в заказе (с фиксацией исторической цены и названия)
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name_snapshot VARCHAR(255) NOT NULL,
    unit_price_snapshot NUMERIC(12,2) NOT NULL,
    quantity INTEGER NOT NULL,
    total_price NUMERIC(12,2) NOT NULL,

    CONSTRAINT chk_order_item_quantity CHECK (quantity > 0),
    CONSTRAINT chk_order_item_prices CHECK (unit_price_snapshot >= 0 AND total_price >= 0)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);

-- ============================================================================
-- 3. ДОМЕН: НАЛОГИ И ФИНАНСОВАЯ УСТОЙЧИВОСТЬ (FINANCE DOMAIN)
-- ============================================================================

-- Реестр первоисточников нормативных актов
CREATE TABLE IF NOT EXISTS legal_sources (
    id VARCHAR(64) PRIMARY KEY,                    -- 'nk_rf', 'fz_54_kkt', 'fz_422_npd' и др.
    name VARCHAR(255) NOT NULL,
    official_url VARCHAR(512) NOT NULL,
    publisher VARCHAR(128) NOT NULL,
    source_type VARCHAR(32) NOT NULL,              -- 'code_rf', 'federal_law', 'gov_decree'
    last_verified_at DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
);

-- Справочник налоговых режимов РФ
CREATE TABLE IF NOT EXISTS tax_regimes (
    code VARCHAR(32) PRIMARY KEY,                  -- 'npd', 'usn_income', 'usn_income_expense', 'ausn'
    name VARCHAR(128) NOT NULL,
    short_name VARCHAR(32) NOT NULL,
    description TEXT,
    allowed_legal_forms TEXT[] NOT NULL,           -- ARRAY['self_employed'], ARRAY['ip', 'ooo']
    max_annual_revenue NUMERIC(14,2) NOT NULL,
    max_employees INTEGER NOT NULL,
    base_tax_rate_percent NUMERIC(5,2) NOT NULL,
    has_regional_discounts BOOLEAN NOT NULL DEFAULT FALSE,
    fixed_insurance_required BOOLEAN NOT NULL DEFAULT FALSE,
    fixed_insurance_amount_annual NUMERIC(10,2) DEFAULT 0.00, -- Взносы ИП за себя (на 2026 год)
    reporting_frequency VARCHAR(20) NOT NULL DEFAULT 'annual', -- 'none', 'quarterly', 'annual'
    kkt_required_by_default BOOLEAN NOT NULL DEFAULT TRUE,
    legal_source_id VARCHAR(64) REFERENCES legal_sources(id) ON DELETE SET NULL,
    effective_year INTEGER NOT NULL DEFAULT 2026
);

-- Региональные налоговые ставки (дифференцированные ставки УСН по субъектам РФ)
CREATE TABLE IF NOT EXISTS tax_regional_rates (
    id BIGSERIAL PRIMARY KEY,
    tax_regime_code VARCHAR(32) NOT NULL REFERENCES tax_regimes(code) ON DELETE CASCADE,
    region_code VARCHAR(10) NOT NULL,
    rate_percent NUMERIC(5,2) NOT NULL,
    conditions_description TEXT,
    legal_act_title VARCHAR(255),
    effective_from DATE NOT NULL,
    effective_to DATE,

    CONSTRAINT chk_tax_regional_rate CHECK (rate_percent >= 0 AND rate_percent <= 100)
);

CREATE INDEX IF NOT EXISTS idx_tax_reg_rates ON tax_regional_rates (region_code, tax_regime_code);

-- История и результаты расчетов налогового калькулятора
CREATE TABLE IF NOT EXISTS tax_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_profile_id UUID NOT NULL REFERENCES business_profiles(id) ON DELETE CASCADE,
    input_parameters JSONB NOT NULL,               -- Входные параметры (доходы, расходы, ФОТ, работники)
    results_comparison JSONB NOT NULL,             -- Сравнение режимов (НПД vs УСН6 vs УСН15)
    recommended_regime_code VARCHAR(32) NOT NULL REFERENCES tax_regimes(code),
    estimated_annual_tax NUMERIC(14,2) NOT NULL,
    estimated_annual_insurance NUMERIC(14,2) NOT NULL,
    assumptions TEXT[] NOT NULL,                   -- Допущения и оговорки расчёта
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tax_calc_profile ON tax_calculations (business_profile_id);

-- Оценка жизнеспособности и устойчивости бизнеса (Unit-экономика)
CREATE TABLE IF NOT EXISTS financial_health_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_profile_id UUID NOT NULL REFERENCES business_profiles(id) ON DELETE CASCADE,
    tax_calculation_id UUID REFERENCES tax_calculations(id) ON DELETE SET NULL,
    monthly_revenue NUMERIC(14,2) NOT NULL,
    monthly_fixed_costs NUMERIC(14,2) NOT NULL,
    monthly_variable_costs NUMERIC(14,2) NOT NULL,
    net_monthly_profit NUMERIC(14,2) NOT NULL,
    margin_percent NUMERIC(5,2) NOT NULL,
    break_even_point NUMERIC(14,2) NOT NULL,       -- Точка безубыточности в рублях
    safety_margin_percent NUMERIC(5,2) NOT NULL,   -- Запас прочности в %
    cash_gap_risk_level VARCHAR(16) NOT NULL,      -- 'LOW', 'MEDIUM', 'HIGH'
    verdict VARCHAR(32) NOT NULL,                  -- 'SUSTAINABLE', 'AT_RISK', 'CRITICAL'
    recommendations JSONB NOT NULL,                -- Рекомендации по оптимизации
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_health_risk_level CHECK (cash_gap_risk_level IN ('LOW', 'MEDIUM', 'HIGH')),
    CONSTRAINT chk_health_verdict CHECK (verdict IN ('SUSTAINABLE', 'AT_RISK', 'CRITICAL'))
);

CREATE INDEX IF NOT EXISTS idx_health_profile ON financial_health_assessments (business_profile_id);

-- ============================================================================
-- 4. ДОМЕН: НОРМАТИВНО-ПРАВОВАЯ БАЗА И ТРЕБОВАНИЯ (LEGAL DOMAIN)
-- ============================================================================

-- Каталог верифицированных юридических требований к бизнесу
CREATE TABLE IF NOT EXISTS business_requirements (
    id BIGSERIAL PRIMARY KEY,
    source_id VARCHAR(64) NOT NULL REFERENCES legal_sources(id) ON DELETE CASCADE,
    category VARCHAR(32) NOT NULL,                 -- 'registration', 'taxes', 'kkt_receipts', 'signboards', 'personal_data', 'labeling'
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    legal_citation TEXT NOT NULL,                  -- Точная статья НПА
    applies_to_legal_forms TEXT[] NOT NULL,        -- ARRAY['self_employed', 'ip', 'ooo']
    applies_to_okved_prefixes TEXT[],              -- Фильтр по кодам ОКВЭД (NULL = для всех)
    region_code VARCHAR(10),                       -- Региональное правило (NULL = федеральное)
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE,
    penalty_risk_description TEXT,                 -- Штрафы и риски за несоблюдение
    action_plan JSONB NOT NULL,                    -- Пошаговый алгоритм выполнения
    is_simulated_in_mvp BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_req_category CHECK (category IN ('registration', 'taxes', 'kkt_receipts', 'signboards', 'personal_data', 'labeling', 'sanitary', 'consumer_rights', 'state_support'))
);

CREATE INDEX IF NOT EXISTS idx_requirements_category ON business_requirements (category);
CREATE INDEX IF NOT EXISTS idx_requirements_source ON business_requirements (source_id);
-- Полнотекстовый поиск по правовым требованиям
CREATE INDEX IF NOT EXISTS idx_requirements_fts ON business_requirements USING gin(to_tsvector('russian', title || ' ' || summary));

-- Шаблоны чек-листов запуска и проверок
CREATE TABLE IF NOT EXISTS checklists (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) UNIQUE NOT NULL,              -- 'checklist_self_employed_start', 'checklist_shop_launch'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_legal_form VARCHAR(20) DEFAULT 'all',
    target_okved_prefix VARCHAR(16),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Пункты чек-листов, привязанные к требованиям
CREATE TABLE IF NOT EXISTS checklist_items (
    id BIGSERIAL PRIMARY KEY,
    checklist_id BIGINT NOT NULL REFERENCES checklists(id) ON DELETE CASCADE,
    requirement_id BIGINT REFERENCES business_requirements(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_mandatory BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_chk_items_list ON checklist_items (checklist_id);

-- Прогресс прохождения чек-листов конкретным бизнесом
CREATE TABLE IF NOT EXISTS user_checklist_progress (
    id BIGSERIAL PRIMARY KEY,
    business_profile_id UUID NOT NULL REFERENCES business_profiles(id) ON DELETE CASCADE,
    checklist_item_id BIGINT NOT NULL REFERENCES checklist_items(id) ON DELETE CASCADE,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    notes TEXT,
    UNIQUE (business_profile_id, checklist_item_id)
);

CREATE INDEX IF NOT EXISTS idx_user_chk_progress ON user_checklist_progress (business_profile_id);

-- Модельный справочник: Согласование вывесок и наружной рекламы
CREATE TABLE IF NOT EXISTS signboard_regulations (
    id BIGSERIAL PRIMARY KEY,
    region_code VARCHAR(10) NOT NULL,
    city VARCHAR(128) NOT NULL,
    authority_name VARCHAR(255) NOT NULL,
    legal_act_title VARCHAR(255) NOT NULL,
    max_sign_height_meters NUMERIC(4,2) NOT NULL DEFAULT 0.50,
    max_sign_width_percent_of_facade INTEGER NOT NULL DEFAULT 70,
    requires_special_permit BOOLEAN NOT NULL DEFAULT FALSE,
    required_documents TEXT[] NOT NULL,
    processing_time_days INTEGER NOT NULL DEFAULT 15,
    official_portal_url VARCHAR(512),
    is_mocked_data BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_signboards_region ON signboard_regulations (region_code, city);

-- ============================================================================
-- 5. ДОМЕН: ДИАЛОГИ И АУДИТ АГЕНТОВ MAX (AGENT & AUDIT DOMAIN)
-- ============================================================================

-- История обращений к AI-агентам и контекст сессий
CREATE TABLE IF NOT EXISTS agent_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    business_profile_id UUID REFERENCES business_profiles(id) ON DELETE SET NULL,
    topic VARCHAR(64) NOT NULL,                    -- 'tax_consultation', 'store_setup', 'signboard_help', 'risk_check'
    messages JSONB NOT NULL DEFAULT '[]'::jsonb,   -- История сообщений (роли user/assistant, цитаты НПА)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_agent_topic CHECK (topic IN ('tax_consultation', 'store_setup', 'signboard_help', 'risk_check', 'general_faq'))
);

CREATE INDEX IF NOT EXISTS idx_agent_conv_user ON agent_conversations (user_id);
CREATE INDEX IF NOT EXISTS idx_agent_conv_profile ON agent_conversations (business_profile_id);

-- ============================================================================
-- 6. ТРИГГЕРЫ ОБНОВЛЕНИЯ TIMESTAMPTZ
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_users_updated_at') THEN
        CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_business_profiles_updated_at') THEN
        CREATE TRIGGER trg_business_profiles_updated_at BEFORE UPDATE ON business_profiles FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_stores_updated_at') THEN
        CREATE TRIGGER trg_stores_updated_at BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_products_updated_at') THEN
        CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_orders_updated_at') THEN
        CREATE TRIGGER trg_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_requirements_updated_at') THEN
        CREATE TRIGGER trg_requirements_updated_at BEFORE UPDATE ON business_requirements FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_agent_conv_updated_at') THEN
        CREATE TRIGGER trg_agent_conv_updated_at BEFORE UPDATE ON agent_conversations FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
    END IF;
END $$;

