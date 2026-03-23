---
sidebar_position: 4
title: Keenetic
---

# Keenetic

## Требования

- Роутер Keenetic с поддержкой OPKG
- Установленный Entware (обязательно)

## Установка Entware

### Новые модели (со встроенным хранилищем)

1. Откройте веб-интерфейс роутера
2. Перейдите в **Параметры системы**
3. Включите компонент **Менеджер пакетов OPKG**

### Старые модели (нужен USB-накопитель)

1. Вставьте USB-накопитель в роутер
2. Установите Entware через менеджер пакетов

Подробнее: https://help.keenetic.com/hc/ru/articles/360021214160

## Установка b4

Подключитесь по SSH и выполните:

```bash
curl -fsSL https://raw.githubusercontent.com/DanielLavrushin/b4/main/install.sh | sh
```

## Управление сервисом

```bash
/opt/etc/init.d/S99b4 start
/opt/etc/init.d/S99b4 stop
/opt/etc/init.d/S99b4 restart
```

## Пути

| Что | Где |
| --- | --- |
| Бинарник | `/opt/sbin/b4` |
| Конфигурация | `/opt/etc/b4/b4.json` |
| Сервис | `/opt/etc/init.d/S99b4` |

## Архитектура

- Старые модели (MT7621) — `mipsle_softfloat`
- Новые модели (aarch64) — `arm64`

Установщик определяет архитектуру автоматически.

:::warning Без Entware
Без Entware b4 устанавливается в `/tmp`, который очищается при каждой перезагрузке. Для постоянной работы Entware обязателен.
:::
