import { useMemo } from "react";
import { Grid, Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { LogsIcon } from "@b4.icons";
import { B4Section, B4Select, B4Switch, B4TextField } from "@b4.elements";
import { B4Config, LogLevel } from "@models/config";
import { SettingsPropHandlerType } from "@models/settings";

interface LoggingSettingsProps {
  config: B4Config;
  onChange: (field: string, value: SettingsPropHandlerType) => void;
}

// Timezone list is locale-independent, compute once at module level
const ZONE_ENTRIES: { value: string; label: string }[] = (() => {
  try {
    return Intl.supportedValuesOf("timeZone").map((tz) => {
      const offset =
        new Intl.DateTimeFormat("en", {
          timeZone: tz,
          timeZoneName: "shortOffset",
        })
          .formatToParts()
          .find((p) => p.type === "timeZoneName")?.value ?? "";
      return { value: tz, label: `${tz} (${offset})` };
    });
  } catch {
    return [{ value: "UTC", label: "UTC" }];
  }
})();

export const LoggingSettings = ({ config, onChange }: LoggingSettingsProps) => {
  const { t } = useTranslation();

  const TIMEZONES = useMemo(
    () => [
      { value: "", label: t("settings.Logging.timezoneAuto") },
      ...ZONE_ENTRIES,
    ],
    [t],
  );

  const LOG_LEVELS: Array<{ value: LogLevel; label: string }> = [
    { value: LogLevel.ERROR, label: t("settings.Logging.levelError") },
    { value: LogLevel.INFO, label: t("settings.Logging.levelInfo") },
    { value: LogLevel.TRACE, label: t("settings.Logging.levelTrace") },
    { value: LogLevel.DEBUG, label: t("settings.Logging.levelDebug") },
  ];

  return (
    <B4Section
      title={t("settings.Logging.title")}
      description={t("settings.Logging.description")}
      icon={<LogsIcon />}
    >
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Stack spacing={2}>
            <B4Select
              label={t("settings.Logging.logLevel")}
              value={config.system.logging.level}
              options={LOG_LEVELS}
              onChange={(e) =>
                onChange("system.logging.level", Number(e.target.value))
              }
              helperText={t("settings.Logging.logLevelHelp")}
            />
            <B4TextField
              label={t("settings.Logging.errorFilePath")}
              value={config.system.logging.error_file}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                onChange("system.logging.error_file", e.target.value)
              }
              placeholder={t("settings.Logging.errorFilePathPlaceholder")}
              helperText={t("settings.Logging.errorFilePathHelp")}
            />
            <B4Select
              label={t("settings.Logging.timezone")}
              value={config.system.timezone ?? ""}
              options={TIMEZONES}
              onChange={(e) =>
                onChange("system.timezone", String(e.target.value))
              }
              helperText={t("settings.Logging.timezoneHelp")}
            />
          </Stack>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Stack spacing={2}>
            <B4Switch
              label={t("settings.Logging.instantFlush")}
              checked={config?.system?.logging?.instaflush}
              onChange={(checked: boolean) =>
                onChange("system.logging.instaflush", Boolean(checked))
              }
              description={t("settings.Logging.instantFlushDesc")}
            />
            <B4Switch
              label={t("settings.Logging.syslog")}
              checked={config?.system?.logging?.syslog}
              onChange={(checked: boolean) =>
                onChange("system.logging.syslog", Boolean(checked))
              }
              description={t("settings.Logging.syslogDesc")}
            />
          </Stack>
        </Grid>
      </Grid>
    </B4Section>
  );
};
