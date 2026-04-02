import { http } from "./http";

export const exportTheme = (name?: string) =>
  http.post(`/admin/theme-packages/export${name ? `?name=${name}` : ""}`, {}).then((r) => r.data);

export const importTheme = (themePackage: Record<string, unknown>) =>
  http.post("/admin/theme-packages/import", themePackage).then((r) => r.data);

export const listThemePackages = () =>
  http.get("/admin/theme-packages").then((r) => r.data);

export const applyThemePackage = (id: number) =>
  http.put(`/admin/theme-packages/${id}/apply`, {}).then((r) => r.data);
