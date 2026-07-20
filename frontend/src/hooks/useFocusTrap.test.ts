import { describe, expect, it } from "vitest";

// useFocusTrap is DOM-integration heavy; cover the focusable selector contract lightly
// via a pure helper re-export if needed later. Smoke: module loads.
import { useFocusTrap } from "./useFocusTrap";

describe("useFocusTrap", () => {
  it("exports a hook function", () => {
    expect(typeof useFocusTrap).toBe("function");
  });
});
