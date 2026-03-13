import { useState } from "react";
import {
  Button,
  Stack,
  Typography,
  Box,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  CircularProgress,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { SecurityIcon, ErrorIcon, CheckIcon, RestoreIcon } from "@b4.icons";
import { B4Alert } from "@b4.elements";
import { useConfigReset } from "@hooks/useConfig";
import { colors } from "@design";
import { B4Dialog } from "@common/B4Dialog";

interface ResetDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

type ResetState = "confirm" | "resetting" | "success" | "error";

export const ResetDialog = ({ open, onClose, onSuccess }: ResetDialogProps) => {
  const [state, setState] = useState<ResetState>("confirm");
  const [message, setMessage] = useState("");
  const { resetConfig } = useConfigReset();
  const { t } = useTranslation();

  const handleReset = async () => {
    setState("resetting");
    setMessage(t("settings.ResetDialog.resetting"));

    const response = await resetConfig();

    if (response?.success) {
      setState("success");
      setMessage(t("settings.ResetDialog.success"));
      setTimeout(() => {
        handleClose();
        onSuccess();
      }, 2000);
    } else {
      setState("error");
      setMessage(t("settings.ResetDialog.failed"));
    }
  };

  const handleClose = () => {
    if (state !== "resetting") {
      setState("confirm");
      setMessage("");
      onClose();
    }
  };

  const defaultProps = {
    title: t("settings.ResetDialog.title"),
    subtitle: t("settings.ResetDialog.subtitle"),
    icon: <RestoreIcon />,
  };

  // Dynamic dialog props based on state
  const getDialogProps = () => {
    switch (state) {
      case "confirm":
        return {
          ...defaultProps,
        };
      case "resetting":
        return {
          ...defaultProps,
          title: t("settings.ResetDialog.resettingTitle"),
          subtitle: t("settings.ResetDialog.pleaseWait"),
          icon: <CircularProgress size={24} />,
        };
      case "success":
        return {
          ...defaultProps,
          title: t("settings.ResetDialog.successTitle"),
          subtitle: t("settings.ResetDialog.successSubtitle"),
        };
      case "error":
        return {
          ...defaultProps,
          title: t("settings.ResetDialog.failedTitle"),
          subtitle: t("settings.ResetDialog.failedSubtitle"),
          icon: <ErrorIcon />,
        };
      default:
        return {
          ...defaultProps,
        };
    }
  };

  const getDialogActions = () => {
    switch (state) {
      case "confirm":
        return (
          <>
            <Button onClick={handleClose}>{t("core.cancel")}</Button>
            <Box sx={{ flex: 1 }} />
            <Button
              onClick={() => {
                void handleReset();
              }}
              variant="contained"
              startIcon={<RestoreIcon />}
            >
              {t("settings.ResetDialog.resetButton")}
            </Button>
          </>
        );
      case "error":
        return (
          <Button onClick={handleClose} variant="contained">
            {t("core.close")}
          </Button>
        );

      case "success":
      default:
        return null;
    }
  };

  const getDialogContent = () => {
    switch (state) {
      case "confirm":
        return (
          <>
            <B4Alert>
              {t("settings.ResetDialog.warning")}
            </B4Alert>
            <B4Alert severity="warning">
              {t("settings.ResetDialog.preserveWarning")}
            </B4Alert>
            <List dense>
              <ListItem>
                <ListItemIcon>
                  <SecurityIcon sx={{ color: colors.secondary }} />
                </ListItemIcon>
                <ListItemText
                  primary={t("settings.ResetDialog.domainConfig")}
                  secondary={t("settings.ResetDialog.domainConfigDesc")}
                />
              </ListItem>
              <ListItem>
                <ListItemIcon>
                  <SecurityIcon sx={{ color: colors.secondary }} />
                </ListItemIcon>
                <ListItemText
                  primary={t("settings.ResetDialog.testingConfig")}
                  secondary={t("settings.ResetDialog.testingConfigDesc")}
                />
              </ListItem>
            </List>
          </>
        );

      case "resetting":
        return (
          <Stack spacing={3} alignItems="center" sx={{ py: 4 }}>
            <CircularProgress size={48} sx={{ color: colors.secondary }} />
            <Typography variant="h6" sx={{ color: colors.text.primary }}>
              {message}
            </Typography>
          </Stack>
        );

      case "success":
        return (
          <Stack spacing={3} alignItems="center" sx={{ py: 4 }}>
            <CheckIcon
              sx={{
                fontSize: 64,
                color: colors.secondary,
              }}
            />
            <Typography variant="h6" sx={{ color: colors.text.primary }}>
              {message}
            </Typography>
          </Stack>
        );

      case "error":
        return (
          <Stack spacing={3} alignItems="center" sx={{ py: 4 }}>
            <ErrorIcon sx={{ fontSize: 64, color: colors.quaternary }} />
            <B4Alert severity="error">{message}</B4Alert>
          </Stack>
        );
    }
  };

  return (
    <B4Dialog
      {...getDialogProps()}
      open={open}
      onClose={handleClose}
      actions={getDialogActions()}
    >
      {getDialogContent()}
    </B4Dialog>
  );
};
