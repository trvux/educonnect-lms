"use client";

import { use, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";

import { getGradebook } from "@/lib/api/gradebook";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

// US5.3 — bảng điểm tổng hợp: hàng = học viên, cột = bài tập.
export default function GradebookPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const courseId = Number(id);

  const { data: entries, isLoading } = useQuery({
    queryKey: ["gradebook", courseId],
    queryFn: () => getGradebook(courseId),
  });

  const { students, assignments, scoreByCell } = useMemo(() => {
    const studentMap = new Map<number, string>();
    const assignmentMap = new Map<number, string>();
    const scoreByCell = new Map<string, number | undefined>();

    for (const e of entries ?? []) {
      studentMap.set(e.student_id, e.student_name);
      assignmentMap.set(e.assignment_id, e.assignment_title);
      scoreByCell.set(`${e.student_id}-${e.assignment_id}`, e.score);
    }

    return {
      students: [...studentMap.entries()],
      assignments: [...assignmentMap.entries()],
      scoreByCell,
    };
  }, [entries]);

  return (
    <div className="mx-auto max-w-4xl px-4 py-6 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Bảng điểm tổng hợp</h1>

      {isLoading && <Skeleton className="mt-4 h-40 w-full" />}

      {!isLoading && (
        <Card className="mt-4">
          <CardHeader>
            <CardTitle className="text-base">Điểm theo học viên / bài tập</CardTitle>
          </CardHeader>
          <CardContent>
            {students.length === 0 ? (
              <p className="text-sm text-muted-foreground">Chưa có học viên nào đăng ký khóa học.</p>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Học viên</TableHead>
                      {assignments.map(([id, title]) => (
                        <TableHead key={id}>{title}</TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {students.map(([studentId, name]) => (
                      <TableRow key={studentId}>
                        <TableCell className="font-medium">{name}</TableCell>
                        {assignments.map(([assignmentId]) => {
                          const score = scoreByCell.get(`${studentId}-${assignmentId}`);
                          return (
                            <TableCell key={assignmentId}>
                              {score !== undefined ? score : "—"}
                            </TableCell>
                          );
                        })}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
