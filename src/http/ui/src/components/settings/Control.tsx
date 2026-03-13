import { useState } from "react";
import { Button, Grid } from "@mui/material";
import { useTranslation } from "react-i18next";
import SettingSection from "@common/B4Section";
import { ControlIcon, RestartIcon, RestoreIcon } from "@b4.icons";
import { RestartDialog } from "./RestartDialog";
import { spacing } from "@design";
import { ResetDialog } from "./ResetDialog";

interface ControlSettingsProps {
  loadConfig: () => void;
}

export const ControlSettings = ({ loadConfig }: ControlSettingsProps) => {
  const [saving] = useState(false);
  const [showRestartDialog, setShowRestartDialog] = useState(false);
  const [showResetDialog, setShowResetDialog] = useState(false);
  const { t } = useTranslation();

  const handleResetSuccess = () => {
    loadConfig();
  };

  return (
    <SettingSection
      title={t("settings.Control.title")}
      description={t("settings.Control.description")}
      icon={<ControlIcon />}
    >
      <Grid container spacing={spacing.lg}>
        <Button
          size="small"
          variant="outlined"
          startIcon={<RestartIcon />}
          onClick={() => setShowRestartDialog(true)}
          disabled={saving}
        >
          {t("settings.Control.restartService")}
        </Button>
        <Button
          size="small"
          variant="outlined"
          startIcon={<RestoreIcon />}
          onClick={() => setShowResetDialog(true)}
          disabled={saving}
        >
          {t("settings.Control.resetConfig")}
        </Button>
      </Grid>

      <RestartDialog
        open={showRestartDialog}
        onClose={() => setShowRestartDialog(false)}
      />

      <ResetDialog
        open={showResetDialog}
        onClose={() => setShowResetDialog(false)}
        onSuccess={handleResetSuccess}
      />
    </SettingSection>
  );
};
