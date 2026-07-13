"use client";

import { use } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CheckCircleIcon, LockIcon } from "@phosphor-icons/react";

import { useSession } from "@/lib/auth";
import { getCourse } from "@/lib/api/courses";
import { getCourseOutline, markLessonComplete } from "@/lib/api/curriculum";
import { MaterialsList } from "../../chapters-section";
import { AssignmentsSection } from "../../assignments-section";
import { Button } from "@/components/ui/button";
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
  const queryClient = useQueryClient();

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
  const isStudent = Boolean(session && session.role === "student");

  // US4.10 — đánh dấu hoàn thành xong thì tải lại outline để sidebar cập
  // nhật ngay khóa/mở khóa bài tiếp theo, không cần reload trang.
  const completeMutation = useMutation({
    mutationFn: () => markLessonComplete(lessonIdNum),
    onSuccess: () => {
      toast.success("Đã đánh dấu hoàn thành bài học");
      queryClient.invalidateQueries({ queryKey: ["course-outline", courseId] });
    },
    onError: () => toast.error("Đánh dấu hoàn thành thất bại, vui lòng thử lại"),
  });

  if (isLoading) {
    return <Skeleton className="h-40 w-full" />;
  }
  if (!lesson) {
    return <p className="text-sm text-muted-foreground">Không tìm thấy bài học.</p>;
  }
  if (lesson.locked) {
    return (
      <div className="flex flex-col items-center gap-2 py-10 text-center text-sm text-muted-foreground">
        <LockIcon className="size-8" />
        <p>Bài học này đang bị khóa — hoàn thành các bài học trước đó để mở khóa.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-xl font-semibold tracking-tight">{lesson.title}</h1>
        {isStudent &&
          (lesson.completed ? (
            <span className="flex items-center gap-1 text-sm text-primary">
              <CheckCircleIcon className="size-4" weight="fill" />
              Đã hoàn thành
            </span>
          ) : (
            <Button size="sm" onClick={() => completeMutation.mutate()} disabled={completeMutation.isPending}>
              Đánh dấu hoàn thành
            </Button>
          ))}
      </div>
      <MaterialsList lessonId={lessonIdNum} canManage={canManage} />
      <AssignmentsSection lessonId={lessonIdNum} canManage={canManage} />
    </div>
  );
}
