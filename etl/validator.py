"""
Модуль валидации качества данных (Data Quality & Integrity Validator).
Проверяет бизнес-правила и ограничения схемы.
Ошибки фиксируются и выводятся с подробным контекстом, не роняя пайплайн молча.
"""
from typing import List, Tuple, Dict, Any
from etl.logger import log_warning, log_error

class ValidationError:
    def __init__(self, entity: str, record_id: Any, field: str, message: str, invalid_value: Any = None):
        self.entity = entity
        self.record_id = record_id
        self.field = field
        self.message = message
        self.invalid_value = invalid_value

    def __str__(self):
        val_str = f" [Значение: '{self.invalid_value}']" if self.invalid_value is not None else ""
        return f"{self.entity} ID={self.record_id} -> Поле '{self.field}': {self.message}{val_str}"


def validate_legal_sources(items: List[Dict[str, Any]]) -> Tuple[List[Dict[str, Any]], List[ValidationError]]:
    valid = []
    errors = []
    allowed_types = {'federal_law', 'code_rf', 'gov_decree', 'municipal_act', 'decree', 'official_portal'}

    for idx, item in enumerate(items):
        item_id = item.get('id') or f"row_{idx}"
        errs = []

        if not item.get('id'):
            errs.append(ValidationError("legal_sources", item_id, "id", "Идентификатор обязателен"))
        if not item.get('name'):
            errs.append(ValidationError("legal_sources", item_id, "name", "Наименование источника обязательно"))
        if not item.get('official_url'):
            errs.append(ValidationError("legal_sources", item_id, "official_url", "URL первоисточника обязателен"))
        
        stype = item.get('source_type')
        if not stype or stype not in allowed_types:
            errs.append(ValidationError("legal_sources", item_id, "source_type", f"Недопустимый тип источника. Разрешено: {allowed_types}", stype))

        if errs:
            errors.extend(errs)
            for e in errs:
                log_warning(f"Ошибка валидации: {e}")
        else:
            valid.append(item)

    return valid, errors


def validate_business_requirements(items: List[Dict[str, Any]], known_source_ids: set) -> Tuple[List[Dict[str, Any]], List[ValidationError]]:
    valid = []
    errors = []
    allowed_categories = {'registration', 'taxes', 'kkt_receipts', 'consumer_rights', 'personal_data', 'labeling', 'signboards', 'state_support', 'sanitary'}

    for idx, item in enumerate(items):
        item_id = item.get('id') or f"row_{idx}"
        errs = []

        if item.get('id') is None:
            errs.append(ValidationError("business_requirements", item_id, "id", "ID требования обязателен"))
        
        source_id = item.get('source_id')
        if not source_id:
            errs.append(ValidationError("business_requirements", item_id, "source_id", "Ссылка на первоисточник обязательна (DQ-06)"))
        elif known_source_ids and source_id not in known_source_ids:
            errs.append(ValidationError("business_requirements", item_id, "source_id", f"Источник '{source_id}' не найден в legal_sources", source_id))

        category = item.get('category')
        if not category or category not in allowed_categories:
            errs.append(ValidationError("business_requirements", item_id, "category", f"Категория '{category}' недопустима. Разрешены: {allowed_categories}", category))

        if not item.get('title'):
            errs.append(ValidationError("business_requirements", item_id, "title", "Заголовок требования обязателен"))
        if not item.get('legal_citation'):
            errs.append(ValidationError("business_requirements", item_id, "legal_citation", "Норма права (статья НПА) обязательна"))

        if errs:
            errors.extend(errs)
            for e in errs:
                log_warning(f"Ошибка валидации: {e}")
        else:
            valid.append(item)

    return valid, errors


