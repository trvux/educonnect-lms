import { apiClient } from "@/lib/api-client";
import type { UserProfile } from "@/lib/types";

// US1.4
export async function getMe() {
  const res = await apiClient.get<UserProfile>("/me");
  return res.data;
}

export async function updateMe(input: { full_name: string; phone: string; student_code: string }) {
  const res = await apiClient.patch<UserProfile>("/me", input);
  return res.data;
}

// US1.4 — không set Content-Type thủ công, để axios tự gắn boundary đúng
// (xem lib/api/materials.ts, cùng lý do).
export async function uploadAvatar(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const res = await apiClient.post<UserProfile>("/me/avatar", formData);
  return res.data;
}

// Backend lưu avatar_path tương đối (vd "avatars/4/anh.png"); ghép với gốc
// API để có URL ảnh thật, giống lib/api/materials.ts#materialDownloadUrl.
export function avatarUrl(avatarPath: string) {
  const base = (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api").replace(
    /\/api\/?$/,
    ""
  );
  return `${base}/uploads/${avatarPath}`;
}

// US1.5
export async function changePassword(input: { current_password: string; new_password: string }) {
  await apiClient.post("/auth/change-password", input);
}

// US1.8
export async function forgotUsername(phone: string) {
  const res = await apiClient.post<{ masked_email: string }>("/auth/forgot-username", { phone });
  return res.data;
}

// US1.6
export async function forgotPassword(email: string) {
  await apiClient.post("/auth/forgot-password", { email });
}

export async function resetPassword(input: { email: string; otp: string; new_password: string }) {
  await apiClient.post("/auth/reset-password", input);
}

// US1.9 — xác thực email lúc đăng ký (tách biệt khoá tài khoản US1.3: xem
// backend user.ErrEmailNotVerified vs user.ErrInactive, trả HTTP 428 vs 403).
export async function verifyEmail(input: { email: string; otp: string }) {
  await apiClient.post("/auth/verify-email", input);
}

export async function resendVerification(email: string) {
  await apiClient.post("/auth/resend-verification", { email });
}
