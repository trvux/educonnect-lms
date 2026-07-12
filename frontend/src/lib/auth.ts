"use client";

import { useEffect, useState } from "react";
import { decodeJwt, type Role } from "./jwt";

const TOKEN_KEY = "token";

export function saveToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token);
  window.dispatchEvent(new Event("auth-changed"));
}

export function clearToken() {
  window.localStorage.removeItem(TOKEN_KEY);
  window.dispatchEvent(new Event("auth-changed"));
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export type Session = {
  userId: number;
  role: Role;
} | null;

function readSession(): Session {
  const token = getToken();
  if (!token) return null;
  const payload = decodeJwt(token);
  if (!payload) return null;
  // exp là Unix timestamp (giây); token hết hạn thì coi như chưa đăng nhập.
  if (payload.exp * 1000 < Date.now()) return null;
  return { userId: payload.uid, role: payload.role };
}

// Hook đọc trạng thái đăng nhập hiện tại (US1.2/US1.3), tự cập nhật khi
// login/logout xảy ra ở tab hiện tại (event "auth-changed") hoặc tab khác
// (event "storage").
export function useSession(): Session {
  const [session, setSession] = useState<Session>(null);

  useEffect(() => {
    setSession(readSession());
    const onChange = () => setSession(readSession());
    window.addEventListener("auth-changed", onChange);
    window.addEventListener("storage", onChange);
    return () => {
      window.removeEventListener("auth-changed", onChange);
      window.removeEventListener("storage", onChange);
    };
  }, []);

  return session;
}
