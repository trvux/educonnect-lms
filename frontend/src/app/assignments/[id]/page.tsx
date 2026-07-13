"use client";

import { use, useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { TimerIcon } from "@phosphor-icons/react";

import { getAssignment, startQuizAttempt } from "@/lib/api/assignments";
import { getMySubmission, submitAssignment } from "@/lib/api/submissions";
import { useSession } from "@/lib/auth";
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
  const queryClient = useQueryClient();
  const isStudent = session?.role === "student";

  const [content, setContent] = useState("");
  const [answers, setAnswers] = useState<Record<number, number>>({});

  const { data: assignment, isLoading } = useQuery({
    queryKey: ["assignment", assignmentId],
    queryFn: () => getAssignment(assignmentId),
  });

  // US5.2 — biết ngay trạng thái đã nộp/điểm khi vào trang, không cần đợi
  // bấm Nộp bài mới biết qua lỗi 409.
  const { data: mySubmission, isLoading: isMySubmissionLoading } = useQuery({
    queryKey: ["my-submission", assignmentId],
    queryFn: () => getMySubmission(assignmentId),
    enabled: isStudent,
  });

  // US5.4 — chỉ ghi nhận "bắt đầu làm bài" khi thật sự cần đếm giờ (quiz có
  // time_limit_minutes, học viên chưa nộp). Idempotent: refresh trang không
  // reset đồng hồ vì backend trả lại đúng started_at gốc.
  const hasTimeLimit = assignment?.kind === "quiz" && !!assignment.time_limit_minutes;
  const { data: quizAttempt } = useQuery({
    queryKey: ["quiz-attempt", assignmentId],
    queryFn: () => startQuizAttempt(assignmentId),
    enabled: isStudent && hasTimeLimit && !isMySubmissionLoading && !mySubmission,
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
      queryClient.setQueryData(["my-submission", assignmentId], data);
    },
    onError: (error: unknown) => {
      const status = (error as { response?: { status?: number } })?.response?.status;
      if (status === 409) {
        toast.info("Bạn đã nộp bài này rồi");
        queryClient.invalidateQueries({ queryKey: ["my-submission", assignmentId] });
      } else if (status === 400) {
        toast.error("Đã quá hạn nộp hoặc dữ liệu không hợp lệ");
      } else {
        toast.error("Nộp bài thất bại, vui lòng thử lại");
      }
    },
  });

  // US5.4 — đồng hồ đếm ngược tính từ started_at (server) + time_limit_minutes;
  // khi hết giờ, tự động gọi Nộp bài với đáp án hiện có (submitMutation.mutate
  // luôn dùng answers/content mới nhất tại thời điểm gọi, không cần ref).
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);
  const deadline =
    quizAttempt && assignment?.time_limit_minutes
      ? new Date(quizAttempt.started_at).getTime() + assignment.time_limit_minutes * 60_000
      : null;

  useEffect(() => {
    if (!deadline || mySubmission) return;
    const interval = setInterval(() => {
      const secondsLeft = Math.max(0, Math.round((deadline - Date.now()) / 1000));
      setRemainingSeconds(secondsLeft);
      if (secondsLeft <= 0) {
        clearInterval(interval);
        toast.info("Đã hết giờ làm bài, đang tự động nộp bài...");
        submitMutation.mutate();
      }
    }, 1000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deadline, mySubmission]);

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

  const canManage = session?.role === "teacher" || session?.role === "admin";
  const quizComplete =
    assignment.kind !== "quiz" || assignment.questions.every((_, i) => answers[i] !== undefined);

  return (
    <div className="mx-auto max-w-2xl px-4 py-6 sm:py-10">
      <div className="flex flex-wrap items-center gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">{assignment.title}</h1>
        <Badge variant="secondary">{assignment.kind === "quiz" ? "Trắc nghiệm" : "Tự luận"}</Badge>
        {isStudent && !mySubmission && remainingSeconds !== null && (
          <Badge variant={remainingSeconds <= 60 ? "destructive" : "outline"} className="gap-1">
            <TimerIcon className="size-3.5" />
            Còn lại: {String(Math.floor(remainingSeconds / 60)).padStart(2, "0")}:
            {String(remainingSeconds % 60).padStart(2, "0")}
          </Badge>
        )}
      </div>
      {assignment.description && (
        <p className="mt-2 text-sm text-muted-foreground">{assignment.description}</p>
      )}
      {assignment.due_at && (
        <p className="mt-1 text-sm text-muted-foreground">
          Hạn nộp: {new Date(assignment.due_at).toLocaleString("vi-VN")}
        </p>
      )}

      {isStudent && isMySubmissionLoading && <Skeleton className="mt-6 h-32 w-full" />}

      {isStudent && !isMySubmissionLoading && !mySubmission && (
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

      {isStudent && !isMySubmissionLoading && mySubmission && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="text-base">Đã nộp bài</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            {mySubmission.graded ? (
              <>
                <p>
                  Điểm: <span className="font-semibold">{mySubmission.score}</span> / 10
                </p>
                {mySubmission.feedback && <p>Nhận xét: {mySubmission.feedback}</p>}
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
