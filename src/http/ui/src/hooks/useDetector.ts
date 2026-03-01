import { useState, useCallback, useEffect, useRef } from "react";
import { ApiError } from "@api/apiClient";
import { detectorApi } from "@api/detector";
import type { DetectorSuite, DetectorTestType } from "@models/detector";

export function useDetector() {
  const [running, setRunning] = useState(false);
  const [suiteId, setSuiteId] = useState<string | null>(null);
  const [suite, setSuite] = useState<DetectorSuite | null>(null);
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    const saved = localStorage.getItem("detector_suiteId");
    if (saved) {
      setSuiteId(saved);
      setRunning(true);
    }
  }, []);

  useEffect(() => {
    if (suiteId) {
      localStorage.setItem("detector_suiteId", suiteId);
    }
  }, [suiteId]);

  useEffect(() => {
    if (!suiteId || !running) return;

    const fetchStatus = async () => {
      try {
        const data = await detectorApi.status(suiteId);
        setSuite(data);
        if (["complete", "failed", "canceled"].includes(data.status)) {
          setRunning(false);
          localStorage.removeItem("detector_suiteId");
        }
      } catch (e) {
        if (e instanceof ApiError && e.status === 404) {
          setRunning(false);
          localStorage.removeItem("detector_suiteId");
          setSuiteId(null);
          return;
        }
        setError(e instanceof Error ? e.message : "Unknown error");
        setRunning(false);
      }
    };

    pollRef.current = setInterval(() => void fetchStatus(), 1500);
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [suiteId, running]);

  const startDetector = useCallback(
    async (tests: DetectorTestType[]) => {
      setError(null);
      setSuite(null);
      setRunning(true);
      try {
        const res = await detectorApi.start(tests);
        setSuiteId(res.id);
      } catch (e) {
        setRunning(false);
        setError(e instanceof Error ? e.message : "Failed to start detector");
      }
    },
    [],
  );

  const cancelDetector = useCallback(async () => {
    if (!suiteId) return;
    try {
      await detectorApi.cancel(suiteId);
      setRunning(false);
    } catch (e) {
      console.error("Failed to cancel detector:", e);
    }
  }, [suiteId]);

  const resetDetector = useCallback(() => {
    localStorage.removeItem("detector_suiteId");
    setSuiteId(null);
    setSuite(null);
    setError(null);
    setRunning(false);
  }, []);

  return {
    running,
    suiteId,
    suite,
    error,
    startDetector,
    cancelDetector,
    resetDetector,
  };
}
