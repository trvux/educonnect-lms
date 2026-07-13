"use client";

import { use } from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { apiClient } from "@/lib/api-client";
import { useSession } from "@/lib/auth";
import type { Course } from "@/lib/types";
import { submitCourseForReview, approveCourse } from "@/lib/api/courses";
import { enrollInCourse, listEnrolledStudents } from "@/lib/api/enrollment";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChaptersSection } from "./chapters-section";
import { ForumSection } from "./forum-section";
import { NotificationSendDialog } from "./notification-send-dialog";

async function getCourse(id: number) {
  const res = await apiClient.get<Course>(`/courses/${id}`);
  return res.data;
}

const statusLabel: Record<Course["status"], string> = {
  draft: "Nháp",
  pending_review: "Chờ duyệt",
  approved: "Đã duyệt",
};

export default function CourseDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const courseId = Number(id);
  const session = useSession();
  const queryClient = useQueryClient();

  const { data: course, isLoading } = useQuery({
    queryKey: ["course", courseId],
    queryFn: () => getCourse(courseId),
  });

  const isOwner = Boolean(session && course && session.userId === course.teacher_id);
  const canManage = isOwner || session?.role === "admin";

  const enrollMutation = useMutation({
    mutationFn: () => enrollInCourse(courseId),
    onSuccess: () => toast.success("Đăng ký khóa học thành công"),
    onError: (error: unknown) => {
      const status = (error as { response?: { status?: number } })?.response?.status;
      if (status === 409) {
        toast.info("Bạn đã đăng ký khóa học này rồi");
      } else {
        toast.error("Đăng ký thất bại, vui lòng thử lại");
      }
    },
  });

  const submitMutation = useMutation({
    mutationFn: () => submitCourseForReview(courseId),
    onSuccess: () => {
      toast.success("Đã gửi khóa học để quản trị viên duyệt");
      queryClient.invalidateQueries({ queryKey: ["course", courseId] });
    },
    onError: () => toast.error("Gửi duyệt thất bại"),
  });

  const approveMutation = useMutation({
    mutationFn: () => approveCourse(courseId),
    onSuccess: () => {
      toast.success("Đã duyệt khóa học, học viên có thể tìm thấy ngay bây giờ");
      queryClient.invalidateQueries({ queryKey: ["course", courseId] });
    },
    onError: () => toast.error("Duyệt khóa học thất bại"),
  });

  const { data: students } = useQuery({
    queryKey: ["students", courseId],
    queryFn: () => listEnrolledStudents(courseId),
    enabled: isOwner,
  });

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6 sm:py-10">
        <Skeleton className="h-8 w-2/3" />
        <Skeleton className="mt-4 h-24 w-full" />
      </div>
    );
  }

  if (!course) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-10 text-center text-sm text-muted-foreground">
        Không tìm thấy khóa học.
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 sm:py-10">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{course.title}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{course.description}</p>
        </div>
        <Badge variant={course.status === "approved" ? "default" : "secondary"}>
          {statusLabel[course.status]}
        </Badge>
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {session && session.role === "student" && (
          <Button onClick={() => enrollMutation.mutate()} disabled={enrollMutation.isPending}>
            {enrollMutation.isPending ? "Đang đăng ký..." : "Đăng ký khóa học"}
          </Button>
        )}
        {isOwner && course.status === "draft" && (
          <Button variant="outline" onClick={() => submitMutation.mutate()} disabled={submitMutation.isPending}>
            Gửi duyệt
          </Button>
        )}
        {session?.role === "admin" && course.status === "pending_review" && (
          <Button onClick={() => approveMutation.mutate()} disabled={approveMutation.isPending}>
            Duyệt khóa học
          </Button>
        )}
        {canManage && (
          <>
            <Button variant="outline" size="sm" nativeButton={false} render={<Link href={`/courses/${courseId}/gradebook`} />}>
              Bảng điểm
            </Button>
            <NotificationSendDialog courseId={courseId} />
          </>
        )}
      </div>

      <Separator className="my-6" />

      <ChaptersSection courseId={courseId} canManage={canManage} />

      {isOwner && (
        <>
          <Separator className="my-6" />
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Danh sách học viên (US3.3)</CardTitle>
            </CardHeader>
            <CardContent>
              {!students || students.length === 0 ? (
                <p className="text-sm text-muted-foreground">Chưa có học viên nào đăng ký.</p>
              ) : (
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Họ tên</TableHead>
                        <TableHead>Email</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {students.map((s) => (
                        <TableRow key={s.StudentID}>
                          <TableCell>{s.FullName}</TableCell>
                          <TableCell>{s.Email}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}

      <Separator className="my-6" />
      <ForumSection courseId={courseId} />
    </div>
  );
}
