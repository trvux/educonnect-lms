export type Role = "student" | "teacher" | "admin";

export type JwtPayload = {
  uid: number;
  role: Role;
  exp: number;
  iat: number;
};

// Giải mã payload JWT phía client chỉ để đọc thông tin hiển thị (role, uid).
// KHÔNG dùng để xác thực chữ ký — việc đó backend Go đã làm ở middleware.
export function decodeJwt(token: string): JwtPayload | null {
  try {
    const payload = token.split(".")[1];
    const base64 = payload.replace(/-/g, "+").replace(/_/g, "/");
    return JSON.parse(atob(base64));
  } catch {
    return null;
  }
}
