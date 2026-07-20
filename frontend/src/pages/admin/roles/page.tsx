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
import { AdminButton, AdminErrorBanner, AdminPageHeader } from "@/components/admin/ui";
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

  const [deleteTarget, setDeleteTarget] = useState<RoleDTO | null>(null);
  const [deleting, setDeleting] = useState(false);

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

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await deleteRole(deleteTarget.id);
      setDeleteTarget(null);
      fetchData();
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

      {/* Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">角色名称</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">标识</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">描述</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">权限数</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {loading ? (
              <tr>
                <td colSpan={7} className="px-6 py-8 text-center text-gray-500">
                  加载中...
                </td>
              </tr>
            ) : roles.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-6 py-8 text-center text-gray-500">
                  暂无角色
                </td>
              </tr>
            ) : (
              roles.map((role) => (
                <tr key={role.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">
                    <div className="flex items-center gap-2">
                      {role.is_system && (
                        <svg className="w-4 h-4 text-amber-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                        </svg>
                      )}
                      {role.display_name}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 font-mono">{role.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-500 max-w-xs truncate">{role.description || "-"}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{role.permissions.length}</td>
                  <td className="px-6 py-4 text-sm">
                    {role.is_system ? (
                      <span className="inline-flex px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-700">
                        系统角色
                      </span>
                    ) : (
                      <span className="inline-flex px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-700">
                        自定义
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {new Date(role.created_at).toLocaleDateString("zh-CN")}
                  </td>
                  <td className="px-6 py-4 text-sm text-right space-x-2">
                    <button
                      onClick={() => openEdit(role)}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      编辑
                    </button>
                    {!role.is_system && (
                      <button
                        onClick={() => setDeleteTarget(role)}
                        className="text-red-600 hover:text-red-800"
                      >
                        删除
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Create/Edit Dialog */}
      {showDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={() => setShowDialog(false)} />
          <div className="relative bg-white rounded-xl shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6 space-y-4">
              <h2 className="text-lg font-bold text-gray-900">
                {editingRole ? "编辑角色" : "创建角色"}
              </h2>

              {formError && (
                <div className="p-3 bg-red-50 text-red-700 rounded-lg text-sm">{formError}</div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">角色标识 <span className="text-red-500">*</span></label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono"
                  placeholder="如: editor, reviewer"
                  disabled={editingRole?.is_system}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">角色名称 <span className="text-red-500">*</span></label>
                <input
                  type="text"
                  value={form.display_name}
                  onChange={(e) => setForm((f) => ({ ...f, display_name: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="如: 编辑员、审核员"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">描述</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="角色的功能说明"
                  rows={2}
                />
              </div>

              {/* Permissions grouped by resource */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">权限配置</label>
                {Object.keys(groupedPermissions).length === 0 ? (
                  <p className="text-sm text-gray-500">暂无可配置权限</p>
                ) : (
                  <div className="space-y-3 border border-gray-200 rounded-lg p-3">
                    {Object.entries(groupedPermissions).map(([resource, perms]) => {
                      const keys = perms.map((p) => `${p.resource}:${p.action}`);
                      const allSelected = keys.every((k) => form.permissions.includes(k));
                      return (
                        <div key={resource}>
                          <div className="flex items-center gap-2 mb-1">
                            <button
                              type="button"
                              onClick={() => toggleResourceGroup(resource)}
                              className={`text-xs font-semibold px-2 py-0.5 rounded ${allSelected ? "bg-blue-100 text-blue-700" : "bg-gray-100 text-gray-600"} hover:opacity-80`}
                            >
                              {resource}
                            </button>
                          </div>
                          <div className="grid grid-cols-2 sm:grid-cols-3 gap-1.5 pl-2">
                            {perms.map((perm) => {
                              const key = `${perm.resource}:${perm.action}`;
                              return (
                                <label
                                  key={perm.id}
                                  className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer hover:bg-gray-50 px-2 py-1 rounded"
                                >
                                  <input
                                    type="checkbox"
                                    checked={form.permissions.includes(key)}
                                    onChange={() => togglePermission(key)}
                                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                  />
                                  <span>{perm.action}</span>
                                  {perm.description && (
                                    <span className="text-gray-400 text-xs truncate">{perm.description}</span>
                                  )}
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
                <div className="p-3 bg-amber-50 text-amber-700 rounded-lg text-sm">
                  系统角色的标识不可修改，但可以调整权限配置。
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <button
                  onClick={() => setShowDialog(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50"
                >
                  取消
                </button>
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {saving ? "保存中..." : "保存"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={() => setDeleteTarget(null)} />
          <div className="relative bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6 space-y-4">
            <h2 className="text-lg font-bold text-gray-900">确认删除</h2>
            <p className="text-sm text-gray-600">
              确定要删除角色 <strong>{deleteTarget.display_name}</strong> 吗？此操作不可撤销。
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setDeleteTarget(null)}
                className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={handleDelete}
                disabled={deleting}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50"
              >
                {deleting ? "删除中..." : "删除"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
