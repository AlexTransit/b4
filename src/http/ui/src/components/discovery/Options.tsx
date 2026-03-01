import { useState, useEffect } from "react";
import {
  Box,
  Stack,
  Typography,
  Collapse,
  Autocomplete,
  Paper,
  ToggleButtonGroup,
  ToggleButton,
  Button,
} from "@mui/material";
import { FilterIcon, ExpandIcon, CollapseIcon } from "@b4.icons";
import {
  B4Badge,
  B4FormGroup,
  B4Slider,
  B4Switch,
  B4TextField,
} from "@b4.elements";
import { colors } from "@design";
import { Capture } from "@b4.capture";

export type TLSVersion = "auto" | "tls12" | "tls13";

export interface DiscoveryOptions {
  skipDNS: boolean;
  skipCache: boolean;
  payloadFiles: string[];
  validationTries: number;
  tlsVersion: TLSVersion;
}

interface DiscoveryOptionsPanelProps {
  options: DiscoveryOptions;
  onChange: (options: DiscoveryOptions) => void;
  onClearCache?: () => void;
  captures: Capture[];
  disabled?: boolean;
}

export const DiscoveryOptionsPanel = ({
  options,
  onChange,
  onClearCache,
  captures,
  disabled = false,
}: DiscoveryOptionsPanelProps) => {
  const [expanded, setExpanded] = useState(() => {
    return localStorage.getItem("b4_discovery_options_expanded") === "true";
  });

  useEffect(() => {
    localStorage.setItem("b4_discovery_options_expanded", String(expanded));
  }, [expanded]);

  const tlsCaptures = captures.filter((c) => c.protocol === "tls");
  const hasOptions =
    options.skipDNS ||
    options.skipCache ||
    options.payloadFiles.length > 0 ||
    options.validationTries > 1 ||
    options.tlsVersion !== "auto";

  return (
    <Box
      sx={{
        border: `1px solid ${colors.border.default}`,
        borderRadius: 1,
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <Box
        onClick={() => setExpanded((e) => !e)}
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          p: 1.5,
          cursor: "pointer",
          bgcolor: colors.background.dark,
          "&:hover": { bgcolor: colors.accent.primary },
        }}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <FilterIcon sx={{ fontSize: 18, color: colors.text.secondary }} />
          <Typography variant="body2" sx={{ color: colors.text.secondary }}>
            Discovery Options
          </Typography>
          {!expanded && hasOptions && (
            <B4Badge
              label={getOptionsSummary(options)}
              sx={{
                height: 20,
                fontSize: "0.7rem",
                bgcolor: colors.accent.secondary,
                color: colors.secondary,
              }}
            />
          )}
        </Stack>
        {expanded ? (
          <CollapseIcon sx={{ fontSize: 18, color: colors.text.secondary }} />
        ) : (
          <ExpandIcon sx={{ fontSize: 18, color: colors.text.secondary }} />
        )}
      </Box>

      {/* Content */}
      <Collapse in={expanded}>
        <Paper
          sx={{
            p: 3,
            bgcolor: colors.background.paper,
            border: `1px solid ${colors.border.default}`,
            display: "flex",
            flexDirection: "column",
          }}
          variant="outlined"
        >
          <B4FormGroup label="Discovery Options" columns={2}>
            <B4Switch
              label="Skip DNS Discovery"
              checked={options.skipDNS}
              onChange={(checked) => onChange({ ...options, skipDNS: checked })}
              disabled={disabled}
            />

            {/* Cache Controls */}
            <Box>
              <B4Switch
                label="Skip Cached Strategies"
                checked={options.skipCache}
                onChange={(checked) =>
                  onChange({ ...options, skipCache: checked })
                }
                disabled={disabled}
              />
              <Stack
                direction="row"
                alignItems="center"
                spacing={1}
                sx={{ mt: 0.5 }}
              >
                <Typography variant="caption" color="text.secondary">
                  Previously successful configurations are tested first.
                </Typography>
                {onClearCache && (
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={onClearCache}
                    disabled={disabled}
                  >
                    Clear Cache
                  </Button>
                )}
              </Stack>
            </Box>

            {/* TLS Version */}
            <Box>
              <Typography variant="body1" sx={{ mb: 1 }}>
                TLS Version
              </Typography>
              <Typography
                variant="caption"
                color="text.secondary"
                sx={{ mb: 1, display: "block" }}
              >
                TLS version used for discovery probes. Some DPI systems handle
                TLS 1.2 and 1.3 differently.
              </Typography>
              <ToggleButtonGroup
                value={options.tlsVersion}
                exclusive
                onChange={(_, value) => {
                  if (value !== null) {
                    onChange({ ...options, tlsVersion: value as TLSVersion });
                  }
                }}
                disabled={disabled}
                size="small"
                sx={{
                  "& .MuiToggleButton-root": {
                    color: colors.text.secondary,
                    borderColor: colors.border.default,
                    textTransform: "none",
                    px: 2,
                    "&.Mui-selected": {
                      bgcolor: colors.accent.secondary,
                      color: colors.secondary,
                      borderColor: colors.secondary,
                      "&:hover": { bgcolor: colors.accent.secondary },
                    },
                  },
                }}
              >
                <ToggleButton value="auto">Auto</ToggleButton>
                <ToggleButton value="tls12">TLS 1.2</ToggleButton>
                <ToggleButton value="tls13">TLS 1.3</ToggleButton>
              </ToggleButtonGroup>
            </Box>

            {/* Custom Payloads */}
            {tlsCaptures.length > 0 && (
              <Box>
                <Typography variant="body1" sx={{ mb: 1 }}>
                  Custom Payloads
                </Typography>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  sx={{ mb: 1, display: "block" }}
                >
                  Test with generated TLS ClientHello (SNI-first) instead of
                  built-in payloads
                </Typography>
                <Autocomplete
                  multiple
                  size="small"
                  options={tlsCaptures.map((c) => c.domain)}
                  value={options.payloadFiles}
                  onChange={(_, newValue) =>
                    onChange({ ...options, payloadFiles: newValue })
                  }
                  disabled={disabled}
                  renderInput={(params) => (
                    <B4TextField
                      {...params}
                      placeholder={
                        options.payloadFiles.length === 0
                          ? "Select captured payloads..."
                          : ""
                      }
                      size="small"
                    />
                  )}
                  renderValue={(value, getTagProps) =>
                    value.map((domain, index) => (
                      <B4Badge
                        {...getTagProps({ index })}
                        key={domain}
                        label={domain}
                        size="small"
                        sx={{
                          bgcolor: colors.accent.secondary,
                          border: `1px solid ${colors.secondary}`,
                        }}
                      />
                    ))
                  }
                />
              </Box>
            )}
            {/* Validation Tries */}
            <Box>
              <B4Slider
                label="Validation Tries"
                value={options.validationTries}
                onChange={(value: number) =>
                  onChange({ ...options, validationTries: value })
                }
                min={1}
                max={5}
                step={1}
                helperText="Number of successful connection attempts required to validate a preset"
              />
            </Box>

            {tlsCaptures.length === 0 && (
              <Typography variant="caption" color="text.secondary">
                No captured payloads available.{" "}
                <a href="/settings/capture" style={{ color: colors.secondary }}>
                  Capture payloads
                </a>{" "}
                to test with custom TLS ClientHello.
              </Typography>
            )}
          </B4FormGroup>
        </Paper>
      </Collapse>
    </Box>
  );
};

function getOptionsSummary(options: DiscoveryOptions): string {
  const parts: string[] = [];
  if (options.skipDNS) parts.push("Skip DNS");
  if (options.skipCache) parts.push("Skip Cache");
  if (options.tlsVersion === "tls12") parts.push("TLS 1.2");
  if (options.tlsVersion === "tls13") parts.push("TLS 1.3");
  if (options.validationTries > 1)
    parts.push(`${options.validationTries} tries`);
  if (options.payloadFiles.length > 0) {
    parts.push(
      `${options.payloadFiles.length} payload${
        options.payloadFiles.length > 1 ? "s" : ""
      }`,
    );
  }
  return parts.join(", ");
}
