import { useState, useEffect } from "react";
import { getEmailSettings, updateEmailSettings, sendTestEmail } from "@/api/emailSettings";
import type { EmailConfig } from "./types";
import { defaultEmailConfig } from "./defaults";
import SmtpConfigTab from "./SmtpConfigTab";
import TemplateEditorTab from "./TemplateEditorTab";
import { AdminLoading, AdminPageHeader } from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

type TabKey = "smtp" | "autoReply" | "forward";

interface Tab {
  key: TabKey;
  label: string;
}

const TABS: Tab[] = [
  { key: "smtp", label: "SMTP 配置" },
  { key: "autoReply", label: "自动回复模板" },
  { key: "forward", label: "转发通知模板" },
];

function deepMerge(defaults: EmailConfig, loaded: Partial<EmailConfig>): EmailConfig {
  return {
    smtp: { ...defaults.smtp, ...loaded.smtp },
    receiver: { ...defaults.receiver, ...loaded.receiver },
    autoReply: { ...defaults.autoReply, ...loaded.autoReply },
    templates: {
      autoReply: {
        ...defaults.templates.autoReply,
        ...loaded.templates?.autoReply,
      },
      forward: {
        ...defaults.templates.forward,
        ...loaded.templates?.forward,
      },
    },
  };
}

export default function AdminEmailSettingsPage() {
  useDocumentTitle("邮箱设置");
  const [config, setConfig] = useState<EmailConfig>(defaultEmailConfig);
  const [activeTab, setActiveTab] = useState<TabKey>("smtp");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [status, setStatus] = useState<{ type: "success" | "error"; message: string } | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await getEmailSettings();
        if (!cancelled) {
          setConfig(deepMerge(defaultEmailConfig, data));
        }
      } catch {
        // API may 404 if never configured — use defaults
        if (!cancelled) {
          setConfig(defaultEmailConfig);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => { cancelled = true; };
  }, []);

  const showStatus = (type: "success" | "error", message: string) => {
    setStatus({ type, message });
    setTimeout(() => setStatus(null), 4000);
  };

  const handleSave = async () => {
    setSaving(true);
    setStatus(null);
    try {
      const updated = await updateEmailSettings(config);
      setConfig(deepMerge(defaultEmailConfig, updated));
      showStatus("success", "配置保存成功");
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || err?.response?.data?.message || "保存失败，请重试";
      showStatus("error", msg);
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    const to = prompt("请输入测试收件人邮箱地址：", config.smtp.from || "");
    if (!to) return;

    setTesting(true);
    setStatus(null);
    try {
      const result = await sendTestEmail(to);
      showStatus(result.success ? "success" : "error", result.message || "测试邮件已发送");
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || err?.response?.data?.message || "发送测试邮件失败";
      showStatus("error", msg);
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="邮箱设置"
        description="配置 SMTP 服务器、自动回复和转发通知邮件模板"
      />

      {/* Status Message */}
      {status && (
        <div
          className={`p-3 rounded-lg text-sm flex items-center gap-2 ${
            status.type === "success"
              ? "bg-green-50 text-green-700"
              : "bg-red-50 text-red-700"
          }`}
        >
          <svg className="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            {status.type === "success" ? (
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            ) : (
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            )}
          </svg>
          {status.message}
        </div>
      )}

      {loading ? (
        <AdminLoading />
      ) : (
        <div className="bg-white rounded-lg shadow">
          {/* Tab Navigation */}
          <div className="border-b border-gray-200">
            <nav className="flex -mb-px">
              {TABS.map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setActiveTab(tab.key)}
                  className={`px-6 py-3.5 text-sm font-medium border-b-2 transition-colors ${
                    activeTab === tab.key
                      ? "border-blue-600 text-blue-600"
                      : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>

          {/* Tab Content */}
          <div className="p-6">
            {activeTab === "smtp" && (
              <SmtpConfigTab
                config={config}
                onChange={setConfig}
                onSave={handleSave}
                onTest={handleTest}
                isSaving={saving}
                isTesting={testing}
              />
            )}
            {activeTab === "autoReply" && (
              <TemplateEditorTab
                templates={config.templates.autoReply}
                onChange={(autoReplyTemplates) =>
                  setConfig((prev) => ({
                    ...prev,
                    templates: { ...prev.templates, autoReply: autoReplyTemplates },
                  }))
                }
                onSave={handleSave}
                isSaving={saving}
              />
            )}
            {activeTab === "forward" && (
              <TemplateEditorTab
                templates={config.templates.forward}
                onChange={(forwardTemplates) =>
                  setConfig((prev) => ({
                    ...prev,
                    templates: { ...prev.templates, forward: forwardTemplates },
                  }))
                }
                onSave={handleSave}
                isSaving={saving}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
}
