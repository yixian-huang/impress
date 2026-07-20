import { describe, expect, it } from "vitest";
import {
  buildModeSwitchConfirmMessage,
  detectModeSwitchLoss,
  detectModeSwitchLossFromBodies,
  shouldConfirmModeSwitch,
} from "./modeSwitchLoss";

describe("detectModeSwitchLoss", () => {
  it("returns empty for plain TipTap HTML (safe to switch silently)", () => {
    expect(detectModeSwitchLoss("<p>Hello <strong>world</strong></p>")).toEqual([]);
    expect(
      detectModeSwitchLoss("<h2>Title</h2><p>A <em>line</em> with <a href=\"/x\">link</a>.</p>"),
    ).toEqual([]);
    expect(detectModeSwitchLoss("<p></p>")).toEqual([]);
    expect(detectModeSwitchLoss("")).toEqual([]);
    expect(detectModeSwitchLoss(null)).toEqual([]);
  });

  it("detects text color without matching background-color alone as color", () => {
    expect(detectModeSwitchLoss('<p><span style="color: #e11">red</span></p>')).toContain("color");
    // background-color alone should not flag color (negative lookbehind)
    expect(
      detectModeSwitchLoss('<p><span style="background-color: yellow">x</span></p>'),
    ).not.toContain("color");
  });

  it("detects highlight, font size, line height, and alignment", () => {
    expect(detectModeSwitchLoss('<p><mark data-color="#ff0">hi</mark></p>')).toEqual(
      expect.arrayContaining(["highlight"]),
    );
    expect(detectModeSwitchLoss('<p><span style="font-size: 20px">big</span></p>')).toContain(
      "fontSize",
    );
    expect(detectModeSwitchLoss('<p style="line-height: 2">spaced</p>')).toContain("lineHeight");
    expect(detectModeSwitchLoss('<p style="text-align: center">c</p>')).toContain("textAlign");
  });

  it("detects columns, details, gallery, sub/sup, and media width", () => {
    expect(
      detectModeSwitchLoss('<div data-type="columns" class="columns cols-2"><div data-type="column">a</div></div>'),
    ).toEqual(expect.arrayContaining(["columns"]));
    expect(detectModeSwitchLoss("<details><summary>s</summary><p>b</p></details>")).toContain(
      "details",
    );
    expect(
      detectModeSwitchLoss('<div data-type="image-gallery" data-images="[]"></div>'),
    ).toContain("gallery");
    expect(detectModeSwitchLoss("<p>H<sub>2</sub>O and x<sup>2</sup></p>")).toContain("subSup");
    expect(detectModeSwitchLoss('<img src="/a.png" alt="" width="320">')).toContain("mediaWidth");
    expect(
      detectModeSwitchLoss('<img src="/a.png" alt="" style="width: 50%">'),
    ).toContain("mediaWidth");
  });

  it("unions losses across language bodies in stable order", () => {
    const keys = detectModeSwitchLossFromBodies(
      '<p style="text-align: right">zh</p>',
      '<p><span style="color: blue">en</span></p>',
    );
    expect(keys).toEqual(["color", "textAlign"]);
    expect(shouldConfirmModeSwitch("<p>plain</p>", "<p></p>")).toBe(false);
    expect(shouldConfirmModeSwitch('<p style="text-align: left">x</p>')).toBe(true);
  });
});

describe("buildModeSwitchConfirmMessage", () => {
  it("lists Chinese labels for detected losses", () => {
    const msg = buildModeSwitchConfirmMessage(["color", "columns"]);
    expect(msg).toContain("文字颜色");
    expect(msg).toContain("多栏布局");
    expect(msg).toContain("是否继续");
  });

  it("falls back when empty", () => {
    expect(buildModeSwitchConfirmMessage([])).toMatch(/是否继续/);
  });
});
