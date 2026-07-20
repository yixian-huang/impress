import { useState } from "react";
import {
  listUsers,
  createUser,
  updateUser,
  deleteUser,
  type UserDTO,
  type CreateUserRequest,
  type UpdateUserRequest,
} from "@/api/users";
import {
  AdminBadge,
  AdminButton,
  AdminCheckbox,
  AdminErrorBanner,
  AdminField,
  AdminInput,
  AdminLoading,
  AdminModal,
  AdminPageHeader,
  AdminPagination,
  AdminSelect,
  AdminTable,
  AdminTableBody,
  AdminTableHead,
  AdminTd,
  AdminTextButton,
  AdminTh,
  AdminTr,
  useAdminConfirm,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { invalidateAdminQueryPrefix, useAdminQuery } from "@/lib/adminQuery";
import { adminQueryKeys } from "@/lib/adminQueryKeys";

const ALL_PERMISSIONS = [
  { key: "dashboard", label: "仪表盘" },
  { key: "content", label: "内容管理" },
  { key: "pages", label: "页面管理" },
  { key: "articles", label: "文章管理" },
  { key: "media", label: "媒体管理" },
  { key: "form-submissions", label: "表单提交" },
  { key: "menus", label: "菜单管理" },
  { key: "theme", label: "主题" },
  { key: "analytics", label: "访问统计" },
  { key: "audit-logs", label: "审计日志" },
  { key: "backups", label: "数据备份" },
  { key: "users", label: "用户管理" },
];

interface UserFormData {
  username: string;
  password: string;
  role: string;
  permissions: string[];
}

const emptyForm: UserFormData = {
  username: "",
  password: "",
  role: "editor",
  permissions: [],
};

export default function AdminUsersPage() {
  useDocumentTitle("用户管理");
  const [page, setPage] = useState(1);

  // Dialog state
  const [showDialog, setShowDialog] = useState(false);
  const [editingUser, setEditingUser] = useState<UserDTO | null>(null);
  const [form, setForm] = useState<UserFormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");

  const { confirm, confirmDialog } = useAdminConfirm();
  const pageSize = 20;

  const { data, error, loading, refetch } = useAdminQuery(
    [...adminQueryKeys.users, page, pageSize],
    () => listUsers(page, pageSize),
  );
  const users = data?.items ?? [];
  const total = data?.total ?? 0;

  const fetchUsers = async () => {
    invalidateAdminQueryPrefix(adminQueryKeys.users);
    await refetch({ force: true });
  };

  const openCreate = () => {
    setEditingUser(null);
    setForm(emptyForm);
    setFormError("");
    setShowDialog(true);
  };

  const openEdit = (user: UserDTO) => {
    setEditingUser(user);
    setForm({
      username: user.username,
      password: "",
      role: user.role,
      permissions: [...user.permissions],
    });
    setFormError("");
    setShowDialog(true);
  };

  const handleSave = async () => {
    setFormError("");
    if (!form.username.trim()) {
      setFormError("请输入用户名");
      return;
    }
    if (!editingUser && form.password.length < 6) {
      setFormError("密码长度不能少于6位");
      return;
    }
    if (editingUser && form.password && form.password.length < 6) {
      setFormError("密码长度不能少于6位");
      return;
    }

    setSaving(true);
    try {
      if (editingUser) {
        const data: UpdateUserRequest = {
          username: form.username,
          role: form.role,
          permissions: form.permissions,
        };
        if (form.password) {
          data.password = form.password;
        }
        await updateUser(editingUser.id, data);
      } else {
        const data: CreateUserRequest = {
          username: form.username,
          password: form.password,
          role: form.role,
          permissions: form.permissions,
        };
        await createUser(data);
      }
      setShowDialog(false);
      fetchUsers();
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || "保存失败";
      setFormError(msg);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (user: UserDTO) => {
    const ok = await confirm({
      title: "删除用户",
      message: `确定删除用户「${user.username}」吗？此操作不可撤销。`,
      confirmLabel: "删除",
      danger: true,
    });
    if (!ok) return;
    try {
      await deleteUser(user.id);
      await fetchUsers();
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || "删除失败";
      alert(msg);
    }
  };

  const togglePermission = (key: string) => {
    setForm((prev) => ({
      ...prev,
      permissions: prev.permissions.includes(key)
        ? prev.permissions.filter((p) => p !== key)
        : [...prev.permissions, key],
    }));
  };

  const toggleAllPermissions = () => {
    const allKeys = ALL_PERMISSIONS.map((p) => p.key);
    const allSelected = allKeys.every((k) => form.permissions.includes(k));
    setForm((prev) => ({
      ...prev,
      permissions: allSelected ? [] : [...allKeys],
    }));
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-6">
      {confirmDialog}
      <AdminPageHeader
        title="用户管理"
        description="管理后台账号与基础权限"
        actions={
          <AdminButton size="sm" onClick={openCreate}>
            创建用户
          </AdminButton>
        }
      />

      {error && <AdminErrorBanner message={error.message || "加载用户列表失败"} />}

      {loading ? (
        <AdminLoading />
      ) : (
        <>
          <AdminTable>
            <AdminTableHead>
              <tr>
                <AdminTh>用户名</AdminTh>
                <AdminTh>角色</AdminTh>
                <AdminTh>超管</AdminTh>
                <AdminTh>权限数</AdminTh>
                <AdminTh>创建时间</AdminTh>
                <AdminTh className="text-right">操作</AdminTh>
              </tr>
            </AdminTableHead>
            <AdminTableBody>
              {users.length === 0 ? (
                <tr>
                  <AdminTd colSpan={6} className="py-8 text-center text-slate-500">
                    暂无用户
                  </AdminTd>
                </tr>
              ) : (
                users.map((u) => (
                  <AdminTr key={u.id}>
                    <AdminTd className="font-medium text-slate-900">{u.username}</AdminTd>
                    <AdminTd>{u.role === "admin" ? "管理员" : "编辑"}</AdminTd>
                    <AdminTd>
                      {u.isSuperAdmin ? (
                        <AdminBadge tone="warning">超级管理员</AdminBadge>
                      ) : (
                        <span className="text-slate-400">—</span>
                      )}
                    </AdminTd>
                    <AdminTd className="tabular-nums">
                      {u.isSuperAdmin ? "全部" : u.permissions.length}
                    </AdminTd>
                    <AdminTd className="whitespace-nowrap text-slate-500">
                      {new Date(u.createdAt).toLocaleDateString("zh-CN")}
                    </AdminTd>
                    <AdminTd className="space-x-3 text-right">
                      <AdminTextButton onClick={() => openEdit(u)}>编辑</AdminTextButton>
                      {!u.isSuperAdmin && (
                        <AdminTextButton tone="danger" onClick={() => handleDelete(u)}>
                          删除
                        </AdminTextButton>
                      )}
                    </AdminTd>
                  </AdminTr>
                ))
              )}
            </AdminTableBody>
          </AdminTable>

          <AdminPagination
            page={page}
            totalPages={totalPages}
            total={total}
            onPageChange={setPage}
          />
        </>
      )}

      <AdminModal
        open={showDialog}
        title={editingUser ? "编辑用户" : "创建用户"}
        onClose={() => setShowDialog(false)}
        footer={
          <>
            <AdminButton variant="secondary" size="sm" onClick={() => setShowDialog(false)}>
              取消
            </AdminButton>
            <AdminButton size="sm" onClick={handleSave} disabled={saving}>
              {saving ? "保存中…" : "保存"}
            </AdminButton>
          </>
        }
      >
        <div className="space-y-4">
          {formError ? <AdminErrorBanner message={formError} className="mb-0" /> : null}

          <AdminField label="用户名">
            <AdminInput
              type="text"
              value={form.username}
              onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
              placeholder="输入用户名"
              disabled={editingUser?.isSuperAdmin}
            />
          </AdminField>

          <AdminField label={editingUser ? "密码（留空不修改）" : "密码"}>
            <AdminInput
              type="password"
              value={form.password}
              onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              placeholder={editingUser ? "留空不修改" : "输入密码（至少6位）"}
            />
          </AdminField>

          <AdminField label="角色">
            <AdminSelect
              value={form.role}
              onChange={(e) => setForm((f) => ({ ...f, role: e.target.value }))}
              disabled={editingUser?.isSuperAdmin}
              className="w-full"
            >
              <option value="admin">管理员</option>
              <option value="editor">编辑</option>
            </AdminSelect>
          </AdminField>

          {!editingUser?.isSuperAdmin && (
            <div>
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-medium text-slate-700">权限</span>
                <AdminTextButton onClick={toggleAllPermissions} className="text-xs">
                  {ALL_PERMISSIONS.every((p) => form.permissions.includes(p.key))
                    ? "取消全选"
                    : "全选"}
                </AdminTextButton>
              </div>
              <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-3">
                {ALL_PERMISSIONS.map((perm) => (
                  <label
                    key={perm.key}
                    className="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
                  >
                    <AdminCheckbox
                      checked={form.permissions.includes(perm.key)}
                      onChange={() => togglePermission(perm.key)}
                    />
                    {perm.label}
                  </label>
                ))}
              </div>
            </div>
          )}

          {editingUser?.isSuperAdmin && (
            <div className="rounded-xl border border-amber-200/80 bg-amber-50 px-3 py-2.5 text-sm text-amber-800">
              超级管理员拥有全部权限，不可修改权限设置。
            </div>
          )}
        </div>
      </AdminModal>
    </div>
  );
}
