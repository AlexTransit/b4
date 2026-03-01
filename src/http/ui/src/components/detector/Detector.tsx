import { useState, useCallback } from "react";
import {
  Box,
  Button,
  Collapse,
  Stack,
  Typography,
  LinearProgress,
  Paper,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@mui/material";
import {
  StartIcon,
  StopIcon,
  RefreshIcon,
  SecurityIcon,
  DnsIcon,
  DomainIcon,
  NetworkIcon,
  CheckCircleIcon,
  WarningIcon,
  ErrorIcon,
  BlockIcon,
  ExpandIcon,
  CollapseIcon,
} from "@b4.icons";
import { colors } from "@design";
import { B4Alert, B4Badge, B4Section, B4Switch } from "@b4.elements";
import { useDetector } from "@hooks/useDetector";
import type {
  DetectorTestType,
  DNSStatus,
  DomainStatus,
  TCPStatus,
  DNSDomainResult,
  DomainCheckResult,
  TCPTargetResult,
} from "@models/detector";

const testNames: Record<DetectorTestType, string> = {
  dns: "DNS Integrity",
  domains: "Domain Accessibility",
  tcp: "TCP Connection Drop",
};

const testDescriptions: Record<DetectorTestType, string> = {
  dns: "Compares UDP DNS vs DoH to detect spoofing and interception",
  domains: "Probes blocked domains via TLS 1.3, TLS 1.2, and HTTP",
  tcp: "Downloads from CDN endpoints to detect TSPU connection drops at 16-20KB",
};

const testIcons: Record<DetectorTestType, React.ReactNode> = {
  dns: <DnsIcon fontSize="small" />,
  domains: <DomainIcon fontSize="small" />,
  tcp: <NetworkIcon fontSize="small" />,
};

function StatusChip({ status, label }: { status: string; label?: string }) {
  const display = label || status;
  let color: "primary" | "error" | "info" | "secondary" | "default" = "default";

  switch (status) {
    case "OK":
      color = "primary";
      break;
    case "DNS_SPOOFING":
    case "TLS_DPI":
    case "DETECTED":
    case "BLOCKED":
      color = "error";
      break;
    case "DNS_INTERCEPTION":
    case "TLS_MITM":
    case "ISP_PAGE":
    case "DNS_FAKE":
    case "MIXED":
      color = "secondary";
      break;
    case "TIMEOUT":
    case "ERROR":
      color = "info";
      break;
  }

  return <B4Badge label={display} size="small" color={color} />;
}

function SummaryIcon({ ok }: { ok: boolean }) {
  return ok ? (
    <CheckCircleIcon sx={{ color: colors.secondary, fontSize: 20 }} />
  ) : (
    <WarningIcon sx={{ color: "#f44336", fontSize: 20 }} />
  );
}

export const DetectorRunner = () => {
  const {
    running,
    suiteId,
    suite,
    error,
    startDetector,
    cancelDetector,
    resetDetector,
  } = useDetector();

  const [selectedTests, setSelectedTests] = useState<Record<DetectorTestType, boolean>>({
    dns: true,
    domains: true,
    tcp: true,
  });

  const isReconnecting = suiteId && running && !suite;

  const progress = suite
    ? Math.min((suite.completed_checks / Math.max(suite.total_checks, 1)) * 100, 100)
    : 0;

  const handleStart = useCallback(() => {
    const tests = (Object.entries(selectedTests) as [DetectorTestType, boolean][])
      .filter(([, v]) => v)
      .map(([k]) => k);
    if (tests.length > 0) {
      void startDetector(tests);
    }
  }, [selectedTests, startDetector]);

  const anyTestSelected = Object.values(selectedTests).some(Boolean);

  return (
    <Stack spacing={3}>
      <B4Section
        title="TSPU / DPI Detector"
        description="Detect ISP-level Deep Packet Inspection and blocking"
        icon={<SecurityIcon />}
      >
        <B4Alert icon={<SecurityIcon />}>
          <strong>DPI Detector:</strong> Runs diagnostic tests to detect TSPU
          (Technical System for Countering Threats) and ISP-level internet
          blocking. Tests DNS integrity, domain accessibility via TLS/HTTP,
          and TCP connection drops at characteristic byte thresholds.
        </B4Alert>

        {/* Test selection */}
        {!running && !suite && (
          <Stack spacing={1}>
            {(["dns", "domains", "tcp"] as DetectorTestType[]).map((test) => (
              <Box
                key={test}
                sx={{
                  display: "flex",
                  alignItems: "center",
                  gap: 2,
                  px: 2,
                  py: 1,
                  borderRadius: 1,
                  bgcolor: selectedTests[test]
                    ? colors.accent.primary
                    : "transparent",
                  border: `1px solid ${
                    selectedTests[test] ? colors.border.medium : colors.border.light
                  }`,
                }}
              >
                {testIcons[test]}
                <Box sx={{ flexGrow: 1 }}>
                  <Typography variant="body2" sx={{ fontWeight: 600 }}>
                    {testNames[test]}
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{ color: colors.text.secondary }}
                  >
                    {testDescriptions[test]}
                  </Typography>
                </Box>
                <B4Switch
                  label=""
                  checked={selectedTests[test]}
                  onChange={(checked) =>
                    setSelectedTests((prev) => ({ ...prev, [test]: checked }))
                  }
                />
              </Box>
            ))}
          </Stack>
        )}

        {/* Action buttons */}
        <Box sx={{ display: "flex", gap: 1 }}>
          {!running && !suite && (
            <Button
              startIcon={<StartIcon />}
              variant="contained"
              onClick={handleStart}
              disabled={!anyTestSelected}
            >
              Start Detection
            </Button>
          )}
          {(running || isReconnecting) && (
            <Button
              variant="outlined"
              color="secondary"
              startIcon={<StopIcon />}
              onClick={() => void cancelDetector()}
            >
              Cancel
            </Button>
          )}
          {suite && !running && (
            <Button
              variant="outlined"
              startIcon={<RefreshIcon />}
              onClick={resetDetector}
            >
              New Detection
            </Button>
          )}
        </Box>

        {error && <B4Alert severity="error">{error}</B4Alert>}

        {isReconnecting && (
          <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
            <CircularProgress size={20} sx={{ color: colors.secondary }} />
            <Typography variant="body2" sx={{ color: colors.text.secondary }}>
              Reconnecting to running detection...
            </Typography>
          </Box>
        )}

        {/* Progress */}
        {running && suite && (
          <Box>
            <Box sx={{ display: "flex", justifyContent: "space-between", mb: 1 }}>
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <Typography variant="body2" color="text.secondary">
                  {suite.current_test && (
                    <B4Badge
                      label={testNames[suite.current_test]}
                      size="small"
                      color="primary"
                      sx={{ mr: 1 }}
                    />
                  )}
                  {suite.completed_checks} of {suite.total_checks} checks
                </Typography>
              </Box>
              <Typography variant="body2" color="text.secondary">
                {progress.toFixed(0)}%
              </Typography>
            </Box>
            <LinearProgress
              variant="determinate"
              value={progress}
              sx={{
                height: 8,
                borderRadius: 4,
                bgcolor: colors.background.dark,
                "& .MuiLinearProgress-bar": {
                  bgcolor: colors.secondary,
                  borderRadius: 4,
                },
              }}
            />
          </Box>
        )}
      </B4Section>

      {/* DNS Results */}
      {suite?.dns_result && (
        <ResultSection
          title="DNS Integrity Check"
          icon={<DnsIcon />}
          summary={suite.dns_result.summary}
          ok={suite.dns_result.status === "OK"}
        >
          {suite.dns_result.doh_blocked && (
            <B4Alert severity="error">
              All DNS-over-HTTPS servers are blocked. Your ISP is filtering encrypted DNS.
            </B4Alert>
          )}
          {suite.dns_result.udp_blocked && (
            <B4Alert severity="error">
              All UDP DNS servers (port 53) are blocked.
            </B4Alert>
          )}
          {suite.dns_result.stub_ips && suite.dns_result.stub_ips.length > 0 && (
            <B4Alert severity="warning" icon={<WarningIcon />}>
              Stub/sinkhole IPs detected: {suite.dns_result.stub_ips.join(", ")}.
              Multiple blocked domains resolve to these IPs.
            </B4Alert>
          )}
          {suite.dns_result.domains && suite.dns_result.domains.length > 0 && (
            <DNSTable domains={suite.dns_result.domains} />
          )}
        </ResultSection>
      )}

      {/* Domain Results */}
      {suite?.domains_result && (
        <ResultSection
          title="Domain Accessibility"
          icon={<DomainIcon />}
          summary={suite.domains_result.summary}
          ok={suite.domains_result.blocked_count === 0}
        >
          {suite.domains_result.domains && suite.domains_result.domains.length > 0 && (
            <DomainsTable domains={suite.domains_result.domains} />
          )}
        </ResultSection>
      )}

      {/* TCP Results */}
      {suite?.tcp_result && (
        <ResultSection
          title="TCP Connection Drop Test"
          icon={<NetworkIcon />}
          summary={suite.tcp_result.summary}
          ok={suite.tcp_result.detected_count === 0}
        >
          {suite.tcp_result.targets && suite.tcp_result.targets.length > 0 && (
            <TCPTable targets={suite.tcp_result.targets} />
          )}
        </ResultSection>
      )}
    </Stack>
  );
};