def validate_tax_regimes(items: List[Dict[str, Any]]) -> Tuple[List[Dict[str, Any]], List[ValidationError]]:
    valid = []
    errors = []

    for idx, item in enumerate(items):
        code = item.get('code') or f"row_{idx}"
        errs = []

        if not item.get('code'):
            errs.append(ValidationError("tax_regimes", code, "code", "Код режима обязателен"))
        if not item.get('name'):
            errs.append(ValidationError("tax_regimes", code, "name", "Название режима обязательно"))

        rate = item.get('base_tax_rate_percent')
        if rate is None or rate < 0 or rate > 100:
            errs.append(ValidationError("tax_regimes", code, "base_tax_rate_percent", "Ставка налога должна быть от 0 до 100%", rate))

        revenue_cap = item.get('max_annual_revenue')
        if revenue_cap is None or revenue_cap < 0:
            errs.append(ValidationError("tax_regimes", code, "max_annual_revenue", "Лимит выручки не может быть отрицательным", revenue_cap))

        if errs:
            errors.extend(errs)
            for e in errs:
                log_warning(f"Ошибка валидации: {e}")
        else:
            valid.append(item)

    return valid, errors


def validate_business_profiles(items: List[Dict[str, Any]]) -> Tuple[List[Dict[str, Any]], List[ValidationError]]:
    valid = []
    errors = []
    allowed_forms = {'self_employed', 'ip', 'ooo'}

    for idx, item in enumerate(items):
        item_id = item.get('id') or f"row_{idx}"
        errs = []

        legal_form = item.get('legal_form')
        if not legal_form or legal_form not in allowed_forms:
            errs.append(ValidationError("business_profiles", item_id, "legal_form", f"Недопустимая форма бизнеса. Разрешены: {allowed_forms}", legal_form))

        # DQ-01: Лимит выручки самозанятого (2.4 млн руб. в год)
        monthly_rev = float(item.get('monthly_revenue_est', 0.0))
        if legal_form == 'self_employed' and (monthly_rev * 12 > 2400000.0):
            errs.append(ValidationError(
                "business_profiles", item_id, "monthly_revenue_est",
                f"Нарушение правила DQ-01: годовой доход самозанятого ({monthly_rev * 12:,.2f} руб.) превышает лимит 2 400 000 руб.",
                monthly_rev
            ))

        # DQ-02: Запрет наемных работников для самозанятого
        has_employees = item.get('has_employees', False)
        emp_count = int(item.get('employee_count', 0))
        if legal_form == 'self_employed' and (has_employees or emp_count > 0):
            errs.append(ValidationError(
                "business_profiles", item_id, "employee_count",
                "Нарушение правила DQ-02: самозанятый не имеет права нанимать сотрудников по ТК РФ",
                emp_count
            ))

        if monthly_rev < 0:
            errs.append(ValidationError("business_profiles", item_id, "monthly_revenue_est", "Выручка не может быть отрицательной", monthly_rev))

        if errs:
            errors.extend(errs)
            for e in errs:
                log_warning(f"Ошибка валидации профиля бизнеса: {e}")
        else:
            valid.append(item)

    return valid, errors


def validate_products(items: List[Dict[str, Any]]) -> Tuple[List[Dict[str, Any]], List[ValidationError]]:
    valid = []
    errors = []

    for idx, item in enumerate(items):
        item_id = item.get('id') or f"row_{idx}"
        errs = []

        price = float(item.get('price', -1))
        if price < 0:
            errs.append(ValidationError("products", item_id, "price", "Цена товара не может быть отрицательной (DQ-03)", price))

        if not item.get('name'):
            errs.append(ValidationError("products", item_id, "name", "Наименование товара обязательно"))

        ptype = item.get('type', 'good')
        if ptype not in ('good', 'service'):
            errs.append(ValidationError("products", item_id, "type", "Тип позиции должен быть 'good' или 'service'", ptype))

        if errs:
            errors.extend(errs)
            for e in errs:
                log_warning(f"Ошибка валидации товара: {e}")
        else:
            valid.append(item)

    return valid, errors
