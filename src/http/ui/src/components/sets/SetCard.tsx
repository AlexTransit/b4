import { useState } from "react";
import {
  Box,
  Card,
  CardContent,
  CardActionArea,
  Checkbox,
  Typography,
  Stack,
  IconButton,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Switch,
  Tooltip,
  Divider,
} from "@mui/material";
import {
  EditIcon,
  CopyIcon,
  CompareIcon,
  ClearIcon,
  DomainIcon,
  IpIcon,
  DragIcon,
  DnsIcon,
  FakingIcon,
  TcpIcon,
  CheckIcon,
  CloseIcon,
  EscalateOutIcon,
  EscalateInIcon,
} from "@b4.icons";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import { B4Badge } from "@b4.elements";
import { colors, radius } from "@design";
import { B4SetConfig } from "@models/config";
import { useTranslation } from "react-i18next";
import { SetStats } from "./Manager";

interface EscalationLink {
  id: string;
  name: string;
}

interface SetCardProps {
  set: B4SetConfig;
  stats?: SetStats;
  index: number;
  onEdit: () => void;
  onDuplicate: () => void;
  onCompare: () => void;
  onDelete: () => void;
  onToggleEnabled: (enabled: boolean) => void;
  dragHandleProps?: React.HTMLAttributes<HTMLDivElement>;
  selectionMode?: boolean;
  selected?: boolean;
  onSelect?: () => void;
  escalatesTo?: EscalationLink;
  escalatedFrom?: EscalationLink[];
  highlighted?: boolean;
  onEscalationHover?: (setId: string | null) => void;
  onEscalationClick?: (setId: string) => void;
}

interface TargetBadgeProps {
  label: string;
  type: "geosite" | "geoip" | "domain" | "ip";
}

const TargetBadge = ({ label, type }: TargetBadgeProps) => {
  // Truncate long labels
  const maxLen = type === "ip" ? 18 : 14;
  const truncated =
    label.length > maxLen ? `${label.slice(0, maxLen)}…` : label;

  const isGeo = type === "geosite" || type === "geoip";

  return (
    <Tooltip title={label}>
      <Box sx={{ maxWidth: 120 }}>
        <B4Badge
          label={truncated}
          size="small"
          icon={
            type === "ip" || type === "geoip" ? (
              <IpIcon sx={{ fontSize: 12 }} />
            ) : undefined
          }
          color={isGeo ? "secondary" : undefined}
          variant={isGeo ? undefined : "outlined"}
          sx={{
            "& .MuiChip-label": {
              overflow: "hidden",
              textOverflow: "ellipsis",
            },
          }}
        />
      </Box>
    </Tooltip>
  );
};

const STRATEGY_LABELS: Record<string, string> = {
  combo: "COMBO",
  hybrid: "HYBRID",
  disorder: "DISORDER",
  overlap: "OVERLAP",
  extsplit: "EXT SPLIT",
  firstbyte: "1ST BYTE",
  tcp: "TCP FRAG",
  ip: "IP FRAG",
  tls: "TLS REC",
  oob: "OOB",
  none: "NONE",
};

