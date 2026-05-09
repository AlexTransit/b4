import { Box, Paper, Typography } from "@mui/material";
import { colors, fonts, radiusPx } from "@design";
import { useTranslation } from "react-i18next";
import { EscalationEntry } from "./Page";

interface EscalationsProps {
  escalations: EscalationEntry[];
  total: number;
}

const formatTimeLeft = (expiresAt: string): string => {
  const expiry = new Date(expiresAt).getTime();
  if (!Number.isFinite(expiry)) return "";
  const diffMs = expiry - Date.now();
  if (diffMs <= 0) return "0m";
  const minutes = Math.floor(diffMs / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  const remMin = minutes % 60;
  return remMin === 0 ? `${hours}h` : `${hours}h ${remMin}m`;
};

export const Escalations = ({ escalations, total }: EscalationsProps) => {
  const { t } = useTranslation();
  if (escalations.length === 0) return null;

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
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          p: "12px 14px 6px",
        }}
      >
        <Typography
          variant="metricLabel"
          sx={{ color: colors.text.secondary, opacity: 0.8 }}
        >
          {t("dashboard.escalations.title")}
        </Typography>
        <Typography
          variant="caption"
          sx={{ color: colors.text.secondary, opacity: 0.7 }}
        >
          {t("dashboard.escalations.totalCount", { count: total })}
        </Typography>
      </Box>
      {escalations.map((e) => (
        <Box
          key={e.host}
          sx={{
            display: "flex",
            alignItems: "center",
            gap: "10px",
            p: "8px 14px",
            borderBottom: `1px solid ${colors.border.light}`,
            "&:last-of-type": { borderBottom: 0 },
          }}
        >
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
            title={e.host}
          >
            {e.host}
          </Box>
          <Typography
            variant="caption"
            sx={{
              color: colors.text.secondary,
              fontFamily: fonts.mono,
              whiteSpace: "nowrap",
            }}
            title={t("dashboard.escalations.viaSet")}
          >
            → {e.to_set}
          </Typography>
          {e.hops > 1 && (
            <Typography
              variant="caption"
              sx={{
                color: colors.text.secondary,
                opacity: 0.6,
                whiteSpace: "nowrap",
              }}
              title={t("dashboard.escalations.hops")}
            >
              ×{e.hops}
            </Typography>
          )}
          <Typography
            variant="caption"
            sx={{
              color: colors.text.secondary,
              opacity: 0.7,
              fontFamily: fonts.mono,
              whiteSpace: "nowrap",
              minWidth: 50,
              textAlign: "right",
            }}
            title={t("dashboard.escalations.expiresIn")}
          >
            {formatTimeLeft(e.expires_at)}
          </Typography>
        </Box>
      ))}
    </Paper>
  );
};
