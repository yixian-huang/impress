import React from "react";
import ReactDOM from "react-dom";
import * as ReactRouterDOM from "react-router-dom";
import * as ReactI18next from "react-i18next";
import * as themeHost from "@/theme-host";

/**
 * Expose shared dependencies on window for external theme bundles.
 * External themes use these instead of bundling their own React / host APIs.
 *
 * - `React`, `ReactDOM`, `ReactRouterDOM`, `ReactI18next` — framework peers
 * - `host` — `@inkless/theme-host` surface (hooks, chrome primitives, blog UI)
 */
const sharedDependencies = {
  React,
  ReactDOM,
  ReactRouterDOM,
  ReactI18next,
  host: themeHost,
};

(window as any).__INKLESS_SHARED__ = sharedDependencies;
// UMD builds of theme packages map peers + host via these globals
// (see packages/theme-blog-first/vite.config.ts `output.globals`).
(window as any).React = React;
(window as any).ReactDOM = ReactDOM;
(window as any).ReactRouterDOM = ReactRouterDOM;
(window as any).ReactI18next = ReactI18next;
(window as any).InklessThemeHost = themeHost;

if (!Object.prototype.hasOwnProperty.call(window, "__IMPRESS_SHARED__")) {
  Object.defineProperty(window, "__IMPRESS_SHARED__", {
    configurable: true,
    get: () => (window as any).__INKLESS_SHARED__,
    set: (value) => {
      (window as any).__INKLESS_SHARED__ = value;
    },
  });
}
