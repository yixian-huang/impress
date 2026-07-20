import { useEffect, useMemo, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { ChevronDown, PanelLeftClose, PanelLeftOpen, Search, X } from "lucide-react";
import { ProductLogo } from "@/components/product/ProductLogo";
import { useAuth } from "@/contexts/AuthContext";
import { BROWSER_STORAGE_KEYS } from "@/lib/browserStorage";
import { prefetchAdminRouteWithEditors } from "@/pages/admin/adminRoutePrefetch";
import {
  filterNavGroups,
  isGroupActive,
  isNavItemActive,
  type AdminNavGroup,
  type AdminNavGroupId,
  type AdminNavItem,
} from "@/pages/admin/nav/adminNav";

function handlePrefetchPath(path: string) {
  prefetchAdminRouteWithEditors(path);
}

interface AdminSidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  mobileOpen: boolean;
  onMobileClose: () => void;
}

type CollapsedMap = Partial<Record<AdminNavGroupId, boolean>>;

function readCollapsedMap(): CollapsedMap {
  try {
    const raw = localStorage.getItem(BROWSER_STORAGE_KEYS.adminNavGroupCollapsed);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as CollapsedMap;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

function writeCollapsedMap(map: CollapsedMap) {
  localStorage.setItem(BROWSER_STORAGE_KEYS.adminNavGroupCollapsed, JSON.stringify(map));
}

function NavLinkItem({
  item,
  collapsed,
  active,
  onNavigate,
}: {
  item: AdminNavItem;
  collapsed: boolean;
  active: boolean;
  onNavigate?: () => void;
}) {
  const Icon = item.icon;
  const location = useLocation();
  const showChildren =
    !collapsed &&
    item.children &&
    item.children.length > 0 &&
    (active || item.children.some((c) => location.pathname.startsWith(c.path)));

  return (
    <div>
      <Link
        to={item.path}
        onClick={onNavigate}
        onMouseEnter={() => handlePrefetchPath(item.path)}
        onFocus={() => handlePrefetchPath(item.path)}
        onTouchStart={() => handlePrefetchPath(item.path)}
        title={collapsed ? item.label : undefined}
        className={`group flex items-center rounded-xl transition-all duration-150 ${
          collapsed ? "justify-center px-2 py-2.5" : "px-3 py-2"
        } ${
          active
            ? "bg-blue-500/15 text-blue-100 shadow-[inset_3px_0_0_0] shadow-blue-400"
            : "text-slate-300 hover:bg-white/[0.06] hover:text-white"
        }`}
      >
        <Icon
          className={`h-[1.125rem] w-[1.125rem] shrink-0 ${
            active ? "text-blue-300" : "text-slate-400 group-hover:text-slate-200"
          }`}
        />
        {!collapsed && <span className="ml-3 truncate text-[13px] font-medium">{item.label}</span>}
      </Link>
      {showChildren && (
        <div className="ml-8 mt-0.5 space-y-0.5 border-l border-white/10 pl-3">
          {item.children!.map((child) => {
            const childActive =
              location.pathname === child.path || location.pathname.startsWith(`${child.path}/`);
            return (
              <Link
                key={child.path}
                to={child.path}
                onClick={onNavigate}
                onMouseEnter={() => handlePrefetchPath(child.path)}
                onFocus={() => handlePrefetchPath(child.path)}
                onTouchStart={() => handlePrefetchPath(child.path)}
                className={`block rounded-lg px-2 py-1.5 text-xs transition-colors ${
                  childActive
                    ? "font-medium text-blue-300"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                {child.label}
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}

function SidebarContent({
  collapsed,
  onToggle,
  onMobileClose,
}: {
  collapsed: boolean;
  onToggle: () => void;
  onMobileClose?: () => void;
}) {
  const location = useLocation();
  const { hasPermission } = useAuth();
  const [query, setQuery] = useState("");
  const [collapsedMap, setCollapsedMap] = useState<CollapsedMap>(() => readCollapsedMap());

  const filteredGroups = useMemo(
    () => filterNavGroups(hasPermission, { query }),
    [hasPermission, query],
  );

  // Auto-expand group containing the active route
  useEffect(() => {
    setCollapsedMap((prev) => {
      let changed = false;
      const next = { ...prev };
      for (const group of filteredGroups) {
        if (isGroupActive(location.pathname, group) && next[group.id]) {
          next[group.id] = false;
          changed = true;
        }
      }
      if (changed) writeCollapsedMap(next);
      return changed ? next : prev;
    });
  }, [location.pathname, filteredGroups]);

  const isGroupCollapsed = (group: AdminNavGroup): boolean => {
    if (query) return false;
    if (collapsedMap[group.id] !== undefined) return Boolean(collapsedMap[group.id]);
    if (isGroupActive(location.pathname, group)) return false;
    return Boolean(group.defaultCollapsed);
  };

  const toggleGroup = (groupId: AdminNavGroupId) => {
    setCollapsedMap((prev) => {
      const currently = prev[groupId];
      const group = filteredGroups.find((g) => g.id === groupId);
      const defaultCollapsed = Boolean(group?.defaultCollapsed);
      const active = group ? isGroupActive(location.pathname, group) : false;
      const effective = currently !== undefined ? currently : active ? false : defaultCollapsed;
      const next = { ...prev, [groupId]: !effective };
      writeCollapsedMap(next);
      return next;
    });
  };

  const handleNavClick = () => {
    onMobileClose?.();
  };

  return (
    <div className="relative flex h-full flex-col overflow-hidden bg-[#0b1220] text-white">
      {/* Soft brand glow */}
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-40 bg-gradient-to-b from-blue-500/10 to-transparent"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -left-16 top-24 h-40 w-40 rounded-full bg-blue-600/10 blur-3xl"
        aria-hidden
      />

      <div className="relative flex h-14 shrink-0 items-center border-b border-white/[0.06] px-4">
        <ProductLogo
          collapsed={collapsed}
          className={collapsed ? "mx-auto" : ""}
          variant="onDark"
        />
      </div>

      {!collapsed && (
        <div className="relative shrink-0 px-3 pt-3">
          <label className="relative block">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-500" />
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="过滤菜单…"
              className="w-full rounded-xl border border-white/10 bg-white/[0.06] py-2 pl-8 pr-8 text-xs text-slate-200 placeholder:text-slate-500 outline-none transition focus:border-blue-400/40 focus:bg-white/[0.08] focus:ring-2 focus:ring-blue-500/20"
            />
            {query ? (
              <button
                type="button"
                onClick={() => setQuery("")}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                aria-label="清除过滤"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            ) : null}
          </label>
        </div>
      )}

      <nav className="relative flex-1 overflow-y-auto px-2 py-3 scrollbar-thin">
        {filteredGroups.length === 0 ? (
          <p className="px-3 py-6 text-center text-xs text-slate-500">无匹配菜单</p>
        ) : (
          filteredGroups.map((group, groupIndex) => {
            const groupCollapsed = isGroupCollapsed(group);
            const showHeader = !collapsed && Boolean(group.label);

            return (
              <div key={group.id} className={groupIndex > 0 ? "mt-3.5" : ""}>
                {showHeader ? (
                  <button
                    type="button"
                    onClick={() => toggleGroup(group.id)}
                    className="mb-1 flex w-full items-center justify-between rounded-lg px-3 py-1.5 text-left text-[10px] font-semibold uppercase tracking-[0.08em] text-slate-500 transition hover:text-slate-300"
                  >
                    <span>{group.label}</span>
                    <ChevronDown
                      className={`h-3.5 w-3.5 transition-transform ${groupCollapsed ? "-rotate-90" : ""}`}
                    />
                  </button>
                ) : null}

                {collapsed && groupIndex > 0 ? (
                  <div className="mx-2 mb-2 border-t border-white/[0.06]" />
                ) : null}

                {!groupCollapsed || collapsed ? (
                  <div className="space-y-0.5">
                    {group.items.map((item) => (
                      <NavLinkItem
                        key={item.path}
                        item={item}
                        collapsed={collapsed}
                        active={isNavItemActive(location.pathname, item)}
                        onNavigate={handleNavClick}
                      />
                    ))}
                  </div>
                ) : null}
              </div>
            );
          })
        )}
      </nav>

      <div className="relative shrink-0 border-t border-white/[0.06] p-2">
        <button
          type="button"
          onClick={onToggle}
          className="flex w-full items-center justify-center gap-2 rounded-xl py-2.5 text-slate-400 transition-colors hover:bg-white/[0.06] hover:text-white"
          title={collapsed ? "展开侧边栏" : "收起侧边栏"}
        >
          {collapsed ? (
            <PanelLeftOpen className="h-5 w-5" />
          ) : (
            <>
              <PanelLeftClose className="h-5 w-5" />
              <span className="text-xs font-medium">收起</span>
            </>
          )}
        </button>
      </div>
    </div>
  );
}

export default function AdminSidebar({
  collapsed,
  onToggle,
  mobileOpen,
  onMobileClose,
}: AdminSidebarProps) {
  return (
    <>
      <aside
        className={`fixed top-0 left-0 z-20 hidden h-screen transition-all duration-200 md:block ${
          collapsed ? "w-16" : "w-64"
        }`}
      >
        <SidebarContent collapsed={collapsed} onToggle={onToggle} />
      </aside>

      {mobileOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <div
            className="absolute inset-0 bg-slate-950/55 backdrop-blur-[2px]"
            onClick={onMobileClose}
          />
          <aside className="absolute top-0 left-0 h-full w-64 overflow-hidden shadow-2xl shadow-slate-950/40">
            <SidebarContent collapsed={false} onToggle={onToggle} onMobileClose={onMobileClose} />
          </aside>
        </div>
      )}
    </>
  );
}
