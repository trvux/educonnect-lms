"use client";

import { use, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { getAssignment } from "@/lib/api/assignments";
import { submitAssignment } from "@/lib/api/submissions";
import { useSession } from "@/lib/auth";
import type { Submission } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { SubmissionsGrading } from "./submissions-grading";

export default function AssignmentDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const assignmentId = Number(id);
  const session = useSession();

  const [content, setContent] = useState("");
  const [answers, setAnswers] = useState<Record<number, number>>({});
  const [submitted, setSubmitted] = useState<Submission | null>(null);
  const [alreadySubmitted, setAlreadySubmitted] = useState(false);

  const { data: assignment, isLoading } = useQuery({
    queryKey: ["assignment", assignmentId],
    queryFn: () => getAssignment(assignmentId),
  });

  const submitMutation = useMutation({
    mutationFn: () => {
      if (assignment?.kind === "quiz") {
        const orderedAnswers = assignment.questions.map((_, i) => answers[i] ?? -1);
        return submitAssignment(assignmentId, { answers: orderedAnswers });
      }
      return submitAssignment(assignmentId, { content });
    },
    onSuccess: (data) => {
      toast.success("Nộp bài thành công");
      setSubmitted(data);
    },
    onError: (error: unknown) => {
      const status = (error as { response?: { status?: number } })?.response?.status;
      if (status === 409) {
        toast.info("Bạn đã nộp bài này rồi");
        setAlreadySubmitted(true);
      } else if (status === 400) {
        toast.error("Đã quá hạn nộp hoặc dữ liệu không hợp lệ");
      } else {
        toast.error("Nộp bài thất bại, vui lòng thử lại");
      }
    },
  });

  if (isLoading) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-6 sm:py-10">
        <Skeleton className="h-8 w-2/3" />
        <Skeleton className="mt-4 h-24 w-full" />
      </div>
    );
  }

  if (!assignment) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-10 text-center text-sm text-muted-foreground">
        Không tìm thấy bài tập.
      </div>
    );
  }

  const isStudent = session?.role === "student";
  const canManage = session?.role === "teacher" || session?.role === "admin";
  const quizComplete =
    assignment.kind !== "quiz" || assignment.questions.every((_, i) => answers[i] !== undefined);

  return (
    <div className="mx-auto max-w-2xl px-4 py-6 sm:py-10">
      <div className="flex flex-wrap items-center gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">{assignment.title}</h1>
        <Badge variant="secondary">{assignment.kind === "quiz" ? "Trắc nghiệm" : "Tự luận"}</Badge>
      </div>
      {assignment.description && (
        <p className="mt-2 text-sm text-muted-foreground">{assignment.description}</p>
      )}
      {assignment.due_at && (
        <p className="mt-1 text-sm text-muted-foreground">
          Hạn nộp: {new Date(assignment.due_at).toLocaleString("vi-VN")}
        </p>
      )}

      {isStudent && !submitted && !alreadySubmitted && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="text-base">Làm bài</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {assignment.kind === "essay" ? (
              <div className="flex flex-col gap-2">
                <Label>Bài làm của bạn</Label>
                <Textarea
                  rows={6}
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="Nhập nội dung bài làm..."
                />
              </div>
            ) : (
              assignment.questions.map((q, qIndex) => (
                <div key={qIndex} className="flex flex-col gap-2">
                  <Label>
                    Câu {qIndex + 1}: {q.content}
                  </Label>
                  <RadioGroup
                    value={String(answers[qIndex] ?? -1)}
                    onValueChange={(v) =>
                      setAnswers((a) => ({ ...a, [qIndex]: Number(v) }))
                    }
                  >
                    {q.options.map((opt, oIndex) => (
                      <div key={oIndex} className="flex items-center gap-2">
                        <RadioGroupItem value={String(oIndex)} id={`ans-${qIndex}-${oIndex}`} />
                        <Label htmlFor={`ans-${qIndex}-${oIndex}`} className="font-normal">
                          {opt}
                        </Label>
                      </div>
                    ))}
                  </RadioGroup>
                </div>
              ))
            )}

            <Button
              className="w-fit"
              onClick={() => submitMutation.mutate()}
              disabled={
                submitMutation.isPending ||
                (assignment.kind === "essay" ? content.trim().length === 0 : !quizComplete)
              }
            >
              {submitMutation.isPending ? "Đang nộp..." : "Nộp bài"}
            </Button>
          </CardContent>
        </Card>
      )}

      {isStudent && alreadySubmitted && !submitted && (
        <Card className="mt-6">
          <CardContent className="pt-6 text-sm text-muted-foreground">
            Bạn đã nộp bài tập này rồi.
          </CardContent>
        </Card>
      )}

      {isStudent && submitted && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="text-base">Đã nộp bài</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            {submitted.graded ? (
              <>
                <p>
                  Điểm: <span className="font-semibold">{submitted.score}</span> / 10
                </p>
                {submitted.feedback && <p>Nhận xét: {submitted.feedback}</p>}
              </>
            ) : (
              <p className="text-muted-foreground">Bài làm đang chờ giảng viên chấm điểm.</p>
            )}
          </CardContent>
        </Card>
      )}

      {canManage && (
        <div className="mt-6">
          <SubmissionsGrading assignmentId={assignmentId} kind={assignment.kind} />
        </div>
      )}
    </div>
  );
}
