"use client";

import { useQuery } from "@tanstack/react-query";
import { Bar, BarChart, CartesianGrid, XAxis } from "recharts";

import { getCourseReports } from "@/lib/api/reports";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ChartConfig, ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";

const chartConfig = {
  average_completion: {
    label: "Tỉ lệ hoàn thành (%)",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

// US7.2 — báo cáo thống kê học viên/khóa học cho giảng viên (chỉ khóa học
// mình sở hữu) và quản trị viên (toàn hệ thống).
export default function ReportsPage() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ["course-reports"],
    queryFn: getCourseReports,
  });

  return (
    <div className="mx-auto max-w-4xl px-4 py-6 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Báo cáo thống kê</h1>

      {isLoading && <Skeleton className="mt-6 h-64 w-full" />}

      {!isLoading && stats?.length === 0 && (
        <p className="mt-6 text-sm text-muted-foreground">Chưa có khóa học nào để thống kê.</p>
      )}

      {!isLoading && stats && stats.length > 0 && (
        <>
          <Card className="mt-6">
            <CardHeader>
              <CardTitle className="text-base">Tỉ lệ hoàn thành trung bình theo khóa học</CardTitle>
            </CardHeader>
            <CardContent>
              <ChartContainer config={chartConfig} className="max-h-64 w-full">
                <BarChart data={stats}>
                  <CartesianGrid vertical={false} />
                  <XAxis
                    dataKey="course_title"
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                  />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Bar dataKey="average_completion" fill="var(--color-average_completion)" radius={4} />
                </BarChart>
              </ChartContainer>
            </CardContent>
          </Card>

          <Card className="mt-4">
            <CardHeader>
              <CardTitle className="text-base">Chi tiết theo khóa học</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Khóa học</TableHead>
                      <TableHead>Học viên</TableHead>
                      <TableHead>Bài tập</TableHead>
                      <TableHead>Hoàn thành TB</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {stats.map((s) => (
                      <TableRow key={s.course_id}>
                        <TableCell className="font-medium">{s.course_title}</TableCell>
                        <TableCell>{s.enrolled_students}</TableCell>
                        <TableCell>{s.total_assignments}</TableCell>
                        <TableCell>{Math.round(s.average_completion)}%</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
