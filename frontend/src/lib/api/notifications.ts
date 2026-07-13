import { apiClient } from "@/lib/api-client";
import type { Notification } from "@/lib/types";

// US6.2
export async function listMyNotifications() {
  const res = await apiClient.get<Notification[]>(`/notifications`);
  return res.data;
}

export async function getUnreadCount() {
  const res = await apiClient.get<{ unread_count: number }>(`/notifications/unread-count`);
  return res.data.unread_count;
}

export async function markNotificationRead(id: number) {
  await apiClient.post(`/notifications/${id}/read`);
}

export async function sendNotification(courseId: number, title: string, message: string) {
  const res = await apiClient.post<Notification[]>(`/courses/${courseId}/notifications`, {
    title,
    message,
  });
  return res.data;
}
