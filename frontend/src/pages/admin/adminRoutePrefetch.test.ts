import { describe, expect, it, beforeEach } from "vitest";
import {
  __resetAdminPrefetchCacheForTests,
  resolveAdminLoaderPath,
  adminRouteLoaders,
} from "./adminRoutePrefetch";

describe("adminRoutePrefetch", () => {
  beforeEach(() => {
    __resetAdminPrefetchCacheForTests();
  });

  it("resolves longest matching loader path", () => {
    expect(resolveAdminLoaderPath("/admin")).toBe("/admin");
    expect(resolveAdminLoaderPath("/admin/articles")).toBe("/admin/articles");
    expect(resolveAdminLoaderPath("/admin/articles/categories")).toBe(
      "/admin/articles/categories",
    );
    expect(resolveAdminLoaderPath("/admin/articles/edit/12")).toBe("/admin/articles/edit");
    expect(resolveAdminLoaderPath("/admin/pages/edit/3")).toBe("/admin/pages/edit");
    expect(resolveAdminLoaderPath("/admin/settings")).toBe("/admin/settings");
    expect(resolveAdminLoaderPath("/blog")).toBeNull();
  });

  it("does not map nested admin paths to the dashboard loader", () => {
    expect(resolveAdminLoaderPath("/admin/media")).toBe("/admin/media");
    expect(resolveAdminLoaderPath("/admin/unknown-route")).toBeNull();
  });

  it("registers loaders for core and secondary routes", () => {
    expect(adminRouteLoaders["/admin"]).toBeTypeOf("function");
    expect(adminRouteLoaders["/admin/articles"]).toBeTypeOf("function");
    expect(adminRouteLoaders["/admin/pages"]).toBeTypeOf("function");
    expect(adminRouteLoaders["/admin/media"]).toBeTypeOf("function");
    expect(adminRouteLoaders["/admin/settings"]).toBeTypeOf("function");
    expect(adminRouteLoaders["/admin/theme"]).toBeTypeOf("function");
  });

  it("resolves editor loaders for new and edit paths", () => {
    expect(resolveAdminLoaderPath("/admin/articles/new")).toBe("/admin/articles/new");
    expect(resolveAdminLoaderPath("/admin/articles/edit/9")).toBe("/admin/articles/edit");
    expect(resolveAdminLoaderPath("/admin/pages/new")).toBe("/admin/pages/new");
    expect(resolveAdminLoaderPath("/admin/pages/edit/2")).toBe("/admin/pages/edit");
  });
});
