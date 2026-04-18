import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Box } from "@mui/material";
import { ConnectionIcon } from "@b4.icons";
import {
  B4FormGroup,
  B4Section,
  B4Switch,
  B4TextField,
  B4Alert,
} from "@b4.elements";
import { B4Config } from "@models/config";

interface MTProtoSettingsProps {
  config: B4Config;
  onChange: (
    field: string,
    value: number | boolean | string | string[],
  ) => void;
}

export const MTProtoSettings = ({ config, onChange }: MTProtoSettingsProps) => {
  const { t } = useTranslation();
  const [generating, setGenerating] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [refreshResult, setRefreshResult] = useState<
    | { ok: true; count: number; dcs: Record<string, string> }
    | { ok: false; error: string }
    | null
  >(null);

  const handleRefreshDCs = async () => {
    setRefreshing(true);
    setRefreshResult(null);
    try {
      const res = await fetch("/api/mtproto/refresh-dcs", { method: "POST" });
      const data = (await res.json()) as {
        success: boolean;
        count?: number;
        dcs?: Record<string, string>;
        error?: string;
      };
      if (data.success && typeof data.count === "number" && data.dcs) {
        setRefreshResult({ ok: true, count: data.count, dcs: data.dcs });
      } else {
        setRefreshResult({ ok: false, error: data.error || "unknown error" });
      }
    } catch (e) {
      setRefreshResult({ ok: false, error: String(e) });
    } finally {
      setRefreshing(false);
    }
  };

  const handleGenerateSecret = async () => {
    const sni = config.system.mtproto?.fake_sni || "storage.googleapis.com";
    setGenerating(true);
    try {
      const res = await fetch("/api/mtproto/generate-secret", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ fake_sni: sni }),
      });
      const data = (await res.json()) as { success: boolean; secret?: string };
      if (data.success && data.secret) {
        onChange("system.mtproto.secret", data.secret);
      }
    } finally {
      setGenerating(false);
    }
  };

  return (
    <B4Section
      title={t("settings.MTProto.title")}
      description={t("settings.MTProto.description")}
      icon={<ConnectionIcon />}
    >
      <B4FormGroup label={t("settings.MTProto.settings")} columns={2}>
        <B4Switch
          label={t("settings.MTProto.enable")}
          checked={config.system.mtproto?.enabled ?? false}
          onChange={(checked: boolean) =>
            onChange("system.mtproto.enabled", checked)
          }
          description={t("settings.MTProto.enableDesc")}
        />
        <B4TextField
          label={t("settings.MTProto.bindAddress")}
          value={config.system.mtproto?.bind_address || "0.0.0.0"}
          onChange={(e) =>
            onChange("system.mtproto.bind_address", e.target.value)
          }
          placeholder={t("settings.MTProto.bindAddressPlaceholder")}
          disabled={!config.system.mtproto?.enabled}
          helperText={t("settings.MTProto.bindAddressHelp")}
        />
        <B4TextField
          label={t("settings.MTProto.port")}
          type="number"
          value={config.system.mtproto?.port ?? 3128}
          onChange={(e) =>
            onChange("system.mtproto.port", Number(e.target.value))
          }
          disabled={!config.system.mtproto?.enabled}
          helperText={t("settings.MTProto.portHelp")}
        />
        <B4TextField
          label={t("settings.MTProto.fakeSNI")}
          value={config.system.mtproto?.fake_sni || "storage.googleapis.com"}
          onChange={(e) => onChange("system.mtproto.fake_sni", e.target.value)}
          disabled={!config.system.mtproto?.enabled}
          helperText={t("settings.MTProto.fakeSNIHelp")}
        />
        <B4TextField
          label={t("settings.MTProto.dcRelay")}
          value={config.system.mtproto?.dc_relay || ""}
          onChange={(e) => onChange("system.mtproto.dc_relay", e.target.value)}
          placeholder="vps-ip:7007"
          disabled={!config.system.mtproto?.enabled}
          helperText={t("settings.MTProto.dcRelayHelp")}
        />
        {config.system.mtproto?.enabled && config.system.mtproto?.dc_relay && (
          <B4Alert severity="info">
            <span
              dangerouslySetInnerHTML={{ __html: t("settings.MTProto.relaySetup") }}
            />
          </B4Alert>
        )}
        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <B4TextField
            label={t("settings.MTProto.secret")}
            value={config.system.mtproto?.secret || ""}
            onChange={(e) => onChange("system.mtproto.secret", e.target.value)}
            disabled={!config.system.mtproto?.enabled}
            helperText={t("settings.MTProto.secretHelp")}
            autoComplete="off"
          />
          <Button
            variant="outlined"
            size="small"
            onClick={() => void handleGenerateSecret()}
            disabled={!config.system.mtproto?.enabled || generating}
          >
            {t("settings.MTProto.generateSecret")}
          </Button>
        </Box>
        {config.system.mtproto?.enabled && (
          <B4Alert severity="info">{t("settings.MTProto.restartNote")}</B4Alert>
        )}
        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <Button
            variant="outlined"
            size="small"
            onClick={() => void handleRefreshDCs()}
            disabled={refreshing}
          >
            {refreshing
              ? t("settings.MTProto.refreshingDCs")
              : t("settings.MTProto.refreshDCs")}
          </Button>
          {refreshResult?.ok && (
            <B4Alert severity="success">
              {t("settings.MTProto.refreshDCsOk", { count: refreshResult.count })}
              <Box
                component="ul"
                sx={{ m: 0, pl: 2, fontFamily: "monospace", fontSize: "0.8rem" }}
              >
                {Object.entries(refreshResult.dcs)
                  .sort((a, b) => Number(a[0]) - Number(b[0]))
                  .map(([id, addr]) => (
                    <li key={id}>
                      DC{id} → {addr}
                    </li>
                  ))}
              </Box>
            </B4Alert>
          )}
          {refreshResult && !refreshResult.ok && (
            <B4Alert severity="error">
              {t("settings.MTProto.refreshDCsErr", { error: refreshResult.error })}
            </B4Alert>
          )}
        </Box>
      </B4FormGroup>
    </B4Section>
  );
};
