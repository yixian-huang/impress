import { afterEach, describe, expect, it } from "vitest";
import {
  clearLocalDraft,
  isLocalDraftUseful,
  localDraftStorageKey,
  readLocalDraft,
  shouldOfferLocalDraft,
  writeLocalDraft,
  type LocalDraftSnapshot,
} from "./localDraft";

function makeDraft(partial: Partial<LocalDraftSnapshot> = {}): LocalDraftSnapshot {
  return {
    v: 1,
    key: "new",
    savedAt: new Date().toISOString(),
    baseUpdatedAt: null,
    editorMode: "richtext",
    enabledLangs: ["zh"],
    zhTitle: "",
    enTitle: "",
    slug: "",
    coverImage: "",
    zhBody: "<p></p>",
    enBody: "<p></p>",
    zhSeoTitle: "",
    enSeoTitle: "",
    zhMetaDescription: "",
    enMetaDescription: "",
    ogImage: "",
    author: "",
    markdownZh: "",
    markdownEn: "",
    ...partial,
  };
}

afterEach(() => {
  clearLocalDraft(null);
  clearLocalDraft("42");
});

describe("localDraftStorageKey", () => {
  it("uses new for missing id", () => {
    expect(localDraftStorageKey(undefined)).toContain("new");
    expect(localDraftStorageKey("12")).toContain("12");
  });
});

describe("isLocalDraftUseful", () => {
  it("rejects empty shells", () => {
    expect(isLocalDraftUseful(makeDraft())).toBe(false);
    expect(isLocalDraftUseful(makeDraft({ zhBody: "<p><br></p>" }))).toBe(false);
  });

  it("accepts title or body content", () => {
    expect(isLocalDraftUseful(makeDraft({ zhTitle: "Hello" }))).toBe(true);
    expect(isLocalDraftUseful(makeDraft({ zhBody: "<p>正文</p>" }))).toBe(true);
    expect(isLocalDraftUseful(makeDraft({ markdownZh: "# hi" }))).toBe(true);
  });
});

describe("shouldOfferLocalDraft", () => {
  it("offers when local body differs from server", () => {
    const local = makeDraft({ zhBody: "<p>local</p>", zhTitle: "T" });
    expect(
      shouldOfferLocalDraft(local, {
        zhTitle: "T",
        enTitle: "",
        zhBody: "<p>server</p>",
        enBody: "",
        baseUpdatedAt: "2026-01-01T00:00:00Z",
      }),
    ).toBe(true);
  });

  it("skips when identical to server", () => {
    const local = makeDraft({
      zhTitle: "Same",
      zhBody: "<p>x</p>",
      baseUpdatedAt: "2026-01-01T00:00:00Z",
    });
    expect(
      shouldOfferLocalDraft(local, {
        zhTitle: "Same",
        enTitle: "",
        zhBody: "<p>x</p>",
        enBody: "",
        baseUpdatedAt: "2026-01-01T00:00:00Z",
      }),
    ).toBe(false);
  });
});

describe("read/write/clear", () => {
  it("round-trips through localStorage", () => {
    const draft = makeDraft({ key: "42", zhTitle: "Draft title", zhBody: "<p>hi</p>" });
    expect(writeLocalDraft(draft)).toBe(true);
    const read = readLocalDraft("42");
    expect(read?.zhTitle).toBe("Draft title");
    expect(read?.zhBody).toBe("<p>hi</p>");
    clearLocalDraft("42");
    expect(readLocalDraft("42")).toBeNull();
  });
});
