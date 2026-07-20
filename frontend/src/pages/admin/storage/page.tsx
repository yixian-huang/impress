import { useState, useEffect, useCallback } from "react";
import {
  getStorageConfig,
  updateStorageConfig,
  testStorageConnection,
  type StorageConfig,
  type UpdateStorageConfigRequest,
} from "@/api/storage";
import {
  AdminErrorBanner,
  AdminLoading,
  AdminPageHeader,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

type Strategy = "local" | "s3" | "oss";

interface StorageFormData {
  strategy: Strategy;
  bucket: string;
  region: string;
  endpoint: string;
  access_key: string;
  secret_key: string;
  base_path: string;
}

const emptyForm: StorageFormData = {
  strategy: "local",
  bucket: "",
  region: "",
  endpoint: "",
  access_key: "",
  secret_key: "",
  base_path: "",
};

const STRATEGIES = [
  { value: "local" as Strategy, label: "本地存储", description: "文件存储在服务器本地磁盘" },
  { value: "s3" as Strategy, label: "Amazon S3", description: "AWS S3 或兼容 S3 的对象存储" },
  { value: "oss" as Strategy, label: "阿里云 OSS", description: "阿里云对象存储服务" },
];

export default function AdminStoragePage() {
  useDocumentTitle("存储配置");
  const [config, setConfig] = useState<StorageConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [form, setForm] = useState<StorageFormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [saveSuccess, setSaveSuccess] = useState(false);

  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);

  const fetchConfig = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await getStorageConfig();
      setConfig(data);
      setForm({
        strategy: (data.strategy as Strategy) || "local",
        bucket: data.bucket || "",
        region: data.region || "",
        endpoint: data.endpoint || "",
        access_key: data.accessKey || "",
        secret_key: "",
        base_path: data.basePath || "",
      });
    } catch {
      setError("加载存储配置失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  const handleSave = async () => {
    setSaveError("");
    setSaveSuccess(false);
    if (form.strategy !== "local" && !form.bucket.trim()) {
      setSaveError("请填写存储桶名称");
      return;
    }
    setSaving(true);
    try {
      const data: UpdateStorageConfigRequest = {
        strategy: form.strategy,
        bucket: form.bucket,
        region: form.region,
        endpoint: form.endpoint,
        accessKey: form.access_key,
        basePath: form.base_path,
      };
      if (form.secret_key) {
        data.secretKey = form.secret_key;
      }
      const updated = await updateStorageConfig(data);
      setConfig(updated);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || "保存失败";
      setSaveError(msg);
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const result = await testStorageConnection();
      setTestResult(result);
    } catch (err: any) {
      setTestResult({
        success: false,
        message: err?.response?.data?.error?.message || "连接测试失败",
      });
    } finally {
      setTesting(false);
    }
  };

  const isRemote = form.strategy === "s3" || form.strategy === "oss";

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="存储配置"
        description="配置文件上传和存储方式"
      />

      {error && <AdminErrorBanner message={error} />}

      {loading ? (
        <AdminLoading />
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Config Form */}
          <div className="lg:col-span-2 bg-white rounded-lg shadow p-6 space-y-6">
            {/* Strategy Selector */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-3">存储策略</label>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {STRATEGIES.map((s) => (
                  <label
                    key={s.value}
                    className={`relative flex flex-col gap-1 p-4 border-2 rounded-lg cursor-pointer transition-colors ${
                      form.strategy === s.value
                        ? "border-blue-500 bg-blue-50"
                        : "border-gray-200 hover:border-gray-300"
                    }`}
                  >
                    <input
                      type="radio"
                      name="strategy"
                      value={s.value}
                      checked={form.strategy === s.value}
                      onChange={(e) => {
                        setForm((f) => ({ ...f, strategy: e.target.value as Strategy }));
                        setTestResult(null);
                      }}
                      className="sr-only"
                    />
                    <span className="text-sm font-semibold text-gray-900">{s.label}</span>
                    <span className="text-xs text-gray-500">{s.description}</span>
                    {form.strategy === s.value && (
                      <div className="absolute top-2 right-2 w-4 h-4 bg-blue-500 rounded-full flex items-center justify-center">
                        <svg className="w-2.5 h-2.5 text-white" fill="currentColor" viewBox="0 0 12 12">
                          <path d="M10.28 2.28L3.989 8.575 1.695 6.28A1 1 0 00.28 7.695l3 3a1 1 0 001.414 0l7-7A1 1 0 0010.28 2.28z" />
                        </svg>
                      </div>
                    )}
                  </label>
                ))}
              </div>
            </div>

            {/* Remote Storage Fields */}
            {isRemote && (
              <div className="space-y-4">
                <div className="border-t border-gray-100 pt-4">
                  <h3 className="text-sm font-semibold text-gray-700 mb-4">
                    {form.strategy === "s3" ? "S3 配置" : "OSS 配置"}
                  </h3>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        存储桶 (Bucket) <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="text"
                        value={form.bucket}
                        onChange={(e) => setForm((f) => ({ ...f, bucket: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                        placeholder="my-bucket"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">地区 (Region)</label>
                      <input
                        type="text"
                        value={form.region}
                        onChange={(e) => setForm((f) => ({ ...f, region: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                        placeholder={form.strategy === "s3" ? "us-east-1" : "oss-cn-hangzhou"}
                      />
                    </div>

                    <div className="sm:col-span-2">
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Endpoint {form.strategy === "s3" ? "(自定义端点，可选)" : ""}
                      </label>
                      <input
                        type="text"
                        value={form.endpoint}
                        onChange={(e) => setForm((f) => ({ ...f, endpoint: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                        placeholder={form.strategy === "s3" ? "https://s3.amazonaws.com" : "https://oss-cn-hangzhou.aliyuncs.com"}
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Access Key</label>
                      <input
                        type="text"
                        value={form.access_key}
                        onChange={(e) => setForm((f) => ({ ...f, access_key: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                        placeholder="输入 Access Key"
                        autoComplete="off"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Secret Key{" "}
                        {config?.hasSecretKey && (
                          <span className="text-xs text-green-600 font-normal">(已配置，留空不修改)</span>
                        )}
                      </label>
                      <input
                        type="password"
                        value={form.secret_key}
                        onChange={(e) => setForm((f) => ({ ...f, secret_key: e.target.value }))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                        placeholder={config?.hasSecretKey ? "留空不修改" : "输入 Secret Key"}
                        autoComplete="new-password"
                      />
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Base Path (for all strategies) */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">基础路径 (Base Path)</label>
              <input
                type="text"
                value={form.base_path}
                onChange={(e) => setForm((f) => ({ ...f, base_path: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                placeholder={form.strategy === "local" ? "uploads/" : "media/"}
              />
              <p className="mt-1 text-xs text-gray-500">
                {form.strategy === "local" ? "本地存储的相对目录路径" : "存储桶内的前缀路径"}
              </p>
            </div>

            {/* Save Error / Success */}
            {saveError && (
              <div className="p-3 bg-red-50 text-red-700 rounded-lg text-sm">{saveError}</div>
            )}
            {saveSuccess && (
              <div className="p-3 bg-green-50 text-green-700 rounded-lg text-sm flex items-center gap-2">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                配置已保存
              </div>
            )}

            {/* Actions */}
            <div className="flex items-center justify-between pt-2 border-t border-gray-100">
              <div className="flex items-center gap-3">
                {isRemote && (
                  <button
                    onClick={handleTest}
                    disabled={testing}
                    className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 transition-colors"
                  >
                    {testing ? "测试中..." : "测试连接"}
                  </button>
                )}
                {testResult && (
                  <span className={`text-sm ${testResult.success ? "text-green-600" : "text-red-600"} flex items-center gap-1`}>
                    <svg
                      className="w-4 h-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      {testResult.success ? (
                        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                      ) : (
                        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                      )}
                    </svg>
                    {testResult.message}
                  </span>
                )}
              </div>
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-6 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                {saving ? "保存中..." : "保存配置"}
              </button>
            </div>
          </div>

          {/* Current Config Info */}
          <div className="bg-white rounded-lg shadow p-6 space-y-4 h-fit">
            <h3 className="text-sm font-semibold text-gray-700">当前配置</h3>
            {config ? (
              <dl className="space-y-3">
                <div>
                  <dt className="text-xs text-gray-500">存储策略</dt>
                  <dd className="text-sm font-medium text-gray-900 mt-0.5">
                    {STRATEGIES.find((s) => s.value === config.strategy)?.label || config.strategy}
                  </dd>
                </div>
                {config.bucket && (
                  <div>
                    <dt className="text-xs text-gray-500">存储桶</dt>
                    <dd className="text-sm font-mono text-gray-900 mt-0.5">{config.bucket}</dd>
                  </div>
                )}
                {config.region && (
                  <div>
                    <dt className="text-xs text-gray-500">地区</dt>
                    <dd className="text-sm font-mono text-gray-900 mt-0.5">{config.region}</dd>
                  </div>
                )}
                {config.endpoint && (
                  <div>
                    <dt className="text-xs text-gray-500">Endpoint</dt>
                    <dd className="text-sm font-mono text-gray-900 mt-0.5 break-all">{config.endpoint}</dd>
                  </div>
                )}
                {config.accessKey && (
                  <div>
                    <dt className="text-xs text-gray-500">Access Key</dt>
                    <dd className="text-sm font-mono text-gray-900 mt-0.5">
                      {config.accessKey.slice(0, 8)}{"*".repeat(Math.max(0, config.accessKey.length - 8))}
                    </dd>
                  </div>
                )}
                <div>
                  <dt className="text-xs text-gray-500">Secret Key</dt>
                  <dd className="text-sm mt-0.5">
                    {config.hasSecretKey ? (
                      <span className="text-green-600">已配置</span>
                    ) : (
                      <span className="text-gray-400">未配置</span>
                    )}
                  </dd>
                </div>
                {config.basePath && (
                  <div>
                    <dt className="text-xs text-gray-500">基础路径</dt>
                    <dd className="text-sm font-mono text-gray-900 mt-0.5">{config.basePath}</dd>
                  </div>
                )}
                {config.updatedAt && (
                  <div className="pt-2 border-t border-gray-100">
                    <dt className="text-xs text-gray-500">最后更新</dt>
                    <dd className="text-xs text-gray-600 mt-0.5">
                      {new Date(config.updatedAt).toLocaleString("zh-CN")}
                    </dd>
                  </div>
                )}
              </dl>
            ) : (
              <p className="text-sm text-gray-500">暂无配置</p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
