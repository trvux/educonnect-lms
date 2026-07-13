import { apiClient } from "@/lib/api-client";
import type { Assignment, AssignmentKind, Question, QuizAttempt } from "@/lib/types";

// US5.1
export async function listAssignments(lessonId: number) {
  const res = await apiClient.get<Assignment[]>(`/lessons/${lessonId}/assignments`);
  return res.data;
}

export async function getAssignment(id: number) {
  const res = await apiClient.get<Assignment>(`/assignments/${id}`);
  return res.data;
}

export async function createAssignment(
  lessonId: number,
  input: {
    title: string;
    description: string;
    kind: AssignmentKind;
    questions: Question[];
    due_at?: string;
    time_limit_minutes?: number;
  }
) {
  const res = await apiClient.post<Assignment>(`/lessons/${lessonId}/assignments`, input);
  return res.data;
}

// US5.4 — ghi nhận thời điểm học viên bắt đầu làm bài trắc nghiệm có giới
// hạn thời gian; idempotent, gọi lại (refresh trang) trả về đúng
// started_at gốc, không reset đồng hồ đếm ngược.
export async function startQuizAttempt(assignmentId: number) {
  const res = await apiClient.post<QuizAttempt>(`/assignments/${assignmentId}/start-attempt`);
  return res.data;
}
