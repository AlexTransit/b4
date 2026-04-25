import { colors, radiusPx } from "@design";
import { B4SetConfig } from "@models/config";
import {
  Circle as CircleIcon,
  FolderOpen as FolderIcon,
} from "@mui/icons-material";
import { Box, Stack, Typography } from "@mui/material";
import { useNavigate } from "react-router";
import { useTranslation } from "react-i18next";

interface ActiveSetsProps {
  sets: B4SetConfig[];
}

export const ActiveSets = ({ sets }: ActiveSetsProps) => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  if (sets.length === 0) return null;

  return (
    <Box
      sx={{
        mb: 1.5,
        p: "14px",
        borderRadius: `${radiusPx.md}px`,
        bgcolor: colors.background.paper,
        border: `1px solid ${colors.border.default}`,
      }}
    >
      <Typography
        variant="metricLabel"
        sx={{ display: "block", color: colors.text.secondary, opacity: 0.8 }}
      >
        {t("dashboard.activeSets.title")}
      </Typography>
      <Stack
        direction="row"
        spacing={1}
        flexWrap="wrap"
        useFlexGap
        sx={{ mt: 1 }}
      >
        {sets.map((set) => {
          const domainCount =
            (set.targets.sni_domains?.length || 0) +
            (set.targets.geosite_categories?.length || 0);
          const ipCount =
            (set.targets.ip?.length || 0) +
            (set.targets.geoip_categories?.length || 0);
          const totalTargets = domainCount + ipCount;

          return (
            <Box
              key={set.id}
              role="button"
              onClick={() => navigate(`/sets/${set.id}`)?.catch(() => {})}
              sx={{
                display: "inline-flex",
                alignItems: "center",
                gap: "6px",
                height: 22,
                padding: "0 10px",
                borderRadius: "11px",
                fontSize: 12,
                fontWeight: 600,
                cursor: "pointer",
                bgcolor: set.enabled
                  ? colors.accent.secondary
                  : colors.accent.primaryStrong,
                color: set.enabled ? colors.secondary : colors.text.disabled,
                transition: "background-color 120ms ease",
                "&:hover": {
                  bgcolor: set.enabled
                    ? colors.accent.secondaryHover
                    : colors.accent.primary,
                },
              }}
            >
              {set.enabled ? (
                <CircleIcon
                  sx={{
                    fontSize: 8,
                    color: colors.state.success,
                    flexShrink: 0,
                  }}
                />
              ) : (
                <FolderIcon
                  sx={{
                    fontSize: 14,
                    color: colors.text.disabled,
                    flexShrink: 0,
                  }}
                />
              )}
              <Box component="span">
                {set.name} · {totalTargets} {t("dashboard.activeSets.targets")}
              </Box>
            </Box>
          );
        })}
      </Stack>
    </Box>
  );
};
