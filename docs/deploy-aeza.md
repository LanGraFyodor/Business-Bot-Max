# VPS: MAX Бизнес рядом с VPN

## Текущее размещение

- VPS: `213.108.21.198`, Ubuntu 24.04.
- Каталог: `/opt/max-business`.
- URL для формы MAX: **https://max-213-108-21-198.sslip.io:8443/apps**.
- Проверка: https://max-213-108-21-198.sslip.io:8443/api/health.
- Второй интерфейс: https://max-213-108-21-198.sslip.io:8443/guide.
- Проект Docker Compose: `maxbusiness`. Конфигурация из [deploy/compose.server.yaml](../deploy/compose.server.yaml), Caddy из [deploy/Caddyfile](../deploy/Caddyfile).

**VPN Xray занимает TCP 443. Этот порт, конфигурация Xray и служба x-ui не менялись.**
Приложение слушает `127.0.0.1:18080`. Caddy использует TCP 8443 для HTTPS и TCP 80 для подтверждения/продления сертификата; TLS-ALPN challenge на 443 отключён. Admin API Caddy слушает только `127.0.0.1:12019`. Контейнеры используют host network: правила публикации портов Docker не добавляются. Лимиты памяти по 256 МБ; CPU приложения 0.5, Caddy 0.25.

Сертификат доверенный, проверен без `-k`, текущий срок — до 28 декабря 2026 (UTC). Caddy автоматически продлевает его при работающем сервере, DNS и доступном порте 80. Домен sslip.io связан с IP; это сторонний бесплатный DNS, гарантии его доступности на месяц нет. На своей стороне оставьте VPS оплаченным и запущенным.

## Что отправить организаторам

Передайте `https://max-213-108-21-198.sslip.io:8443/apps`, имя `@t68_hakaton_max_bot` и ID `394154549`. Формат HTTPS технически корректен; поддержка порта 8443 конкретной формой и политикой MAX должна быть проверена организаторами. Если форма запрещает порт, понадобится отдельный IP/VPS на стандартном 443 либо стабильный внешний HTTPS-прокси. Переносить VPN с 443 ради этого без отдельного согласования не нужно.

После привязки URL запускается **одно зарегистрированное мини-приложение**. Оно выбирает два самостоятельных экрана по параметрам `docspace` и `guide`. Две кнопки отправляет наш бот через Bot API при `/start` или другом ответе; форма сама эти две кнопки не создаёт. Проверить нужно на телефоне и в веб-клиенте MAX.

## Данные и роли жюри

Приложение запущено с `DEMO_MODE=false`: идентификация по MAX, без переключателя демопрофилей. В `data/live` находится отдельная копия тестовой компании. Оригинал на ноутбуке `data/demo` не изменялся.

Порядок вступления и сценарий — [JURY.md](JURY.md). Два многоразовых кода записаны локально в `data/jury-september-2026/REVIEWER-ACCESS.private.md`; секреты не включаются в архив исходников. Они действуют для семи тестовых команд до 31 октября 2026, 00:20:37 МСК. Обычные одноразовые приглашения на 7 дней сохранены.

Не копируйте начальный `state.json` поверх работающего сервера: это уничтожит изменения проверяющих. Повторная подготовка — только в новый каталог:
```powershell
python scripts/prepare-review.py --source data/demo --target data/new-review-copy --days 31
```

## Эксплуатация сервера

Выполнять через SSH:
```bash
cd /opt/max-business
docker compose ps
docker compose logs --tail 50 docspace
docker compose logs --tail 50 https
curl --fail https://max-213-108-21-198.sslip.io:8443/api/health
```

Перезапустить только приложение: `docker compose restart docspace`.
Остановить сайт и бота: `docker compose down` в **этом каталоге**; VPN останется отдельной службой. Повторный запуск: `docker compose up -d`. Не используйте `down -v`: сертификаты находятся в постоянных томах.

В контейнер импортирован заранее собранный образ `max-business:20260930`, поэтому сервер с 1 CPU не тратит ресурсы на сборку. Обновление: локально собрать и сохранить образ, передать tar через SCP, на сервере выполнить `docker load -i max-business-image.tar`, затем `docker compose up -d --no-deps docspace`. Рабочие данные при обновлении образа сохраняются.

Консистентная резервная копия данных (краткая пауза только бота/приложения):
```bash
cd /opt/max-business
mkdir -p backups
chmod 700 backups
docker compose stop docspace
tar -czf backups/live-$(date +%Y%m%d-%H%M%S).tar.gz data/live
docker compose start docspace
```

Файл `.env` с токеном имеет режим 600, родительский каталог — 700. Не публикуйте его. На ноутбуке остановлен только прежний live-poller на 8084; демо на 8080 сохранено. Не запускайте второй poller с тем же токеном.

Источники: [MAX — подключение мини-приложения](https://dev.max.ru/docs/webapps/introduction), [sslip.io](https://sslip.io/), [Caddy Automatic HTTPS](https://caddyserver.com/docs/automatic-https), [HTTP-01 и отключение TLS-ALPN](https://caddyserver.com/docs/caddyfile/directives/tls).

