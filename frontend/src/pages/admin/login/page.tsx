import { useState, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { ADMIN_DEFAULT_PATH } from "@/router/adminAccess";
import { PRODUCT_BRAND } from "@/config/productBrand";
import { AdminButton, AdminInput, AdminErrorBanner } from "@/components/admin/ui";

export default function LoginPage() {
  useDocumentTitle(`${PRODUCT_BRAND.name} 登录`);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await login(username, password);
      navigate(ADMIN_DEFAULT_PATH);
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败，请重试");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="admin-scope relative flex min-h-screen items-center justify-center overflow-hidden bg-[#f4f6f9] px-4 py-12 sm:px-6">
      {/* Ambient background */}
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,_rgba(37,99,235,0.12),_transparent_55%),radial-gradient(ellipse_at_bottom_right,_rgba(99,102,241,0.08),_transparent_45%)]"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -top-24 left-1/2 h-64 w-[36rem] -translate-x-1/2 rounded-full bg-blue-500/10 blur-3xl"
        aria-hidden
      />

      <div className="relative w-full max-w-[420px]">
        <div className="mb-8 text-center">
          <div className="mb-5 flex justify-center">
            <div className="rounded-2xl bg-white p-3 shadow-[0_8px_30px_rgba(15,23,42,0.08)] ring-1 ring-slate-200/80">
              <img className="h-10 w-10" src="/brand/inkless-mark.svg" alt="Inkless" />
            </div>
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
            {PRODUCT_BRAND.name} 管理后台
          </h1>
          <p className="mt-2 text-sm text-slate-500">使用管理员账号登录以继续</p>
        </div>

        <div className="rounded-2xl border border-slate-200/80 bg-white/90 p-6 shadow-[0_16px_48px_rgba(15,23,42,0.08)] backdrop-blur sm:p-8">
          <form className="space-y-5" onSubmit={handleSubmit}>
            {error ? (
              <AdminErrorBanner message={error} className="mb-0" onDismiss={() => setError("")} />
            ) : null}

            <div>
              <label htmlFor="username" className="mb-1.5 block text-sm font-medium text-slate-700">
                用户名
              </label>
              <AdminInput
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="输入用户名"
              />
            </div>

            <div>
              <label htmlFor="password" className="mb-1.5 block text-sm font-medium text-slate-700">
                密码
              </label>
              <AdminInput
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="输入密码"
              />
            </div>

            <AdminButton type="submit" disabled={loading} size="lg" className="w-full">
              {loading ? "登录中…" : "登录"}
            </AdminButton>
          </form>
        </div>

        <p className="mt-6 text-center text-xs text-slate-400">
          {PRODUCT_BRAND.fullName} · {PRODUCT_BRAND.domain}
        </p>
      </div>
    </div>
  );
}
