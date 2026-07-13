import { apiClient } from "@/lib/api-client";
import type { CourseProgress, CourseStats } from "@/lib/types";

// US7.1
export async function getMyProgress() {
  const res = await apiClient.get<CourseProgress[]>(`/me/progress`);
  return res.data;
}

// US7.2
export async function getCourseReports() {
  const res = await apiClient.get<CourseStats[]>(`/reports/courses`);
  return res.data;
}
