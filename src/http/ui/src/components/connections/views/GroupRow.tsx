import { memo } from "react";
import { Box, Stack, Typography, Tooltip } from "@mui/material";
import { AddIcon, NetworkIcon } from "@b4.icons";
import { B4Badge } from "@common/B4Badge";
import { ProtocolChip } from "@common/ProtocolChip";
import { colors } from "@design";
import { Sparkline } from "./Sparkline";
import { formatRelativeShort } from "@utils";
import type { EnrichedGroup } from "@hooks/useConnectionGroups";
import { useTranslation } from "react-i18next";

export const ROW_HEIGHT = 48;

interface Props {
  group: EnrichedGroup;
  now: number;
  selected: boolean;
  onSelect: (key: string) => void;
  onAddDomain: (domain: string) => void;
  onAddIp: (ip: string) => void;
  onEnrichAsn: (ip: string) => void;
  enrichingIps: Set<string>;
}

export const GroupRow = memo<Props>(
  ({
    group,
    now,
    selected,
    onSelect,
    onAddDomain,
    onAddIp,
    onEnrichAsn,
    enrichingIps,
  }) => {
    const { t } = useTranslation();
    const ipBase = group.destIp.split(":")[0].replaceAll(/[[\]]/g, "");
    const isEnriching = enrichingIps.has(ipBase);
    const hasDomain = !!group.domain;
    const deviceLabel = group.deviceName || group.mac;
    const matched = !!group.hostSet || !!group.ipSet;

    return (
      <Box
        onClick={() => onSelect(group.key)}
        sx={{
          height: ROW_HEIGHT,
          display: "flex",
          alignItems: "center",
          gap: 1.5,
          px: 2,
          borderBottom: `1px solid ${colors.border.light}`,
          cursor: "pointer",
          bgcolor: selected ? colors.accent.primary : "transparent",
          "&:hover": {
            bgcolor: selected
              ? colors.accent.primaryHover
              : colors.accent.primaryStrong,
          },
          transition: "background-color 80ms",
        }}
      >
        <Box sx={{ width: 170, flexShrink: 0, overflow: "hidden" }}>
          <ProtocolChip protocol={group.protocol} flags={group.flags} />
        </Box>

        <Box sx={{ flex: 2, minWidth: 0 }}>
          <Stack
            direction="row"
            spacing={1}
            alignItems="center"
            sx={{ minWidth: 0 }}
          >
            {group.tls && (
              <B4Badge variant="outlined" color="primary" label={group.tls} />
            )}
            <Typography
              sx={{
                color: hasDomain ? colors.text.primary : colors.text.disabled,
                fontSize: 13,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
                fontStyle: hasDomain ? "normal" : "italic",
                minWidth: 0,
              }}
            >
              {group.domain || t("connections.aggregated.noDomain")}
            </Typography>
            {hasDomain && !group.hostSet && (
              <Tooltip
                title={t("connections.aggregated.addDomain")}
                arrow
                placement="top"
              >
                <AddIcon
                  onClick={(e) => {
                    e.stopPropagation();
                    onAddDomain(group.domain);
                  }}
                  sx={{
                    fontSize: 16,
                    bgcolor: `${colors.secondary}88`,
                    color: colors.background.default,
                    borderRadius: "50%",
                    "&:hover": { bgcolor: colors.secondary },
                  }}
                />
              </Tooltip>
            )}
          </Stack>
        </Box>

        <Box sx={{ flex: 1.5, minWidth: 0 }}>
          <Stack
            direction="row"
            spacing={1}
            alignItems="center"
            sx={{ minWidth: 0 }}
          >
            <Typography
              sx={{
                fontFamily: "monospace",
                fontSize: 12,
                color: colors.text.secondary,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {group.destIp || "—"}
            </Typography>
            {group.asnName ? (
              <B4Badge
                variant="outlined"
                color="primary"
                label={group.asnName}
              />
            ) : group.destIp ? (
              <Tooltip
                title={t("connections.table.enrichAsn")}
                arrow
                placement="top"
              >
                <NetworkIcon
                  onClick={(e) => {
                    e.stopPropagation();
                    if (!isEnriching) onEnrichAsn(group.destIp);
                  }}
                  sx={{
                    fontSize: 14,
                    color: isEnriching
                      ? colors.text.disabled
                      : `${colors.secondary}88`,
                    "&:hover": { color: colors.secondary },
                  }}
                />
              </Tooltip>
            ) : null}
            {group.destIps.length > 1 && (
              <B4Badge
                label={`+${group.destIps.length - 1}`}
                variant="outlined"
              />
            )}
            {!group.ipSet && group.destIp && (
              <Tooltip
                title={t("connections.aggregated.addIp")}
                arrow
                placement="top"
              >
                <AddIcon
                  onClick={(e) => {
                    e.stopPropagation();
                    onAddIp(group.destIp);
                  }}
                  sx={{
                    fontSize: 14,
                    bgcolor: `${colors.secondary}88`,
                    color: colors.background.default,
                    borderRadius: "50%",
                    "&:hover": { bgcolor: colors.secondary },
                  }}
                />
              </Tooltip>
            )}
          </Stack>
        </Box>

        <Box sx={{ width: 130, flexShrink: 0 }}>
          <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
            {group.hostSet && (
              <B4Badge color="secondary" label={group.hostSet} />
            )}
            {group.ipSet && group.ipSet !== group.hostSet && (
              <B4Badge color="secondary" label={group.ipSet} />
            )}
          </Stack>
        </Box>

        <Box sx={{ width: 100, flexShrink: 0 }}>
          <Tooltip title={deviceLabel || ""} arrow placement="top">
            <Typography
              sx={{
                fontSize: 12,
                color: deviceLabel
                  ? colors.text.secondary
                  : colors.text.disabled,
                fontFamily: group.deviceName ? "inherit" : "monospace",
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {deviceLabel || "—"}
            </Typography>
          </Tooltip>
        </Box>

        <Box sx={{ width: 120, flexShrink: 0 }}>
          <Sparkline data={group.buckets} width={120} height={20} />
        </Box>

        <Box sx={{ width: 60, flexShrink: 0, textAlign: "right" }}>
          <Typography
            sx={{
              fontFamily: "monospace",
              fontSize: 12,
              color: colors.text.secondary,
            }}
          >
            {group.packets}
          </Typography>
        </Box>

        <Box sx={{ width: 48, flexShrink: 0, textAlign: "right" }}>
          <Typography
            sx={{
              fontFamily: "monospace",
              fontSize: 11,
              color: colors.text.disabled,
            }}
          >
            {formatRelativeShort(t, group.lastSeen, now)}
          </Typography>
        </Box>
      </Box>
    );
  },
  (prev, next) =>
    prev.selected === next.selected &&
    prev.enrichingIps === next.enrichingIps &&
    prev.group.packets === next.group.packets &&
    prev.group.lastSeen === next.group.lastSeen &&
    prev.group.asnName === next.group.asnName &&
    prev.group.destIp === next.group.destIp &&
    prev.group.hostSet === next.group.hostSet &&
    prev.group.ipSet === next.group.ipSet &&
    prev.group.buckets === next.group.buckets,
);

GroupRow.displayName = "GroupRow";
