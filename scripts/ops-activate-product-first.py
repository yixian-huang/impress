#!/usr/bin/env python3
"""Activate product-first on the product ops DB only (default: /opt/inkless-ops)."""
from __future__ import annotations

import json
import os
import shutil
import sqlite3
import time
from datetime import datetime, timezone
from pathlib import Path

DB_PATH = Path(os.environ.get("INKLESS_DB", "/opt/inkless-ops/data/inkless.db"))

PRODUCT_PAGES = [
    ("home", "home", 0, {"zh": "首页", "en": "Home"}, {"showInHeader": True, "showInFooter": True}),
    ("features", "features", 1, {"zh": "能力", "en": "Features"}, {"showInHeader": True, "showInFooter": True}),
    ("contact", "contact", 10, {"zh": "联系", "en": "Contact"}, {"showInHeader": True, "showInFooter": True}),
]

FEATURES = {
    "publicPages": {
        "home": True,
        "blog": False,
        "contact": True,
        "about": False,
        "advantages": False,
        "coreServices": False,
        "cases": False,
        "experts": False,
    },
    "blog": {"comments": False, "rss": False},
}

THEME_TOKENS = {
    "colors": {
        "primary": "#111827",
        "primaryDark": "#0b1220",
        "accent": "#14b8a6",
        "accentHover": "#0d9488",
        "surface": "#ffffff",
        "surfaceAlt": "#f8fafc",
        "onPrimary": "#ffffff",
        "onSurface": "#111827",
        "onSurfaceMuted": "#64748b",
        "border": "#e2e8f0",
    },
    "fonts": {
        "sans": "ui-sans-serif, system-ui, -apple-system, \"Segoe UI\", sans-serif",
        "heading": "ui-sans-serif, system-ui, -apple-system, \"Segoe UI\", sans-serif",
        "mono": "ui-monospace, SF Mono, Menlo, Monaco, Consolas, monospace",
    },
    "layout": {
        "maxWidth": "72rem",
        "borderRadius": "0.5rem",
        "contentPadding": "1.5rem",
        "sectionSpacing": "4.5rem",
        "contentGap": "2rem",
    },
}

# Product home schema (not corporate blocks)
HOME_CFG = {
    "hero": {
        "eyebrow": {"zh": "Inkless CMS", "en": "Inkless CMS"},
        "title": {
            "zh": "用主题驱动你的产品站与内容运营",
            "en": "Theme-driven product sites and content ops",
        },
        "subtitle": {
            "zh": "开源自托管 CMS：产品介绍、博客与企业站由主题决定呈现。",
            "en": "Open-source self-hosted CMS — product, blog, or corporate shape is owned by themes.",
        },
        "primaryCta": {"label": {"zh": "快速开始", "en": "Get started"}, "href": "#install"},
        "secondaryCta": {
            "label": {"zh": "GitHub", "en": "GitHub"},
            "href": "https://github.com/yixian-huang/inkless",
        },
    },
    "features": {
        "title": {"zh": "核心能力", "en": "Capabilities"},
        "items": [
            {
                "title": {"zh": "主题系统", "en": "Themes"},
                "description": {
                    "zh": "product-first / blog-first / corporate 一键切换形态。",
                    "en": "Swap product-first, blog-first, or corporate presentation.",
                },
            },
            {
                "title": {"zh": "内容运营", "en": "Content ops"},
                "description": {
                    "zh": "页面、文章、媒体与发布版本管理。",
                    "en": "Pages, articles, media, and versioned publish.",
                },
            },
            {
                "title": {"zh": "可扩展", "en": "Extensible"},
                "description": {
                    "zh": "插件与主题契约，定制留在扩展层。",
                    "en": "Plugins and theme contract keep customization outside core.",
                },
            },
        ],
    },
    "howItWorks": {
        "title": {"zh": "如何开始", "en": "How it works"},
        "steps": [
            {
                "title": {"zh": "部署实例", "en": "Deploy"},
                "description": {"zh": "自托管 artifact / compose 部署。", "en": "Self-host via artifact or compose."},
            },
            {
                "title": {"zh": "选择主题", "en": "Pick a theme"},
                "description": {
                    "zh": "激活 product-first 作为产品运营站。",
                    "en": "Activate product-first for product ops sites.",
                },
            },
            {
                "title": {"zh": "发布内容", "en": "Publish"},
                "description": {"zh": "配置品牌与页面，对外发布。", "en": "Configure brand and pages, then publish."},
            },
        ],
    },
    "install": {
        "title": {"zh": "快速开始", "en": "Quick start"},
        "code": "# See repository README for current install path\ngit clone https://github.com/yixian-huang/inkless.git\n# deploy with your ops pipeline (e.g. npc deploy)",
        "caption": {
            "zh": "完整安装与文档由独立文档服务提供（配置 docsUrl）。",
            "en": "Full install docs live on your external docs service (set docsUrl).",
        },
    },
    "bottomCta": {
        "title": {"zh": "开始构建你的产品站", "en": "Build your product site"},
        "subtitle": {
            "zh": "现代、可扩展的内容管理系统。",
            "en": "A modern, extensible content management system.",
        },
        "primaryCta": {
            "label": {"zh": "查看源码", "en": "View source"},
            "href": "https://github.com/yixian-huang/inkless",
        },
    },
}

