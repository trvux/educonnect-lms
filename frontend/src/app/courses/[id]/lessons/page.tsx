"use client";

import { use, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { getCourseOutline } from "@/lib/api/curriculum";
import { Skeleton } from "@/components/ui/skeleton";

// US4.9 — vào /courses/[id]/lessons mà chưa chọn bài học cụ thể nào thì tự
// chuyển tới bài học đầu tiên của khóa học (giống bấm "Vào học" trên
// Coursera/Udemy luôn mở bài học đầu tiên).
export default function LessonsIndexPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const courseId = Number(id);
  const router = useRouter();

  const { data: outline, isLoading } = useQuery({
    queryKey: ["course-outline", courseId],
    queryFn: () => getCourseOutline(courseId),
  });

  const firstLesson = outline?.find((chapter) => chapter.lessons.length > 0)?.lessons[0];

  useEffect(() => {
    if (firstLesson) {
      router.replace(`/courses/${courseId}/lessons/${firstLesson.id}`);
    }
  }, [firstLesson, courseId, router]);

  if (isLoading || firstLesson) {
    return <Skeleton className="h-40 w-full" />;
  }

  return (
    <p className="text-sm text-muted-foreground">
      Khóa học chưa có bài học nào. Quay lại trang khóa học để thêm nội dung.
    </p>
  );
}