export const SetCard = ({
  set,
  stats,
  index,
  onEdit,
  onDuplicate,
  onCompare,
  onDelete,
  onToggleEnabled,
  dragHandleProps,
  selectionMode,
  selected,
  onSelect,
  escalatesTo,
  escalatedFrom,
  highlighted,
  onEscalationHover,
  onEscalationClick,
}: SetCardProps) => {
  const { t } = useTranslation();
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const strategy = set.fragmentation.strategy;
  const isSelected = !!(selectionMode && selected);
  const borderColor =
    highlighted || isSelected ? colors.secondary : colors.border.default;

  const domainCount = stats?.total_domains ?? set.targets.sni_domains.length;
  const ipCount = stats?.total_ips ?? set.targets.ip.length;

  const handleMenuOpen = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setMenuAnchor(e.currentTarget);
  };

  const handleMenuClose = () => setMenuAnchor(null);

  const handleAction = (action: () => void) => {
    handleMenuClose();
    action();
  };

  return (
    <Card
      elevation={1}
      sx={{
        position: "relative",
        opacity: set.enabled ? 1 : 0.5,
        transition: "all 0.2s ease",
        border: `1px solid ${borderColor}`,
        borderRadius: radius.md,
        bgcolor: set.enabled ? colors.background.paper : colors.background.dark,
        boxShadow: highlighted
          ? `0 0 0 2px ${colors.secondary}, 0 8px 24px ${colors.accent.primary}`
          : undefined,
        "&:hover": {
          borderColor: colors.secondary,
          transform: "translateY(-2px)",
          boxShadow: `0 8px 24px ${colors.accent.primary}`,
        },
      }}
    >
      {/* Top accent bar */}
      <Box
        sx={{
          height: 4,
          bgcolor: colors.secondary,
          borderRadius: `${radius.md}px ${radius.md}px 0 0`,
        }}
      />

      {/* Header row */}
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          px: 2,
          pt: 1.5,
          pb: 1,
        }}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          {selectionMode ? (
            <Checkbox
              size="small"
              checked={selected}
              onChange={(e) => {
                e.stopPropagation();
                onSelect?.();
              }}
              onClick={(e) => e.stopPropagation()}
              sx={{
                color: colors.text.secondary,
                "&.Mui-checked": { color: colors.secondary },
                p: 0.5,
              }}
            />
          ) : (
            <Box
              {...dragHandleProps}
              sx={{
                cursor: "grab",
                color: colors.text.secondary,
                display: "flex",
                "&:hover": { color: colors.secondary },
              }}
            >
              <DragIcon fontSize="small" />
            </Box>
          )}

          <Tooltip title={set.enabled ? t("core.disable") : t("core.enable")}>
            <Switch
              size="small"
              checked={set.enabled}
              onChange={(e) => {
                e.stopPropagation();
                onToggleEnabled(e.target.checked);
              }}
              onClick={(e) => e.stopPropagation()}
            />
          </Tooltip>

        </Stack>

        {!selectionMode && (
          <IconButton size="small" onClick={handleMenuOpen}>
            <MoreVertIcon fontSize="small" />
          </IconButton>
        )}

        <Menu
          anchorEl={menuAnchor}
          open={Boolean(menuAnchor)}
          onClose={handleMenuClose}
          transformOrigin={{ horizontal: "right", vertical: "top" }}
          anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
        >
          <MenuItem onClick={() => handleAction(onEdit)}>
            <ListItemIcon>
              <EditIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>{t("core.edit")}</ListItemText>
          </MenuItem>
          <MenuItem onClick={() => handleAction(onDuplicate)}>
            <ListItemIcon>
              <CopyIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>{t("core.duplicate")}</ListItemText>
          </MenuItem>
          <MenuItem onClick={() => handleAction(onCompare)}>
            <ListItemIcon>
              <CompareIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>{t("core.compare")}</ListItemText>
          </MenuItem>
          <Divider />
          <MenuItem
            onClick={() => handleAction(onDelete)}
            sx={{ color: colors.secondary }}
          >
            <ListItemIcon>
              <ClearIcon fontSize="small" sx={{ color: colors.secondary }} />
            </ListItemIcon>
            <ListItemText>{t("core.delete")}</ListItemText>
          </MenuItem>
        </Menu>
      </Box>

      {/* Clickable content area */}
      <CardActionArea onClick={selectionMode ? onSelect : onEdit} sx={{ borderRadius: 0 }}>
        <CardContent sx={{ pt: 0, pb: 2 }}>
          {/* Name */}
          <Typography
            variant="h6"
            sx={{
              fontWeight: 600,
              my: 1,
              textTransform: "uppercase",
              color: set.enabled ? colors.text.primary : colors.text.secondary,
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {set.name}
          </Typography>

          {/* Target preview */}
          <Box
            sx={{
              p: 2,
              mb: 2,
              borderRadius: radius.sm,
              bgcolor: colors.background.dark,
              border: `1px solid ${colors.border.light}`,
              minHeight: 48,
            }}
          >
            {set.targets.geosite_categories.length > 0 ||
            set.targets.sni_domains.length > 0 ||
            set.targets.geoip_categories.length > 0 ||
            set.targets.ip.length > 0 ? (
              <Stack direction="row" flexWrap="wrap" gap={0.5}>
                {/* Geosite categories first */}
                {set.targets.geosite_categories.slice(0, 2).map((cat) => (
                  <TargetBadge key={cat} label={cat} type="geosite" />
                ))}

                {/* Then domains if room */}
                {set.targets.geosite_categories.length < 2 &&
                  set.targets.sni_domains
                    .slice(0, 2 - set.targets.geosite_categories.length)
                    .map((domain) => (
                      <TargetBadge key={domain} label={domain} type="domain" />
                    ))}

                {/* GeoIP categories */}
                {set.targets.geosite_categories.length +
                  set.targets.sni_domains.length <
                  2 &&
                  set.targets.geoip_categories
                    .slice(
                      0,
                      2 -
                        set.targets.geosite_categories.length -
                        set.targets.sni_domains.length,
                    )
                    .map((cat) => (
                      <TargetBadge key={cat} label={cat} type="geoip" />
                    ))}

                {/* Manual IPs last */}
                {set.targets.geosite_categories.length +
                  set.targets.sni_domains.length +
                  set.targets.geoip_categories.length <
                  2 &&
                  set.targets.ip
                    .slice(
                      0,
                      2 -
                        set.targets.geosite_categories.length -
                        set.targets.sni_domains.length -
                        set.targets.geoip_categories.length,
                    )
                    .map((ip) => <TargetBadge key={ip} label={ip} type="ip" />)}

                {/* +N more */}
                {set.targets.geosite_categories.length +
                  set.targets.sni_domains.length +
                  set.targets.geoip_categories.length +
                  set.targets.ip.length >
                  2 && (
                  <B4Badge
                    label={`+${
                      set.targets.geosite_categories.length +
                      set.targets.sni_domains.length +
                      set.targets.geoip_categories.length +
                      set.targets.ip.length -
                      2
                    }`}
                    size="small"
                    variant="outlined"
                  />
                )}
              </Stack>
            ) : (
              <Typography
                variant="caption"
                color="text.secondary"
                sx={{ fontStyle: "italic" }}
              >
                {t("sets.card.noTargets")}
              </Typography>
            )}
          </Box>

          {/* Active techniques */}
          <Stack direction="row" flexWrap="wrap" gap={0.5} sx={{ mb: 2 }}>
            <B4Badge
              label={STRATEGY_LABELS[strategy]}
              size="small"
              sx={{ bgcolor: colors.primary, color: colors.text.primary }}
            />
            {set.faking.sni && (
              <B4Badge label="FAKE" size="small" color="primary" />
            )}
            {set.dns?.enabled && (
              <B4Badge
                label="DNS"
                size="small"
                variant="outlined"
                color="secondary"
              />
            )}
            {set.fragmentation.reverse_order && (
              <B4Badge label="REV" size="small" variant="outlined" />
            )}
          </Stack>

          {/* Domain/IP counts */}
          <Stack direction="row" spacing={2} sx={{ mb: 2 }}>
            <Tooltip
              title={`${stats?.manual_domains || 0} ${t("sets.card.manual")}, ${
                stats?.geosite_domains || 0
              } ${t("sets.card.geosite")}`}
            >
              <Stack
                direction="row"
                alignItems="center"
                spacing={0.5}
                sx={{ flex: 1 }}
              >
                <DomainIcon
                  sx={{ fontSize: 16, color: colors.text.secondary }}
                />
                <Typography
                  variant="body2"
                  fontWeight={600}
                  color="text.primary"
                >
                  {domainCount.toLocaleString()}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {t("core.domains")}
                </Typography>
              </Stack>
            </Tooltip>
            <Tooltip
              title={`${stats?.manual_ips || 0} ${t("sets.card.manual")}, ${
                stats?.geoip_ips || 0
              } ${t("sets.card.geoip")}`}
            >
              <Stack
                direction="row"
                alignItems="center"
                spacing={0.5}
                sx={{ flex: 1 }}
              >
                <IpIcon sx={{ fontSize: 16, color: colors.text.secondary }} />
                <Typography
                  variant="body2"
                  fontWeight={600}
                  color="text.primary"
                >
                  {ipCount.toLocaleString()}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {t("core.ips")}
                </Typography>
              </Stack>
            </Tooltip>
          </Stack>

          {/* Quick flags */}
          <Box
            sx={{
              display: "flex",
              gap: 1,
              p: 1,
              borderRadius: radius.sm,
              bgcolor: colors.background.dark,
              border: `1px solid ${colors.border.light}`,
            }}
          >
            <QuickFlag
              icon={<TcpIcon />}
              label={`${set.tcp.conn_bytes_limit}B`}
              tooltip={t("sets.card.tcpBytesLimit")}
            />
            <QuickFlag
              icon={<FakingIcon />}
              enabled={set.faking.sni}
              tooltip={set.faking.sni ? t("sets.card.sniFakingOn") : t("sets.card.sniFakingOff")}
            />
            <QuickFlag
              icon={<DnsIcon />}
              enabled={set.dns?.enabled}
              tooltip={
                set.dns?.enabled ? `DNS → ${set.dns.target_dns}` : t("sets.card.dnsOff")
              }
            />
          </Box>
        </CardContent>
      </CardActionArea>

      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          gap: 1,
          px: 2,
          pb: 1.5,
          minHeight: 30,
        }}
      >
        <Box
          sx={{
            display: "flex",
            flexWrap: "wrap",
            gap: 0.5,
            flex: 1,
            minWidth: 0,
          }}
        >
          {escalatesTo && (
            <EscalationChip
              icon={<EscalateOutIcon sx={{ fontSize: 12 }} />}
              prefix={t("sets.card.escalatesTo")}
              link={escalatesTo}
              onHover={onEscalationHover}
              onClick={onEscalationClick}
            />
          )}
          {escalatedFrom?.map((link) => (
            <EscalationChip
              key={link.id}
              icon={<EscalateInIcon sx={{ fontSize: 12 }} />}
              prefix={t("sets.card.escalatedFrom")}
              link={link}
              onHover={onEscalationHover}
              onClick={onEscalationClick}
              variant="outlined"
            />
          ))}
        </Box>

        <Box
          sx={{
            flexShrink: 0,
            width: 22,
            height: 22,
            borderRadius: "50%",
            bgcolor: colors.accent.primary,
            border: `1px solid ${colors.border.default}`,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <Typography
            variant="caption"
            color="text.secondary"
            fontWeight={600}
            sx={{ fontSize: "0.65rem" }}
          >
            {index + 1}
          </Typography>
        </Box>
      </Box>
    </Card>
  );
};

