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

// US4.6
export async function renameChapter(chapterId: number, title: string) {
  const res = await apiClient.patch<Chapter>(`/chapters/${chapterId}`, { title });
  return res.data;
}

export async function deleteChapter(chapterId: number) {
  await apiClient.delete(`/chapters/${chapterId}`);
}

export async function listLessons(chapterId: number) {
  const res = await apiClient.get<Lesson[]>(`/chapters/${chapterId}/lessons`);
  return res.data;
}

export async function createLesson(chapterId: number, title: string) {
  const res = await apiClient.post<Lesson>(`/chapters/${chapterId}/lessons`, { title });
  return res.data;
}

// US4.6
export async function renameLesson(lessonId: number, title: string) {
  const res = await apiClient.patch<Lesson>(`/lessons/${lessonId}`, { title });
  return res.data;
}

export async function deleteLesson(lessonId: number) {
  await apiClient.delete(`/lessons/${lessonId}`);
}
