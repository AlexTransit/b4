import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Box,
  IconButton,
  InputAdornment,
  Tooltip,
  Typography,
  Chip,
  Collapse,
  Stack,
  Grid,
} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import CheckIcon from "@mui/icons-material/Check";
import IosShareIcon from "@mui/icons-material/IosShare";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import AutorenewIcon from "@mui/icons-material/Autorenew";
import RefreshIcon from "@mui/icons-material/Refresh";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import { QRCodeSVG } from "qrcode.react";
import { ConnectionIcon } from "@b4.icons";
import {
  B4FormGroup,
  B4Section,
  B4Switch,
  B4TextField,
  B4Alert,
  B4Dialog,
} from "@b4.elements";
import { copyText } from "@utils";
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
  const [shareOpen, setShareOpen] = useState(false);
  const [shareHost, setShareHost] = useState("");
  const [copied, setCopied] = useState(false);
  const [showDcs, setShowDcs] = useState(false);

  const port = config.system.mtproto?.port ?? 3128;
  const secret = config.system.mtproto?.secret || "";
  const shareLink = useMemo(() => {
    const host = (shareHost || "").trim();
    if (!host || !secret) return "";
    return `tg://proxy?server=${encodeURIComponent(host)}&port=${port}&secret=${encodeURIComponent(secret)}`;
  }, [shareHost, port, secret]);
  const canShare =
    typeof navigator !== "undefined" && typeof navigator.share === "function";

  const openShare = () => {
    const bind = config.system.mtproto?.bind_address || "";
    const isAnyAddr = !bind || bind === "0.0.0.0" || bind === "::";
    setShareHost(isAnyAddr ? globalThis.location.hostname : bind);
    setCopied(false);
    setShareOpen(true);
  };

  const handleCopy = async () => {
    if (!shareLink) return;
    if (await copyText(shareLink)) {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    }
  };

  const handleNativeShare = async () => {
    if (!shareLink || !canShare) return;
    try {
      await navigator.share({
        title: t("settings.MTProto.title"),
        url: shareLink,
      });
    } catch {
      /* user cancelled */
    }
  };

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
        setShowDcs(true);
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

  let dcAlertSeverity: "success" | "error" | "info" = "info";
  let dcAlertText: string = t("settings.MTProto.refreshDCsHint");
  if (refreshResult?.ok) {
    dcAlertSeverity = "success";
    dcAlertText = t("settings.MTProto.refreshDCsOk", {
      count: refreshResult.count,
    });
  } else if (refreshResult && !refreshResult.ok) {
    dcAlertSeverity = "error";
    dcAlertText = t("settings.MTProto.refreshDCsErr", {
      error: refreshResult.error,
    });
  }

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
        {config.system.mtproto?.enabled && (
          <B4Alert severity="warning">
            {t("settings.MTProto.restartNote")}
          </B4Alert>
        )}
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
        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <B4TextField
            label={t("settings.MTProto.secret")}
            value={config.system.mtproto?.secret || ""}
            onChange={(e) => onChange("system.mtproto.secret", e.target.value)}
            disabled={!config.system.mtproto?.enabled}
            helperText={t("settings.MTProto.secretHelp")}
            autoComplete="off"
            slotProps={{
              input: {
                endAdornment: (
                  <InputAdornment position="end">
                    <Chip
                      size="small"
                      icon={
                        <AutorenewIcon
                          sx={{
                            animation: generating
                              ? "spin 1s linear infinite"
                              : "none",
                            "@keyframes spin": {
                              from: { transform: "rotate(0deg)" },
                              to: { transform: "rotate(360deg)" },
                            },
                          }}
                        />
                      }
                      label={
                        generating
                          ? t("settings.MTProto.generating")
                          : t("settings.MTProto.generateSecret")
                      }
                      onClick={() => void handleGenerateSecret()}
                      disabled={!config.system.mtproto?.enabled || generating}
                      sx={{ cursor: "pointer" }}
                    />
                  </InputAdornment>
                ),
              },
            }}
          />
          <Button
            variant="contained"
            size="small"
            startIcon={<IosShareIcon />}
            onClick={openShare}
            disabled={!config.system.mtproto?.enabled || !secret}
            sx={{ alignSelf: "flex-start" }}
          >
            {t("settings.MTProto.shareLink")}
          </Button>
        </Box>

        <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
          <B4Alert
            severity={dcAlertSeverity}
            action={
              <Stack direction="row" alignItems="center" gap={0.5}>
                {refreshResult?.ok && (
                  <Tooltip
                    title={
                      showDcs
                        ? t("settings.MTProto.hideDCs")
                        : t("settings.MTProto.showDCs")
                    }
                  >
                    <IconButton
                      size="small"
                      color="inherit"
                      onClick={() => setShowDcs((v) => !v)}
                      sx={{
                        transition: "transform 0.2s",
                        transform: showDcs ? "rotate(180deg)" : "rotate(0deg)",
                      }}
                    >
                      <ExpandMoreIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                )}
                <Button
                  color="inherit"
                  size="small"
                  startIcon={
                    <RefreshIcon
                      sx={{
                        animation: refreshing
                          ? "spin 1s linear infinite"
                          : "none",
                        "@keyframes spin": {
                          from: { transform: "rotate(0deg)" },
                          to: { transform: "rotate(360deg)" },
                        },
                      }}
                    />
                  }
                  onClick={() => void handleRefreshDCs()}
                  disabled={refreshing}
                >
                  {refreshing
                    ? t("settings.MTProto.refreshingDCs")
                    : t("settings.MTProto.refreshDCs")}
                </Button>
              </Stack>
            }
          >
            {dcAlertText}
            {refreshResult?.ok && (
              <Collapse in={showDcs} unmountOnExit>
                <Box
                  component="ul"
                  sx={{
                    m: 0,
                    mt: 1,
                    pl: 2,
                    fontFamily: "monospace",
                    fontSize: "0.8rem",
                  }}
                >
                  {Object.entries(refreshResult.dcs)
                    .sort((a, b) => Number(a[0]) - Number(b[0]))
                    .map(([id, addr]) => (
                      <li key={id}>
                        DC{id} → {addr}
                      </li>
                    ))}
                </Box>
              </Collapse>
            )}
          </B4Alert>
        </Box>
        {config.system.mtproto?.enabled && config.system.mtproto?.dc_relay && (
          <B4Alert severity="info">
            <span
              dangerouslySetInnerHTML={{
                __html: t("settings.MTProto.relaySetup"),
              }}
            />
          </B4Alert>
        )}
      </B4FormGroup>
      <B4Dialog
        open={shareOpen}
        onClose={() => setShareOpen(false)}
        fullWidth
        maxWidth="sm"
        title={t("settings.MTProto.shareDialogTitle")}
        icon={<IosShareIcon />}
        actions={
          <>
            <Button onClick={() => setShareOpen(false)}>
              {t("core.close")}
            </Button>
            <Box sx={{ flex: 1 }} />
            <Button
              component="a"
              variant="outlined"
              href={shareLink || "#"}
              target="_blank"
              rel="noreferrer"
              startIcon={<OpenInNewIcon />}
              disabled={!shareLink}
            >
              {t("settings.MTProto.shareOpen")}
            </Button>
            {canShare && (
              <Button
                variant="contained"
                startIcon={<IosShareIcon />}
                onClick={() => void handleNativeShare()}
                disabled={!shareLink}
              >
                {t("settings.MTProto.shareNative")}
              </Button>
            )}
            <Button
              variant="contained"
              startIcon={copied ? <CheckIcon /> : <ContentCopyIcon />}
              onClick={() => void handleCopy()}
              disabled={!shareLink}
            >
              {copied ? t("core.copied") : t("core.copy")}
            </Button>
          </>
        }
      >
        <B4TextField
          sx={{ mt: 3 }}
          label={t("settings.MTProto.shareHost")}
          value={shareHost}
          onChange={(e) => setShareHost(e.target.value)}
          helperText={t("settings.MTProto.shareHostHelp")}
          autoFocus
        />
        <B4TextField
          label={t("settings.MTProto.shareLinkLabel")}
          value={shareLink}
          slotProps={{
            input: {
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copied ? t("core.copied") : t("core.copy")}>
                    <span>
                      <IconButton
                        size="small"
                        onClick={() => void handleCopy()}
                        disabled={!shareLink}
                      >
                        {copied ? (
                          <CheckIcon fontSize="small" color="success" />
                        ) : (
                          <ContentCopyIcon fontSize="small" />
                        )}
                      </IconButton>
                    </span>
                  </Tooltip>
                </InputAdornment>
              ),
            },
          }}
          helperText={t("settings.MTProto.shareLinkHelp")}
        />
        {shareLink && (
          <Box
            sx={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              gap: 1,
              alignSelf: "center",
            }}
          >
            <Box sx={{ px: 1, pt: 1, bgcolor: "#fff", borderRadius: 2 }}>
              <QRCodeSVG
                value={shareLink}
                size={220}
                level="H"
                marginSize={0}
                imageSettings={{
                  src: "/favicon.svg",
                  height: 32,
                  width: 32,
                  excavate: true,
                }}
              />
            </Box>
            <Typography variant="caption" color="text.secondary">
              {t("settings.MTProto.shareQrHelp")}
            </Typography>
          </Box>
        )}
      </B4Dialog>
    </B4Section>
  );
};
