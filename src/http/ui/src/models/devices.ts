import { B4Config, Device } from "@b4.settings";

export interface DeviceInfo {
  mac: string;
  ip: string;
  hostname: string;
  vendor: string;
  alias?: string;
  country: string;
  is_manual?: boolean;
}

export interface DevicesResponse {
  available: boolean;
  source?: string;
  devices: DeviceInfo[];
}

export interface DevicesSettingsProps {
  config: B4Config;
  onChange: (
    field: string,
    value: boolean | string | string[] | number | Device[],
  ) => void;
}
