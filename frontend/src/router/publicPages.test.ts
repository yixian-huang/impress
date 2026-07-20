import { describe, expect, it } from "vitest";
import { resolveAutomaticNavigation, resolvePublicRoutingPages } from "./publicPages";

const manifestPages = [
  {
    slug: "home",
    contentKey: "home",
    renderMode: "hardcoded" as const,
    nav: { label: "Home", labelZh: "首页", order: 1, showInHeader: true },
  },
];

describe("resolvePublicRoutingPages", () => {
  it("uses unified pages as route and navigation truth", () => {
    const pages = resolvePublicRoutingPages({
      unifiedPages: [
        {
          id: 9,
          slug: "launch",
          title: { zh: "发布页", en: "Launch" },
          mode: "composable",
          sortOrder: 3,
          showInNav: true,
          status: "published",
          publishedVersion: 1,
        },
      ],
      themePages: [],
      manifestPages,
    });

    expect(pages).toEqual(expect.arrayContaining([
      expect.objectContaining({
        slug: "launch",
        renderMode: "dynamic",
        showInHeader: true,
        showInFooter: true,
      }),
      expect.objectContaining({ slug: "home" }),
    ]));
  });

  it("keeps hardcoded theme rendering for a matching unified slug", () => {
    const pages = resolvePublicRoutingPages({
      unifiedPages: [
        {
          id: 1,
          slug: "home",
          title: { zh: "新首页", en: "New Home" },
          mode: "template",
          sortOrder: 1,
          showInNav: false,
          status: "published",
          publishedVersion: 2,
        },
      ],
      themePages: [],
      manifestPages,
    });

    expect(pages[0]).toEqual(expect.objectContaining({
      slug: "home",
      contentKey: "home",
      renderMode: "hardcoded",
      showInHeader: false,
    }));
  });

  it("merges DB theme rows with missing manifest pages", () => {
    const pages = resolvePublicRoutingPages({
      unifiedPages: [],
      themePages: [
        {
          id: 2,
          slug: "about",
          title: { zh: "关于", en: "About" },
          contentKey: "about",
          renderMode: "dynamic",
          isThemePage: true,
          themeId: "theme-a",
          navConfig: { showInHeader: true },
          sortOrder: 2,
          status: "published",
        },
      ],
      manifestPages,
      activeThemeId: "theme-a",
    });

    // DB owns "about"; manifest fills "home" which is not in DB.
    expect(pages.map((page) => page.slug).sort()).toEqual(["about", "home"]);
  });

  it("fills new theme package pages not yet in DB (e.g. author)", () => {
    const pages = resolvePublicRoutingPages({
      unifiedPages: [],
      themePages: [
        {
          id: 1,
          slug: "home",
          title: { zh: "首页", en: "Home" },
          contentKey: "home",
          renderMode: "hardcoded",
          isThemePage: true,
          themeId: "blog-first",
          navConfig: { showInHeader: true },
          sortOrder: 0,
          status: "published",
        },
      ],
      manifestPages: [
        ...manifestPages,
        {
          slug: "author",
          contentKey: "author",
          renderMode: "hardcoded",
          nav: { label: "About", labelZh: "关于", order: 10, showInHeader: true },
        },
      ],
      activeThemeId: "blog-first",
    });

    expect(pages.map((page) => page.slug)).toEqual(["home", "author"]);
    expect(resolveAutomaticNavigation([], pages as never, "zh", "header", [
      {
        slug: "author",
        contentKey: "author",
        renderMode: "hardcoded",
        nav: { label: "About", labelZh: "关于", order: 10, showInHeader: true },
      },
    ])).toEqual(expect.arrayContaining([
      expect.objectContaining({ path: "/author", label: "关于" }),
    ]));
  });

  it("merges legacy routes and navigation by slug during migration", () => {
    const unifiedPages = [
      {
        id: 9,
        slug: "about",
        title: { zh: "新版关于", en: "New About" },
        mode: "composable" as const,
        sortOrder: 3,
        showInNav: false,
        status: "published",
        publishedVersion: 1,
      },
    ];
    const themePages = [
      {
        id: 2,
        slug: "about",
        title: { zh: "旧版关于", en: "Old About" },
        contentKey: "about",
        renderMode: "dynamic" as const,
        isThemePage: true,
        themeId: "theme-a",
        navConfig: { showInHeader: true },
        sortOrder: 2,
        status: "published",
      },
      {
        id: 3,
        slug: "contact",
        title: { zh: "联系", en: "Contact" },
        contentKey: "contact",
        renderMode: "dynamic" as const,
        isThemePage: true,
        themeId: "theme-a",
        navConfig: { showInHeader: true },
        sortOrder: 4,
        status: "published",
      },
    ];

    const routes = resolvePublicRoutingPages({
      unifiedPages,
      themePages,
      manifestPages,
      activeThemeId: "theme-a",
    });
    // Manifest fills "home"; unified owns "about"; theme DB fills "contact".
    expect(routes.map((page) => page.slug)).toEqual(["home", "about", "contact"]);
    expect(resolveAutomaticNavigation(unifiedPages, themePages, "zh", "header")).toEqual([
      { label: "联系", path: "/contact", sortOrder: 4 },
    ]);
  });
});
