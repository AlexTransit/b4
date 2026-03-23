---
sidebar_position: 2
title: OpenWRT
---

# OpenWRT

## Требования

- OpenWRT 19.07 и выше
- Внешнее хранилище (USB или extroot) — рекомендуется, так как на внутренней памяти роутера может не хватить места

:::warning Место на диске
На роутерах с OpenWRT внутренняя память ограничена (overlay). Если доступно менее 2 МБ, установщик предупредит об этом. Рекомендуется использовать extroot или USB-накопитель.

Инструкция по настройке extroot: https://openwrt.org/docs/guide-user/additional-software/extroot_configuration
:::

## Установка

Подключитесь к роутеру по SSH и выполните:

```bash
curl -fsSL https://raw.githubusercontent.com/DanielLavrushin/b4/main/install.sh | sh
```

Если `curl` не установлен:

```bash
opkg update && opkg install curl ca-certificates
```

Или через `wget`:

```bash
wget -qO- https://raw.githubusercontent.com/DanielLavrushin/b4/main/install.sh | sh
```

:::info wget на OpenWRT
Стандартный `wget` в OpenWRT (BusyBox) не поддерживает HTTPS. Установите полную версию:
```bash
opkg update && opkg install wget-ssl ca-certificates
```
:::

## Управление сервисом

```bash
/etc/init.d/b4 start
/etc/init.d/b4 stop
/etc/init.d/b4 restart
/etc/init.d/b4 enable     # автозапуск при загрузке
```

## Пути

При наличии `/opt` (extroot/USB):

| Что | Где |
| --- | --- |
| Бинарник | `/opt/bin/b4` |
| Конфигурация | `/opt/etc/b4/b4.json` |

Без внешнего хранилища (fallback):

| Что | Где |
| --- | --- |
| Бинарник | `/usr/bin/b4` |
| Конфигурация | `/etc/b4/b4.json` |

## LuCI-приложение

Существует сторонний пакет [luci-app-b4](https://github.com/BugOldfag/luci-app-b4), который добавляет управление b4 в интерфейс LuCI. Проект находится в стадии alpha и покрывает часть функций. Основной веб-интерфейс b4 (порт 7000) по-прежнему доступен.

## Модули ядра

На современных версиях OpenWRT (с apk):

```bash
apk add kmod-nft-queue kmod-nft-nat kmod-nft-compat
```

На старых версиях (с opkg):

```bash
opkg install kmod-nfnetlink-queue kmod-ipt-nfqueue iptables-mod-nfqueue iptables-mod-conntrack-extra
```
