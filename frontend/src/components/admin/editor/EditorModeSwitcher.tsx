interface EditorModeSwitcherProps {
  mode: "richtext" | "markdown";
  onModeChange: (mode: "richtext" | "markdown") => void;
}

export default function EditorModeSwitcher({ mode, onModeChange }: EditorModeSwitcherProps) {
  return (
    <div className="flex items-center gap-1 bg-slate-100 rounded-lg p-0.5">
      <button
        onClick={() => onModeChange("richtext")}
        className={`px-3 py-1 text-xs rounded-xl transition ${
          mode === "richtext" ? "bg-white shadow text-slate-900" : "text-slate-500 hover:text-slate-700"
        }`}
      >
        Rich Text
      </button>
      <button
        onClick={() => onModeChange("markdown")}
        className={`px-3 py-1 text-xs rounded-xl transition ${
          mode === "markdown" ? "bg-white shadow text-slate-900" : "text-slate-500 hover:text-slate-700"
        }`}
      >
        Markdown
      </button>
    </div>
  );
}
