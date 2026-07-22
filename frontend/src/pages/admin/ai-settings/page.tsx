import { useCallback, useEffect, useState } from "react";
import {
  getAIConfig,
  testAIConfig,
  updateAIConfig,
  type AIConfigInput,
  type AIHealthResult,
  type AIProviderName,
} from "@/api/aiConfig";
import {
  AdminButton,
  AdminCard,
  AdminErrorBanner,
  AdminField,
  AdminInput,
  AdminLoading,
  AdminPageHeader,
  AdminSuccessBanner,
} from "@/components/admin/ui";
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
    return <AdminLoading />;
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="AI 配置"
        description="统一管理翻译、文章元数据（标题/slug/SEO）、知识问答和建站向导使用的 AI 提供方"
      />

      <AdminCard>
        <div className="grid gap-3 md:grid-cols-3">
          {PROVIDERS.map((provider) => {
            const active = form.provider === provider.value;
            return (
              <label
                key={provider.value}
                className={`cursor-pointer rounded-2xl border-2 p-4 transition-all ${
                  active
                    ? "border-blue-500 bg-blue-50/80 shadow-sm shadow-blue-600/10"
                    : "border-slate-200 hover:border-slate-300 hover:bg-slate-50/50"
                }`}
              >
                <input
                  className="sr-only"
                  type="radio"
                  name="provider"
                  value={provider.value}
                  checked={active}
                  onChange={() => {
                    setForm((current) => ({ ...current, provider: provider.value }));
                    setHealth(null);
                  }}
                />
                <span className="block text-sm font-semibold text-slate-900">{provider.label}</span>
                <span className="mt-1 block text-xs leading-5 text-slate-500">
                  {provider.description}
                </span>
              </label>
            );
          })}
        </div>

        {enabled && (
          <div className="mt-6 grid gap-4 md:grid-cols-2">
            <AdminField label="API Key" className="md:col-span-2">
              <AdminInput
                type="password"
                autoComplete="new-password"
                value={form.apiKey}
                onChange={(event) =>
                  setForm((current) => ({ ...current, apiKey: event.target.value }))
                }
                placeholder={
                  hasAPIKey ? `已配置 ${maskedAPIKey}，留空保持不变` : "请输入 API Key"
                }
                className="font-mono"
              />
            </AdminField>

            <AdminField label="Base URL">
              <AdminInput
                type="url"
                value={form.baseUrl}
                onChange={(event) =>
                  setForm((current) => ({ ...current, baseUrl: event.target.value }))
                }
                placeholder={
                  form.provider === "anthropic"
                    ? "https://api.anthropic.com"
                    : "https://api.openai.com/v1"
                }
                className="font-mono"
              />
            </AdminField>

            <AdminField label="默认模型">
              <AdminInput
                type="text"
                value={form.model}
                onChange={(event) =>
                  setForm((current) => ({ ...current, model: event.target.value }))
                }
                placeholder={form.provider === "anthropic" ? "claude-sonnet-4-5" : "gpt-4o-mini"}
                className="font-mono"
              />
            </AdminField>
          </div>
        )}

        {message?.kind === "success" ? (
          <AdminSuccessBanner message={message.text} className="mt-5 mb-0" />
        ) : null}
        {message?.kind === "error" ? (
          <AdminErrorBanner message={message.text} className="mt-5 mb-0" />
        ) : null}

        {health && (
          <div
            className={`mt-5 rounded-2xl border px-4 py-3 text-sm ${
              health.healthy
                ? "border-emerald-200 bg-emerald-50 text-emerald-800"
                : "border-amber-200 bg-amber-50 text-amber-800"
            }`}
          >
            {health.healthy ? "连接正常" : "当前未启用"}：{health.message || health.provider}
            {health.model ? `（${health.model}）` : ""}
          </div>
        )}

        <div className="mt-6 flex justify-end gap-2 border-t border-slate-100 pt-4">
          {enabled && (
            <AdminButton variant="secondary" onClick={handleTest} disabled={testing}>
              {testing ? "测试中…" : "测试连接"}
            </AdminButton>
          )}
          <AdminButton onClick={handleSave} disabled={saving}>
            {saving ? "保存中…" : "保存并应用"}
          </AdminButton>
        </div>
      </AdminCard>
    </div>
  );
}
