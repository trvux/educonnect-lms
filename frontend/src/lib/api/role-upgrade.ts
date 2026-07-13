import { apiClient } from "@/lib/api-client";
import type { RoleUpgradeRequest } from "@/lib/types";

// US1.7
export async function createRoleUpgradeRequest(reason: string) {
  const res = await apiClient.post<RoleUpgradeRequest>("/me/role-upgrade-request", { reason });
  return res.data;
}

export async function listPendingRoleUpgrades() {
  const res = await apiClient.get<RoleUpgradeRequest[]>("/admin/role-upgrade-requests");
  return res.data;
}

export async function approveRoleUpgrade(id: number) {
  const res = await apiClient.post<RoleUpgradeRequest>(`/admin/role-upgrade-requests/${id}/approve`);
  return res.data;
}

export async function rejectRoleUpgrade(id: number) {
  const res = await apiClient.post<RoleUpgradeRequest>(`/admin/role-upgrade-requests/${id}/reject`);
  return res.data;
}
