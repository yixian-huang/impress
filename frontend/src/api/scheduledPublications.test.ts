import { describe, expect, it, vi, beforeEach } from "vitest";
import { http } from "@/api/http";
import {
  cancelScheduledPublication,
  createScheduledPublication,
  dateTimeLocalToISOString,
  getResourceScheduledPublication,
  listScheduledPublications,
  retryScheduledPublication,
  toDateTimeLocalValue,
  updateScheduledPublication,
} from "./scheduledPublications";

vi.mock("@/api/http", () => ({
  http: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

describe("scheduled publications api", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("lists scheduled publications with filters", async () => {
    vi.mocked(http.get).mockResolvedValue({
      data: { items: [], total: 0, page: 1, pageSize: 20 },
    } as never);

    await listScheduledPublications({
      page: 2,
      pageSize: 20,
      status: "pending",
      resourceType: "article",
      resourceId: 12,
    });

    expect(http.get).toHaveBeenCalledWith(
      "/admin/scheduled-publications?page=2&pageSize=20&status=pending&resourceType=article&resourceId=12",
    );
  });

  it("selects the active resource schedule from list results", async () => {
    vi.mocked(http.get).mockImplementation(async (url) => ({
      data: {
        items: String(url).includes("status=pending")
          ? [{ id: 3, status: "pending" }]
          : String(url).includes("status=failed")
            ? [{ id: 2, status: "failed" }]
            : [],
        total: 1,
        page: 1,
        pageSize: 1,
      },
    } as never));

    const item = await getResourceScheduledPublication("page", 7);

    expect(http.get).toHaveBeenCalledTimes(3);
    expect(http.get).toHaveBeenCalledWith(
      "/admin/scheduled-publications?page=1&pageSize=1&status=pending&resourceType=page&resourceId=7",
    );
    expect(http.get).toHaveBeenCalledWith(
      "/admin/scheduled-publications?page=1&pageSize=1&status=running&resourceType=page&resourceId=7",
    );
    expect(http.get).toHaveBeenCalledWith(
      "/admin/scheduled-publications?page=1&pageSize=1&status=failed&resourceType=page&resourceId=7",
    );
    expect(item?.id).toBe(3);
  });

  it("uses create, update, cancel, and retry endpoints", async () => {
    vi.mocked(http.post).mockResolvedValue({ data: { id: 1 } } as never);
    vi.mocked(http.put).mockResolvedValue({ data: { id: 1 } } as never);
    vi.mocked(http.delete).mockResolvedValue({ data: { id: 1 } } as never);

    await createScheduledPublication({
      resourceType: "article",
      resourceId: 1,
      scheduledAt: "2026-07-17T02:00:00.000Z",
      publishPayload: { status: "published" },
    });
    await updateScheduledPublication(1, { scheduledAt: "2026-07-18T02:00:00.000Z" });
    await cancelScheduledPublication(1);
    await retryScheduledPublication(1);

    expect(http.post).toHaveBeenNthCalledWith(1, "/admin/scheduled-publications", {
      resourceType: "article",
      resourceId: 1,
      scheduledAt: "2026-07-17T02:00:00.000Z",
      publishPayload: { status: "published" },
    });
    expect(http.put).toHaveBeenCalledWith("/admin/scheduled-publications/1", {
      scheduledAt: "2026-07-18T02:00:00.000Z",
    });
    expect(http.delete).toHaveBeenCalledWith("/admin/scheduled-publications/1");
    expect(http.post).toHaveBeenNthCalledWith(2, "/admin/scheduled-publications/1/retry");
  });

  it("converts local datetime input values", () => {
    const iso = dateTimeLocalToISOString("2026-07-17T10:30");

    expect(iso).toMatch(/^2026-07-17T/);
    expect(toDateTimeLocalValue(iso)).toBe("2026-07-17T10:30");
  });
});
