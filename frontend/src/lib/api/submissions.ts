import { apiClient } from "@/lib/api-client";
import type { Submission } from "@/lib/types";

// US5.2
export async function submitAssignment(
  assignmentId: number,
  input: { content?: string; answers?: number[] }
) {
  const res = await apiClient.post<Submission>(`/assignments/${assignmentId}/submit`, input);
  return res.data;
}

// US5.3 — giảng viên xem danh sách bài nộp để chấm điểm.
export async function listSubmissions(assignmentId: number) {
  const res = await apiClient.get<Submission[]>(`/assignments/${assignmentId}/submissions`);
  return res.data;
}

export async function gradeSubmission(submissionId: number, score: number, feedback: string) {
  const res = await apiClient.post<Submission>(`/submissions/${submissionId}/grade`, {
    score,
    feedback,
  });
  return res.data;
}
