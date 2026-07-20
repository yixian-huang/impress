import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes, TextareaHTMLAttributes } from "react";
import { adminTheme } from "./adminTheme";

function cx(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export function AdminLabel({
  children,
  htmlFor,
  className = "",
}: {
  children: ReactNode;
  htmlFor?: string;
  className?: string;
}) {
  return (
    <label htmlFor={htmlFor} className={cx("mb-1.5 block", adminTheme.label, className)}>
      {children}
    </label>
  );
}

export function AdminInput({ className = "", ...rest }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cx(adminTheme.input, className)} {...rest} />;
}

export function AdminSelect({
  className = "",
  children,
  ...rest
}: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select className={cx(adminTheme.select, className)} {...rest}>
      {children}
    </select>
  );
}

export function AdminTextarea({
  className = "",
  ...rest
}: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className={cx(adminTheme.textarea, className)} {...rest} />;
}

export function AdminToolbar({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return <div className={cx(adminTheme.toolbar, className)}>{children}</div>;
}
