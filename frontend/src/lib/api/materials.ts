import { apiClient } from "@/lib/api-client";
import type { Material } from "@/lib/types";

// US4.2
export async function listMaterials(lessonId: number) {
  const res = await apiClient.get<Material[]>(`/lessons/${lessonId}/materials`);
  return res.data;
}

// US4.1 — không set Content-Type thủ công, để axios tự gắn boundary đúng.
export async function uploadMaterial(lessonId: number, file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const res = await apiClient.post<Material>(`/lessons/${lessonId}/materials`, formData);
  return res.data;
}

// US4.3 — tài liệu không còn phục vụ qua link tĩnh công khai (lỗ hổng bảo
// mật cũ: ai có link cũng tải được, không cần đăng nhập/đăng ký khóa học).
// Endpoint /materials/:id/download giờ yêu cầu Authorization header, nên
// không thể dùng thẻ <a href=...> trực tiếp (không tự gắn header, và
// thuộc tính `download` bị trình duyệt bỏ qua với link cross-origin) — phải
// fetch qua axios (đã tự gắn Bearer token) rồi tự tạo Blob URL để tải.
export async function downloadMaterial(materialId: number, fileName: string) {
  const res = await apiClient.get(`/materials/${materialId}/download`, { responseType: "blob" });
  const blobUrl = URL.createObjectURL(res.data as Blob);
  const a = document.createElement("a");
  a.href = blobUrl;
  a.download = fileName;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(blobUrl);
}

// US4.8 — chỉ giảng viên sở hữu khóa học/admin xóa được (kiểm tra thật ở
// backend); học viên không bao giờ thấy nút xóa (canManage ở FE).
export async function deleteMaterial(materialId: number) {
  await apiClient.delete(`/materials/${materialId}`);
}

// US4.5 — thẻ <video src="..."> không gửi được header Authorization nên
// không thể trỏ thẳng vào /materials/:id/download; backend mint sẵn 1
// stream_token ngắn hạn (30 phút) kèm theo material loại video trong
// listMaterials(), ghép vào query string để endpoint /stream xác thực.
export function materialStreamUrl(materialId: number, streamToken: string) {
  const base = apiClient.defaults.baseURL ?? "";
  return `${base}/materials/${materialId}/stream?token=${encodeURIComponent(streamToken)}`;
}
