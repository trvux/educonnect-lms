import { apiClient } from "@/lib/api-client";
import type { ForumPost } from "@/lib/types";

// US6.1
export async function listForumPosts(courseId: number) {
  const res = await apiClient.get<ForumPost[]>(`/courses/${courseId}/forum-posts`);
  return res.data;
}

export async function createForumPost(courseId: number, content: string, parentId?: number) {
  const res = await apiClient.post<ForumPost>(`/courses/${courseId}/forum-posts`, {
    content,
    parent_id: parentId,
  });
  return res.data;
}
