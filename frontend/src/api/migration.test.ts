import { beforeEach, describe, expect, it, vi } from "vitest";
import { http } from "@/api/http";
import {
  getMigrationJob,
  getMigrationJobs,
  importData,
  retryMigrationJob,
  streamMigrationJob,
} from "./migration";

vi.mock("@/api/http", () => ({
  http: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

describe("migration api", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    const values = new Map<string, string>();
    vi.stubGlobal("localStorage", {
      getItem: (key: string) => values.get(key) ?? null,
      setItem: (key: string, value: string) => values.set(key, value),
      removeItem: (key: string) => values.delete(key),
      clear: () => values.clear(),
    });
  });

  it("uses migration job endpoints", async () => {
    vi.mocked(http.get).mockResolvedValueOnce({ data: { jobs: [] } } as never);
    vi.mocked(http.get).mockResolvedValueOnce({ data: { jobId: "mig-1" } } as never);
    vi.mocked(http.post).mockResolvedValueOnce({
      data: { jobId: "mig-1", message: "queued", totalArticles: 1, parseErrors: [] },
    } as never);
    vi.mocked(http.post).mockResolvedValueOnce({ data: { jobId: "mig-1" } } as never);

    await getMigrationJobs();
    await getMigrationJob("mig-1");
    await importData(new File(["content"], "export.zip"), "markdown");
    await retryMigrationJob("mig-1");

    expect(http.get).toHaveBeenNthCalledWith(1, "/admin/migration/jobs");
    expect(http.get).toHaveBeenNthCalledWith(2, "/admin/migration/jobs/mig-1");
    expect(http.post).toHaveBeenNthCalledWith(
      1,
      "/admin/migration/import",
      expect.any(FormData),
      {},
    );
    expect(http.post).toHaveBeenNthCalledWith(2, "/admin/migration/jobs/mig-1/retry");
  });

  it("streams progress with bearer authentication and no token query parameter", async () => {
    localStorage.setItem("accessToken", "secret-token");
    const fetchMock = vi.fn().mockResolvedValue(new Response(
      [
        'event: progress\ndata: {"jobId":"mig-1","phase":"importing","processed":1}\n\n',
        'data: {"jobId":"mig-1","phase":"done","processed":2}\n\n',
      ].join(""),
      {
        status: 200,
        headers: { "Content-Type": "text/event-stream" },
      },
    ));
    vi.stubGlobal("fetch", fetchMock);
    const updates: Array<{ phase?: string; processed?: number }> = [];

    await streamMigrationJob("mig-1", new AbortController().signal, (job) => updates.push(job));

    expect(fetchMock).toHaveBeenCalledWith(
      "/admin/migration/jobs/mig-1/stream",
      expect.objectContaining({
        headers: expect.objectContaining({
          Accept: "text/event-stream",
          Authorization: "Bearer secret-token",
        }),
      }),
    );
    expect(fetchMock.mock.calls[0][0]).not.toContain("secret-token");
    expect(updates).toEqual([
      expect.objectContaining({ phase: "importing", processed: 1 }),
      expect.objectContaining({ phase: "done", processed: 2 }),
    ]);
  });
});
