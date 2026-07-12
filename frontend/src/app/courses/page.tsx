"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { searchCourses } from "@/lib/api/courses";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

// US3.1: Học viên tìm kiếm khóa học. Mobile-first: danh sách card 1 cột
// trên mobile, lưới nhiều cột từ breakpoint md trở lên.
export default function CoursesPage() {
  const [keyword, setKeyword] = useState("");

  const { data: courses, isLoading } = useQuery({
    queryKey: ["courses", keyword],
    queryFn: () => searchCourses(keyword),
  });

  return (
    <div className="mx-auto max-w-5xl px-4 py-6 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Khám phá khóa học</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Tìm khóa học đã được quản trị viên duyệt và công khai.
      </p>

      <Input
        className="mt-6"
        placeholder="Tìm theo tên khóa học..."
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
      />

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {isLoading &&
          Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-32 w-full" />)}

        {!isLoading && courses?.length === 0 && (
          <p className="col-span-full text-sm text-muted-foreground">
            Chưa có khóa học nào phù hợp.
          </p>
        )}

        {courses?.map((course) => (
          <Link key={course.id} href={`/courses/${course.id}`}>
            <Card className="h-full transition-colors hover:border-primary">
              <CardHeader>
                <div className="flex items-start justify-between gap-2">
                  <CardTitle className="text-base">{course.title}</CardTitle>
                  <Badge variant="secondary">Đã duyệt</Badge>
                </div>
              </CardHeader>
              <CardContent>
                <p className="line-clamp-3 text-sm text-muted-foreground">
                  {course.description || "Chưa có mô tả."}
                </p>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
