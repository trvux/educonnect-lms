"use client";

import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { apiClient } from "@/lib/api-client";
import { approveCourse } from "@/lib/api/courses";
import { useSession } from "@/lib/auth";
import type { Course } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

async function listPendingCourses() {
  const res = await apiClient.get<Course[]>("/admin/courses/pending");
  return res.data;
}

// US2.3: Quản trị viên xem hàng chờ và duyệt khóa học.
export default function AdminCoursesPage() {
  const session = useSession();
  const queryClient = useQueryClient();

  const { data: courses, isLoading } = useQuery({
    queryKey: ["admin-courses-pending"],
    queryFn: listPendingCourses,
    enabled: session?.role === "admin",
  });

  const approveMutation = useMutation({
    mutationFn: (courseId: number) => approveCourse(courseId),
    onSuccess: () => {
      toast.success("Đã duyệt khóa học");
      queryClient.invalidateQueries({ queryKey: ["admin-courses-pending"] });
    },
    onError: () => toast.error("Duyệt khóa học thất bại"),
  });

  if (session && session.role !== "admin") {
    return (
      <div className="mx-auto max-w-md px-4 py-10 text-center text-sm text-muted-foreground">
        Chỉ Quản trị viên mới truy cập được trang này.
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Hàng chờ duyệt khóa học</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Danh sách khóa học đang ở trạng thái chờ duyệt (pending_review).
      </p>

      <div className="mt-6 flex flex-col gap-3">
        {isLoading && <Skeleton className="h-24 w-full" />}
        {!isLoading && courses?.length === 0 && (
          <p className="text-sm text-muted-foreground">Không có khóa học nào đang chờ duyệt.</p>
        )}
        {courses?.map((course) => (
          <Card key={course.id}>
            <CardHeader>
              <CardTitle className="text-base">
                <Link href={`/courses/${course.id}`} className="hover:underline">
                  {course.title}
                </Link>
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-muted-foreground">{course.description || "Chưa có mô tả."}</p>
              <Button
                size="sm"
                onClick={() => approveMutation.mutate(course.id)}
                disabled={approveMutation.isPending}
                className="w-fit"
              >
                Duyệt
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
