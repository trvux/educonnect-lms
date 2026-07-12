import { apiClient } from "@/lib/api-client";
import type { EnrolledStudent } from "@/lib/types";

// US3.2
export async function enrollInCourse(courseId: number) {
  const res = await apiClient.post(`/courses/${courseId}/enroll`);
  return res.data;
}

// US3.3
export async function listEnrolledStudents(courseId: number) {
  const res = await apiClient.get<EnrolledStudent[]>(`/courses/${courseId}/students`);
  return res.data;
}
