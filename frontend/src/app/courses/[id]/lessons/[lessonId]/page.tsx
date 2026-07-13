"use client";

import { use } from "react";
import { useQuery } from "@tanstack/react-query";

import { useSession } from "@/lib/auth";
import { getCourse } from "@/lib/api/courses";
import { getCourseOutline } from "@/lib/api/curriculum";
import { MaterialsList } from "../../chapters-section";
import { AssignmentsSection } from "../../assignments-section";
import { Skeleton } from "@/components/ui/skeleton";

// US4.9 — nội dung 1 bài học trong course player: tài liệu + bài tập, dùng
// lại đúng component MaterialsList/AssignmentsSection đã có ở trang quản lý
// chương/bài học (chapters-section.tsx) để tránh trùng lặp logic quyền hạn.
// Tiêu đề bài học lấy từ cache "course-outline" (layout.tsx đã fetch cùng
// query key) — không gọi API riêng vì backend chưa có GET /lessons/:id.
export default function LessonContentPage({
  params,
}: {
  params: Promise<{ id: string; lessonId: string }>;
}) {
  const { id, lessonId } = use(params);
  const courseId = Number(id);
  const lessonIdNum = Number(lessonId);
  const session = useSession();

  const { data: course } = useQuery({
    queryKey: ["course", courseId],
    queryFn: () => getCourse(courseId),
  });
  const { data: outline, isLoading } = useQuery({
    queryKey: ["course-outline", courseId],
    queryFn: () => getCourseOutline(courseId),
  });

  const lesson = outline?.flatMap((chapter) => chapter.lessons).find((l) => l.id === lessonIdNum);
  const canManage = Boolean(
    session && course && (session.userId === course.teacher_id || session.role === "admin")
  );

  if (isLoading) {
    return <Skeleton className="h-40 w-full" />;
  }
  if (!lesson) {
    return <p className="text-sm text-muted-foreground">Không tìm thấy bài học.</p>;
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold tracking-tight">{lesson.title}</h1>
      <MaterialsList lessonId={lessonIdNum} canManage={canManage} />
      <AssignmentsSection lessonId={lessonIdNum} canManage={canManage} />
    </div>
  );
}
