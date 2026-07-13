"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useSession, clearToken } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { NotificationBell } from "@/components/notification-bell";

const HIDDEN_ON = ["/login", "/register", "/forgot-password", "/forgot-username"];

export function Navbar() {
  const session = useSession();
  const router = useRouter();
  const pathname = usePathname();

  if (HIDDEN_ON.includes(pathname)) return null;

  function handleLogout() {
    clearToken();
    router.push("/login");
  }

  return (
    <header className="border-b bg-background">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between gap-4 px-4">
        <Link href="/courses" className="font-semibold">
          EduConnect LMS
        </Link>

        <nav className="flex items-center gap-2 sm:gap-4">
          <Button variant="ghost" size="sm" nativeButton={false} render={<Link href="/courses" />}>
            Khóa học
          </Button>

          {(session?.role === "teacher" || session?.role === "admin") && (
            <Button variant="ghost" size="sm" nativeButton={false} render={<Link href="/courses/new" />}>
              Tạo khóa học
            </Button>
          )}

          {session?.role === "admin" && (
            <Button variant="ghost" size="sm" nativeButton={false} render={<Link href="/admin/courses" />}>
              Duyệt khóa học
            </Button>
          )}

          {session?.role === "student" && (
            <Button variant="ghost" size="sm" nativeButton={false} render={<Link href="/dashboard" />}>
              Tiến độ
            </Button>
          )}

          {(session?.role === "teacher" || session?.role === "admin") && (
            <Button variant="ghost" size="sm" nativeButton={false} render={<Link href="/reports" />}>
              Báo cáo
            </Button>
          )}

          {session?.role === "admin" && (
            <Button
              variant="ghost"
              size="sm"
              nativeButton={false}
              render={<Link href="/admin/role-upgrade-requests" />}
            >
              Duyệt nâng cấp
            </Button>
          )}

          {session && <NotificationBell />}

          {session && (
            <Button variant="ghost" size="sm" nativeButton={false} render={<Link href="/profile" />}>
              Hồ sơ
            </Button>
          )}

          {session ? (
            <Button variant="outline" size="sm" onClick={handleLogout}>
              Đăng xuất
            </Button>
          ) : (
            <Button size="sm" nativeButton={false} render={<Link href="/login" />}>
              Đăng nhập
            </Button>
          )}
        </nav>
      </div>
    </header>
  );
}
