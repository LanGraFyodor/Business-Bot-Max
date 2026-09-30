"""Build a reproducible source-and-screenshots archive for hackathon submission."""
from pathlib import Path
import hashlib
import zipfile

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "output"
OUT.mkdir(exist_ok=True)
ARCHIVE = OUT / "Docspace-source.zip"
TOP_FILES = [
    ".dockerignore", ".env.example", ".gitignore", "Dockerfile",
    "README.md", "README-PRESENTATION.md", "DATA-API.yaml", "compose.yaml", "go.mod", "start.ps1",
    "web/index.html", "web/package.json", "web/package-lock.json",
    "web/tsconfig.json", "web/vite.config.ts", "web/e2e.mjs", "web/invites-e2e.mjs", "web/document-workflow-e2e.mjs", "web/office-e2e.mjs", "web/company-e2e.mjs", "web/hierarchy-e2e.mjs", "web/escalation-e2e.mjs", "web/report-workflow-e2e.mjs", "web/redesign-smoke.mjs", "web/guide-e2e.mjs", "web/capture-presentation.mjs",
]
SOURCE_DIRS = ["cmd", "internal", "web/src", "web/public", "docs", "certs", "next/Business-Bot-Max", "presentation/screenshots", "deploy"]
EXTRAS = ["scripts/max-setup.ps1", "scripts/start-live-bot.ps1", "scripts/package.py", "scripts/send-report.py", "scripts/sync-guide-data.py", "scripts/prepare-review.py"]
files = [ROOT / path for path in TOP_FILES + EXTRAS]
for directory in SOURCE_DIRS:
    files.extend(path for path in (ROOT / directory).rglob("*") if path.is_file())
files = sorted((path for path in set(files) if ".git" not in path.relative_to(ROOT).parts and (path.suffix.lower() != ".pdf" or "demo_assets" in path.relative_to(ROOT).parts) and path.name != ".env" and not path.name.endswith(".private.md")), key=lambda path: path.relative_to(ROOT).as_posix())
for path in files:
    if not path.is_file():
        raise SystemExit(f"Missing source file: {path}")
with zipfile.ZipFile(ARCHIVE, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
    for path in files:
        relative = path.relative_to(ROOT).as_posix()
        archive.write(path, relative)
digest = hashlib.sha256(ARCHIVE.read_bytes()).hexdigest()
(OUT / "Docspace-source.sha256").write_text(
    f"{digest}  {ARCHIVE.name}\n", encoding="utf-8"
)
print(f"{ARCHIVE.name}: {len(files)} files, {ARCHIVE.stat().st_size} bytes")
print(f"SHA256 {digest}")

