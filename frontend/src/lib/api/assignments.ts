import { apiClient } from "@/lib/api-client";
import type { Assignment, AssignmentKind, Question } from "@/lib/types";

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
  input: { title: string; description: string; kind: AssignmentKind; questions: Question[]; due_at?: string }
) {
  const res = await apiClient.post<Assignment>(`/lessons/${lessonId}/assignments`, input);
  return res.data;
}
