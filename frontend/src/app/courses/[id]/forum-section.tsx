"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ArrowBendUpLeftIcon } from "@phosphor-icons/react";

import { createForumPost, listForumPosts } from "@/lib/api/forum";
import { useSession } from "@/lib/auth";
import type { ForumPost } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";

// US6.1 — diễn đàn hỏi đáp theo khóa học, comment lồng 1 cấp (câu hỏi gốc +
// trả lời) vì Post.ParentID chỉ trỏ về bài gốc, không lồng sâu hơn.
export function ForumSection({ courseId }: { courseId: number }) {
  const session = useSession();
  const queryClient = useQueryClient();
  const [newQuestion, setNewQuestion] = useState("");
  const [replyTo, setReplyTo] = useState<number | null>(null);
  const [replyContent, setReplyContent] = useState("");

  const { data: posts, isLoading } = useQuery({
    queryKey: ["forum-posts", courseId],
    queryFn: () => listForumPosts(courseId),
  });

  const postMutation = useMutation({
    mutationFn: (input: { content: string; parentId?: number }) =>
      createForumPost(courseId, input.content, input.parentId),
    onSuccess: () => {
      setNewQuestion("");
      setReplyContent("");
      setReplyTo(null);
      queryClient.invalidateQueries({ queryKey: ["forum-posts", courseId] });
    },
    onError: () => toast.error("Đăng bài thất bại"),
  });

  const questions = posts?.filter((p) => !p.parent_id) ?? [];
  const repliesByParent = new Map<number, ForumPost[]>();
  for (const p of posts ?? []) {
    if (p.parent_id) {
      repliesByParent.set(p.parent_id, [...(repliesByParent.get(p.parent_id) ?? []), p]);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Diễn đàn hỏi đáp</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {isLoading && <p className="text-sm text-muted-foreground">Đang tải...</p>}
        {!isLoading && questions.length === 0 && (
          <p className="text-sm text-muted-foreground">Chưa có câu hỏi nào.</p>
        )}

        <div className="flex flex-col gap-4">
          {questions.map((q) => (
            <div key={q.id} className="flex flex-col gap-2 rounded-md border p-3">
              <div className="flex items-baseline justify-between gap-2">
                <p className="text-sm font-medium">{q.author_name ?? `User #${q.author_id}`}</p>
                <p className="text-xs text-muted-foreground">
                  {new Date(q.created_at).toLocaleString("vi-VN")}
                </p>
              </div>
              <p className="text-sm">{q.content}</p>

              <div className="ml-4 flex flex-col gap-2 border-l pl-3">
                {repliesByParent.get(q.id)?.map((r) => (
                  <div key={r.id} className="flex flex-col gap-1">
                    <div className="flex items-baseline justify-between gap-2">
                      <p className="text-sm font-medium">{r.author_name ?? `User #${r.author_id}`}</p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(r.created_at).toLocaleString("vi-VN")}
                      </p>
                    </div>
                    <p className="text-sm">{r.content}</p>
                  </div>
                ))}
              </div>

              {session &&
                (replyTo === q.id ? (
                  <div className="ml-4 flex flex-col gap-2">
                    <Textarea
                      value={replyContent}
                      onChange={(e) => setReplyContent(e.target.value)}
                      placeholder="Trả lời..."
                      rows={2}
                    />
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        disabled={!replyContent.trim() || postMutation.isPending}
                        onClick={() => postMutation.mutate({ content: replyContent, parentId: q.id })}
                      >
                        Gửi
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => setReplyTo(null)}>
                        Hủy
                      </Button>
                    </div>
                  </div>
                ) : (
                  <Button
                    size="sm"
                    variant="ghost"
                    className="ml-4 w-fit"
                    onClick={() => setReplyTo(q.id)}
                  >
                    <ArrowBendUpLeftIcon className="size-4" />
                    Trả lời
                  </Button>
                ))}
            </div>
          ))}
        </div>

        {session && (
          <div className="flex flex-col gap-2 border-t pt-4">
            <Textarea
              value={newQuestion}
              onChange={(e) => setNewQuestion(e.target.value)}
              placeholder="Đặt câu hỏi mới cho khóa học này..."
              rows={2}
            />
            <Button
              className="w-fit"
              disabled={!newQuestion.trim() || postMutation.isPending}
              onClick={() => postMutation.mutate({ content: newQuestion })}
            >
              {postMutation.isPending ? "Đang đăng..." : "Đăng câu hỏi"}
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
