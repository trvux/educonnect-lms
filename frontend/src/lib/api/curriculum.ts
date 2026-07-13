import { apiClient } from "@/lib/api-client";
import type { Chapter, CourseOutlineChapter, Lesson, LessonProgressEntry } from "@/lib/types";

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

// US4.7 — ids là toàn bộ ID chương của khóa học theo đúng thứ tự mới sau
// khi kéo-thả (server validate phải là hoán vị đầy đủ, không thiếu/thừa).
export async function reorderChapters(courseId: number, ids: number[]) {
  const res = await apiClient.patch<Chapter[]>(`/courses/${courseId}/chapters/reorder`, { ids });
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

// US4.6
export async function renameLesson(lessonId: number, title: string) {
  const res = await apiClient.patch<Lesson>(`/lessons/${lessonId}`, { title });
  return res.data;
}

export async function deleteLesson(lessonId: number) {
  await apiClient.delete(`/lessons/${lessonId}`);
}

// US4.7 — tương tự reorderChapters nhưng cho bài học trong 1 chương.
export async function reorderLessons(chapterId: number, ids: number[]) {
  const res = await apiClient.patch<Lesson[]>(`/chapters/${chapterId}/lessons/reorder`, { ids });
  return res.data;
}

// US4.10 — trạng thái hoàn thành + khóa của mọi bài học trong khóa học,
// theo góc nhìn người dùng hiện tại (backend tự tính, không tin client).
export async function getLessonProgress(courseId: number) {
  const res = await apiClient.get<LessonProgressEntry[]>(`/courses/${courseId}/lesson-progress`);
  return res.data;
}

// US4.10 — học viên đánh dấu đã học xong 1 bài; backend chặn (403) nếu bài
// đang bị khóa (bài trước chưa hoàn thành), idempotent nếu gọi lại.
export async function markLessonComplete(lessonId: number) {
  await apiClient.post(`/lessons/${lessonId}/complete`);
}

// US4.9/US4.10 — cây chương/bài học đầy đủ của khóa học kèm trạng thái
// hoàn thành/khóa từng bài, dùng để dựng sidebar "course player". Không có
// endpoint gộp sẵn ở backend nên gọi tuần tự: chapters + lesson-progress
// song song, rồi lessons của từng chương, ghép completed/locked vào.
export async function getCourseOutline(courseId: number): Promise<CourseOutlineChapter[]> {
  const [chapters, progress] = await Promise.all([listChapters(courseId), getLessonProgress(courseId)]);
  const progressByLessonId = new Map(progress.map((p) => [p.lesson_id, p]));

  return Promise.all(
    chapters.map(async (chapter) => {
      const lessons = await listLessons(chapter.id);
      return {
        ...chapter,
        lessons: lessons.map((l) => ({
          ...l,
          completed: progressByLessonId.get(l.id)?.completed ?? false,
          locked: progressByLessonId.get(l.id)?.locked ?? false,
        })),
      };
    })
  );
}
