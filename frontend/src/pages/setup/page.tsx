import { FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  completeSetup,
  saveSetupEnv,
  testDatabase,
  type DatabaseConfig,
} from "@/api/setup";
import { clearSetupStatusCache, useSetupStatus } from "@/hooks/useSetupStatus";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

type Step = "welcome" | "database" | "restart" | "site" | "admin" | "content" | "done";

export default function SetupPage() {
  const { t } = useTranslation("setup");
  useDocumentTitle(t("title"));
  const navigate = useNavigate();
  const { status, loading } = useSetupStatus();

  const [step, setStep] = useState<Step>("welcome");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [testingDb, setTestingDb] = useState(false);
  const [dbTestOk, setDbTestOk] = useState(false);
  const [savedEnvPath, setSavedEnvPath] = useState("");

  const [dbType, setDbType] = useState<"sqlite" | "postgres">("sqlite");
  const [sqlitePath, setSqlitePath] = useState("./data/impress.db");
  const [pgHost, setPgHost] = useState("localhost");
  const [pgPort, setPgPort] = useState("5432");
  const [pgUser, setPgUser] = useState("impress");
  const [pgPassword, setPgPassword] = useState("");
  const [pgDbName, setPgDbName] = useState("impress");
  const [pgSslMode, setPgSslMode] = useState("disable");

  const [nameZh, setNameZh] = useState("");
  const [nameEn, setNameEn] = useState("");
  const [defaultLocale, setDefaultLocale] = useState<"zh" | "en">("zh");
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [seedMode, setSeedMode] = useState<"blank" | "demo">("blank");

  const needsEnvConfig = status?.needsEnvConfig === true;

  const databaseConfig = useMemo((): DatabaseConfig => {
    if (dbType === "postgres") {
      return {
        type: "postgres",
        postgres: {
          host: pgHost.trim(),
          port: Number(pgPort) || 5432,
          user: pgUser.trim(),
          password: pgPassword,
          dbname: pgDbName.trim(),
          sslmode: pgSslMode.trim() || "disable",
        },
      };
    }
    return { type: "sqlite", sqlitePath: sqlitePath.trim() || "./data/impress.db" };
  }, [dbType, sqlitePath, pgHost, pgPort, pgUser, pgPassword, pgDbName, pgSslMode]);

  const stepOrder = useMemo((): Step[] => {
    if (needsEnvConfig) {
      return ["welcome", "database", "restart"];
    }
    return ["welcome", "site", "admin", "content"];
  }, [needsEnvConfig]);

  useEffect(() => {
    if (!loading && status?.installed) {
      navigate("/admin/login", { replace: true });
    }
  }, [loading, status, navigate]);

  const handleWelcomeNext = () => {
    setError("");
    setStep(needsEnvConfig ? "database" : "site");
  };

  const handleTestDatabase = async () => {
    setError("");
    setDbTestOk(false);
    setTestingDb(true);
    try {
      await testDatabase(databaseConfig);
      setDbTestOk(true);
    } catch (err) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error
          ?.message || t("errors.setupFailed");
      setError(message);
    } finally {
      setTestingDb(false);
    }
  };

  const handleSaveEnv = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      const result = await saveSetupEnv({ port: 8088, database: databaseConfig });
      setSavedEnvPath(result.envPath);
      clearSetupStatusCache();
      setStep("restart");
    } catch (err) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error
          ?.message || t("errors.setupFailed");
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleFinish = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (!nameZh.trim() && !nameEn.trim()) {
      setError(t("errors.siteNameRequired"));
      return;
    }
    if (password !== confirmPassword) {
      setError(t("errors.passwordMismatch"));
      return;
    }

    setSubmitting(true);
    try {
      await completeSetup({
        admin: { username: username.trim(), password },
        site: {
          name: { zh: nameZh.trim(), en: nameEn.trim() },
          defaultLocale,
        },
        seedMode,
      });
      clearSetupStatusCache();
      setStep("done");
      setTimeout(() => navigate("/admin/login", { replace: true }), 1500);
    } catch (err) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error
          ?.message || t("errors.setupFailed");
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading || status?.installed) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 text-gray-600">
        …
      </div>
    );
  }

  const stepIndex = stepOrder.indexOf(step);

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">{t("title")}</h1>
          <p className="mt-2 text-sm text-gray-600">{t("subtitle")}</p>
        </div>

        {step !== "done" && stepIndex >= 0 && (
          <div className="flex flex-wrap justify-center gap-2 mb-6 text-xs text-gray-500">
            {stepOrder.map((s, i) => (
              <span key={s} className={i <= stepIndex ? "text-blue-600 font-medium" : ""}>
                {t(`steps.${s}`)}
                {i < stepOrder.length - 1 ? " → " : ""}
              </span>
            ))}
          </div>
        )}

        <div className="bg-white shadow rounded-lg p-6">
          {error && (
            <div className="mb-4 rounded-md bg-red-50 p-3 text-sm text-red-800">{error}</div>
          )}
          {dbTestOk && step === "database" && (
            <div className="mb-4 rounded-md bg-green-50 p-3 text-sm text-green-800">
              {t("database.testSuccess")}
            </div>
          )}

          {step === "welcome" && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold text-gray-900">{t("welcome.heading")}</h2>
              <p className="text-sm text-gray-600">
                {needsEnvConfig ? t("welcome.body") : t("welcome.bodyConfigured")}
              </p>
              {!needsEnvConfig && (
                <p className="text-sm text-gray-500">
                  {t("welcome.database")}:{" "}
                  <span className="font-mono">{status?.databaseType ?? "sqlite"}</span>
                </p>
              )}
              <button
                type="button"
                onClick={handleWelcomeNext}
                className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {t("welcome.next")}
              </button>
            </div>
          )}

          {step === "database" && (
            <form className="space-y-4" onSubmit={handleSaveEnv}>
              <h2 className="text-lg font-medium text-gray-900">{t("database.heading")}</h2>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("database.type")}
                </label>
                <select
                  value={dbType}
                  onChange={(e) => {
                    setDbType(e.target.value as "sqlite" | "postgres");
                    setDbTestOk(false);
                  }}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                >
                  <option value="sqlite">{t("database.sqlite")}</option>
                  <option value="postgres">{t("database.postgres")}</option>
                </select>
              </div>

              {dbType === "sqlite" ? (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    {t("database.sqlitePath")}
                  </label>
                  <input
                    value={sqlitePath}
                    onChange={(e) => {
                      setSqlitePath(e.target.value);
                      setDbTestOk(false);
                    }}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 font-mono text-sm"
                  />
                </div>
              ) : (
                <div className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div className="col-span-2 sm:col-span-1">
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        {t("database.host")}
                      </label>
                      <input
                        value={pgHost}
                        onChange={(e) => {
                          setPgHost(e.target.value);
                          setDbTestOk(false);
                        }}
                        className="w-full border border-gray-300 rounded-md px-3 py-2"
                      />
                    </div>
                    <div className="col-span-2 sm:col-span-1">
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        {t("database.port")}
                      </label>
                      <input
                        value={pgPort}
                        onChange={(e) => {
                          setPgPort(e.target.value);
                          setDbTestOk(false);
                        }}
                        className="w-full border border-gray-300 rounded-md px-3 py-2"
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      {t("database.user")}
                    </label>
                    <input
                      value={pgUser}
                      onChange={(e) => {
                        setPgUser(e.target.value);
                        setDbTestOk(false);
                      }}
                      className="w-full border border-gray-300 rounded-md px-3 py-2"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      {t("database.password")}
                    </label>
                    <input
                      type="password"
                      value={pgPassword}
                      onChange={(e) => {
                        setPgPassword(e.target.value);
                        setDbTestOk(false);
                      }}
                      className="w-full border border-gray-300 rounded-md px-3 py-2"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        {t("database.dbname")}
                      </label>
                      <input
                        value={pgDbName}
                        onChange={(e) => {
                          setPgDbName(e.target.value);
                          setDbTestOk(false);
                        }}
                        className="w-full border border-gray-300 rounded-md px-3 py-2"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        {t("database.sslmode")}
                      </label>
                      <input
                        value={pgSslMode}
                        onChange={(e) => {
                          setPgSslMode(e.target.value);
                          setDbTestOk(false);
                        }}
                        className="w-full border border-gray-300 rounded-md px-3 py-2"
                      />
                    </div>
                  </div>
                </div>
              )}

              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setStep("welcome")}
                  className="flex-1 py-2 border border-gray-300 rounded-md"
                >
                  {t("actions.back")}
                </button>
                <button
                  type="button"
                  onClick={handleTestDatabase}
                  disabled={testingDb || submitting}
                  className="flex-1 py-2 border border-blue-600 text-blue-600 rounded-md disabled:opacity-60"
                >
                  {testingDb ? t("database.testing") : t("database.test")}
                </button>
              </div>
              <button
                type="submit"
                disabled={submitting}
                className="w-full py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-60"
              >
                {submitting ? t("database.saving") : t("database.save")}
              </button>
            </form>
          )}

          {step === "restart" && (
            <div className="space-y-4">
              <h2 className="text-xl font-semibold text-gray-900">{t("restart.heading")}</h2>
              <p className="text-sm text-gray-600">
                {t("restart.body", {
                  path: savedEnvPath || status?.envFilePath || ".env",
                })}
              </p>
              <button
                type="button"
                onClick={() => window.location.reload()}
                className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {t("restart.reload")}
              </button>
            </div>
          )}

          {step === "site" && (
            <form
              className="space-y-4"
              onSubmit={(e) => {
                e.preventDefault();
                if (!nameZh.trim() && !nameEn.trim()) {
                  setError(t("errors.siteNameRequired"));
                  return;
                }
                setError("");
                setStep("admin");
              }}
            >
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("site.nameZh")}
                </label>
                <input
                  value={nameZh}
                  onChange={(e) => setNameZh(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("site.nameEn")}
                </label>
                <input
                  value={nameEn}
                  onChange={(e) => setNameEn(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("site.defaultLocale")}
                </label>
                <select
                  value={defaultLocale}
                  onChange={(e) => setDefaultLocale(e.target.value as "zh" | "en")}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                >
                  <option value="zh">{t("site.localeZh")}</option>
                  <option value="en">{t("site.localeEn")}</option>
                </select>
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setStep("welcome")}
                  className="flex-1 py-2 border border-gray-300 rounded-md"
                >
                  {t("actions.back")}
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  {t("actions.next")}
                </button>
              </div>
            </form>
          )}

          {step === "admin" && (
            <form
              className="space-y-4"
              onSubmit={(e) => {
                e.preventDefault();
                if (password !== confirmPassword) {
                  setError(t("errors.passwordMismatch"));
                  return;
                }
                setError("");
                setStep("content");
              }}
            >
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("admin.username")}
                </label>
                <input
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  autoComplete="username"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("admin.password")}
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  autoComplete="new-password"
                />
                <p className="mt-1 text-xs text-gray-500">{t("admin.hint")}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t("admin.confirmPassword")}
                </label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  autoComplete="new-password"
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setStep("site")}
                  className="flex-1 py-2 border border-gray-300 rounded-md"
                >
                  {t("actions.back")}
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  {t("actions.next")}
                </button>
              </div>
            </form>
          )}

          {step === "content" && (
            <form className="space-y-4" onSubmit={handleFinish}>
              <h2 className="text-lg font-medium text-gray-900">{t("content.heading")}</h2>
              <label className="flex items-start gap-3 p-3 border rounded-md cursor-pointer has-[:checked]:border-blue-500">
                <input
                  type="radio"
                  name="seedMode"
                  value="blank"
                  checked={seedMode === "blank"}
                  onChange={() => setSeedMode("blank")}
                  className="mt-1"
                />
                <span>
                  <span className="font-medium block">{t("content.blank")}</span>
                  <span className="text-sm text-gray-500">{t("content.blankDesc")}</span>
                </span>
              </label>
              <label className="flex items-start gap-3 p-3 border rounded-md cursor-pointer has-[:checked]:border-blue-500">
                <input
                  type="radio"
                  name="seedMode"
                  value="demo"
                  checked={seedMode === "demo"}
                  onChange={() => setSeedMode("demo")}
                  className="mt-1"
                />
                <span>
                  <span className="font-medium block">{t("content.demo")}</span>
                  <span className="text-sm text-gray-500">{t("content.demoDesc")}</span>
                </span>
              </label>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setStep("admin")}
                  className="flex-1 py-2 border border-gray-300 rounded-md"
                  disabled={submitting}
                >
                  {t("actions.back")}
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="flex-1 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-60"
                >
                  {submitting ? t("actions.finishing") : t("actions.finish")}
                </button>
              </div>
            </form>
          )}

          {step === "done" && (
            <div className="text-center space-y-2">
              <h2 className="text-xl font-semibold text-green-700">{t("success.heading")}</h2>
              <p className="text-sm text-gray-600">{t("success.body")}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
