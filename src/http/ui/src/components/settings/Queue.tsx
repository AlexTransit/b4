import { useTranslation } from "react-i18next";
import { NetworkIcon } from "@b4.icons";
import { B4FormGroup, B4Section, B4TextField, B4Slider } from "@b4.elements";
import { B4Config } from "@models/config";

interface QueueSettingsProps {
  config: B4Config;
  onChange: (
    field: string,
    value: number | boolean | string | string[],
  ) => void;
}

export const QueueSettings = ({ config, onChange }: QueueSettingsProps) => {
  const { t } = useTranslation();

  return (
    <B4Section
      title={t("settings.Queue.title")}
      description={t("settings.Queue.description")}
      icon={<NetworkIcon />}
    >
      <B4FormGroup label={t("settings.Queue.groupLabel")} columns={2}>
        <B4TextField
          label={t("settings.Queue.queueStart")}
          type="number"
          value={config.queue.start_num}
          onChange={(e) => onChange("queue.start_num", Number(e.target.value))}
          helperText={t("settings.Queue.queueStartHelp")}
        />
        <B4TextField
          label={t("settings.Queue.packetMark")}
          type="number"
          value={config.queue.mark}
          onChange={(e) => onChange("queue.mark", Number(e.target.value))}
          helperText={t("settings.Queue.packetMarkHelp")}
        />
        <B4Slider
          label={t("settings.Queue.workerThreads")}
          value={config.queue.threads}
          onChange={(value) => onChange("queue.threads", value)}
          min={1}
          max={16}
          step={1}
          helperText={t("settings.Queue.workerThreadsHelp")}
        />
        <B4Slider
          label={t("settings.Queue.tcpLimit")}
          value={config.queue.tcp_conn_bytes_limit}
          onChange={(value) => onChange("queue.tcp_conn_bytes_limit", value)}
          min={1}
          max={100}
          step={1}
          helperText={t("settings.Queue.tcpLimitHelp")}
        />
        <B4Slider
          label={t("settings.Queue.udpLimit")}
          value={config.queue.udp_conn_bytes_limit}
          onChange={(value) => onChange("queue.udp_conn_bytes_limit", value)}
          min={1}
          max={30}
          step={1}
          helperText={t("settings.Queue.udpLimitHelp")}
        />
      </B4FormGroup>
    </B4Section>
  );
};
