import { Grid } from "@mui/material";
import { StatCard } from "./StatCard";
import { formatNumber } from "@utils";
import { useTranslation } from "react-i18next";
import type { Metrics } from "./Page";

interface MetricsCardsProps {
  metrics: Metrics;
}

export const MetricsCards = ({ metrics }: MetricsCardsProps) => {
  const { t } = useTranslation();
  const targetRate =
    metrics.total_connections > 0
      ? ((metrics.targeted_connections / metrics.total_connections) * 100).toFixed(1)
      : "0.0";

  const isIdle = metrics.rst_dropped === 0;

  return (
    <Grid container spacing={1.5} sx={{ height: "100%" }} alignItems="stretch">
      <Grid size={{ xs: 12, sm: 4 }} sx={{ display: "flex" }}>
        <StatCard
          label={t("dashboard.metrics.targeted")}
          value={formatNumber(metrics.targeted_connections)}
          sub={`${targetRate}% ${t("dashboard.metrics.ofTotal")}`}
          tone="secondary"
        />
      </Grid>

      <Grid size={{ xs: 12, sm: 4 }} sx={{ display: "flex" }}>
        <StatCard
          label={t("dashboard.metrics.rstDropped")}
          value={formatNumber(metrics.rst_dropped)}
          sub={isIdle ? t("dashboard.metrics.idle") : undefined}
          tone={isIdle ? "muted" : "primary"}
        />
      </Grid>

      <Grid size={{ xs: 12, sm: 4 }} sx={{ display: "flex" }}>
        <StatCard
          label={t("dashboard.metrics.packets")}
          value={formatNumber(metrics.packets_processed)}
          sub={`${metrics.current_pps.toFixed(1)} ${t("dashboard.metrics.pktPerSec")}`}
          tone="primary"
        />
      </Grid>
    </Grid>
  );
};
