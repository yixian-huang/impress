/**
 * Minimal line-based diff (Myers-inspired LCS) for version comparison UI.
 * No external dependency — good enough for article body/title snapshots.
 */

export type DiffOp = "equal" | "add" | "remove";

export interface DiffLine {
  op: DiffOp;
  text: string;
  leftLine?: number;
  rightLine?: number;
}

/** Compute a line-level unified diff between two strings. */
export function diffLines(left: string, right: string): DiffLine[] {
  const a = left.replace(/\r\n/g, "\n").split("\n");
  const b = right.replace(/\r\n/g, "\n").split("\n");
  const n = a.length;
  const m = b.length;

  // LCS DP table (compact: only two rows)
  // For long bodies we cap to avoid O(n*m) blowups — fall back to coarse blocks.
  const MAX = 4000;
  if (n * m > MAX * MAX) {
    return coarseDiff(a, b);
  }

  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0));
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      if (a[i] === b[j]) dp[i][j] = dp[i + 1][j + 1] + 1;
      else dp[i][j] = Math.max(dp[i + 1][j], dp[i][j + 1]);
    }
  }

  const result: DiffLine[] = [];
  let i = 0;
  let j = 0;
  let leftLine = 1;
  let rightLine = 1;
  while (i < n && j < m) {
    if (a[i] === b[j]) {
      result.push({ op: "equal", text: a[i], leftLine, rightLine });
      i++;
      j++;
      leftLine++;
      rightLine++;
    } else if (dp[i + 1][j] >= dp[i][j + 1]) {
      result.push({ op: "remove", text: a[i], leftLine });
      i++;
      leftLine++;
    } else {
      result.push({ op: "add", text: b[j], rightLine });
      j++;
      rightLine++;
    }
  }
  while (i < n) {
    result.push({ op: "remove", text: a[i], leftLine });
    i++;
    leftLine++;
  }
  while (j < m) {
    result.push({ op: "add", text: b[j], rightLine });
    j++;
    rightLine++;
  }
  return result;
}

function coarseDiff(a: string[], b: string[]): DiffLine[] {
  // When content is huge, emit whole-file remove/add if different.
  if (a.join("\n") === b.join("\n")) {
    return a.map((text, idx) => ({
      op: "equal" as const,
      text,
      leftLine: idx + 1,
      rightLine: idx + 1,
    }));
  }
  const out: DiffLine[] = [];
  a.forEach((text, idx) => out.push({ op: "remove", text, leftLine: idx + 1 }));
  b.forEach((text, idx) => out.push({ op: "add", text, rightLine: idx + 1 }));
  return out;
}

/** Strip simple HTML tags for readable text comparison of rich-text bodies. */
export function htmlToPlainText(html: string): string {
  if (!html) return "";
  return html
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<\/p>/gi, "\n")
    .replace(/<\/h[1-6]>/gi, "\n")
    .replace(/<\/li>/gi, "\n")
    .replace(/<\/div>/gi, "\n")
    .replace(/<[^>]+>/g, "")
    .replace(/&nbsp;/g, " ")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&")
    .replace(/&quot;/g, '"')
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}
