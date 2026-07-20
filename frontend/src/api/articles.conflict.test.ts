import { describe, expect, it } from "vitest";
import { isArticleVersionConflict } from "./articles";

describe("isArticleVersionConflict", () => {
  it("detects 409 version_conflict", () => {
    const err = {
      response: {
        status: 409,
        data: {
          error: {
            code: "version_conflict",
            message: "conflict",
            currentUpdatedAt: "2026-01-01T00:00:00Z",
          },
        },
      },
    };
    const r = isArticleVersionConflict(err);
    expect(r.conflict).toBe(true);
    expect(r.currentUpdatedAt).toBe("2026-01-01T00:00:00Z");
  });

  it("ignores non-409", () => {
    expect(isArticleVersionConflict({ response: { status: 400 } }).conflict).toBe(false);
  });
});
