import { http } from "./http";

export interface PermissionDTO {
  id: number;
  resource: string;
  action: string;
  description: string;
}

export interface RoleDTO {
  id: number;
  name: string;
  display_name: string;
  description: string;
  is_system: boolean;
  permissions: string[];
  created_at: string;
}

export interface RoleListResponse {
  items: RoleDTO[];
}

export interface PermissionListResponse {
  items: PermissionDTO[];
}

export interface CreateRoleRequest {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
}

export interface UpdateRoleRequest {
  name?: string;
  display_name?: string;
  description?: string;
  permissions?: string[];
}

export async function listRoles() {
  const res = await http.get<RoleListResponse>("/admin/roles", {

  });
  return res.data;
}

export async function createRole(data: CreateRoleRequest) {
  const res = await http.post<RoleDTO>("/admin/roles", data, {

  });
  return res.data;
}

export async function updateRole(id: number, data: UpdateRoleRequest) {
  const res = await http.put<RoleDTO>(`/admin/roles/${id}`, data, {

  });
  return res.data;
}

export async function deleteRole(id: number) {
  await http.delete(`/admin/roles/${id}`, {

  });
}

export async function listPermissions() {
  const res = await http.get<PermissionListResponse>("/admin/permissions", {

  });
  return res.data;
}

export async function assignRole(user_id: number, role_id: number) {
  const res = await http.post("/admin/roles/assign", { user_id, role_id }, {

  });
  return res.data;
}

export async function unassignRole(user_id: number, role_id: number) {
  const res = await http.post("/admin/roles/unassign", { user_id, role_id }, {

  });
  return res.data;
}
