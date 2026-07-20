import type { ReactNode } from "react";
import { adminTheme } from "./adminTheme";

export interface AdminTableProps {
  children: ReactNode;
  className?: string;
}

export function AdminTable({ children, className = "" }: AdminTableProps) {
  return (
    <div className={`${adminTheme.card} overflow-hidden ${className}`}>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-100 text-sm">{children}</table>
      </div>
    </div>
  );
}

export function AdminTableHead({ children }: { children: ReactNode }) {
  return <thead className={adminTheme.tableHead}>{children}</thead>;
}

export function AdminTableBody({ children }: { children: ReactNode }) {
  return <tbody className="divide-y divide-slate-100/90 bg-white">{children}</tbody>;
}

export function AdminTh({
  children,
  className = "",
  colSpan,
}: {
  children: ReactNode;
  className?: string;
  colSpan?: number;
}) {
  return (
    <th colSpan={colSpan} className={`${adminTheme.tableCellHead} ${className}`}>
      {children}
    </th>
  );
}

export function AdminTd({
  children,
  className = "",
  colSpan,
  title,
}: {
  children?: ReactNode;
  className?: string;
  colSpan?: number;
  title?: string;
}) {
  return (
    <td colSpan={colSpan} title={title} className={`${adminTheme.tableCell} ${className}`}>
      {children}
    </td>
  );
}
