"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { gradeSubmission, listSubmissions } from "@/lib/api/submissions";
import type { AssignmentKind, Submission } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

export function SubmissionsGrading({
  assignmentId,
  kind,
}: {
  assignmentId: number;
  kind: AssignmentKind;
}) {
  const { data: submissions, isLoading } = useQuery({
    queryKey: ["submissions", assignmentId],
    queryFn: () => listSubmissions(assignmentId),
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Bài nộp của học viên</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading && <p className="text-sm text-muted-foreground">Đang tải...</p>}
        {!isLoading && submissions?.length === 0 && (
          <p className="text-sm text-muted-foreground">Chưa có học viên nào nộp bài.</p>
        )}
        <div className="flex flex-col gap-3">
          {submissions?.map((s) => (
            <SubmissionRow key={s.id} submission={s} kind={kind} assignmentId={assignmentId} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function SubmissionRow({
  submission,
  kind,
  assignmentId,
}: {
  submission: Submission;
  kind: AssignmentKind;
  assignmentId: number;
}) {
  const queryClient = useQueryClient();
  const [score, setScore] = useState(submission.score?.toString() ?? "");
  const [feedback, setFeedback] = useState(submission.feedback ?? "");

  const gradeMutation = useMutation({
    mutationFn: () => gradeSubmission(submission.id, Number(score), feedback),
    onSuccess: () => {
      toast.success("Đã chấm điểm");
      queryClient.invalidateQueries({ queryKey: ["submissions", assignmentId] });
    },
    onError: () => toast.error("Chấm điểm thất bại"),
  });

  return (
    <div className="rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm font-medium">Học viên #{submission.student_id}</p>
        <Badge variant={submission.graded ? "default" : "secondary"}>
          {submission.graded ? `Đã chấm: ${submission.score}/10` : "Chưa chấm"}
        </Badge>
      </div>

      {kind === "essay" && submission.content && (
        <p className="mt-2 text-sm text-muted-foreground">{submission.content}</p>
      )}
      {kind === "quiz" && (
        <p className="mt-2 text-sm text-muted-foreground">
          Đáp án: {submission.answers.join(", ")}
        </p>
      )}

      {kind === "essay" && (
        <div className="mt-3 flex flex-col gap-2 sm:flex-row sm:items-end">
          <div className="flex flex-col gap-1 sm:w-24">
            <Input
              type="number"
              min={0}
              max={10}
              step={0.5}
              value={score}
              onChange={(e) => setScore(e.target.value)}
              placeholder="Điểm"
            />
          </div>
          <Textarea
            className="sm:flex-1"
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            placeholder="Nhận xét (tùy chọn)"
            rows={1}
          />
          <Button
            size="sm"
            onClick={() => gradeMutation.mutate()}
            disabled={score === "" || gradeMutation.isPending}
          >
            {gradeMutation.isPending ? "Đang lưu..." : "Chấm điểm"}
          </Button>
        </div>
      )}
    </div>
  );
}
