"""Create a separate MAX-authenticated review copy; never overwrite source or target."""
import argparse
import hashlib
import json
import secrets
import shutil
from datetime import datetime, timedelta, timezone
from pathlib import Path

def prepare(source: Path, target: Path, days: int):
    source, target = source.resolve(), target.resolve()
    if target.exists() or source == target or source in target.parents:
        raise SystemExit("Use a new target outside the source directory; nothing was overwritten.")
    original = (source / "state.json").read_bytes()
    state = json.loads(original.decode("utf-8-sig"))
    if not state.get("teams") or not state.get("documents"):
        raise SystemExit("Source must contain a populated demo company.")
    files = {}
    for doc in state["documents"]:
        for version in doc["versions"]:
            key = version["id"]
            if not key.isalnum():
                raise SystemExit("Unsafe version ID")
            file = source / "files" / key
            if not file.is_file():
                raise SystemExit("Missing source file: " + key)
            files[key] = file
    current = datetime.now(timezone.utc)
    stamp = current.isoformat(timespec="seconds").replace("+00:00", "Z")
    until = (current + timedelta(days=days)).isoformat(timespec="seconds").replace("+00:00", "Z")
    def synthetic(user):
        user["id"] = "sample:" + user["id"].removeprefix("sample:")
    for team in state["teams"]:
        team["reviewUntil"] = until
        team["description"] = "Тестовая компания MAX. Синтетические данные. " + team.get("description", "")
        for member in team["members"]:
            synthetic(member)
        for key in ("invites", "inviteCode", "inviteExpires"):
            team.pop(key, None)
    for doc in state["documents"]:
        synthetic(doc["owner"])
        for comment in doc.get("comments", []):
            synthetic(comment["author"])
    for message in state.get("messages", []):
        synthetic(message["author"])
    for key in ("accounts", "outbox", "integrations", "pollMarker", "demoSeedVersion"):
        state.pop(key, None)
    root = next((t for t in state["teams"] if not t.get("parentTeamId")), state["teams"][0])
    codes = {role: secrets.token_hex(12) for role in ("manager", "employee")}
    root["invites"] = [
        {"id": secrets.token_hex(12), "code": code, "role": role, "createdBy": "review-bootstrap",
         "created": stamp, "expires": until, "maxUses": 100,
         "workspaceTeamIds": [t["id"] for t in state["teams"]]}
        for role, code in codes.items()
    ]
    live = target / "live"
    (live / "files").mkdir(parents=True)
    for key, file in files.items():
        shutil.copy2(file, live / "files" / key)
    (live / "state.json").write_text(json.dumps(state, ensure_ascii=False, indent=2), encoding="utf-8")
    private = f"""# Доступ к тестовой компании MAX — приватная инструкция

Бот: https://max.ru/t68_hakaton_max_bot
Срок доступа всей тестовой компании: **{until} (UTC)**.
Команд: {len(state["teams"])}; каналов: {len(state["channels"])}; документов: {len(state["documents"])}.

Откройте бота и отправьте **одну** команду по выбранной роли:

- Руководитель: `/join {codes["manager"]}`
- Сотрудник: `/join {codes["employee"]}`

Каждый код многоразовый: до 100 разных аккаунтов. Его можно передать команде и жюри.
Код руководителя даёт управление всеми семью тестовыми командами; храните его среди участников проверки.
Повторный ввод тем же аккаунтом ничего не дублирует. Чтобы испытать обе роли одновременно, используйте два MAX-аккаунта.
После вступления нажмите «Документы». Если приложение уже открыто, закройте и откройте его снова.
Начните в «Орбита»: сотрудник загружает свой файл и отправляет на согласование, руководитель получает уведомление и принимает решение.
Для примеров 1С откройте «Отчёты»; для справочника нажмите вторую кнопку бота.
Вход не подменяет личность: код даёт права реальному MAX ID. После указанной даты доступ к копии компании прекращается.
Это не токен бота и не пароль сервера. Не размещайте этот файл в публичном репозитории.

Прямые ссылки после привязки мини-приложения:
- Руководитель: https://max.ru/t68_hakaton_max_bot?startapp=i_{codes["manager"]}
- Сотрудник: https://max.ru/t68_hakaton_max_bot?startapp=i_{codes["employee"]}

Сначала компания должна быть перенесена на сервер из папки `live`.
"""
    (target / "REVIEWER-ACCESS.private.md").write_text(private, encoding="utf-8")
    manifest = {"created": stamp, "expires": until, "teams": len(state["teams"]),
                "channels": len(state["channels"]), "documents": len(state["documents"]),
                "files": len(files), "source_state_sha256": hashlib.sha256(original).hexdigest()}
    (target / "manifest.json").write_text(json.dumps(manifest, indent=2), encoding="utf-8")
    assert (source / "state.json").read_bytes() == original
    print(json.dumps(manifest))
    print("Private instructions: " + str(target / "REVIEWER-ACCESS.private.md"))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, default=Path("data/demo"))
    parser.add_argument("--target", type=Path, required=True)
    parser.add_argument("--days", type=int, default=31)
    args = parser.parse_args()
    if not 1 <= args.days <= 31:
        parser.error("days must be 1..31")
    prepare(args.source, args.target, args.days)

