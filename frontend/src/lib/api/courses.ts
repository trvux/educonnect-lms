import { apiClient } from "@/lib/api-client";
import type { Course } from "@/lib/types";

// US3.1: tìm kiếm khóa học (public, chỉ trả về course đã Approved).
export async function searchCourses(keyword: string) {
  const res = await apiClient.get<Course[]>("/courses", { params: { search: keyword } });
  return res.data;
}

// Xem chi tiết 1 khóa học, public.
export async function getCourse(id: number) {
  const res = await apiClient.get<Course>(`/courses/${id}`);
  return res.data;
}

// US2.1: giảng viên tạo khóa học (Draft).
export async function createCourse(input: { title: string; description: string }) {
  const res = await apiClient.post<Course>("/courses", input);
  return res.data;
}

// US2.1 (submit): giảng viên gửi khóa học của mình để admin duyệt.
export async function submitCourseForReview(courseId: number) {
  const res = await apiClient.post<Course>(`/courses/${courseId}/submit`);
  return res.data;
}

// US2.3: quản trị viên duyệt khóa học đang pending_review.
export async function approveCourse(courseId: number) {
  const res = await apiClient.post<Course>(`/admin/courses/${courseId}/approve`);
  return res.data;
}
