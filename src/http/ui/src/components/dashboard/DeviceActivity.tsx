import { useEffect, useState, useMemo } from "react";
import {
  Box,
  Paper,
  Typography,
  IconButton,
  Collapse,
  Tooltip,
  Menu,
  MenuItem,
} from "@mui/material";
import {
  ExpandMore as ExpandMoreIcon,
  AddCircleOutline as AddIcon,
  Check as CheckIcon,
} from "@mui/icons-material";
import { colors, fonts, radiusPx } from "@design";
import { formatNumber } from "@utils";
import { B4SetConfig } from "@models/config";
import { setsApi } from "@b4.sets";
import { B4ConfidencePill, B4CountPill } from "@b4.elements";
import { useTranslation } from "react-i18next";

interface DeviceInfo {
  mac: string;
  ip: string;
  hostname: string;
  vendor: string;
  alias?: string;
}

interface DeviceActivityProps {
  deviceDomains: Record<string, Record<string, number>>;
  domainTLS: Record<string, string>;
  sets: B4SetConfig[];
  targetedDomains: Set<string>;
  onRefreshSets: () => void;
}

const ROW_GRID = "200px 1fr 100px 100px 24px";

export const DeviceActivity = ({
  deviceDomains,
  domainTLS,
  sets,
  targetedDomains,
  onRefreshSets,
}: DeviceActivityProps) => {
  const { t } = useTranslation();
  const [devices, setDevices] = useState<DeviceInfo[]>([]);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  useEffect(() => {
    fetch("/api/devices")
      .then((r) => r.json())
      .then((data: { devices?: DeviceInfo[] }) => {
        if (data?.devices) setDevices(data.devices);
      })
      .catch(() => {});
  }, []);

  const isDomainTargeted = (domain: string): boolean => {
    if (targetedDomains.has(domain)) return true;
    const parts = domain.split(".");
    for (let i = 1; i < parts.length; i++) {
      if (targetedDomains.has(parts.slice(i).join("."))) return true;
    }
    return false;
  };

  const deviceMap = useMemo(() => {
    const map: Record<string, DeviceInfo> = {};
    for (const d of devices) {
      map[d.mac] = d;
    }
    return map;
  }, [devices]);

  const sortedDevices = useMemo(() => {
    return Object.entries(deviceDomains)
      .map(([mac, domains]) => ({
        mac,
        domains,
        total: Object.values(domains).reduce((s, c) => s + c, 0),
        domainCount: Object.keys(domains).length,
      }))
      .sort((a, b) => b.total - a.total);
  }, [deviceDomains]);

  const toggleExpand = (mac: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(mac)) next.delete(mac);
      else next.add(mac);
      return next;
    });
  };

  const getDeviceName = (mac: string): string => {
    const dev = deviceMap[mac];
    if (dev?.alias) return dev.alias;
    if (dev?.hostname) return dev.hostname;
    if (dev?.vendor && dev.vendor !== "Private")
      return `${dev.vendor} (${mac})`;
    return mac;
  };

  const getDeviceMeta = (mac: string): string => {
    const dev = deviceMap[mac];
    if (!dev) return "";
    const parts: string[] = [];
    if (dev.ip) parts.push(dev.ip);
    if (dev.vendor && dev.vendor !== "Private") parts.push(dev.vendor);
    return parts.join(" · ");
  };

  if (sortedDevices.length === 0) return null;

  return (
    <Paper
      variant="outlined"
      sx={{
        bgcolor: colors.background.paper,
        borderColor: colors.border.default,
        borderRadius: `${radiusPx.md}px`,
        p: 0,
        overflow: "hidden",
        height: "100%",
      }}
    >
      <Typography
        variant="metricLabel"
        sx={{
          display: "block",
          color: colors.text.secondary,
          opacity: 0.8,
          p: "12px 14px 6px",
        }}
      >
        {t("dashboard.deviceActivity.title")}
      </Typography>
      {sortedDevices.map(({ mac, domains, total, domainCount }) => {
        const isExpanded = expanded.has(mac);
        const sortedDomains = Object.entries(domains).sort(
          (a, b) => b[1] - a[1],
        );
        return (
          <Box key={mac}>
            <Box
              role="button"
              tabIndex={0}
              aria-expanded={isExpanded}
              onClick={() => toggleExpand(mac)}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  toggleExpand(mac);
                }
              }}
              sx={{
                display: "grid",
                gridTemplateColumns: ROW_GRID,
                alignItems: "center",
                gap: "18px",
                p: "12px 14px",
                borderBottom: `1px solid ${colors.border.light}`,
                cursor: "pointer",
                transition: "background-color 120ms ease",
                "&:hover": { bgcolor: "rgba(255, 255, 255, 0.025)" },
              }}
            >
              <Box
                sx={{
                  display: "flex",
                  flexDirection: "column",
                  lineHeight: 1.2,
                  minWidth: 0,
                }}
              >
                <Box
                  component="span"
                  sx={{
                    color: colors.text.primary,
                    fontSize: 13,
                    fontWeight: 600,
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                  title={getDeviceName(mac)}
                >
                  {getDeviceName(mac)}
                </Box>
                <Box
                  component="span"
                  sx={{
                    fontFamily: fonts.mono,
                    fontSize: 11,
                    color: colors.text.secondary,
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                  title={getDeviceMeta(mac)}
                >
                  {getDeviceMeta(mac)}
                </Box>
              </Box>
              <Box />
              <Box
                sx={{
                  fontFamily: fonts.mono,
                  fontSize: 12,
                  color: colors.text.primary,
                  textAlign: "right",
                }}
              >
                {domainCount} {t("core.domains")}
              </Box>
              <Box
                sx={{
                  fontFamily: fonts.mono,
                  fontSize: 12,
                  color: colors.text.primary,
                  textAlign: "right",
                }}
              >
                {formatNumber(total)}
              </Box>
              <ExpandMoreIcon
                sx={{
                  color: colors.text.secondary,
                  fontSize: 18,
                  transition: "transform 150ms ease",
                  transform: isExpanded ? "rotate(180deg)" : "rotate(0)",
                  justifySelf: "center",
                }}
              />
            </Box>
            <Collapse in={isExpanded} unmountOnExit>
              <Box sx={{ bgcolor: colors.background.default }}>
                {sortedDomains.map(([domain, count]) => (
                  <DeviceDomainRow
                    key={domain}
                    domain={domain}
                    count={count}
                    tls={domainTLS[domain]}
                    isTargeted={isDomainTargeted(domain)}
                    sets={sets}
                    onAdded={onRefreshSets}
                  />
                ))}
              </Box>
            </Collapse>
          </Box>
        );
      })}
    </Paper>
  );
};

