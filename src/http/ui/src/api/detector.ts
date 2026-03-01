import { apiPost, apiGet, apiDelete } from "./apiClient";
import type { DetectorResponse, DetectorSuite, DetectorTestType } from "@models/detector";

export const detectorApi = {
  start: (tests: DetectorTestType[]) =>
    apiPost<DetectorResponse>("/api/detector/start", { tests }),
  status: (id: string) =>
    apiGet<DetectorSuite>(`/api/detector/status/${id}`),
  cancel: (id: string) =>
    apiDelete(`/api/detector/cancel/${id}`),
};