// Result section wrapper
function ResultSection({
  title,
  icon,
  summary,
  ok,
  children,
}: {
  title: string;
  icon: React.ReactNode;
  summary: string;
  ok: boolean;
  children: React.ReactNode;
}) {
  const [expanded, setExpanded] = useState(true);

  return (
    <Paper
      elevation={0}
      sx={{
        bgcolor: colors.background.paper,
        border: `1px solid ${colors.border.default}`,
        borderRadius: 2,
        overflow: "hidden",
      }}
    >
      <Box
        onClick={() => setExpanded((v) => !v)}
        sx={{
          p: 2,
          bgcolor: colors.accent.primary,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          cursor: "pointer",
          userSelect: "none",
          "&:hover": { bgcolor: colors.accent.primaryHover },
        }}
      >
        <Box sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
          {icon}
          <Typography variant="h6" sx={{ color: colors.text.primary }}>
            {title}
          </Typography>
        </Box>
        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <SummaryIcon ok={ok} />
          <Typography
            variant="body2"
            sx={{
              color: ok ? colors.secondary : "#f44336",
              fontWeight: 600,
            }}
          >
            {summary}
          </Typography>
          {expanded ? (
            <CollapseIcon sx={{ color: colors.text.secondary, fontSize: 20 }} />
          ) : (
            <ExpandIcon sx={{ color: colors.text.secondary, fontSize: 20 }} />
          )}
        </Box>
      </Box>
      <Collapse in={expanded}>
        <Box sx={{ p: 2 }}>
          <Stack spacing={2}>{children}</Stack>
        </Box>
      </Collapse>
    </Paper>
  );
}

