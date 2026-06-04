import { useCallback, useEffect, useState } from "react";

const STORAGE_KEY = "impress.comment.guest";

export interface CommentGuestIdentity {
  authorName: string;
  authorEmail: string;
}

function readStored(): CommentGuestIdentity {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { authorName: "", authorEmail: "" };
    const parsed = JSON.parse(raw) as CommentGuestIdentity;
    return {
      authorName: typeof parsed.authorName === "string" ? parsed.authorName : "",
      authorEmail: typeof parsed.authorEmail === "string" ? parsed.authorEmail : "",
    };
  } catch {
    return { authorName: "", authorEmail: "" };
  }
}

export function saveCommentGuestIdentity(identity: CommentGuestIdentity): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(identity));
  } catch {
    /* private mode / quota */
  }
}

/** Persists display name + email for repeat visits (not authentication). */
export function useCommentIdentity() {
  const [identity, setIdentity] = useState<CommentGuestIdentity>({ authorName: "", authorEmail: "" });

  useEffect(() => {
    setIdentity(readStored());
  }, []);

  const persist = useCallback((next: CommentGuestIdentity) => {
    setIdentity(next);
    saveCommentGuestIdentity(next);
  }, []);

  return { identity, persist };
}
