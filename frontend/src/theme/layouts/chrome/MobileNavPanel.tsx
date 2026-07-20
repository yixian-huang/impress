import { useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { isExternalNavPath, type SiteNavItem } from "./useSiteNavigation";

interface MobileNavPanelProps {
  items: SiteNavItem[];
  open: boolean;
  onNavigate: () => void;
  variant?: "corporate" | "blog";
}

function MobileNavLink({
  item,
  className,
  onNavigate,
  children,
}: {
  item: SiteNavItem;
  className: string;
  onNavigate: () => void;
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
        onClick={onNavigate}
      >
        {children}
      </a>
    );
  }
  return (
    <Link to={href} className={className} onClick={onNavigate}>
      {children}
    </Link>
  );
}

function CorporateMobileItem({ item, depth = 0, onNavigate }: {
  item: SiteNavItem;
  depth?: number;
  onNavigate: () => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const location = useLocation();
  const hasChildren = item.children && item.children.length > 0;
  const isActive =
    !isExternalNavPath(item.path) && item.path ? location.pathname === item.path : false;

  return (
    <div>
      <div className="flex items-center" style={{ paddingLeft: depth * 20 }}>
        <MobileNavLink
          item={item}
          className={`flex-1 py-2.5 text-sm font-medium transition-colors cursor-pointer ${
            isActive
              ? "text-blue-600"
              : depth === 0
                ? "text-gray-800 hover:text-blue-600"
                : "text-gray-500 hover:text-blue-600"
          }`}
          onNavigate={onNavigate}
        >
          {item.label}
        </MobileNavLink>
        {hasChildren && (
          <button
            type="button"
            onClick={() => setExpanded(!expanded)}
            className="p-2 text-gray-400 hover:text-gray-600 cursor-pointer"
          >
            <svg className={`w-4 h-4 transition-transform duration-200 ${expanded ? "rotate-90" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        )}
      </div>
      {hasChildren && expanded && (
        <div>
          {item.children!.map((child, ci) => (
            <CorporateMobileItem
              key={child.path || child.label || String(ci)}
              item={child}
              depth={depth + 1}
              onNavigate={onNavigate}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function BlogMobileItem({ item, onNavigate }: { item: SiteNavItem; onNavigate: () => void }) {
  const location = useLocation();
  const isActive =
    !isExternalNavPath(item.path) && item.path ? location.pathname === item.path : false;

  return (
    <MobileNavLink
      item={item}
      className={`block py-2.5 text-sm font-medium transition-colors ${
        isActive ? "text-primary" : "text-on-surface hover:text-primary"
      }`}
      onNavigate={onNavigate}
    >
      {item.label}
    </MobileNavLink>
  );
}

export default function MobileNavPanel({ items, open, onNavigate, variant = "blog" }: MobileNavPanelProps) {
  return (
    <div className={`lg:hidden overflow-hidden transition-all duration-300 ${
      open ? "max-h-[80vh] opacity-100 mt-4" : "max-h-0 opacity-0"
    }`}>
      <div className="bg-white rounded-xl shadow-xl ring-1 ring-black/5 p-4 divide-y divide-gray-100">
        <div className="pb-2 space-y-0.5">
          {items.map((item, index) =>
            variant === "corporate" ? (
              <CorporateMobileItem
                key={item.path || item.label || String(index)}
                item={item}
                onNavigate={onNavigate}
              />
            ) : (
              <BlogMobileItem
                key={item.path || item.label || String(index)}
                item={item}
                onNavigate={onNavigate}
              />
            ),
          )}
        </div>
      </div>
    </div>
  );
}
