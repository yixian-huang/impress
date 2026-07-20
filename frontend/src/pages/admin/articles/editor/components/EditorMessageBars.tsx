export function EditorMessageBars({
  error,
  onClearError,
  successMessage,
  scheduleMessage,
  onClearSuccess,
}: {
  error: string | null;
  onClearError: () => void;
  successMessage: string;
  scheduleMessage: string;
  onClearSuccess: () => void;
}) {
  return (
    <>
      {error && (
        <div className="px-4 py-2 bg-red-50 border-b border-red-200 text-red-800 text-sm flex-shrink-0">
          {error}
          <button type="button" onClick={onClearError} className="ml-2 text-red-600 hover:text-red-800">
            &times;
          </button>
        </div>
      )}
      {(scheduleMessage || successMessage) && (
        <div className="px-4 py-2 bg-green-50 border-b border-green-200 text-green-800 text-sm flex-shrink-0">
          {successMessage || scheduleMessage}
          <button type="button" onClick={onClearSuccess} className="ml-2 text-green-600 hover:text-green-800">
            &times;
          </button>
        </div>
      )}
    </>
  );
}
