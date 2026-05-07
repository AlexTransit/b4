---
sidebar_position: 12
title: MTProto Proxy
---

# MTProto Proxy for Telegram

B4 includes a built-in Telegram MTProto proxy that disguises traffic as a regular HTTPS connection to a popular website.

## Two deployment scenarios

![mtproto](/img/mtproto/20260322135051.png)

### Scenario 1: B4 on a VPS abroad (simple)

B4 runs on a server outside the censored zone. Users from inside Russia connect Telegram directly to the VPS.

```text
Phone (Russia) ──────▶ B4 on VPS ──────▶ Telegram
                TSPU sees
            "HTTPS to google.com"
              (not blocked)
```

Setup takes 2 minutes. No extra software needed.

### Scenario 2: B4 on a router inside Russia (with relay)

B4 runs on a router or machine inside Russia. An additional VPS is required to forward traffic, because TSPU blocks all direct IP-level connections to Telegram servers.

```text
Phone ──────▶ B4 (router) ──────▶ VPS ──────▶ Telegram
       TSPU sees                 TSPU sees
   "HTTPS to google.com"      "traffic to VPS"
       (not blocked)            (not blocked)
```

The VPS only needs a simple TCP forwarder — no keys, no MTProto-specific software.

---

## Scenario 1: B4 on a VPS

### Step 1: Configure B4

In the B4 web UI → **Settings** → **General** → **MTProto Proxy**:

1. **Enable MTProto Proxy** — turn it on
2. **Port** — listen port (recommended: `443`)
3. **Fake SNI Domain** — domain to impersonate (e.g. `storage.googleapis.com`)
4. Click **Generate Secret**
5. Copy the **Secret** value
6. Save settings and restart B4

Leave **DC Relay** empty — B4 on the VPS reaches Telegram directly.

### Step 2: Configure Telegram

1. Open **Telegram** → **Settings** → **Data and Storage** → **Proxy**
2. Tap **Add Proxy**
3. Choose **MTProto**
4. Fill in:
   - **Server**: VPS IP or hostname
   - **Port**: the port from step 1
   - **Secret**: the copied secret
5. Tap **Done** and enable the proxy

![telegra](/img/mtproto/20260322135130.png)

---

## Scenario 2: B4 on a router inside Russia

### Step 1: Install socat on the VPS

On any VPS abroad, install socat:

```bash
apt install -y socat
```

B4 will generate the actual forwarding commands in the next step. There's no point hardcoding IPs in this guide — Telegram periodically changes DC addresses and adds new ones (for example, the media DC `203`).

### Step 2: Configure B4

In the B4 web UI → **Settings** → **General** → **MTProto Proxy**:

1. **Enable MTProto Proxy** — turn it on
2. **Port** — listen port (e.g. `7002`)
3. **Fake SNI Domain** — domain to impersonate (e.g. `storage.googleapis.com`)
4. **DC Relay** — VPS address with the base port (e.g. `my-vps.com:7007`)
5. Click **Generate Secret**
6. Copy the **Secret** value

### Step 3: Get the socat commands for the VPS

Click the **?** button next to the **DC Relay** field. The "DC Relay socat setup" dialog opens, listing the current Telegram DCs and ready-to-run `socat` commands for each one (including the media DC).

Click **Copy all**, switch to the VPS, and run the commands.

:::info Why the helper
The DC list is fetched live from `getProxyConfig` — Telegram's own published list. B4 computes the relay port as `base_port + |DC| - 1`. If Telegram adds a new DC or changes an IP, the helper shows the up-to-date commands without needing to update this guide.
:::

:::warning VPS firewall
Open every port the helper shows on the VPS firewall (the "Open these ports on the VPS firewall" line at the bottom of the dialog). Currently this is typically 6 ports: five for the main DCs (1–5) and one for the media DC `203`.
:::

:::tip
To auto-start `socat`, add the commands to `/etc/rc.local` or create a systemd service.
:::

### Step 4: Save B4

Once socat is running on the VPS, click **Save** in B4 and restart B4.

### Step 5: Configure Telegram

1. Open **Telegram** → **Settings** → **Data and Storage** → **Proxy**
2. Tap **Add Proxy**
3. Choose **MTProto**
4. Fill in:
   - **Server**: IP of the router/machine running B4
   - **Port**: the port from step 2
   - **Secret**: the copied secret
5. Tap **Done** and enable the proxy

---

## Choosing a fake SNI domain

The domain should be:

- popular in Russia
- not blocked
- critically important (so blocking it would break other services)

:::info
If someone connects to the B4 port without the correct secret, B4 transparently forwards them to the real site (the one configured in Fake SNI). A scanner sees an ordinary site, not a proxy.
:::

## Troubleshooting

### Telegram shows "Connecting…"

- Make sure `socat` is running on the VPS and the ports are reachable (Scenario 2)
- Double-check the VPS address in the DC Relay field
- B4 logs should show `MTProto fake-TLS handshake OK` and `MTProto relay` lines

### Wrong secret

In the logs: `HMAC verification failed`

The secret in Telegram doesn't match the one configured in B4.

### Clock skew

In the logs: `timestamp out of range`

The clocks on the device and the B4 machine disagree. Sync them (NTP).

### VPS unreachable

In the logs: `dial DC ... i/o timeout`

- VPS is off, or `socat` is not running
- VPS firewall blocks inbound connections on the required ports

### No response from Telegram

In the logs: `DC->client: 0 bytes`

- If DC Relay is **not set**: Telegram servers are blocked by IP. You need a VPS relay (Scenario 2).
- If DC Relay **is set**: `socat` is not running on the VPS, or the wrong port was specified.
