import { beforeEach, describe, expect, it, vi } from "vitest";
import { BROWSER_STORAGE_KEYS, migrateLegacyBrowserStorage } from "./browserStorage";

describe("migrateLegacyBrowserStorage", () => {
  beforeEach(() => {
    const createStorage = () => {
      const values = new Map<string, string>();
      return {
        getItem: (key: string) => values.get(key) ?? null,
        setItem: (key: string, value: string) => values.set(key, value),
        removeItem: (key: string) => values.delete(key),
        clear: () => values.clear(),
      };
    };
    vi.stubGlobal("localStorage", createStorage());
    vi.stubGlobal("sessionStorage", createStorage());
    localStorage.clear();
    sessionStorage.clear();
  });

  it("migrates valid legacy keys and removes legacy entries", () => {
    localStorage.setItem("accessToken", "access");
    localStorage.setItem("refreshToken", "refresh");
    localStorage.setItem("admin_sidebar_collapsed", "true");
    localStorage.setItem("impress.comment.guest", JSON.stringify({ name: "Ada" }));
    sessionStorage.setItem("impress.setup.step", "admin");
    sessionStorage.setItem("impress.setup.draft", JSON.stringify({ username: "root" }));

    migrateLegacyBrowserStorage();

    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.authAccessToken)).toBe("access");
    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.authRefreshToken)).toBe("refresh");
    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.adminSidebarCollapsed)).toBe("true");
    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.commentGuest)).toBe('{"name":"Ada"}');
    expect(sessionStorage.getItem(BROWSER_STORAGE_KEYS.setupStep)).toBe("admin");
    expect(sessionStorage.getItem(BROWSER_STORAGE_KEYS.setupDraft)).toBe('{"username":"root"}');
    expect(localStorage.getItem("accessToken")).toBeNull();
    expect(localStorage.getItem("refreshToken")).toBeNull();
    expect(localStorage.getItem("admin_sidebar_collapsed")).toBeNull();
    expect(localStorage.getItem("impress.comment.guest")).toBeNull();
    expect(sessionStorage.getItem("impress.setup.step")).toBeNull();
    expect(sessionStorage.getItem("impress.setup.draft")).toBeNull();
  });

  it("keeps canonical values ahead of legacy values", () => {
    localStorage.setItem(BROWSER_STORAGE_KEYS.authAccessToken, "canonical");
    localStorage.setItem("accessToken", "legacy");

    migrateLegacyBrowserStorage();

    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.authAccessToken)).toBe("canonical");
    expect(localStorage.getItem("accessToken")).toBeNull();
  });

  it("drops invalid legacy payloads instead of copying them", () => {
    localStorage.setItem("admin_sidebar_collapsed", "yes");
    localStorage.setItem("impress.comment.guest", "not-json");
    sessionStorage.setItem("impress.setup.step", "unknown");

    migrateLegacyBrowserStorage();

    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.adminSidebarCollapsed)).toBeNull();
    expect(localStorage.getItem(BROWSER_STORAGE_KEYS.commentGuest)).toBeNull();
    expect(sessionStorage.getItem(BROWSER_STORAGE_KEYS.setupStep)).toBeNull();
    expect(localStorage.getItem("admin_sidebar_collapsed")).toBeNull();
    expect(localStorage.getItem("impress.comment.guest")).toBeNull();
    expect(sessionStorage.getItem("impress.setup.step")).toBeNull();
  });
});
