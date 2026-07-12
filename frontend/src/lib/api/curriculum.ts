import { apiClient } from "@/lib/api-client";
import type { Chapter, Lesson } from "@/lib/types";

// US2.2
export async function listChapters(courseId: number) {
  const res = await apiClient.get<Chapter[]>(`/courses/${courseId}/chapters`);
  return res.data;
}

export async function createChapter(courseId: number, title: string) {
  const res = await apiClient.post<Chapter>(`/courses/${courseId}/chapters`, { title });
  return res.data;
}

export async function listLessons(chapterId: number) {
  const res = await apiClient.get<Lesson[]>(`/chapters/${chapterId}/lessons`);
  return res.data;
}

export async function createLesson(chapterId: number, title: string) {
  const res = await apiClient.post<Lesson>(`/chapters/${chapterId}/lessons`, { title });
  return res.data;
}
