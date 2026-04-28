import { useState } from "react";
import {
  Button,
  DialogContent,
  DialogContentText,
  Grid,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import SettingSection from "@common/B4Section";
import { ControlIcon, RestartIcon, InfoIcon, RestoreIcon } from "@b4.icons";
import { RestartDialog } from "./RestartDialog";
import { SystemInfoDialog } from "./SystemInfoDialog";
import { B4Dialog } from "@b4.elements";
import { useSnackbar } from "@context/SnackbarProvider";
import { configApi } from "@b4.settings";
import { spacing } from "@design";

export const ControlSettings = () => {
  const [showRestartDialog, setShowRestartDialog] = useState(false);
  const [showSysInfoDialog, setShowSysInfoDialog] = useState(false);
  const [showResetDialog, setShowResetDialog] = useState(false);
  const [resetting, setResetting] = useState(false);
  const { showError, showSuccess } = useSnackbar();
  const { t } = useTranslation();

  const handleResetConfirm = async () => {
    try {
      setResetting(true);
      await configApi.reset();
      showSuccess(t("settings.Control.resetSuccess"));
      setTimeout(() => globalThis.window.location.reload(), 800);
    } catch (error) {
      showError(error instanceof Error ? error.message : t("settings.Control.resetError"));
      setResetting(false);
      setShowResetDialog(false);
    }
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
        >
          {t("settings.Control.restartService")}
        </Button>
        <Button
          size="small"
          variant="outlined"
          startIcon={<InfoIcon />}
          onClick={() => setShowSysInfoDialog(true)}
        >
          {t("settings.Control.systemInfo")}
        </Button>
        <Button
          size="small"
          variant="outlined"
          color="warning"
          startIcon={<RestoreIcon />}
          onClick={() => setShowResetDialog(true)}
        >
          {t("settings.Control.resetConfig")}
        </Button>
      </Grid>

      <RestartDialog
        open={showRestartDialog}
        onClose={() => setShowRestartDialog(false)}
      />

      <SystemInfoDialog
        open={showSysInfoDialog}
        onClose={() => setShowSysInfoDialog(false)}
      />

      <B4Dialog
        title={t("settings.Control.resetConfig")}
        open={showResetDialog}
        onClose={() => !resetting && setShowResetDialog(false)}
        actions={
          <>
            <Button onClick={() => setShowResetDialog(false)} disabled={resetting}>
              {t("core.cancel")}
            </Button>
            <Button
              onClick={() => {
                void handleResetConfirm();
              }}
              variant="contained"
              color="warning"
              disabled={resetting}
            >
              {resetting ? t("core.saving") : t("settings.Control.resetConfig")}
            </Button>
          </>
        }
      >
        <DialogContent>
          <DialogContentText>
            {t("settings.Control.resetConfirm")}
          </DialogContentText>
        </DialogContent>
      </B4Dialog>
    </SettingSection>
  );
};
