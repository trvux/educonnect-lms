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

// Backend lưu file_path tương đối (vd "lesson-1/slide.pdf"); ghép với gốc
// API để có URL tải file thật (route static file phục vụ ở backend sau).
export function materialDownloadUrl(filePath: string) {
  const base = (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api").replace(
    /\/api\/?$/,
    ""
  );
  return `${base}/uploads/${filePath}`;
}
