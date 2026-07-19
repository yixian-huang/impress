import { http } from "@/api/http";
import { getStoredAccessToken } from "@/lib/browserStorage";

const apiBaseURL = (import.meta.env.VITE_API_BASE_URL || "").trim().replace(/\/+$/, "");

export type MigrationFormat = "wordpress" | "halo" | "markdown";
export type MigrationJobPhase = "parsing" | "importing" | "done" | "failed";

export interface MigrationJob {
  jobId: string;
  source: MigrationFormat;
  phase: MigrationJobPhase;
  total: number;
  processed: number;
  succeeded: number;
  failed: number;
  errors: string[];
  attempt: number;
  retryable: boolean;
  startedAt: string;
  finishedAt?: string;
}

export interface MigrationJobsResponse {
  jobs: MigrationJob[];
}

export async function importData(
  file: File,
  format: MigrationFormat
): Promise<{ jobId: string; message: string; totalArticles: number; parseErrors: string[] }> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("source", format);
  const response = await http.post<{ jobId: string; message: string; totalArticles: number; parseErrors: string[] }>(
    "/admin/migration/import",
    formData,
    {}
  );
  return response.data;
}

export async function getMigrationJobs(): Promise<MigrationJob[]> {
  const response = await http.get<MigrationJobsResponse>("/admin/migration/jobs");
  return response.data.jobs || [];
}

export async function getMigrationJob(jobId: string): Promise<MigrationJob> {
  const response = await http.get<MigrationJob>(`/admin/migration/jobs/${jobId}`);
  return response.data;
}

export async function retryMigrationJob(jobId: string): Promise<MigrationJob> {
  const response = await http.post<MigrationJob>(`/admin/migration/jobs/${jobId}/retry`);
  return response.data;
}

export async function streamMigrationJob(
  jobId: string,
  signal: AbortSignal,
  onProgress: (job: Partial<MigrationJob>) => void,
): Promise<void> {
  const accessToken = getStoredAccessToken();
  const response = await fetch(
    `${apiBaseURL}/admin/migration/jobs/${encodeURIComponent(jobId)}/stream`,
    {
      headers: {
        Accept: "text/event-stream",
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      },
      signal,
    },
  );
  if (!response.ok) {
    throw new Error(`migration stream failed with status ${response.status}`);
  }
  if (!response.body) {
    throw new Error("migration stream response has no body");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let eventName = "message";
  let dataLines: string[] = [];

  const dispatch = () => {
    if (dataLines.length === 0) {
      eventName = "message";
      return;
    }
    if (eventName === "message" || eventName === "progress") {
      onProgress(JSON.parse(dataLines.join("\n")) as Partial<MigrationJob>);
    }
    eventName = "message";
    dataLines = [];
  };

  const consumeLine = (rawLine: string) => {
    const line = rawLine.endsWith("\r") ? rawLine.slice(0, -1) : rawLine;
    if (line === "") {
      dispatch();
    } else if (line.startsWith("event:")) {
      eventName = line.slice(6).trimStart();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice(5).trimStart());
    }
  };

  while (true) {
    const { done, value } = await reader.read();
    buffer += decoder.decode(value, { stream: !done });
    let newlineIndex = buffer.indexOf("\n");
    while (newlineIndex >= 0) {
      consumeLine(buffer.slice(0, newlineIndex));
      buffer = buffer.slice(newlineIndex + 1);
      newlineIndex = buffer.indexOf("\n");
    }
    if (done) {
      if (buffer !== "") consumeLine(buffer);
      dispatch();
      return;
    }
  }
}
