"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { BellIcon } from "@phosphor-icons/react";

import { sendNotification } from "@/lib/api/notifications";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

// US6.2 — giảng viên gửi thông báo trong hệ thống tới toàn bộ học viên đã
// đăng ký khóa học.
export function NotificationSendDialog({ courseId }: { courseId: number }) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState("");

  const sendMutation = useMutation({
    mutationFn: () => sendNotification(courseId, title, message),
    onSuccess: (sent) => {
      toast.success(`Đã gửi thông báo tới ${sent.length} học viên`);
      setTitle("");
      setMessage("");
      setOpen(false);
    },
    onError: () => toast.error("Gửi thông báo thất bại"),
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant="outline" size="sm" />}>
        <BellIcon className="size-4" />
        Gửi thông báo
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Gửi thông báo tới học viên</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label>Tiêu đề</Label>
            <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Có bài tập mới" />
          </div>
          <div className="flex flex-col gap-2">
            <Label>Nội dung</Label>
            <Textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Nội dung thông báo..."
            />
          </div>
        </div>
        <DialogFooter>
          <Button onClick={() => sendMutation.mutate()} disabled={!title.trim() || sendMutation.isPending}>
            {sendMutation.isPending ? "Đang gửi..." : "Gửi"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