interface QuickFlagProps {
  icon: React.ReactNode;
  label?: string;
  enabled?: boolean;
  tooltip: string;
}

const QuickFlag = ({ icon, label, enabled, tooltip }: QuickFlagProps) => {
  const isBoolean = enabled !== undefined;
  const isActive = isBoolean ? enabled : true;
  const color = isActive ? colors.secondary : colors.text.disabled;

  return (
    <Tooltip title={tooltip}>
      <Stack
        direction="row"
        alignItems="center"
        spacing={0.5}
        sx={{
          flex: 1,
          justifyContent: "center",
          py: 0.5,
          px: 0.5,
          borderRadius: radius.sm,
          bgcolor: isActive ? colors.accent.secondary : "transparent",
        }}
      >
        <Box sx={{ color, display: "flex", "& svg": { fontSize: 14 } }}>
          {icon}
        </Box>
        {label ? (
          <Typography
            variant="caption"
            fontWeight={600}
            sx={{ color, fontSize: "0.7rem" }}
          >
            {label}
          </Typography>
        ) : (
          <Box sx={{ color, display: "flex", "& svg": { fontSize: 12 } }}>
            {enabled ? <CheckIcon /> : <CloseIcon />}
          </Box>
        )}
      </Stack>
    </Tooltip>
  );
};

interface EscalationChipProps {
  icon: React.ReactNode;
  prefix: string;
  link: { id: string; name: string };
  variant?: "filled" | "outlined";
  onHover?: (setId: string | null) => void;
  onClick?: (setId: string) => void;
}

const EscalationChip = ({
  icon,
  prefix,
  link,
  variant,
  onHover,
  onClick,
}: EscalationChipProps) => (
  <Tooltip title={`${prefix}: ${link.name}`}>
    <B4Badge
      icon={icon as React.ReactElement}
      label={link.name}
      size="small"
      color="secondary"
      variant={variant}
      clickable
      onMouseEnter={() => onHover?.(link.id)}
      onMouseLeave={() => onHover?.(null)}
      onFocus={() => onHover?.(link.id)}
      onBlur={() => onHover?.(null)}
      onClick={(e) => {
        e.stopPropagation();
        onClick?.(link.id);
      }}
      sx={{
        maxWidth: "100%",
        cursor: "pointer",
        "& .MuiChip-label": {
          overflow: "hidden",
          textOverflow: "ellipsis",
          maxWidth: 110,
        },
      }}
    />
  </Tooltip>
);
