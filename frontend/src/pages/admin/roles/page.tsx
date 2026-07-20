import { useState } from "react";
import {
  listRoles,
  createRole,
  updateRole,
  deleteRole,
  listPermissions,
  type RoleDTO,
  type PermissionDTO,
  type CreateRoleRequest,
  type UpdateRoleRequest,
} from "@/api/roles";
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
  AdminTable,
  AdminTableBody,
  AdminTableHead,
  AdminTd,
  AdminTextarea,
  AdminTextButton,
  AdminTh,
  AdminTr,
  useAdminConfirm,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { invalidateAdminQueryPrefix, useAdminQuery } from "@/lib/adminQuery";
import { adminQueryKeys } from "@/lib/adminQueryKeys";

interface RoleFormData {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
}

const emptyForm: RoleFormData = {
  name: "",
  display_name: "",
  description: "",
  permissions: [],
};

export default function AdminRolesPage() {
  useDocumentTitle("角色管理");

  const [showDialog, setShowDialog] = useState(false);
  const [editingRole, setEditingRole] = useState<RoleDTO | null>(null);
  const [form, setForm] = useState<RoleFormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");
  const [deleting, setDeleting] = useState(false);

  const { confirm, confirmDialog } = useAdminConfirm();

  const { data, error, loading, refetch } = useAdminQuery(
    adminQueryKeys.roles,
    async () => {
      const [rolesData, permsData] = await Promise.all([listRoles(), listPermissions()]);
      return {
        roles: rolesData.items as RoleDTO[],
        permissions: permsData.items as PermissionDTO[],
      };
    },
  );
  const roles = data?.roles ?? [];
  const permissions = data?.permissions ?? [];

  const fetchData = async () => {
    invalidateAdminQueryPrefix(adminQueryKeys.roles);
    await refetch({ force: true });
  };

  const openCreate = () => {
    setEditingRole(null);
    setForm(emptyForm);
    setFormError("");
    setShowDialog(true);
  };

  const openEdit = (role: RoleDTO) => {
    setEditingRole(role);
    setForm({
      name: role.name,
      display_name: role.display_name,
      description: role.description,
      permissions: [...role.permissions],
    });
    setFormError("");
    setShowDialog(true);
  };

  const handleSave = async () => {
    setFormError("");
    if (!form.name.trim()) {
      setFormError("请输入角色标识");
      return;
    }
    if (!form.display_name.trim()) {
      setFormError("请输入角色名称");
      return;
    }
    setSaving(true);
    try {
      if (editingRole) {
        const data: UpdateRoleRequest = {
          name: form.name,
          display_name: form.display_name,
          description: form.description,
          permissions: form.permissions,
        };
        await updateRole(editingRole.id, data);
      } else {
        const data: CreateRoleRequest = {
          name: form.name,
          display_name: form.display_name,
          description: form.description,
          permissions: form.permissions,
        };
        await createRole(data);
      }
      setShowDialog(false);
      fetchData();
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || "保存失败";
      setFormError(msg);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (role: RoleDTO) => {
    const ok = await confirm({
      title: "删除角色",
      message: `确定要删除角色「${role.display_name}」吗？此操作不可撤销。`,
      confirmLabel: "删除",
      danger: true,
    });
    if (!ok) return;
    setDeleting(true);
    try {
      await deleteRole(role.id);
      await fetchData();
    } catch (err: any) {
      const msg = err?.response?.data?.error?.message || "删除失败";
      alert(msg);
    } finally {
      setDeleting(false);
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

  // Group permissions by resource
  const groupedPermissions = permissions.reduce<Record<string, PermissionDTO[]>>(
    (acc, perm) => {
      if (!acc[perm.resource]) acc[perm.resource] = [];
      acc[perm.resource].push(perm);
      return acc;
    },
    {}
  );

  // Toggle all permissions in a resource group
  const toggleResourceGroup = (resource: string) => {
    const group = groupedPermissions[resource] || [];
    const keys = group.map((p) => `${p.resource}:${p.action}`);
    const allSelected = keys.every((k) => form.permissions.includes(k));
    setForm((prev) => ({
      ...prev,
      permissions: allSelected
        ? prev.permissions.filter((p) => !keys.includes(p))
        : [...new Set([...prev.permissions, ...keys])],
    }));
  };

  return (
    <div className="space-y-6">
      {confirmDialog}
      <AdminPageHeader
        title="角色管理"
        description="配置角色与权限集合"
        actions={
          <AdminButton size="sm" onClick={openCreate}>
            创建角色
          </AdminButton>
        }
      />

      {error && <AdminErrorBanner message={error.message || "加载数据失败"} />}

      {loading ? (
        <AdminLoading />
      ) : (
        <AdminTable>
          <AdminTableHead>
            <tr>
              <AdminTh>角色名称</AdminTh>
              <AdminTh>标识</AdminTh>
              <AdminTh>描述</AdminTh>
              <AdminTh>权限数</AdminTh>
              <AdminTh>类型</AdminTh>
              <AdminTh>创建时间</AdminTh>
              <AdminTh className="text-right">操作</AdminTh>
            </tr>
          </AdminTableHead>
          <AdminTableBody>
            {roles.length === 0 ? (
              <tr>
                <AdminTd colSpan={7} className="py-8 text-center text-slate-500">
                  暂无角色
                </AdminTd>
              </tr>
            ) : (
              roles.map((role) => (
                <AdminTr key={role.id}>
                  <AdminTd className="font-medium text-slate-900">
                    <div className="flex items-center gap-2">
                      {role.is_system && (
                        <svg className="h-4 w-4 shrink-0 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                        </svg>
                      )}
                      {role.display_name}
                    </div>
                  </AdminTd>
                  <AdminTd className="font-mono text-slate-600">{role.name}</AdminTd>
                  <AdminTd className="max-w-xs truncate text-slate-500">{role.description || "-"}</AdminTd>
                  <AdminTd className="text-slate-600">{role.permissions.length}</AdminTd>
                  <AdminTd>
                    {role.is_system ? (
                      <AdminBadge tone="warning">系统角色</AdminBadge>
                    ) : (
                      <AdminBadge tone="info">自定义</AdminBadge>
                    )}
                  </AdminTd>
                  <AdminTd className="text-slate-500">
                    {new Date(role.created_at).toLocaleDateString("zh-CN")}
                  </AdminTd>
                  <AdminTd className="space-x-3 text-right">
                    <AdminTextButton onClick={() => openEdit(role)}>编辑</AdminTextButton>
                    {!role.is_system && (
                      <AdminTextButton tone="danger" onClick={() => handleDelete(role)} disabled={deleting}>
                        删除
                      </AdminTextButton>
                    )}
                  </AdminTd>
                </AdminTr>
              ))
            )}
          </AdminTableBody>
        </AdminTable>
      )}

      <AdminModal
        open={showDialog}
        title={editingRole ? "编辑角色" : "创建角色"}
        onClose={() => setShowDialog(false)}
        widthClass="max-w-2xl"
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

          <AdminField label="角色标识 *">
            <AdminInput
              type="text"
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              className="font-mono"
              placeholder="如: editor, reviewer"
              disabled={editingRole?.is_system}
            />
          </AdminField>

          <AdminField label="角色名称 *">
            <AdminInput
              type="text"
              value={form.display_name}
              onChange={(e) => setForm((f) => ({ ...f, display_name: e.target.value }))}
              placeholder="如: 编辑员、审核员"
            />
          </AdminField>

          <AdminField label="描述">
            <AdminTextarea
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              placeholder="角色的功能说明"
              rows={2}
            />
          </AdminField>

          <div>
            <div className="mb-2 text-sm font-medium text-slate-700">权限配置</div>
            {Object.keys(groupedPermissions).length === 0 ? (
              <p className="text-sm text-slate-500">暂无可配置权限</p>
            ) : (
              <div className="space-y-3 rounded-2xl border border-slate-200 p-3">
                {Object.entries(groupedPermissions).map(([resource, perms]) => {
                  const keys = perms.map((p) => `${p.resource}:${p.action}`);
                  const allSelected = keys.every((k) => form.permissions.includes(k));
                  return (
                    <div key={resource}>
                      <div className="mb-1 flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => toggleResourceGroup(resource)}
                          className={`rounded-full px-2 py-0.5 text-xs font-semibold transition ${
                            allSelected
                              ? "bg-blue-50 text-blue-700 ring-1 ring-blue-600/15"
                              : "bg-slate-100 text-slate-600 hover:bg-slate-200"
                          }`}
                        >
                          {resource}
                        </button>
                      </div>
                      <div className="grid grid-cols-2 gap-1.5 pl-1 sm:grid-cols-3">
                        {perms.map((perm) => {
                          const key = `${perm.resource}:${perm.action}`;
                          return (
                            <label
                              key={perm.id}
                              className="flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1 text-sm text-slate-700 hover:bg-slate-50"
                            >
                              <AdminCheckbox
                                checked={form.permissions.includes(key)}
                                onChange={() => togglePermission(key)}
                              />
                              <span>{perm.action}</span>
                              {perm.description ? (
                                <span className="truncate text-xs text-slate-400">{perm.description}</span>
                              ) : null}
                            </label>
                          );
                        })}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {editingRole?.is_system && (
            <div className="rounded-xl border border-amber-200/80 bg-amber-50 px-3 py-2.5 text-sm text-amber-800">
              系统角色的标识不可修改，但可以调整权限配置。
            </div>
          )}
        </div>
      </AdminModal>
    </div>
  );
}
