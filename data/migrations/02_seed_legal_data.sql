-- ============================================================================
-- Миграция 02_seed_legal_data.sql
-- Наполнение справочников legal_sources и business_requirements
-- ============================================================================

-- 1. Заполнение официальных источников НПА
INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_422_npd', 'Федеральный закон от 27.11.2018 N 422-ФЗ ''О специальном налоговом режиме Налог на профессиональный доход''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102488887', 'ФНС России / Президент РФ', 'federal_law', '2026-08-04', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('nk_rf_usn', 'Налоговый кодекс РФ, Глава 26.2 ''Упрощенная система налогообложения (УСН)''', 'https://www.nalog.gov.ru/rn77/taxation/taxes/usn/', 'ФНС России / Минфин РФ', 'code_rf', '2026-01-01', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('nk_rf_430', 'Налоговый кодекс РФ, Статья 430 ''Размер страховых взносов, уплачиваемых ИП за себя''', 'https://www.nalog.gov.ru/rn77/taxation/taxes/strah/', 'ФНС России / Социальный фонд России', 'code_rf', '2026-01-01', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_129_reg', 'Федеральный закон от 08.08.2001 N 129-ФЗ ''О государственной регистрации юридических лиц и ИП''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102072483', 'ФНС России', 'federal_law', '2026-08-04', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_54_kkt', 'Федеральный закон от 22.05.2003 N 54-ФЗ ''О применении ККТ при осуществлении расчетов в РФ''', 'https://www.nalog.gov.ru/rn77/related_activities/kkt/', 'ФНС России', 'federal_law', '2025-06-24', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('law_2300_1_zozpp', 'Закон РФ от 07.02.1992 N 2300-1 ''О защите прав потребителей''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102014512', 'Роспотребнадзор', 'federal_law', '2026-09-01', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('pp_2463_rules', 'Постановление Правительства РФ от 31.12.2020 N 2463 ''Правила продажи товаров по договору розничной купли-продажи''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102928373', 'Правительство РФ', 'gov_decree', '2026-05-30', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_152_pdn', 'Федеральный закон от 27.07.2006 N 152-ФЗ ''О персональных данных''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102108261', 'Роскомнадзор', 'federal_law', '2026-07-26', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('rasp_792_marking', 'Распоряжение Правительства РФ от 28.04.2018 N 792-р ''Перечень товаров, подлежащих обязательной маркировке Честный ЗНАК''', 'https://честныйзнак.рф', 'Правительство РФ / Минпромторг', 'gov_decree', '2026-04-27', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('pp_moscow_902_signs', 'Постановление Правительства Москвы от 25.12.2013 N 902-ПП ''О размещении информационных конструкций в городе Москве''', 'https://www.mos.ru/authority/documents/doc/31818220/', 'Правительство Москвы', 'municipal_act', '2026-01-01', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_176_tax_reform', 'Федеральный закон от 12.07.2024 N 176-ФЗ ''О налоговой реформе (НДС на УСН)''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=604085449', 'Госдума / Президент РФ', 'federal_law', '2025-11-28', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_248_control', 'Федеральный закон от 31.07.2020 N 248-ФЗ ''О государственном контроле и приложении Инспектор''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102766324', 'Минэкономразвития России', 'federal_law', '2026-08-04', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('minek_142_grants', 'Приказ Минэкономразвития России от 26.03.2021 N 142 ''Гранты молодым предпринимателям до 25 лет''', 'https://economy.gov.ru', 'Минэкономразвития России', 'decree', '2024-05-20', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

INSERT INTO legal_sources (id, name, official_url, publisher, source_type, last_verified_at, status)
VALUES ('fz_178_social_contract', 'Федеральный закон от 17.07.1999 N 178-ФЗ ''О государственной социальной помощи (Соцконтракт)''', 'http://pravo.gov.ru/proxy/ips/?docbody=&nd=102060601', 'Минтруд России', 'federal_law', '2026-07-26', 'active')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, official_url = EXCLUDED.official_url, last_verified_at = EXCLUDED.last_verified_at;

-- 2. Заполнение каталога нормативных требований к бизнесу
INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (1, 'fz_422_npd', 'registration', 'Запрет на привлечение работников по трудовым договорам для самозанятых', 'Плательщик НПД не имеет права заключать трудовые договоры. При необходимости привлекаются исполнители по договорам ГПХ.', 'п. 4 ч. 2 ст. 4 Федерального закона № 422-ФЗ', ARRAY['self_employed']::text[], NULL, NULL, TRUE, 'При найме сотрудника по трудовому договору статус самозанятого аннулируется задним числом с доначислением НДФЛ (13%) и страховых взносов (30%) со штрафом 20-40%.', '["Проверить, требуются ли постоянные сотрудники в штат.","Если требуется штат - зарегистрировать статус ИП и выбрать УСН или АУСН.","Для разовых задач оформлять договоры ГПХ с другими самозанятыми."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (2, 'fz_422_npd', 'taxes', 'Соблюдение годового лимита выручки самозанятого (2.4 млн руб.)', 'Доход самозанятого нарастающим итогом с 1 января по 31 декабря не должен превышать 2 400 000 рублей.', 'п. 8 ч. 2 ст. 4 Федерального закона № 422-ФЗ', ARRAY['self_employed']::text[], NULL, NULL, TRUE, 'С суммы дохода сверх 2.4 млн рублей налог рассчитывается по ставке 13% НДФЛ для физлиц, либо ИП переводится на ОСНО с НДС.', '["Вести регулярный учет выручки в приложении Мой налог.","При приближении к порогу 2 млн руб. заранее подготовить документы на открытие ИП на УСН 6%."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (3, 'fz_422_npd', 'registration', 'Запрет перепродажи чужих товаров для самозанятых', 'Самозанятым разрешено продавать только товары собственного производства (хендмейд, выпечка, пошив, декор). Перепродажа чужих товаров запрещена.', 'п. 2 ч. 2 ст. 4 Федерального закона № 422-ФЗ', ARRAY['self_employed']::text[], ARRAY['47.']::text[], NULL, TRUE, 'Лишение статуса плательщика НПД и доначисление налогов по общей системе налогообложения.', '["Убедиться, что в каталоге магазина представлены только товары ручной работы или собственного изготовления.","Если планируется перепродажа фабричных товаров или оптовых закупок - зарегистрировать ИП на УСН."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (4, 'fz_422_npd', 'kkt_receipts', 'Формирование и отправка чека покупателю через приложение «Мой налог»', 'Самозанятый обязан выбивать чек на каждую продажу в приложении Мой налог и отправлять клиенту ссылку/QR-код.', 'ст. 12 Федерального закона № 422-ФЗ', ARRAY['self_employed']::text[], NULL, NULL, TRUE, 'Штраф 20% от суммы расчета за первое нарушение, 100% от суммы расчета при повторном нарушении в течение 6 месяцев (ст. 129.13 НК РФ).', '["После получения оплаты от покупателя в MAX зайти в приложение Мой налог.","Ввести наименование услуги/товара и сумму, нажать \u0027Выдать чек\u0027.","Скопировать ссылку на чек и отправить покупателю в чат MAX."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (5, 'nk_rf_usn', 'taxes', 'Своевременная подача уведомления о переходе на УСН', 'Уведомление о переходе на УСН подается в течение 30 календарных дней с даты регистрации ИП или ООО.', 'п. 2 ст. 346.13 Налогового кодекса РФ', ARRAY['ip', 'ooo']::text[], NULL, NULL, TRUE, 'Если не подать уведомление в 30-дневный срок, бизнес автоматически оказывается на общей системе (ОСНО) с обязанностью уплаты НДС 20% и налога на прибыль/НДФЛ.', '["Приложить форму уведомления 26.2-1 к пакету документов на регистрацию в налоговую.","Получить в налоговой отметку о принятии уведомления и сохранить ее."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (6, 'nk_rf_usn', 'taxes', 'Уменьшение налога УСН «Доходы» на сумму страховых взносов', 'ИП без сотрудников вправе уменьшить авансовый платеж и итоговый налог УСН 6% на 100% суммы фиксированных взносов за себя.', 'п. 3.1 ст. 346.21 Налогового кодекса РФ', ARRAY['ip']::text[], NULL, NULL, FALSE, 'Переплата налога в бюджет при незнании права на налоговый вычет.', '["Внести данные об уплаченных взносах в калькулятор налогов.","Рассчитать итоговый налог УСН с учетом 100% вычета (налог = начисленный налог минус уплаченные взносы)."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (7, 'nk_rf_430', 'taxes', 'Ежегодная уплата фиксированных страховых взносов ИП за себя', 'ИП обязан уплатить фиксированные взносы на ОПС и ОМС в срок до 31 декабря текущего года (размер в 2026 г. ~53 658 руб.).', 'п. 1.2 ст. 430 Налогового кодекса РФ', ARRAY['ip']::text[], NULL, NULL, TRUE, 'Начисление пеней за каждый день просрочки (1/300 ключевой ставки ЦБ РФ) и блокировка расчетного счета.', '["Разбить сумму фиксированных взносов на 4 квартала для равномерной нагрузки (по ~13 415 руб.).","Оплачивать через Единый налоговый счет (ЕНС) в личном кабинете ИП на сайте nalog.gov.ru."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (8, 'fz_129_reg', 'registration', 'Бесплатная электронная регистрация ИП без уплаты госпошлины', 'При направлении документов на регистрацию ИП в электронном виде (через сайт ФНС, Госуслуги или уполномоченный банк) госпошлина составляет 0 рублей.', 'ст. 9 Закона № 129-ФЗ, подп. 32 п. 3 ст. 333.35 НК РФ', ARRAY['ip']::text[], NULL, NULL, FALSE, 'Экономия 800 рублей и времени на визит в регистрирующую налоговую инспекцию.', '["Сформировать заявление по форме Р21001 через сервис ФНС \u0027Государственная онлайн-регистрация бизнеса\u0027.","Подписать заявление усиленной КЭП физлица (включая приложение \u0027Госключ\u0027) и отправить в ФНС."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (9, 'fz_54_kkt', 'kkt_receipts', 'Применение облачной онлайн-кассы для интернет-магазина', 'При приеме платежей от физлиц в сети Интернет продавец (ИП или ООО) обязан применять онлайн-кассу и формировать электронный чек.', 'п. 2 ст. 1.2, п. 1 ст. 4.3 Федерального закона № 54-ФЗ', ARRAY['ip', 'ooo']::text[], ARRAY['47.']::text[], NULL, TRUE, 'Штраф для ИП от 25% до 50% от суммы выручки мимо кассы (минимум 10 000 руб.), для ООО - от 75% до 100% (минимум 30 000 руб.). При повторном нарушении - приостановление деятельности до 90 суток.', '["Заключить договор с сервисом облачной фискализации (Атол Онлайн, МодульКасса, ЮКасса).","Купить или взять в аренду фискальный накопитель (ФН на 15 или 36 месяцев).","Зарегистрировать кассу в ФНС через личный кабинет налогоплательщика."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (10, 'law_2300_1_zozpp', 'consumer_rights', 'Обеспечение права покупателя на возврат товара в 7-дневный срок', 'В интернет-магазине потребитель вправе отказаться от товара в любое время до доставки, а после передачи товара - в течение 7 дней.', 'п. 4 ст. 26.1 Закона РФ № 2300-1 «О защите прав потребителей»', ARRAY['self_employed', 'ip', 'ooo']::text[], ARRAY['47.']::text[], NULL, TRUE, 'Если информация о 7 днях не предоставлена письменно в момент доставки, срок возврата автоматически увеличивается до 3 месяцев! Неустойка 1% в день за каждый день задержки возврата денег.', '["Разместить в карточке магазина в MAX раздел \u0027Условия возврата\u0027.","Вкладывать в посылку печатную памятку с правилами возврата товара в течение 7 дней.","Возвращать деньги покупателю в течение 10 дней с момента предъявления требования."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (11, 'pp_2463_rules', 'consumer_rights', 'Размещение перечня невозвратных товаров в каталоге магазина', 'Продавец вправе отказать в возврате качественных товаров, входящих в перечень ПП № 2463 (косметика, белье, гигиена, ювелирные изделия).', 'Постановление Правительства РФ от 31.12.2020 № 2463', ARRAY['ip', 'ooo']::text[], ARRAY['47.']::text[], NULL, FALSE, 'Необоснованные финансовые потери продавца при возврате вскрытой косметики или белья клиентами.', '["Маркировать в каталоге товары категорий \u0027Косметика\u0027, \u0027Белье\u0027, \u0027Парфюмерия\u0027 специальным тегом \u0027Не подлежит возврату\u0027.","Информировать покупателя в чат-боте перед подтверждением заказа."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (12, 'fz_152_pdn', 'personal_data', 'Размещение Политики обработки персональных данных в витрине MAX', 'Любая интернет-витрина, запрашивающая имя, телефон или адрес покупателя, обязана обеспечить открытый доступ к Политике конфиденциальности.', 'ч. 2 ст. 18.1 Федерального закона № 152-ФЗ', ARRAY['self_employed', 'ip', 'ooo']::text[], NULL, NULL, TRUE, 'Административный штраф по ч. 3 ст. 13.11 КоАП РФ для ИП от 10 000 до 20 000 руб., для юрлиц - от 30 000 до 60 000 руб.', '["Сгенерировать типовую Политику конфиденциальности с помощью агента-помощника.","Разместить ссылку на Политику в описании магазина и прикрепить чекбокс согласия в корзине заказа."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (13, 'rasp_792_marking', 'labeling', 'Обязательная маркировка одежды и обуви в системе «Честный ЗНАК»', 'Продажа предметов одежды и обуви требует наличия на каждой единице кода Data Matrix и сканирования при продаже.', 'Распоряжение Правительства РФ № 792-р', ARRAY['ip', 'ooo']::text[], ARRAY['47.71', '47.72']::text[], NULL, TRUE, 'Конфискация всей партии немаркированного товара и штраф по ст. 15.12 КоАП РФ до 300 000 рублей.', '["Зарегистрироваться в системе Честный ЗНАК с помощью усиленной ЭЦП.","Подключить оператора электронного документооборота (ЭДО).","Настроить передачу кодов маркировки через кассовый аппарат при формировании чека."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (14, 'pp_moscow_902_signs', 'signboards', 'Соблюдение нормативов габаритов настенных вывесок (высота до 0.5 м)', 'Информационная настенная вывеска на фасаде не требует получения разрешения, если ее высота не превышает 0.50 м, а длина - не более 70% фасада.', 'п. 18.1 Правил размещения информационных конструкций (ПП Москвы № 902-ПП)', ARRAY['ip', 'ooo']::text[], ARRAY['47.', '56.', '96.']::text[], '77', TRUE, 'Штраф ОАТИ от 30 000 до 50 000 руб. для ИП и до 500 000 руб. для юрлиц с принудительным демонтажем конструкции за счет владельца.', '["Измерить длину фасада занимаемого помещения.","Заказать изготовление объемных букв на подложке высотой до 50 см.","Разместить вывеску на единой горизонтальной оси с соседними вывесками здания."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (15, 'fz_176_tax_reform', 'taxes', 'Контроль порога 60 млн руб. для освобождения от уплаты НДС на УСН', 'Бизнес на УСН с доходами до 60 млн руб. в год автоматически освобожден от НДС. При превышении порога возникает обязанность уплаты НДС по ставкам 5% или 20%.', 'ст. 145 НК РФ (в ред. Федерального закона № 176-ФЗ)', ARRAY['ip', 'ooo']::text[], NULL, NULL, TRUE, 'При превышении лимита и невыставлении счетов-фактур - блокировка счетов и доначисление НДС со штрафом 20%.', '["Настроить в аналитике мониторинг приближения к годовому обороту 50 млн руб.","При превышении 60 млн руб. выбрать ставку НДС: 5% (без права на налоговые вычеты) или 20% (с вычетами)."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (16, 'minek_142_grants', 'state_support', 'Получение безвозмездного гранта до 500 тыс. руб. для предпринимателей до 25 лет', 'Индивидуальные предприниматели и учредители ООО в возрасте до 25 лет включительно имеют право на безвозвратную субсидию от 100 до 500 тысяч рублей.', 'Приказ Минэкономразвития России № 142', ARRAY['ip', 'ooo']::text[], NULL, NULL, FALSE, 'Возврат гранта государству в случае нецелевого расходования средств (запрещено снимать наличные).', '["Пройти бесплатное обучение в региональном центре \u0027Мой бизнес\u0027 (сертификат действует 1 год).","Подготовить бизнес-план проекта и предоставить подтверждение 25% софинансирования.","Подать заявку на конкурс грантов через региональный портал госуслуг или МФЦ."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

INSERT INTO business_requirements (id, source_id, category, title, summary, legal_citation, applies_to_legal_forms, applies_to_okved_prefixes, region_code, is_mandatory, penalty_risk_description, action_plan, is_simulated_in_mvp)
VALUES (17, 'fz_178_social_contract', 'state_support', 'Единовременная выплата до 350 тыс. руб. по соцконтракту на открытие дела', 'Граждане со среднедушевым доходом ниже прожиточного минимума могут получить от государства до 350 000 руб. на приобретение оборудования при регистрации самозанятости или ИП.', 'ст. 8.1 Федерального закона № 178-ФЗ', ARRAY['self_employed', 'ip']::text[], NULL, NULL, FALSE, 'Возврат средств при досрочном прекращении статуса самозанятого/ИП в течение первого года работы.', '["Подать заявление в орган соцзащиты или МФЦ по месту жительства.","Защитить бизнес-план и смету расходов на комиссии.","Предоставить чеки и договоры на покупку оборудования."]'::jsonb, FALSE)
ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, summary = EXCLUDED.summary, legal_citation = EXCLUDED.legal_citation, action_plan = EXCLUDED.action_plan;

-- Синхронизация последовательности ID
SELECT setval('business_requirements_id_seq', (SELECT MAX(id) FROM business_requirements));