GLOBAL_CFG = {
    "identity": {
        "name": {"zh": "Inkless", "en": "Inkless"},
        "tagline": {
            "zh": "现代可扩展的内容管理系统",
            "en": "A modern, extensible content management system",
        },
        "localeMode": "bilingual",
        "defaultLocale": "zh",
    },
    "brand": {
        "logo": {
            "light": "/brand/inkless-wordmark.svg",
            "dark": "/brand/inkless-wordmark.svg",
        },
        "favicon": "/brand/favicon.svg",
        "ogImage": "/brand/og-default.png",
        "primaryColor": "#111827",
        "accentColor": "#14b8a6",
    },
    "author": {
        "name": "Inkless CMS",
        "avatar": "",
        "bio": {
            "zh": "面向软件产品与内容运营的开源 CMS。",
            "en": "Open-source CMS for product sites and content ops.",
        },
        "socials": [{"kind": "github", "url": "https://github.com/yixian-huang/inkless"}],
    },
    "header": {
        "brandMode": "logo",
        "showSocials": False,
        "showRssLink": False,
    },
    "footer": {
        "copyright": {
            "zh": "© 2026 Inkless · inkless.run",
            "en": "© 2026 Inkless · inkless.run",
        }
    },
    "seo": {
        "defaultTitle": {
            "zh": "Inkless · 内容管理系统",
            "en": "Inkless · Content Management System",
        },
        "titleTemplate": "{page} · Inkless",
        "defaultDescription": {
            "zh": "主题驱动的开源 CMS，适合产品运营站与内容站点。",
            "en": "Theme-driven open-source CMS for product ops and content sites.",
        },
        "twitterHandle": "",
    },
}


def upsert_site_config(cur: sqlite3.Cursor, key: str, payload: dict, now: str) -> None:
    s = json.dumps(payload, ensure_ascii=False)
    row = cur.execute("select id from site_configs where key=?", (key,)).fetchone()
    if row:
        cur.execute(
            "update site_configs set draft_config=?, published_config=?, "
            "draft_version=draft_version+1, published_version=published_version+1, updated_at=? where key=?",
            (s, s, now, key),
        )
    else:
        cur.execute(
            "insert into site_configs (key, draft_config, draft_version, published_config, published_version, created_at, updated_at) "
            "values (?,?,?,?,?,?,?)",
            (key, s, 1, s, 1, now, now),
        )


def upsert_content(cur: sqlite3.Cursor, page_key: str, payload: dict, now: str) -> None:
    s = json.dumps(payload, ensure_ascii=False)
    row = cur.execute(
        "select published_version, draft_version from content_documents where page_key=?",
        (page_key,),
    ).fetchone()
    if row:
        cur.execute(
            "update content_documents set draft_config=?, published_config=?, draft_version=?, published_version=?, updated_at=? where page_key=?",
            (s, s, row["draft_version"] + 1, row["published_version"] + 1, now, page_key),
        )
    else:
        cur.execute(
            "insert into content_documents (page_key, draft_config, draft_version, published_config, published_version, updated_at) "
            "values (?,?,?,?,?,?)",
            (page_key, s, 1, s, 1, now),
        )


