import { useEffect, type RefObject } from "react";

export function useOutsideClick(
  ref: RefObject<HTMLElement | null>,
  open: boolean,
  onOutside: () => void,
) {
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onOutside();
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [ref, open, onOutside]);
}
