import { useMemo, useState } from "react";
import {
  Box,
  Paper,
  Typography,
  IconButton,
  Tooltip,
  Menu,
  MenuItem,
} from "@mui/material";
import { AddCircleOutline as AddIcon } from "@mui/icons-material";
import { colors, fonts, radiusPx } from "@design";
import { formatNumber } from "@utils";
import { B4SetConfig } from "@models/config";
import { setsApi } from "@b4.sets";
import { B4ConfidencePill, B4CountPill } from "@b4.elements";
import { useTranslation } from "react-i18next";

interface UnmatchedDomainsProps {
  topDomains: Record<string, number>;
  domainTLS: Record<string, string>;
  sets: B4SetConfig[];
  targetedDomains: Set<string>;
  onRefreshSets: () => void;
}

export const UnmatchedDomains = ({
  topDomains,
  domainTLS,
  sets,
  targetedDomains,
  onRefreshSets,
}: UnmatchedDomainsProps) => {
  const { t } = useTranslation();
  const isDomainTargeted = (domain: string): boolean => {
    if (targetedDomains.has(domain)) return true;
    const parts = domain.split(".");
    for (let i = 1; i < parts.length; i++) {
      if (targetedDomains.has(parts.slice(i).join("."))) return true;
    }
    return false;
  };

  const unmatched = useMemo(() => {
    return Object.entries(topDomains)
      .filter(([domain]) => !isDomainTargeted(domain))
      .sort((a, b) => b[1] - a[1])
      .slice(0, 15);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [topDomains, targetedDomains]);

  if (unmatched.length === 0) return null;

  return (
    <Paper
      sx={{
        bgcolor: colors.background.paper,
        borderColor: colors.border.default,
        borderRadius: `${radiusPx.md}px`,
        p: 0,
        overflow: "hidden",
      }}
      variant="outlined"
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
        {t("dashboard.unmatchedDomains.title")}
      </Typography>
      {unmatched.map(([domain, count]) => (
        <UnmatchedRow
          key={domain}
          domain={domain}
          count={count}
          tls={domainTLS[domain]}
          sets={sets}
          onAdded={onRefreshSets}
        />
      ))}
    </Paper>
  );
};

interface UnmatchedRowProps {
  domain: string;
  count: number;
  tls?: string;
  sets: B4SetConfig[];
  onAdded: () => void;
}

const UnmatchedRow = ({
  domain,
  count,
  tls,
  sets,
  onAdded,
}: UnmatchedRowProps) => {
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
        p: "8px 14px",
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
      {enabledSets.length > 0 && (
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
