"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { DownloadIcon, PlusIcon, UploadIcon } from "@phosphor-icons/react";

import {
  listChapters,
  createChapter,
  listLessons,
  createLesson,
} from "@/lib/api/curriculum";
import { listMaterials, uploadMaterial, downloadMaterial } from "@/lib/api/materials";
import { AssignmentsSection } from "./assignments-section";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

// US2.2 (chương/bài học) + US4.1/US4.2 (tài liệu). Mobile-first: mỗi chương
// là 1 Accordion item để tiết kiệm không gian màn hình nhỏ; bài học/tài
// liệu hiện dạng danh sách dọc, không cần bảng vì ít cột dữ liệu.
export function ChaptersSection({ courseId, canManage }: { courseId: number; canManage: boolean }) {
  const queryClient = useQueryClient();
  const [newChapterTitle, setNewChapterTitle] = useState("");
  const [chapterDialogOpen, setChapterDialogOpen] = useState(false);

  const { data: chapters, isLoading } = useQuery({
    queryKey: ["chapters", courseId],
    queryFn: () => listChapters(courseId),
  });

  const createChapterMutation = useMutation({
    mutationFn: () => createChapter(courseId, newChapterTitle),
    onSuccess: () => {
      toast.success("Đã thêm chương mới");
      setNewChapterTitle("");
      setChapterDialogOpen(false);
      queryClient.invalidateQueries({ queryKey: ["chapters", courseId] });
    },
    onError: () => toast.error("Thêm chương thất bại"),
  });

  return (
    <div>
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Nội dung khóa học</h2>
        {canManage && (
          <Dialog open={chapterDialogOpen} onOpenChange={setChapterDialogOpen}>
            <DialogTrigger render={<Button size="sm" variant="outline" />}>
              <PlusIcon className="size-4" />
              Thêm chương
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Thêm chương mới</DialogTitle>
              </DialogHeader>
              <Input
                placeholder="Tên chương"
                value={newChapterTitle}
                onChange={(e) => setNewChapterTitle(e.target.value)}
              />
              <DialogFooter>
                <Button
                  onClick={() => createChapterMutation.mutate()}
                  disabled={!newChapterTitle || createChapterMutation.isPending}
                >
                  Lưu
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {isLoading && <p className="mt-4 text-sm text-muted-foreground">Đang tải...</p>}
      {!isLoading && chapters?.length === 0 && (
        <p className="mt-4 text-sm text-muted-foreground">Chưa có chương nào.</p>
      )}

      <Accordion multiple className="mt-2">
        {chapters?.map((chapter) => (
          <AccordionItem key={chapter.id} value={String(chapter.id)}>
            <AccordionTrigger>{chapter.title}</AccordionTrigger>
            <AccordionContent>
              <LessonsSection chapterId={chapter.id} canManage={canManage} />
            </AccordionContent>
          </AccordionItem>
        ))}
      </Accordion>
    </div>
  );
}

function LessonsSection({ chapterId, canManage }: { chapterId: number; canManage: boolean }) {
  const queryClient = useQueryClient();
  const [newLessonTitle, setNewLessonTitle] = useState("");
  const [lessonDialogOpen, setLessonDialogOpen] = useState(false);

  const { data: lessons, isLoading } = useQuery({
    queryKey: ["lessons", chapterId],
    queryFn: () => listLessons(chapterId),
  });

  const createLessonMutation = useMutation({
    mutationFn: () => createLesson(chapterId, newLessonTitle),
    onSuccess: () => {
      toast.success("Đã thêm bài học mới");
      setNewLessonTitle("");
      setLessonDialogOpen(false);
      queryClient.invalidateQueries({ queryKey: ["lessons", chapterId] });
    },
    onError: () => toast.error("Thêm bài học thất bại"),
  });

  return (
    <div className="flex flex-col gap-3 pl-1">
      {isLoading && <p className="text-sm text-muted-foreground">Đang tải bài học...</p>}
      {!isLoading && lessons?.length === 0 && (
        <p className="text-sm text-muted-foreground">Chưa có bài học nào trong chương này.</p>
      )}

      {lessons?.map((lesson) => (
        <div key={lesson.id} className="rounded-md border p-3">
          <p className="font-medium">{lesson.title}</p>
          <MaterialsList lessonId={lesson.id} canManage={canManage} />
          <AssignmentsSection lessonId={lesson.id} canManage={canManage} />
        </div>
      ))}

      {canManage && (
        <Dialog open={lessonDialogOpen} onOpenChange={setLessonDialogOpen}>
          <DialogTrigger render={<Button size="sm" variant="ghost" className="w-fit" />}>
            <PlusIcon className="size-4" />
            Thêm bài học
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Thêm bài học mới</DialogTitle>
            </DialogHeader>
            <Input
              placeholder="Tên bài học"
              value={newLessonTitle}
              onChange={(e) => setNewLessonTitle(e.target.value)}
            />
            <DialogFooter>
              <Button
                onClick={() => createLessonMutation.mutate()}
                disabled={!newLessonTitle || createLessonMutation.isPending}
              >
                Lưu
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}

function MaterialsList({ lessonId, canManage }: { lessonId: number; canManage: boolean }) {
  const queryClient = useQueryClient();

  const { data: materials } = useQuery({
    queryKey: ["materials", lessonId],
    queryFn: () => listMaterials(lessonId),
  });

  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadMaterial(lessonId, file),
    onSuccess: () => {
      toast.success("Tải tài liệu lên thành công");
      queryClient.invalidateQueries({ queryKey: ["materials", lessonId] });
    },
    onError: () => toast.error("Tải tài liệu lên thất bại"),
  });

  const downloadMutation = useMutation({
    mutationFn: ({ id, fileName }: { id: number; fileName: string }) => downloadMaterial(id, fileName),
    onError: () => toast.error("Tải tài liệu thất bại, vui lòng thử lại"),
  });

  return (
    <div className="mt-2 flex flex-col gap-2">
      {materials?.map((m) => (
        <button
          key={m.id}
          type="button"
          onClick={() => downloadMutation.mutate({ id: m.id, fileName: m.file_name })}
          disabled={downloadMutation.isPending}
          className="flex w-fit items-center gap-2 text-sm text-primary hover:underline disabled:opacity-50"
        >
          <DownloadIcon className="size-4" />
          {m.file_name}
        </button>
      ))}
      {materials?.length === 0 && (
        <p className="text-sm text-muted-foreground">Chưa có tài liệu.</p>
      )}

      {canManage && (
        <label className="flex w-fit cursor-pointer items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <UploadIcon className="size-4" />
          {uploadMutation.isPending ? "Đang tải lên..." : "Upload tài liệu"}
          <input
            type="file"
            className="hidden"
            disabled={uploadMutation.isPending}
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) uploadMutation.mutate(file);
              e.target.value = "";
            }}
          />
        </label>
      )}
    </div>
  );
}
