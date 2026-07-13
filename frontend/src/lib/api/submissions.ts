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

// US5.2 — biết ngay trạng thái đã nộp/điểm khi vào trang làm bài, không
// cần đợi bấm Nộp bài mới biết qua lỗi 409. Trả về null nếu chưa nộp (404).
export async function getMySubmission(assignmentId: number) {
  try {
    const res = await apiClient.get<Submission>(`/assignments/${assignmentId}/my-submission`);
    return res.data;
  } catch (error: unknown) {
    const status = (error as { response?: { status?: number } })?.response?.status;
    if (status === 404) return null;
    throw error;
  }
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
