import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import { resolve } from "node:path";

/**
 * Library build for remote install (UMD) and standalone ESM dist.
 * Host APIs are external — provided at runtime as window.InklessThemeHost.
 */
export default defineConfig({
  plugins: [react()],
  define: {
    "process.env.NODE_ENV": JSON.stringify("production"),
    "import.meta.env.VITE_API_BASE_URL": JSON.stringify(""),
  },
  build: {
    lib: {
      entry: resolve(__dirname, "src/register.ts"),
      name: "InklessThemeEditorialFirm",
      formats: ["umd", "es"],
      fileName: (format) => (format === "umd" ? "theme.umd.js" : "theme.es.js"),
    },
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      output: [
        {
          format: "umd",
          name: "InklessThemeEditorialFirm",
          entryFileNames: "theme.umd.js",
          inlineDynamicImports: true,
          globals: {
            react: "React",
            "react-dom": "ReactDOM",
            "react-router-dom": "ReactRouterDOM",
            "react-i18next": "ReactI18next",
            "@inkless/theme-host": "InklessThemeHost",
          },
        },
        {
          format: "es",
          entryFileNames: "theme.es.js",
          inlineDynamicImports: true,
        },
      ],
      external: [
        "react",
        "react-dom",
        "react-router-dom",
        "react-i18next",
        "@inkless/theme-host",
      ],
    },
  },
});