interface DeviceDomainRowProps {
  domain: string;
  count: number;
  tls?: string;
  isTargeted: boolean;
  sets: B4SetConfig[];
  onAdded: () => void;
}

const DeviceDomainRow = ({
  domain,
  count,
  tls,
  isTargeted,
  sets,
  onAdded,
}: DeviceDomainRowProps) => {
  const { t } = useTranslation();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [adding, setAdding] = useState(false);
  const enabledSets = sets.filter((s) => s.enabled);

  const handleAdd = async (setId: string) => {
    setAnchorEl(null);
    setAdding(true);
    try {
      await setsApi.addDomainToSet(setId, domain);
      onAdded();
    } catch (e) {
      console.error("Failed to add domain:", e);
    } finally {
      setAdding(false);
    }
  };

  return (
    <Box
      sx={{
        display: "flex",
        alignItems: "center",
        gap: "10px",
        p: "8px 14px 8px 28px",
        borderBottom: `1px solid ${colors.border.light}`,
        "&:last-of-type": { borderBottom: 0 },
        transition: "background-color 120ms ease",
        "&:hover": { bgcolor: "rgba(255, 255, 255, 0.025)" },
      }}
    >
      {tls && <B4ConfidencePill score={tls} />}
      <Box
        component="span"
        sx={{
          fontFamily: fonts.mono,
          fontSize: 11,
          letterSpacing: "0.04em",
          color: colors.text.primary,
          textTransform: "uppercase",
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
          flex: 1,
          minWidth: 0,
        }}
        title={domain}
      >
        {domain}
      </Box>
      <B4CountPill value={formatNumber(count)} />
      {isTargeted && (
        <Tooltip title={t("dashboard.deviceActivity.alreadyInSet")}>
          <CheckIcon sx={{ color: colors.state.success, fontSize: 16 }} />
        </Tooltip>
      )}
      {!isTargeted && enabledSets.length > 0 && (
        <>
          <Tooltip title={t("core.addToSet")}>
            <IconButton
              size="small"
              onClick={(e) => setAnchorEl(e.currentTarget)}
              disabled={adding}
              sx={{
                color: colors.text.secondary,
                p: 0.25,
                "&:hover": { color: colors.secondary },
              }}
            >
              <AddIcon sx={{ fontSize: 16 }} />
            </IconButton>
          </Tooltip>
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={() => setAnchorEl(null)}
            slotProps={{
              paper: {
                sx: {
                  bgcolor: colors.background.default,
                  border: `1px solid ${colors.border.default}`,
                },
              },
            }}
          >
            {enabledSets.map((set) => (
              <MenuItem
                key={set.id}
                onClick={() => void handleAdd(set.id)}
                sx={{ color: colors.text.primary, fontSize: "0.8rem" }}
              >
                {set.name}
              </MenuItem>
            ))}
          </Menu>
        </>
      )}
    </Box>
  );
};
