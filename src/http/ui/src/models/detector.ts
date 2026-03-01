export type DetectorTestType = "dns" | "domains" | "tcp";
export type SuiteStatus = "pending" | "running" | "complete" | "failed" | "canceled";

// DNS types
export type DNSStatus = "OK" | "DNS_SPOOFING" | "DNS_INTERCEPTION" | "TIMEOUT" | "BLOCKED";

export interface DNSDomainResult {
  domain: string;
  doh_ip: string;
  udp_ip: string;
  status: DNSStatus;
  is_stub_ip?: boolean;
}

export interface DNSResult {
  status: DNSStatus;
  doh_server: string;
  udp_server: string;
  doh_blocked: boolean;
  udp_blocked: boolean;
  stub_ips?: string[];
  domains: DNSDomainResult[];
  summary: string;
  spoof_count: number;
  intercept_count: number;
  ok_count: number;
}

// Domain accessibility types
export type DomainStatus =
  | "OK"
  | "TLS_DPI"
  | "TLS_MITM"
  | "ISP_PAGE"
  | "BLOCKED"
  | "DNS_FAKE"
  | "TIMEOUT"
  | "ERROR";

export interface TLSProbeResult {
  status: DomainStatus;
  detail?: string;
  latency_ms: number;
}

export interface HTTPProbeResult {
  status: DomainStatus;
  detail?: string;
  status_code?: number;
  redirect_to?: string;
}

export interface DomainCheckResult {
  domain: string;
  ip: string;
  tls13?: TLSProbeResult;
  tls12?: TLSProbeResult;
  http?: HTTPProbeResult;
  is_fake_ip?: boolean;
  overall: DomainStatus;
}

export interface DomainsResult {
  domains: DomainCheckResult[];
  blocked_count: number;
  ok_count: number;
  dpi_count: number;
  summary: string;
}

// TCP 16-20KB types
export type TCPStatus = "OK" | "DETECTED" | "MIXED" | "TIMEOUT" | "ERROR";

export interface TCPTarget {
  id: string;
  url: string;
  asn: string;
  provider: string;
  country: string;
}

export interface TCPTargetResult {
  target: TCPTarget;
  status: TCPStatus;
  bytes_read: number;
  drop_at_kb?: number;
  detail?: string;
}

export interface TCPResult {
  targets: TCPTargetResult[];
  detected_count: number;
  ok_count: number;
  summary: string;
}

// Overall suite
export interface DetectorSuite {
  id: string;
  status: SuiteStatus;
  start_time: string;
  end_time?: string;
  tests: DetectorTestType[];
  current_test?: DetectorTestType;
  total_checks: number;
  completed_checks: number;
  dns_result?: DNSResult;
  domains_result?: DomainsResult;
  tcp_result?: TCPResult;
}

export interface DetectorResponse {
  id: string;
  tests: string[];
  estimated_tests: number;
  message: string;
}
