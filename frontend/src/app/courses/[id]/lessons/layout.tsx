"use client";

import { use } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeftIcon } from "@phosphor-icons/react";

import { getCourse } from "@/lib/api/courses";
import { getCourseOutline } from "@/lib/api/curriculum";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";

// US4.9 — "course player": sidebar cố định liệt kê cây chương → bài học
// (giống VLU E-learning/Moodle/Udemy), khung chính bên phải hiện nội dung
// bài học đang chọn. Thay cho việc xổ hết nội dung trên 1 trang dài
// (ChaptersSection ở /courses/[id] vẫn giữ nguyên, dùng riêng cho giảng
// viên quản lý cấu trúc — xem ghi chú trong chapters-section.tsx).
export default function CoursePlayerLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const courseId = Number(id);
  const pathname = usePathname();

  const { data: course } = useQuery({
    queryKey: ["course", courseId],
    queryFn: () => getCourse(courseId),
  });

  const { data: outline, isLoading } = useQuery({
    queryKey: ["course-outline", courseId],
    queryFn: () => getCourseOutline(courseId),
  });

  return (
    // "transform": Sidebar dùng position: fixed (inset-y-0) vì mặc định tự
    // coi nó là layout ngoài cùng — nhưng app này đã có Navbar ở RootLayout
    // phía trên, nên fixed sẽ đè lên Navbar (định vị theo viewport, không
    // theo layout cha). Thêm "transform" (dù không đổi gì) tạo containing
    // block CSS, khiến Sidebar định vị theo div này thay vì viewport.
    <div className="relative transform">
      {/* min-h-0: bỏ min-h-svh mặc định của SidebarProvider vì đã có
          containing block riêng ở trên, không cần ép full viewport height. */}
      <SidebarProvider className="min-h-0 flex-1">
        <Sidebar collapsible="offcanvas">
          <SidebarHeader>
            <Link
              href={`/courses/${courseId}`}
              className="flex items-center gap-2 px-2 py-1 text-sm font-medium hover:underline"
            >
              <ArrowLeftIcon className="size-4" />
              {course?.title ?? "Quay lại khóa học"}
            </Link>
          </SidebarHeader>
          <SidebarContent>
            {isLoading && (
              <div className="flex flex-col gap-2 p-2">
                <Skeleton className="h-5 w-full" />
                <Skeleton className="h-5 w-full" />
                <Skeleton className="h-5 w-full" />
              </div>
            )}
            {!isLoading && outline?.length === 0 && (
              <p className="p-2 text-sm text-muted-foreground">Khóa học chưa có nội dung.</p>
            )}
            {outline?.map((chapter) => (
              <SidebarGroup key={chapter.id}>
                <SidebarGroupLabel>{chapter.title}</SidebarGroupLabel>
                <SidebarMenu>
                  {chapter.lessons.map((lesson) => {
                    const href = `/courses/${courseId}/lessons/${lesson.id}`;
                    return (
                      <SidebarMenuItem key={lesson.id}>
                        <SidebarMenuButton isActive={pathname === href} render={<Link href={href} />}>
                          {lesson.title}
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    );
                  })}
                  {chapter.lessons.length === 0 && (
                    <p className="px-2 py-1 text-xs text-muted-foreground">Chưa có bài học.</p>
                  )}
                </SidebarMenu>
              </SidebarGroup>
            ))}
          </SidebarContent>
        </Sidebar>
        <SidebarInset>
          <div className="flex items-center gap-2 border-b p-2">
            <SidebarTrigger />
            <span className="text-sm font-medium">Nội dung khóa học</span>
          </div>
          <div className="mx-auto w-full max-w-2xl px-4 py-6 sm:py-10">{children}</div>
        </SidebarInset>
      </SidebarProvider>
    </div>
  );
}
