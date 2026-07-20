import { useState } from "react";

interface EmbedUrlModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (result: { type: "youtube" | "bilibili" | "iframe"; url: string }) => void;
}

function detectEmbedType(url: string): { type: "youtube" | "bilibili" | "iframe"; url: string } {
  // YouTube
  const ytMatch = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]+)/);
  if (ytMatch) return { type: "youtube", url: `https://www.youtube.com/watch?v=${ytMatch[1]}` };

  // Bilibili
  const biliMatch = url.match(/bilibili\.com\/video\/(BV[a-zA-Z0-9]+)/);
  if (biliMatch) return { type: "bilibili", url: `https://player.bilibili.com/player.html?bvid=${biliMatch[1]}&autoplay=0` };

  return { type: "iframe", url };
}

export default function EmbedUrlModal({ open, onClose, onConfirm }: EmbedUrlModalProps) {
  const [url, setUrl] = useState("");

  const handleConfirm = () => {
    const trimmed = url.trim();
    if (!trimmed) return;
    onConfirm(detectEmbedType(trimmed));
    setUrl("");
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/45 p-4 backdrop-blur-[2px]">
      <div className="bg-white rounded-xl shadow-xl w-[90vw] max-w-lg">
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-900">嵌入外部内容</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>
        <div className="p-6">
          <p className="text-sm text-slate-600 mb-3">支持 YouTube、Bilibili 视频链接或任意网页 URL</p>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://..."
            className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            onKeyDown={(e) => { if (e.key === "Enter") handleConfirm(); }}
            autoFocus
          />
        </div>
        <div className="px-6 py-3 border-t border-slate-100 flex justify-end gap-2">
          <button onClick={onClose} className="inline-flex h-9 items-center rounded-xl border border-slate-200 bg-white px-4 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50">取消</button>
          <button
            onClick={handleConfirm}
            disabled={!url.trim()}
            className="inline-flex h-9 items-center rounded-xl bg-blue-600 px-4 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:opacity-50"
          >
            插入
          </button>
        </div>
      </div>
    </div>
  );
}
