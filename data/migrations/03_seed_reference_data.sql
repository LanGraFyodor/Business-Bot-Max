-- ============================================================================
-- Миграция 03_seed_reference_data.sql
-- Наполнение базовых справочников: tax_regimes, tax_regional_rates, checklists, signboard_regulations
-- ============================================================================

-- 1. Справочник налоговых режимов РФ
INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('npd', 'Налог на профессиональный доход', 'НПД (Самозанятость)', 'Специальный налоговый режим для самозанятых граждан и ИП без наемных сотрудников с годовым доходом до 2.4 млн руб.', ARRAY['self_employed']::text[], 2400000, 0, 4, FALSE, FALSE, 0, 'none', FALSE, 'fz_422_npd', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('usn_income', 'Упрощенная система налогообложения (Доходы)', 'УСН 6%', 'Налог со всей полученной выручки по ставке 6%. Выгодно при низких подтвержденных расходах (услуги, IT, консалтинг, аренда).', ARRAY['ip', 'ooo']::text[], 450000000, 130, 6, TRUE, TRUE, 53658, 'annual', TRUE, 'nk_rf_usn', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('usn_income_expense', 'Упрощенная система налогообложения (Доходы минус расходы)', 'УСН 15%', 'Налог с чистой прибыли (доходы минус документально подтвержденные расходы). Выгодно при доле расходов свыше 60-70% (розница, общепит, производство).', ARRAY['ip', 'ooo']::text[], 450000000, 130, 15, TRUE, TRUE, 53658, 'annual', TRUE, 'nk_rf_usn', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('ausn_income', 'Автоматизированная УСН (Доходы)', 'АУСН 8%', 'Экспериментальный режим: налог 8%, полное освобождение от налоговых деклараций и страховых взносов, расчет налога банком.', ARRAY['ip', 'ooo']::text[], 60000000, 5, 8, FALSE, FALSE, 0, 'none', TRUE, 'nk_rf_usn', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('ausn_income_expense', 'Автоматизированная УСН (Доходы минус расходы)', 'АУСН 20%', 'Налог 20% с разницы доходов и расходов, без отчетности и без взносов. Расходы подтверждаются безналично через расчетный счет.', ARRAY['ip', 'ooo']::text[], 60000000, 5, 20, FALSE, FALSE, 0, 'none', TRUE, 'nk_rf_usn', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('patent', 'Патентная система налогообложения', 'ПСН (Патент)', 'Фиксированная стоимость патента, рассчитываемая от потенциального дохода региона. Без сдачи деклараций, только для ИП.', ARRAY['ip']::text[], 60000000, 15, 6, FALSE, TRUE, 53658, 'none', TRUE, 'nk_rf_usn', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

INSERT INTO tax_regimes (code, name, short_name, description, allowed_legal_forms, max_annual_revenue, max_employees, base_tax_rate_percent, has_regional_discounts, fixed_insurance_required, fixed_insurance_amount_annual, reporting_frequency, kkt_required_by_default, legal_source_id, effective_year)
VALUES ('osno', 'Общая система налогообложения', 'ОСНО', 'Классическая система налогообложения с уплатой НДС 20% и налога на прибыль 20-25% (или НДФЛ 13-15% для ИП).', ARRAY['ip', 'ooo']::text[], 999999999999, 99999, 20, FALSE, TRUE, 53658, 'quarterly', TRUE, 'nk_rf_usn', 2026)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, base_tax_rate_percent = EXCLUDED.base_tax_rate_percent, max_annual_revenue = EXCLUDED.max_annual_revenue, fixed_insurance_amount_annual = EXCLUDED.fixed_insurance_amount_annual;

-- 2. Региональные ставки налогов (УСН)
INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (1, 'usn_income_expense', '16', 7, 'Пониженная ставка для всех видов деятельности в Республике Татарстан', 'Закон Республики Татарстан от 17.06.2009 № 19-ЗРТ', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (2, 'usn_income', '16', 6, 'Базовая ставка УСН Доходы в Татарстане', 'НК РФ ст. 346.20', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (3, 'usn_income_expense', '77', 10, 'Льготная ставка в Москве для обрабатывающего производства, научных исследований, спорта и соцуслуг', 'Закон города Москвы от 07.10.2009 № 41', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (4, 'usn_income_expense', '77', 15, 'Базовая ставка УСН Доходы минус расходы в Москве для торговли и услуг', 'НК РФ ст. 346.20', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (5, 'usn_income_expense', '78', 7, 'Пониженная ставка в Санкт-Петербурге для всех налогоплательщиков', 'Закон Санкт-Петербурга от 05.05.2009 № 185-36', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (6, 'usn_income_expense', '66', 7, 'Пониженная ставка в Свердловской области', 'Закон Свердловской области от 20.03.2006 № 31-ОЗ', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (7, 'usn_income_expense', '54', 10, 'Пониженная ставка в Новосибирской области', 'Закон Новосибирской области от 16.10.2003 № 142-ОЗ', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

INSERT INTO tax_regional_rates (id, tax_regime_code, region_code, rate_percent, conditions_description, legal_act_title, effective_from, effective_to)
VALUES (8, 'usn_income_expense', '23', 15, 'Базовая ставка в Краснодарском крае (льгота 5% для аккредитованных IT-компаний)', 'Закон Краснодарского края от 11.06.2008 № 1494-КЗ', '2020-01-01', NULL)
ON CONFLICT (id) DO UPDATE SET rate_percent = EXCLUDED.rate_percent, conditions_description = EXCLUDED.conditions_description;

SELECT setval('tax_regional_rates_id_seq', (SELECT MAX(id) FROM tax_regional_rates));

-- 3. Регламенты вывесок по городам (Казань, Москва, СПб)
INSERT INTO signboard_regulations (id, region_code, city, authority_name, legal_act_title, max_sign_height_meters, max_sign_width_percent_of_facade, requires_special_permit, required_documents, processing_time_days, official_portal_url, is_mocked_data)
VALUES (1, '16', 'Казань', 'Управление архитектуры и градостроительства ИК МО г. Казани', 'Решение Казанской городской Думы от 18.10.2006 № 21-13 (Правила благоустройства Казани)', 0.5, 70, FALSE, ARRAY['Фотофиксация фасада до монтажа', 'Дизайн-проект вывески', 'Документ на право пользования помещением']::text[], 15, 'https://kzn.ru/prinyatie-resheniy/uslugi/soglasovanie-vyvesok/', TRUE)
ON CONFLICT (id) DO UPDATE SET authority_name = EXCLUDED.authority_name, max_sign_height_meters = EXCLUDED.max_sign_height_meters;

INSERT INTO signboard_regulations (id, region_code, city, authority_name, legal_act_title, max_sign_height_meters, max_sign_width_percent_of_facade, requires_special_permit, required_documents, processing_time_days, official_portal_url, is_mocked_data)
VALUES (2, '77', 'Москва', 'Комитет по архитектуре и градостроительству г. Москвы (Москомархитектура)', 'Постановление Правительства Москвы от 25.12.2013 № 902-ПП', 0.5, 70, FALSE, ARRAY['Не требуются при соблюдении типовых параметров Правил размещения']::text[], 0, 'https://www.mos.ru/mka/function/napravleniya-deyatelnosti/razmeshchenie-vyvesok/', FALSE)
ON CONFLICT (id) DO UPDATE SET authority_name = EXCLUDED.authority_name, max_sign_height_meters = EXCLUDED.max_sign_height_meters;

INSERT INTO signboard_regulations (id, region_code, city, authority_name, legal_act_title, max_sign_height_meters, max_sign_width_percent_of_facade, requires_special_permit, required_documents, processing_time_days, official_portal_url, is_mocked_data)
VALUES (3, '78', 'Санкт-Петербург', 'Комитет по печати и взаимодействию со СМИ Санкт-Петербурга', 'Постановление Правительства Санкт-Петербурга от 20.09.2012 № 1002', 0.62, 80, TRUE, ARRAY['Заявление на выдачу разрешения', 'Лист согласования КГА', 'Согласие собственников МКД']::text[], 30, 'https://gu.spb.ru/187979/eservice/', TRUE)
ON CONFLICT (id) DO UPDATE SET authority_name = EXCLUDED.authority_name, max_sign_height_meters = EXCLUDED.max_sign_height_meters;

SELECT setval('signboard_regulations_id_seq', (SELECT MAX(id) FROM signboard_regulations));

-- 4. Шаблоны чек-листов и их пункты
INSERT INTO checklists (id, code, title, description, target_legal_form, target_okved_prefix)
VALUES (1, 'chk_self_employed_start', 'Быстрый старт самозанятого за 1 день', 'Пошаговый план запуска деятельности плательщика НПД (кондитеры, мастера, услуги, хендмейд).', 'self_employed', NULL)
ON CONFLICT (code) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description;

INSERT INTO checklists (id, code, title, description, target_legal_form, target_okved_prefix)
VALUES (2, 'chk_ip_online_shop', 'Запуск интернет-магазина одежды для ИП на УСН', 'Полный цикл запуска онлайн-продаж товаров легпрома с онлайн-кассой, маркировкой и доставкой в MAX.', 'ip', '47.')
ON CONFLICT (code) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description;

INSERT INTO checklists (id, code, title, description, target_legal_form, target_okved_prefix)
VALUES (3, 'chk_cafe_launch', 'Открытие кофейни/общепита для малого бизнеса', 'Чек-лист открытия точки общепита: вывеска, касса, Роспотребнадзор, сотрудники.', 'ooo', '56.')
ON CONFLICT (code) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description;

SELECT setval('checklists_id_seq', (SELECT MAX(id) FROM checklists));

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (1, 1, 1, 'Регистрация в приложении «Мой налог»', 'Скачать приложение ''Мой налог'', зарегистрироваться по паспорту за 5 минут, выбрать регион деятельности.', 10, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (2, 1, 3, 'Проверка каталога на отсутствие перепродажи', 'Убедиться, что в продажу выставлены только товары собственного изготовления (ст. 4 422-ФЗ).', 20, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (3, 1, 12, 'Размещение Политики конфиденциальности в MAX', 'Вставить ссылку на Политику в описании магазина перед сбором телефонов покупателей (152-ФЗ).', 30, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (4, 1, 4, 'Тестовая выдача чека в «Мой налог»', 'Сформировать чек на тестовую сумму, проверить генерацию QR-кода и отправку покупателю в MAX.', 40, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (5, 2, 8, 'Подача заявления Р21001 и уведомления на УСН 6%', 'Бесплатно подать документы через сервис ФНС или банк с использованием Госключа.', 10, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (6, 2, 9, 'Подключение облачной онлайн-кассы (54-ФЗ)', 'Арендовать облачную кассу для интернет-магазина, зарегистрировать в ФНС и подключить фискальный накопитель.', 20, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (7, 2, 13, 'Регистрация в системе маркировки «Честный ЗНАК»', 'Оформить КЭП, зарегистрировать личный кабинет Честный ЗНАК по товарной группе ''Одежда и обувь''.', 30, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (8, 2, 10, 'Размещение публичной оферты и правил возврата 7 дней', 'Опубликовать в магазине MAX правила возврата товаров надлежащего качества по ст. 26.1 ЗоЗПП.', 40, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (9, 2, 6, 'Настройка налогового календаря и вычета взносов', 'Запланировать ежеквартальные авансовые платежи УСН с уменьшением налога на взносы ИП за себя.', 50, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (10, 3, 5, 'Регистрация ООО и выбор УСН «Доходы минус расходы»', 'Подготовить устав, решение учредителя, подать уведомление о применении УСН 15% (или 7% в Татарстане).', 10, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (11, 3, 14, 'Проверка фасадной вывески кофейни', 'Заказать настенную вывеску высотой до 0.5 м без перекрытия окон и мигающих элементов (ПП № 902-ПП).', 20, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (12, 3, 9, 'Установка торговой кассы с эквайрингом и СБП', 'Приобрести стационарную ККТ, подключить терминал эквайринга с комиссией до 0.7% по СБП.', 30, TRUE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

INSERT INTO checklist_items (id, checklist_id, requirement_id, title, description, sort_order, is_mandatory)
VALUES (13, 3, 16, 'Подача заявки на грант молодым предпринимателям', 'Если учредителю до 25 лет - получить до 500 тыс. руб. от Минэкономразвития на оборудование (кофемашина, мебель).', 40, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, sort_order = EXCLUDED.sort_order;

SELECT setval('checklist_items_id_seq', (SELECT MAX(id) FROM checklist_items));