// DNS results table
function DNSTable({ domains }: { domains: DNSDomainResult[] }) {
  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <StyledHeaderCell>Domain</StyledHeaderCell>
            <StyledHeaderCell>DoH IP</StyledHeaderCell>
            <StyledHeaderCell>UDP IP</StyledHeaderCell>
            <StyledHeaderCell>Status</StyledHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {domains.map((d) => (
            <TableRow key={d.domain}>
              <StyledCell>
                {d.domain}
                {d.is_stub_ip && (
                  <B4Badge
                    label="STUB"
                    size="small"
                    color="error"
                    sx={{ ml: 1 }}
                  />
                )}
              </StyledCell>
              <StyledCell sx={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                {d.doh_ip}
              </StyledCell>
              <StyledCell sx={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                {d.udp_ip}
              </StyledCell>
              <StyledCell>
                <StatusChip status={d.status} />
              </StyledCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

// Domain accessibility results table
function DomainsTable({ domains }: { domains: DomainCheckResult[] }) {
  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <StyledHeaderCell>Domain</StyledHeaderCell>
            <StyledHeaderCell>IP</StyledHeaderCell>
            <StyledHeaderCell>TLS 1.3</StyledHeaderCell>
            <StyledHeaderCell>TLS 1.2</StyledHeaderCell>
            <StyledHeaderCell>HTTP</StyledHeaderCell>
            <StyledHeaderCell>Overall</StyledHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {domains.map((d) => (
            <TableRow key={d.domain}>
              <StyledCell>{d.domain}</StyledCell>
              <StyledCell sx={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                {d.ip || "-"}
              </StyledCell>
              <StyledCell>
                {d.tls13 ? (
                  <StatusChip status={d.tls13.status} />
                ) : (
                  "-"
                )}
              </StyledCell>
              <StyledCell>
                {d.tls12 ? (
                  <StatusChip status={d.tls12.status} />
                ) : (
                  "-"
                )}
              </StyledCell>
              <StyledCell>
                {d.http ? (
                  <StatusChip status={d.http.status} />
                ) : (
                  "-"
                )}
              </StyledCell>
              <StyledCell>
                <StatusChip status={d.overall} />
              </StyledCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

// TCP drop test results table
function TCPTable({ targets }: { targets: TCPTargetResult[] }) {
  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <StyledHeaderCell>#</StyledHeaderCell>
            <StyledHeaderCell>Provider</StyledHeaderCell>
            <StyledHeaderCell>ASN</StyledHeaderCell>
            <StyledHeaderCell>Country</StyledHeaderCell>
            <StyledHeaderCell>Status</StyledHeaderCell>
            <StyledHeaderCell>Detail</StyledHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {targets.map((t) => (
            <TableRow key={t.target.id}>
              <StyledCell>{t.target.id}</StyledCell>
              <StyledCell>{t.target.provider}</StyledCell>
              <StyledCell sx={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                {t.target.asn}
              </StyledCell>
              <StyledCell>{t.target.country}</StyledCell>
              <StyledCell>
                <StatusChip status={t.status} />
              </StyledCell>
              <StyledCell
                sx={{
                  fontSize: "0.8rem",
                  color: colors.text.secondary,
                  maxWidth: 300,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {t.detail || "-"}
              </StyledCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

// Styled table cells
function StyledHeaderCell({ children, ...props }: { children: React.ReactNode } & Record<string, unknown>) {
  return (
    <TableCell
      sx={{
        color: colors.text.secondary,
        borderBottom: `1px solid ${colors.border.default}`,
        fontSize: "0.75rem",
        textTransform: "uppercase",
        fontWeight: 600,
        whiteSpace: "nowrap",
        py: 1,
        ...((props.sx as object) || {}),
      }}
    >
      {children}
    </TableCell>
  );
}

function StyledCell({
  children,
  sx: sxProp,
  ...props
}: {
  children: React.ReactNode;
  sx?: Record<string, unknown>;
} & Record<string, unknown>) {
  return (
    <TableCell
      sx={{
        color: colors.text.primary,
        borderBottom: `1px solid ${colors.border.light}`,
        whiteSpace: "nowrap",
        py: 0.75,
        ...(sxProp || {}),
      }}
      {...props}
    >
      {children}
    </TableCell>
  );
}
