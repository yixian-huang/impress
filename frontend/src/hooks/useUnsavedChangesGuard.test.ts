import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useUnsavedChangesGuard } from "./useUnsavedChangesGuard";

describe("useUnsavedChangesGuard", () => {
  const confirmMock = vi.fn(() => true);

  beforeEach(() => {
    confirmMock.mockReset();
    confirmMock.mockReturnValue(true);
    // happy-dom may not define window.confirm
    Object.defineProperty(window, "confirm", {
      configurable: true,
      writable: true,
      value: confirmMock,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("confirmLeave returns true when clean", () => {
    const { result } = renderHook(() => useUnsavedChangesGuard(false));
    expect(result.current.confirmLeave()).toBe(true);
    expect(confirmMock).not.toHaveBeenCalled();
  });

  it("confirmLeave prompts when dirty", () => {
    confirmMock.mockReturnValue(false);
    const { result } = renderHook(() => useUnsavedChangesGuard(true, "leave?"));
    expect(result.current.confirmLeave()).toBe(false);
    expect(confirmMock).toHaveBeenCalledWith("leave?");
  });

  it("registers beforeunload while dirty", () => {
    const add = vi.spyOn(window, "addEventListener");
    const remove = vi.spyOn(window, "removeEventListener");
    const { unmount, rerender } = renderHook(
      ({ dirty }) => useUnsavedChangesGuard(dirty),
      { initialProps: { dirty: true } },
    );
    expect(add).toHaveBeenCalledWith("beforeunload", expect.any(Function));
    rerender({ dirty: false });
    unmount();
    expect(remove).toHaveBeenCalledWith("beforeunload", expect.any(Function));
  });
});
