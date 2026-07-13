"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { getMyProgress } from "@/lib/api/reports";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";

// US7.1 — dashboard tiến độ học tập của học viên đang đăng nhập.
export default function DashboardPage() {
  const { data: progress, isLoading } = useQuery({
    queryKey: ["my-progress"],
    queryFn: getMyProgress,
  });

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Tiến độ học tập</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Tỉ lệ hoàn thành tính theo số bài tập/trắc nghiệm đã nộp trên tổng số bài tập của khóa học.
      </p>

      {isLoading && (
        <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-32 w-full" />
        </div>
      )}

      {!isLoading && progress?.length === 0 && (
        <p className="mt-6 text-sm text-muted-foreground">
          Bạn chưa đăng ký khóa học nào.{" "}
          <Link href="/courses" className="text-primary underline underline-offset-4">
            Tìm khóa học
          </Link>
        </p>
      )}

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
        {progress?.map((p) => (
          <Card key={p.course_id}>
            <CardHeader>
              <CardTitle className="text-base">
                <Link href={`/courses/${p.course_id}`} className="hover:underline">
                  {p.course_title}
                </Link>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between text-sm text-muted-foreground">
                <span>
                  Đã nộp {p.submitted}/{p.total_assignments} bài tập
                </span>
                <span className="font-medium text-foreground">
                  {Math.round(p.percent_complete)}%
                </span>
              </div>
              <Progress value={p.percent_complete} className="mt-2" />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
