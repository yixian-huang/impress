import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { editorialFirmPageConfigs } from "./pageConfigs";

/**
 * Host Go embed (`editorial_firm_seeds.json`) must stay in lockstep with
 * package TS seeds. When this fails, update the JSON after editing pageConfigs.
 */
const hostSeedsPath = resolve(
  dirname(fileURLToPath(import.meta.url)),
  "../../../../backend/internal/builtinthemes/editorial_firm_seeds.json",
);

function sectionFingerprint(
  pages: Record<string, { sections: Array<{ id: string; type: string }> }>,
): Record<string, string[]> {
  const out: Record<string, string[]> = {};
  for (const key of Object.keys(pages).sort()) {
    out[key] = (pages[key]?.sections ?? []).map((s) => `${s.type}:${s.id}`);
  }
  return out;
}

describe("editorial-firm host seed parity", () => {
  it("TS pageConfigs match backend editorial_firm_seeds.json fingerprints", () => {
    const hostRaw = JSON.parse(readFileSync(hostSeedsPath, "utf8")) as Record<
      string,
      { sections: Array<{ id: string; type: string }> }
    >;
    const tsFp = sectionFingerprint(editorialFirmPageConfigs);
    const hostFp = sectionFingerprint(hostRaw);
    expect(hostFp).toEqual(tsFp);
  });
});
