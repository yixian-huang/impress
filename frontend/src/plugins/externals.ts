import React from "react";
import ReactDOM from "react-dom";
import * as ReactRouterDOM from "react-router-dom";
import * as ReactI18next from "react-i18next";

/**
 * Expose shared dependencies on window for external theme bundles.
 * External themes can use these instead of bundling their own React.
 */
const sharedDependencies = {
  React,
  ReactDOM,
  ReactRouterDOM,
  ReactI18next,
};

(window as any).__INKLESS_SHARED__ = sharedDependencies;

if (!Object.prototype.hasOwnProperty.call(window, "__IMPRESS_SHARED__")) {
  Object.defineProperty(window, "__IMPRESS_SHARED__", {
    configurable: true,
    get: () => (window as any).__INKLESS_SHARED__,
    set: (value) => {
      (window as any).__INKLESS_SHARED__ = value;
    },
  });
}
