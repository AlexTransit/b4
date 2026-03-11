import { useState, useRef } from "react";
import { Button, Grid, Stack, Typography, Box } from "@mui/material";
import { BackupIcon, DownloadIcon, UploadIcon } from "@b4.icons";
import { B4Section, B4Alert } from "@b4.elements";
import { useSnackbar } from "@context/SnackbarProvider";
import { RestartDialog } from "./RestartDialog";
import { colors } from "@design";
import { getAuthToken } from "@context/AuthProvider";

export const BackupSettings = () => {
  const { showError, showSuccess } = useSnackbar();
  const [downloading, setDownloading] = useState(false);
  const [restoring, setRestoring] = useState(false);
  const [showRestartDialog, setShowRestartDialog] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDownload = async () => {
    try {
      setDownloading(true);

      const headers: Record<string, string> = {};
      const token = getAuthToken();
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const response = await fetch("/api/backup", { headers });
      if (!response.ok) {
        throw new Error(`Download failed: ${response.statusText}`);
      }

      const blob = await response.blob();
      const disposition = response.headers.get("Content-Disposition");
      const filenameMatch = disposition?.match(/filename="(.+)"/);
      const filename = filenameMatch?.[1] ?? "b4-backup.tar.gz";

      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      showSuccess("Backup downloaded successfully");
    } catch (error) {
      showError(
        error instanceof Error ? error.message : "Failed to download backup",
      );
    } finally {
      setDownloading(false);
    }
  };

  const handleRestore = async (file: File) => {
    try {
      setRestoring(true);

      const formData = new FormData();
      formData.append("file", file);

      const headers: Record<string, string> = {};
      const token = getAuthToken();
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const response = await fetch("/api/backup/restore", {
        method: "POST",
        headers,
        body: formData,
      });

      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(
          (data as { error?: string }).error ||
            `Restore failed: ${response.statusText}`,
        );
      }

      showSuccess("Backup restored successfully. Restart B4 to apply changes.");
      setShowRestartDialog(true);
    } catch (error) {
      showError(
        error instanceof Error ? error.message : "Failed to restore backup",
      );
    } finally {
      setRestoring(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      void handleRestore(file);
    }
  };

  return (
    <Stack spacing={3}>
      <B4Alert icon={<BackupIcon />}>
        Create a backup of your B4 configuration directory or restore from a
        previous backup. Backups include configuration, discovery history,
        detector history, device aliases, and captured payloads. Geodat files
        (.dat) and OUI database are excluded.
      </B4Alert>

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Section
            title="Download Backup"
            description="Generate and download a tar.gz archive of your B4 configuration"
            icon={<DownloadIcon />}
          >
            <Stack spacing={2}>
              <Typography variant="body2" sx={{ color: colors.text.secondary }}>
                The backup will include all files from the configuration
                directory except geodat (.dat) files and the OUI database.
              </Typography>
              <Box>
                <Button
                  variant="contained"
                  startIcon={<DownloadIcon />}
                  onClick={() => {
                    void handleDownload();
                  }}
                  disabled={downloading}
                >
                  {downloading ? "Generating..." : "Download Backup"}
                </Button>
              </Box>
            </Stack>
          </B4Section>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <B4Section
            title="Restore Backup"
            description="Upload a previously downloaded backup to restore your configuration"
            icon={<UploadIcon />}
          >
            <Stack spacing={2}>
              <Typography variant="body2" sx={{ color: colors.text.secondary }}>
                Upload a .tar.gz backup file. Files will be extracted into the
                configuration directory. A service restart is recommended after
                restoring.
              </Typography>
              <Box>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".tar.gz,.tgz"
                  style={{ display: "none" }}
                  onChange={handleFileSelect}
                />
                <Button
                  variant="outlined"
                  startIcon={<UploadIcon />}
                  onClick={() => fileInputRef.current?.click()}
                  disabled={restoring}
                >
                  {restoring ? "Restoring..." : "Upload & Restore"}
                </Button>
              </Box>
            </Stack>
          </B4Section>
        </Grid>
      </Grid>

      <RestartDialog
        open={showRestartDialog}
        onClose={() => setShowRestartDialog(false)}
      />
    </Stack>
  );
};
