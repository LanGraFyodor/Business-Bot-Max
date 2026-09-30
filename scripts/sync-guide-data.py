"""Export only public reference material and synthetic examples; never accounts or orders.
Original files in next/Business-Bot-Max remain untouched.
"""
from pathlib import Path
import hashlib
import json
import re

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "next/Business-Bot-Max/data"
TARGET = ROOT / "web/src/guide/data.json"
PUBLIC = ROOT / "web/public/guide-data/laws"
REFERENCE = ["legal_sources", "legal_fragments", "business_requirements", "checklists", "checklist_items", "legal_forms", "tax_regimes", "regions", "okved"]

def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))

def fields(rows, names):
    return [{name: row.get(name) for name in names.split()} for row in rows]

def main():
    data = {name: read(SOURCE / "reference" / (name + ".json")) for name in REFERENCE}
    seeds = read(SOURCE / "seeds/test_dataset.json")
    data["business_profiles"] = fields(seeds["business_profiles"], "id business_name legal_form city okved_code employee_count monthly_revenue_est monthly_expenses_est")
    data["stores"] = fields(seeds["stores"], "id business_profile_id title description delivery_enabled delivery_terms delivery_fee")
    data["products"] = fields(seeds["products"], "id store_id name description price unit is_in_stock requires_marking")
    data["financial_health_assessments"] = fields(seeds["financial_health_assessments"], "business_profile_id net_monthly_profit monthly_revenue monthly_variable_costs monthly_fixed_costs break_even_point safety_margin_percent cash_gap_risk_level")
    PUBLIC.mkdir(parents=True, exist_ok=True)
    for source in data["legal_sources"]:
        sid = source["id"]
        if not re.fullmatch(r"[a-z0-9_]+", sid):
            raise ValueError("Unsafe source ID")
        original = SOURCE / "raw/laws" / sid / "original.txt"
        if original.exists():
            blob = original.read_bytes()
            (PUBLIC / (sid + ".txt")).write_bytes(blob)
            source["local_text_url"] = "/guide-data/laws/" + sid + ".txt"
            source["sha256"] = hashlib.sha256(blob).hexdigest()
    data["provenance"] = {"source": "next/Business-Bot-Max/data", "mode": "reference-snapshot", "examples": "synthetic", "legalReview": "not-verified-by-frontend"}
    TARGET.parent.mkdir(parents=True, exist_ok=True)
    TARGET.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print("Guide exported: %s sources, %s requirements, %s demo businesses" % (len(data["legal_sources"]), len(data["business_requirements"]), len(data["business_profiles"])))

if __name__ == "__main__":
    main()
