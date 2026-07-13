import { apiClient } from "@/lib/api-client";
import type { GradebookEntry } from "@/lib/types";

// US5.3
export async function getGradebook(courseId: number) {
  const res = await apiClient.get<GradebookEntry[]>(`/courses/${courseId}/gradebook`);
  return res.data;
}