def main() -> None:
    if not DB_PATH.exists():
        raise SystemExit(f"db not found: {DB_PATH} (refusing to touch other paths)")
    if "inkless-ops" not in str(DB_PATH):
        raise SystemExit(
            f"safety: INKLESS_DB must be under inkless-ops, got {DB_PATH}"
        )

    bak = DB_PATH.with_name(f"{DB_PATH.name}.bak-activate-product-first-{int(time.time())}")
    shutil.copy2(DB_PATH, bak)
    print("backup", bak)

    con = sqlite3.connect(str(DB_PATH))
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    now = datetime.now(timezone.utc).isoformat()

    # Ensure theme row
    row = cur.execute(
        "select id from installed_themes where theme_id=?", ("product-first",)
    ).fetchone()
    if not row:
        cur.execute(
            """insert into installed_themes
            (theme_id, name, name_zh, description, author, version, source, external_url, is_active, preview, config, created_at, updated_at)
            values (?,?,?,?,?,?,?,?,?,?,?,?,?)""",
            (
                "product-first",
                "Product First",
                "产品优先",
                "软件产品介绍站：主视觉、能力、安装引导",
                "Inkless CMS",
                "0.1.0",
                "built-in",
                "",
                0,
                "linear-gradient(135deg, #111827 0%, #14b8a6 100%)",
                "{}",
                now,
                now,
            ),
        )
        print("created installed_themes product-first")
    else:
        cur.execute(
            "update installed_themes set name=?, name_zh=?, description=?, updated_at=? where theme_id=?",
            ("Product First", "产品优先", "软件产品介绍站", now, "product-first"),
        )

    cur.execute("update installed_themes set is_active=0")
    cur.execute(
        "update installed_themes set is_active=1, updated_at=? where theme_id=?",
        (now, "product-first"),
    )
    print("active", dict(cur.execute(
        "select theme_id, is_active from installed_themes where is_active=1"
    ).fetchone()))

    # Soft-unpublish non-product theme pages that would clutter nav? Keep rows; reassign product slugs.
    for slug, content_key, sort_order, title, nav in PRODUCT_PAGES:
        title_b = json.dumps(title, ensure_ascii=False)
        nav_b = json.dumps(nav, ensure_ascii=False)
        existing = cur.execute(
            "select id from pages where slug=? and deleted_at is null", (slug,)
        ).fetchone()
        if existing:
            cur.execute(
                """update pages set theme_id=?, content_key=?, render_mode=?, is_theme_page=1,
                   title=?, nav_config=?, status=?, sort_order=?, updated_at=? where id=?""",
                (
                    "product-first",
                    content_key,
                    "hardcoded",
                    title_b,
                    nav_b,
                    "published",
                    sort_order,
                    now,
                    existing["id"],
                ),
            )
            print("page reassigned", slug)
        else:
            cur.execute(
                """insert into pages
                (slug, parent_id, title, template, config, status, sort_order, seo_title, seo_description, keywords,
                 theme_id, content_key, render_mode, is_theme_page, nav_config, cover_image, auto_summary, allow_comments,
                 pinned, visibility, metadata, created_at, updated_at)
                values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)""",
                (
                    slug,
                    None,
                    title_b,
                    "default",
                    "{}",
                    "published",
                    sort_order,
                    "{}",
                    "{}",
                    "{}",
                    "product-first",
                    content_key,
                    "hardcoded",
                    1,
                    nav_b,
                    "",
                    0,
                    0,
                    0,
                    "public",
                    "{}",
                    now,
                    now,
                ),
            )
            print("page created", slug)

    # Hide leftover corporate theme pages on this instance (if any)
    cur.execute(
        """update pages set status='draft', updated_at=?
           where deleted_at is null and theme_id != 'product-first'
           and slug not in ('home','features','contact')""",
        (now,),
    )

    upsert_site_config(cur, "features", FEATURES, now)
    upsert_site_config(cur, "theme", THEME_TOKENS, now)
    upsert_content(cur, "global", GLOBAL_CFG, now)
    upsert_content(cur, "home", HOME_CFG, now)
    print("features + tokens + global + home updated")

    # Theme installed config for CTA defaults (if host stores theme settings on installed_themes.config)
    theme_settings = {
        "header": {
            "brandMode": "logo",
            "docsUrl": "",
            "githubUrl": "https://github.com/yixian-huang/inkless",
            "primaryCtaLabel": "Get started",
            "primaryCtaHref": "#install",
            "showRssLink": False,
            "showSocials": False,
        }
    }
    cur.execute(
        "update installed_themes set config=?, updated_at=? where theme_id=?",
        (json.dumps(theme_settings, ensure_ascii=False), now, "product-first"),
    )

    con.commit()
    print(
        "VERIFY pages",
        cur.execute(
            "select slug, theme_id, status from pages where deleted_at is null order by sort_order"
        ).fetchall(),
    )
    con.close()
    print("DONE product-first activated on", DB_PATH)


if __name__ == "__main__":
    main()
