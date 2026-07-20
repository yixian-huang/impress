import { describe, expect, it } from "vitest";
import { resolveSaveStatus } from "./saveStatusUtils";

describe("resolveSaveStatus", () => {
  it("publish always returns published", () => {
    expect(resolveSaveStatus("publish", "draft")).toBe("published");
    expect(resolveSaveStatus("publish", "published")).toBe("published");
    expect(resolveSaveStatus("publish", "scheduled")).toBe("published");
  });

  it("preserves published on draft/autosave intents", () => {
    expect(resolveSaveStatus("draft", "published")).toBe("published");
    expect(resolveSaveStatus("autosave", "published")).toBe("published");
  });

  it("saves drafts as draft", () => {
    expect(resolveSaveStatus("draft", "draft")).toBe("draft");
    expect(resolveSaveStatus("autosave", "draft")).toBe("draft");
    expect(resolveSaveStatus("autosave", "scheduled")).toBe("draft");
  });
});
