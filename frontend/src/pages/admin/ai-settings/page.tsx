import { useCallback, useEffect, useState } from "react";
import {
  getAIConfig,
  testAIConfig,
  updateAIConfig,
  type AIConfigInput,
  type AIHealthResult,
  type AIProviderName,
} from "@/api/aiConfig";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

interface AIFormState {
  provider: AIProviderName;
  apiKey: string;
  baseUrl: string;
  model: string;
}

const EMPTY_FORM: AIFormState = {
  provider: "disabled",
  apiKey: "",
  baseUrl: "",
  model: "",
};

const PROVIDERS: Array<{ value: AIProviderName; label: string; description: string }> = [
  { value: "disabled", label: "停用 AI", description: "翻译、问答和建站向导会明确返回未配置状态" },
  { value: "openai", label: "OpenAI 兼容接口", description: "支持 OpenAI 及兼容 Chat Completions 的服务" },
  { value: "anthropic", label: "Anthropic", description: "使用 Anthropic Messages API" },
];

function errorMessage(error: unknown, fallback: string): string {
  const response = (error as {
    response?: { data?: { error?: { message?: string } } };
  })?.response;
  return response?.data?.error?.message || fallback;
}

function toRequest(form: AIFormState): AIConfigInput {
  const input: AIConfigInput = {
    provider: form.provider,
    base_url: form.baseUrl.trim(),
    model: form.model.trim(),
  };
  if (form.apiKey.trim()) {
    input.api_key = form.apiKey.trim();
  }
  return input;
}

export default function AdminAISettingsPage() {
  useDocumentTitle("AI 配置");
  const [form, setForm] = useState<AIFormState>(EMPTY_FORM);
  const [hasAPIKey, setHasAPIKey] = useState(false);
  const [maskedAPIKey, setMaskedAPIKey] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [message, setMessage] = useState<{ kind: "success" | "error"; text: string } | null>(null);
  const [health, setHealth] = useState<AIHealthResult | null>(null);

  const loadConfig = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const config = await getAIConfig();
      setForm({
        provider: config.provider || "disabled",
        apiKey: "",
        baseUrl: config.base_url || "",
        model: config.model || "",
      });
      setHasAPIKey(config.has_api_key);
      setMaskedAPIKey(config.api_key_masked || "");
    } catch (error) {
      setMessage({ kind: "error", text: errorMessage(error, "加载 AI 配置失败") });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadConfig();
  }, [loadConfig]);

  const handleSave = async () => {
    setSaving(true);
    setMessage(null);
    setHealth(null);
    try {
      const config = await updateAIConfig(toRequest(form));
      setHasAPIKey(config.has_api_key);
      setMaskedAPIKey(config.api_key_masked || "");
      setForm((current) => ({ ...current, apiKey: "" }));
      setMessage({ kind: "success", text: "配置已保存并立即应用到运行时" });
    } catch (error) {
      setMessage({ kind: "error", text: errorMessage(error, "保存 AI 配置失败") });
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    setMessage(null);
    setHealth(null);
    try {
      const result = await testAIConfig(toRequest(form));
      setHealth(result);
    } catch (error) {
      setMessage({ kind: "error", text: errorMessage(error, "AI 连接测试失败") });
    } finally {
      setTesting(false);
    }
  };

  const enabled = form.provider !== "disabled";

  if (loading) {
    return <div className="rounded-lg bg-white p-8 text-center text-gray-500 shadow">加载中...</div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">AI 配置</h1>
        <p className="mt-1 text-sm text-gray-500">统一管理翻译、知识问答和建站向导使用的 AI 提供方</p>
      </div>

      <div className="rounded-lg bg-white p-6 shadow">
        <div className="grid gap-3 md:grid-cols-3">
          {PROVIDERS.map((provider) => (
            <label
              key={provider.value}
              className={`cursor-pointer rounded-lg border-2 p-4 transition-colors ${
                form.provider === provider.value
                  ? "border-blue-500 bg-blue-50"
                  : "border-gray-200 hover:border-gray-300"
              }`}
            >
              <input
                className="sr-only"
                type="radio"
                name="provider"
                value={provider.value}
                checked={form.provider === provider.value}
                onChange={() => {
                  setForm((current) => ({ ...current, provider: provider.value }));
                  setHealth(null);
                }}
              />
              <span className="block text-sm font-semibold text-gray-900">{provider.label}</span>
              <span className="mt-1 block text-xs leading-5 text-gray-500">{provider.description}</span>
            </label>
          ))}
        </div>

        {enabled && (
          <div className="mt-6 grid gap-4 md:grid-cols-2">
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium text-gray-700">API Key</label>
              <input
                type="password"
                autoComplete="new-password"
                value={form.apiKey}
                onChange={(event) => setForm((current) => ({ ...current, apiKey: event.target.value }))}
                placeholder={hasAPIKey ? `已配置 ${maskedAPIKey}，留空保持不变` : "请输入 API Key"}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">Base URL</label>
              <input
                type="url"
                value={form.baseUrl}
                onChange={(event) => setForm((current) => ({ ...current, baseUrl: event.target.value }))}
                placeholder={form.provider === "anthropic" ? "https://api.anthropic.com" : "https://api.openai.com/v1"}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">默认模型</label>
              <input
                type="text"
                value={form.model}
                onChange={(event) => setForm((current) => ({ ...current, model: event.target.value }))}
                placeholder={form.provider === "anthropic" ? "claude-sonnet-4-5" : "gpt-4o-mini"}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
        )}

        {message && (
          <div
            className={`mt-5 rounded-lg p-3 text-sm ${
              message.kind === "success" ? "bg-green-50 text-green-700" : "bg-red-50 text-red-700"
            }`}
          >
            {message.text}
          </div>
        )}

        {health && (
          <div
            className={`mt-5 rounded-lg p-3 text-sm ${
              health.healthy ? "bg-green-50 text-green-700" : "bg-amber-50 text-amber-700"
            }`}
          >
            {health.healthy ? "连接正常" : "当前未启用"}：{health.message || health.provider}
            {health.model ? `（${health.model}）` : ""}
          </div>
        )}

        <div className="mt-6 flex justify-end gap-3 border-t border-gray-100 pt-4">
          {enabled && (
            <button
              type="button"
              onClick={handleTest}
              disabled={testing}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              {testing ? "测试中..." : "测试连接"}
            </button>
          )}
          <button
            type="button"
            onClick={handleSave}
            disabled={saving}
            className="rounded-lg bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? "保存中..." : "保存并应用"}
          </button>
        </div>
      </div>
    </div>
  );
}
