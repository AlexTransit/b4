import { ReactNode, useMemo, useState } from "react";
import {
  Box,
  Checkbox,
  Chip,
  InputAdornment,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import ClearIcon from "@mui/icons-material/Clear";
import IconButton from "@mui/material/IconButton";
import { B4Badge } from "@b4.elements";
import { DeviceInfo } from "@b4.devices";
import { colors } from "@design";
import { sortDevices } from "@utils";
import { useTranslation } from "react-i18next";

export interface B4DeviceTableColumn {
  header: string;
  renderCell: (device: DeviceInfo) => ReactNode;
}

interface B4DeviceTableProps {
  devices: DeviceInfo[];
  loading: boolean;
  isSelected: (mac: string) => boolean;
  onToggle: (mac: string) => void;
  onSelectAll: (checked: boolean) => void;
  renderNameCell?: (device: DeviceInfo) => ReactNode;
  extraColumns?: B4DeviceTableColumn[];
  showOfflineChip?: boolean;
  maxHeight?: number;
}

export const B4DeviceTable = ({
  devices,
  loading,
  isSelected,
  onToggle,
  onSelectAll,
  renderNameCell,
  extraColumns = [],
  showOfflineChip = false,
  maxHeight = 350,
}: B4DeviceTableProps) => {
  const { t } = useTranslation();
  const [filterText, setFilterText] = useState("");

  const normalizedFilter = filterText.trim().toLowerCase();

  const filteredDevices = useMemo(() => {
    if (!normalizedFilter) return devices;
    return devices.filter((d) => {
      const name = d.alias || d.vendor || d.hostname || "";
      return (
        d.mac?.toLowerCase().includes(normalizedFilter) ||
        d.ip?.toLowerCase().includes(normalizedFilter) ||
        name.toLowerCase().includes(normalizedFilter)
      );
    });
  }, [devices, normalizedFilter]);

  const selectedCount = filteredDevices.filter((d) => isSelected(d.mac)).length;
  const allSelected = filteredDevices.length > 0 && selectedCount === filteredDevices.length;
  const someSelected = selectedCount > 0 && !allSelected;

  const handleSelectAll = (checked: boolean) => {
    if (!normalizedFilter) {
      onSelectAll(checked);
      return;
    }
    filteredDevices.forEach((d) => {
      const currentlySelected = isSelected(d.mac);
      if (checked && !currentlySelected) onToggle(d.mac);
      if (!checked && currentlySelected) onToggle(d.mac);
    });
  };

  const headers = [
    t("core.devices.macAddress"),
    t("core.devices.ip"),
    t("core.devices.deviceName"),
    ...extraColumns.map((c) => c.header),
  ];

  return (
    <Box>
      <TextField
        fullWidth
        size="small"
        placeholder={t("core.devices.filterPlaceholder")}
        value={filterText}
        onChange={(e) => setFilterText(e.target.value)}
        sx={{ mb: 1 }}
        slotProps={{
          input: {
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon fontSize="small" />
              </InputAdornment>
            ),
            endAdornment: filterText ? (
              <InputAdornment position="end">
                <IconButton
                  size="small"
                  aria-label="clear filter"
                  onClick={() => setFilterText("")}
                >
                  <ClearIcon fontSize="small" />
                </IconButton>
              </InputAdornment>
            ) : undefined,
          },
        }}
      />
      <TableContainer
        component={Paper}
        sx={{
          bgcolor: colors.background.paper,
          border: `1px solid ${colors.border.default}`,
          maxHeight,
        }}
      >
        <Table size="small" stickyHeader>
          <TableHead>
            <TableRow>
              <TableCell
                padding="checkbox"
                sx={{ bgcolor: colors.background.dark }}
              >
                <Checkbox
                  color="secondary"
                  indeterminate={someSelected}
                  checked={allSelected}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                />
              </TableCell>
              {headers.map((label) => (
                <TableCell
                  key={label}
                  sx={{
                    bgcolor: colors.background.dark,
                    color: colors.text.secondary,
                  }}
                >
                  {label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredDevices.length === 0 ? (
              <TableRow>
                <TableCell colSpan={headers.length + 1} align="center">
                  {loading
                    ? t("core.devices.loadingDevices")
                    : t("core.devices.noDevices")}
                </TableCell>
              </TableRow>
            ) : (
              sortDevices(filteredDevices, isSelected).map((device) => (
                <TableRow
                  key={device.mac}
                  hover
                  onClick={() => onToggle(device.mac)}
                  sx={{ cursor: "pointer" }}
                >
                  <TableCell padding="checkbox">
                    <Checkbox
                      checked={isSelected(device.mac)}
                      color="secondary"
                      onChange={(event) => {
                        event.stopPropagation();
                        onToggle(device.mac);
                      }}
                    />
                  </TableCell>
                  <TableCell
                    sx={{ fontFamily: "monospace", fontSize: "0.85rem" }}
                  >
                    {device.is_manual ? (
                      <Typography variant="caption" color="text.secondary">
                        {t("core.devices.byIp")}
                      </Typography>
                    ) : (
                      device.mac
                    )}
                  </TableCell>
                  <TableCell
                    sx={{ fontFamily: "monospace", fontSize: "0.85rem" }}
                  >
                    <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                      {device.ip}
                      {device.is_manual && (
                        <Chip
                          label={t("core.devices.manual")}
                          size="small"
                          variant="outlined"
                          sx={{ fontSize: "0.7rem", height: 20 }}
                        />
                      )}
                      {showOfflineChip &&
                        !device.is_manual &&
                        device.is_online === false && (
                          <Chip
                            label={t("core.devices.offline")}
                            size="small"
                            variant="outlined"
                            sx={{
                              fontSize: "0.7rem",
                              height: 20,
                              color: colors.text.secondary,
                            }}
                          />
                        )}
                    </Box>
                  </TableCell>
                  {renderNameCell ? (
                    <TableCell onClick={(e) => e.stopPropagation()}>
                      {renderNameCell(device)}
                    </TableCell>
                  ) : (
                    <TableCell>
                      <B4Badge
                        label={
                          device.alias ||
                          device.vendor ||
                          device.hostname ||
                          t("core.unknown")
                        }
                        color="primary"
                        variant={isSelected(device.mac) ? "filled" : "outlined"}
                      />
                    </TableCell>
                  )}
                  {extraColumns.map((col) => (
                    <TableCell
                      key={col.header}
                      onClick={(e) => e.stopPropagation()}
                    >
                      {col.renderCell(device)}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};
