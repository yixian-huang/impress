import type { ReactNode } from "react";
import { adminTheme } from "./adminTheme";

export interface AdminCardProps {
  children: ReactNode;
  className?: string;
  padded?: boolean;
  title?: string;
  description?: string;
  actions?: ReactNode;
}

export default function AdminCard({
  children,
  className = "",
  padded = true,
  title,
  description,
  actions,
}: AdminCardProps) {
  return (
    <div className={`${adminTheme.card} ${className}`}>
      {(title || actions) && (
        <div className={adminTheme.cardHeader}>
          <div className="min-w-0">
            {title ? <h2 className={adminTheme.sectionTitle}>{title}</h2> : null}
            {description ? <p className={adminTheme.sectionDesc}>{description}</p> : null}
          </div>
          {actions ? <div className="shrink-0">{actions}</div> : null}
        </div>
      )}
      <div className={padded ? adminTheme.cardPad : ""}>{children}</div>
    </div>
  );
}
