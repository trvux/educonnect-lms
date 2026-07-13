"use client";

import { useState } from "react";
import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { PlusIcon, TrashIcon } from "@phosphor-icons/react";

import { createAssignment, listAssignments } from "@/lib/api/assignments";
import type { AssignmentKind } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

const kindLabel: Record<AssignmentKind, string> = {
  essay: "Tự luận",
  quiz: "Trắc nghiệm",
};

type QuestionDraft = { content: string; options: string[]; correct_index: number };

function emptyQuestion(): QuestionDraft {
  return { content: "", options: ["", ""], correct_index: 0 };
}

export function AssignmentsSection({
  lessonId,
  canManage,
}: {
  lessonId: number;
  canManage: boolean;
}) {
  const { data: assignments, isLoading } = useQuery({
    queryKey: ["assignments", lessonId],
    queryFn: () => listAssignments(lessonId),
  });

  return (
    <div className="mt-3 flex flex-col gap-2 border-t pt-3">
      <p className="text-sm font-medium">Bài tập / trắc nghiệm</p>

      {isLoading && <p className="text-sm text-muted-foreground">Đang tải...</p>}
      {!isLoading && assignments?.length === 0 && (
        <p className="text-sm text-muted-foreground">Chưa có bài tập nào.</p>
      )}

      <div className="flex flex-col gap-1">
        {assignments?.map((a) => (
          <Link
            key={a.id}
            href={`/assignments/${a.id}`}
            className="flex flex-wrap items-center gap-2 text-sm text-primary hover:underline"
          >
            {a.title}
            <Badge variant="secondary">{kindLabel[a.kind]}</Badge>
          </Link>
        ))}
      </div>

      {canManage && <CreateAssignmentDialog lessonId={lessonId} />}
    </div>
  );
}

function CreateAssignmentDialog({ lessonId }: { lessonId: number }) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [kind, setKind] = useState<AssignmentKind>("essay");
  const [dueAt, setDueAt] = useState("");
  const [timeLimitMinutes, setTimeLimitMinutes] = useState("");
  const [questions, setQuestions] = useState<QuestionDraft[]>([emptyQuestion()]);

  function resetForm() {
    setTitle("");
    setDescription("");
    setKind("essay");
    setDueAt("");
    setTimeLimitMinutes("");
    setQuestions([emptyQuestion()]);
  }

  const createMutation = useMutation({
    mutationFn: () =>
      createAssignment(lessonId, {
        title,
        description,
        kind,
        questions: kind === "quiz" ? questions : [],
        due_at: dueAt ? new Date(dueAt).toISOString() : undefined,
        time_limit_minutes:
          kind === "quiz" && timeLimitMinutes ? Number(timeLimitMinutes) : undefined,
      }),
    onSuccess: () => {
      toast.success("Đã tạo bài tập");
      resetForm();
      setOpen(false);
      queryClient.invalidateQueries({ queryKey: ["assignments", lessonId] });
    },
    onError: () => toast.error("Tạo bài tập thất bại"),
  });

  const canSubmit =
    title.trim().length > 0 &&
    (kind === "essay" ||
      questions.every((q) => q.content.trim() && q.options.every((o) => o.trim())));

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) resetForm();
      }}
    >
      <DialogTrigger render={<Button size="sm" variant="outline" className="w-fit" />}>
        <PlusIcon className="size-4" />
        Tạo bài tập
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Tạo bài tập / trắc nghiệm</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label>Tiêu đề</Label>
            <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Bài tập tuần 1" />
          </div>

          <div className="flex flex-col gap-2">
            <Label>Mô tả</Label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Yêu cầu, hướng dẫn nộp bài..."
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label>Loại bài tập</Label>
              <Select value={kind} onValueChange={(v) => setKind(v as AssignmentKind)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="essay">Tự luận</SelectItem>
                  <SelectItem value="quiz">Trắc nghiệm</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-2">
              <Label>Hạn nộp (tùy chọn)</Label>
              <Input type="datetime-local" value={dueAt} onChange={(e) => setDueAt(e.target.value)} />
            </div>
          </div>

          {kind === "quiz" && (
            <div className="flex flex-col gap-2">
              <Label>Thời gian làm bài (phút, tùy chọn)</Label>
              <Input
                type="number"
                min={1}
                value={timeLimitMinutes}
                onChange={(e) => setTimeLimitMinutes(e.target.value)}
                placeholder="Không giới hạn nếu để trống"
              />
            </div>
          )}

          {kind === "quiz" && (
            <div className="flex flex-col gap-3">
              <Label>Câu hỏi</Label>
              {questions.map((q, qIndex) => (
                <QuestionEditor
                  key={qIndex}
                  question={q}
                  onChange={(next) =>
                    setQuestions((qs) => qs.map((old, i) => (i === qIndex ? next : old)))
                  }
                  onRemove={
                    questions.length > 1
                      ? () => setQuestions((qs) => qs.filter((_, i) => i !== qIndex))
                      : undefined
                  }
                />
              ))}
              <Button
                type="button"
                size="sm"
                variant="ghost"
                className="w-fit"
                onClick={() => setQuestions((qs) => [...qs, emptyQuestion()])}
              >
                <PlusIcon className="size-4" />
                Thêm câu hỏi
              </Button>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            onClick={() => createMutation.mutate()}
            disabled={!canSubmit || createMutation.isPending}
          >
            {createMutation.isPending ? "Đang tạo..." : "Tạo bài tập"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function QuestionEditor({
  question,
  onChange,
  onRemove,
}: {
  question: QuestionDraft;
  onChange: (next: QuestionDraft) => void;
  onRemove?: () => void;
}) {
  return (
    <div className="flex flex-col gap-2 rounded-md border p-3">
      <div className="flex items-start gap-2">
        <Input
          value={question.content}
          onChange={(e) => onChange({ ...question, content: e.target.value })}
          placeholder="Nội dung câu hỏi"
        />
        {onRemove && (
          <Button type="button" size="icon" variant="ghost" onClick={onRemove}>
            <TrashIcon className="size-4" />
          </Button>
        )}
      </div>

      <RadioGroup
        value={String(question.correct_index)}
        onValueChange={(v) => onChange({ ...question, correct_index: Number(v) })}
        className="flex flex-col gap-2"
      >
        {question.options.map((opt, oIndex) => (
          <div key={oIndex} className="flex items-center gap-2">
            <RadioGroupItem value={String(oIndex)} id={`q-${oIndex}-opt`} />
            <Input
              value={opt}
              onChange={(e) => {
                const options = question.options.map((o, i) => (i === oIndex ? e.target.value : o));
                onChange({ ...question, options });
              }}
              placeholder={`Lựa chọn ${oIndex + 1}`}
              className="flex-1"
            />
            {question.options.length > 2 && (
              <Button
                type="button"
                size="icon"
                variant="ghost"
                onClick={() => {
                  const options = question.options.filter((_, i) => i !== oIndex);
                  const correct_index =
                    question.correct_index >= options.length ? 0 : question.correct_index;
                  onChange({ ...question, options, correct_index });
                }}
              >
                <TrashIcon className="size-4" />
              </Button>
            )}
          </div>
        ))}
      </RadioGroup>
      <Button
        type="button"
        size="sm"
        variant="ghost"
        className="w-fit"
        onClick={() => onChange({ ...question, options: [...question.options, ""] })}
      >
        <PlusIcon className="size-4" />
        Thêm lựa chọn
      </Button>
    </div>
  );
}
