import { useState, useRef, useEffect } from "react";
import { ExternalLink, LogOut, Menu, UserRound } from "lucide-react";
import { useAuth } from "@/contexts/AuthContext";
import { getNavTitle } from "@/pages/admin/nav/adminNav";

interface AdminTopbarProps {
  pathname: string;
  siteName: string;
  onOpenMobileMenu: () => void;
  onLogout: () => void;
}

function roleBadgeLabel(user: { isSuperAdmin?: boolean; role?: string } | null | undefined): string {
  if (user?.isSuperAdmin) return "超级管理员";
  if (user?.role === "admin") return "管理员";
  return "编辑";
}

export default function AdminTopbar({
  pathname,
  siteName,
  onOpenMobileMenu,
  onLogout,
}: AdminTopbarProps) {
  const { user } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const pageTitle = getNavTitle(pathname);

  useEffect(() => {
    if (!menuOpen) return;
    const onPointerDown = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, [menuOpen]);

  const isSuper = Boolean(user?.isSuperAdmin);

  return (
    <header className="sticky top-0 z-10 h-14 border-b border-slate-200/70 bg-white/75 backdrop-blur-xl supports-[backdrop-filter]:bg-white/65">
      <div className="flex h-full items-center justify-between gap-3 px-4 sm:px-6">
        <div className="flex min-w-0 items-center gap-3">
          <button
            type="button"
            onClick={onOpenMobileMenu}
            className="rounded-xl p-1.5 text-slate-500 transition hover:bg-slate-100 hover:text-slate-800 md:hidden"
            aria-label="打开菜单"
          >
            <Menu className="h-5 w-5" />
          </button>
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold tracking-tight text-slate-900">
              {pageTitle}
            </p>
            <p className="hidden truncate text-xs text-slate-500 sm:block">
              {siteName}
              <span className="mx-1.5 text-slate-300">·</span>
              管理后台
            </p>
          </div>
        </div>

        <div className="relative flex items-center gap-2" ref={menuRef}>
          <a
            href="/"
            target="_blank"
            rel="noreferrer"
            className="hidden items-center gap-1.5 rounded-xl px-2.5 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900 sm:inline-flex"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            前台
          </a>

          <button
            type="button"
            onClick={() => setMenuOpen((open) => !open)}
            className="inline-flex items-center gap-2 rounded-xl border border-slate-200/90 bg-white px-2 py-1.5 text-sm shadow-sm transition hover:bg-slate-50 hover:border-slate-300"
            aria-expanded={menuOpen}
            aria-haspopup="menu"
          >
            <span className="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-blue-700 text-white shadow-sm shadow-blue-600/25">
              <UserRound className="h-3.5 w-3.5" />
            </span>
            <span className="hidden max-w-[8rem] truncate text-slate-700 sm:inline">
              {user?.username || "管理员"}
            </span>
            <span
              className={`hidden rounded-full px-1.5 py-0.5 text-[10px] font-medium md:inline ${
                isSuper
                  ? "bg-amber-50 text-amber-800 ring-1 ring-amber-600/15"
                  : "bg-slate-100 text-slate-600 ring-1 ring-slate-500/10"
              }`}
            >
              {roleBadgeLabel(user)}
            </span>
          </button>

          {menuOpen && (
            <div
              role="menu"
              className="absolute right-0 top-full z-30 mt-1.5 w-56 overflow-hidden rounded-2xl border border-slate-200/90 bg-white py-1 shadow-[0_16px_48px_rgba(15,23,42,0.14)]"
            >
              <div className="border-b border-slate-100 px-3.5 py-3">
                <p className="truncate text-sm font-semibold text-slate-900">
                  {user?.username || "管理员"}
                </p>
                <p className="mt-0.5 text-xs text-slate-500">{roleBadgeLabel(user)}</p>
              </div>
              <a
                href="/"
                target="_blank"
                rel="noreferrer"
                role="menuitem"
                className="flex items-center gap-2 px-3.5 py-2.5 text-sm text-slate-700 transition hover:bg-slate-50"
                onClick={() => setMenuOpen(false)}
              >
                <ExternalLink className="h-4 w-4 text-slate-400" />
                打开前台
              </a>
              <button
                type="button"
                role="menuitem"
                onClick={() => {
                  setMenuOpen(false);
                  onLogout();
                }}
                className="flex w-full items-center gap-2 px-3.5 py-2.5 text-sm text-red-600 transition hover:bg-red-50"
              >
                <LogOut className="h-4 w-4" />
                退出登录
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
