import { useState, useEffect, useRef, useCallback, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { isExternalNavPath, type SiteNavItem } from "./useSiteNavigation";

interface DesktopNavLinksProps {
  items: SiteNavItem[];
  /** Corporate hero scroll: light text when not scrolled */
  variant?: "corporate" | "blog";
  scrolled?: boolean;
}

function NavLink({
  item,
  className,
  children,
}: {
  item: SiteNavItem;
  className: string;
  children: ReactNode;
}) {
  const href = item.path || "/";
  const external = isExternalNavPath(href) || item.target === "_blank";
  if (external) {
    return (
      <a
        href={href}
        target={item.target === "_self" ? undefined : item.target || "_blank"}
        rel={item.target === "_self" ? undefined : "noopener noreferrer"}
        className={className}
      >
        {children}
      </a>
    );
  }
  return (
    <Link to={href} className={className}>
      {children}
    </Link>
  );
}

function CorporateNavItem({ item, scrolled, depth = 0 }: {
  item: SiteNavItem;
  scrolled: boolean;
  depth?: number;
}) {
  const [open, setOpen] = useState(false);
  const [flipX, setFlipX] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const location = useLocation();

  const enter = useCallback(() => { clearTimeout(timer.current); setOpen(true); }, []);
  const leave = useCallback(() => { timer.current = setTimeout(() => setOpen(false), 120); }, []);

  useEffect(() => () => clearTimeout(timer.current), []);

  useEffect(() => {
    if (open && dropdownRef.current) {
      const rect = dropdownRef.current.getBoundingClientRect();
      setFlipX(rect.right > window.innerWidth - 8);
    }
  }, [open]);

  const hasChildren = item.children && item.children.length > 0;
  const isActive =
    !isExternalNavPath(item.path) && item.path ? location.pathname === item.path : false;
  const isRoot = depth === 0;

  const rootClass = `text-sm font-medium whitespace-nowrap cursor-pointer transition-colors duration-200 ${
    scrolled
      ? `text-gray-700 hover:text-blue-600 ${isActive ? "text-blue-600" : ""}`
      : `text-white/90 hover:text-white ${isActive ? "text-white" : ""}`
  }`;
  const subClass = `flex items-center justify-between gap-4 px-4 py-2.5 text-sm text-white/80 hover:text-white hover:bg-white/10 whitespace-nowrap cursor-pointer transition-colors ${
    isActive ? "text-white bg-white/10" : ""
  }`;
  const linkClass = isRoot ? rootClass : subClass;

  if (!hasChildren) {
    return (
      <NavLink item={item} className={linkClass}>
        {item.label}
      </NavLink>
    );
  }

  const chevron = (
    <svg
      className={`w-3 h-3 shrink-0 ${isRoot ? "ml-0.5" : ""} ${isRoot ? (scrolled ? "text-gray-400" : "text-white/60") : "text-white/40"}`}
      fill="none" stroke="currentColor" viewBox="0 0 24 24"
    >
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
        d={isRoot ? "M19 9l-7 7-7-7" : (flipX ? "M15 19l-7-7 7-7" : "M9 5l7 7-7 7")} />
    </svg>
  );

  const positionClass = isRoot
    ? (flipX ? "right-0 top-full pt-2" : "left-0 top-full pt-2")
    : (flipX ? "right-full top-0 pr-1" : "left-full top-0 pl-1");

  return (
    <div className="relative" onMouseEnter={enter} onMouseLeave={leave}>
      <NavLink item={item} className={`${linkClass} inline-flex items-center`}>
        {item.label}
        {chevron}
      </NavLink>
      <div
        ref={dropdownRef}
        className={`absolute z-50 transition-all duration-200 ${positionClass} ${open ? "opacity-100 visible translate-y-0" : "opacity-0 invisible -translate-y-1 pointer-events-none"}`}
      >
        <div className="bg-black/70 backdrop-blur-md rounded-lg shadow-xl ring-1 ring-white/10 py-1.5 min-w-[180px]">
          {item.children!.map((child, ci) => (
            <CorporateNavItem key={child.path || child.label || String(ci)} item={child} scrolled={scrolled} depth={depth + 1} />
          ))}
        </div>
      </div>
    </div>
  );
}

function BlogNavItem({ item }: { item: SiteNavItem }) {
  const location = useLocation();
  const isActive =
    !isExternalNavPath(item.path) && item.path ? location.pathname === item.path : false;

  return (
    <NavLink
      item={item}
      className={`text-sm font-medium whitespace-nowrap transition-colors ${
        isActive ? "text-primary" : "text-on-surface-muted hover:text-primary"
      }`}
    >
      {item.label}
    </NavLink>
  );
}

export default function DesktopNavLinks({ items, variant = "blog", scrolled = true }: DesktopNavLinksProps) {
  if (items.length === 0) return null;

  return (
    <div className="hidden lg:flex items-center gap-7">
      {items.map((item, index) =>
        variant === "corporate" ? (
          <CorporateNavItem key={item.path || item.label || String(index)} item={item} scrolled={scrolled} />
        ) : (
          <BlogNavItem key={item.path || item.label || String(index)} item={item} />
        ),
      )}
    </div>
  );
}
