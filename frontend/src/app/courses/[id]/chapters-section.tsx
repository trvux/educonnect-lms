"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { DownloadIcon, PencilIcon, PlusIcon, TrashIcon, UploadIcon } from "@phosphor-icons/react";

import {
  listChapters,
  createChapter,
  renameChapter,
  deleteChapter,
  listLessons,
  createLesson,
  renameLesson,
  deleteLesson,
} from "@/lib/api/curriculum";
import { listMaterials, uploadMaterial, downloadMaterial, deleteMaterial } from "@/lib/api/materials";
import type { Chapter, Lesson, Material, MaterialFileType } from "@/lib/types";
import { AssignmentsSection } from "./assignments-section";
import { Badge } from "@/components/ui/badge";
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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

// US4.4 — whitelist định dạng file, khớp danh sách backend chấp nhận
// (internal/domain/material/material.go#extensionToFileType). Thuộc tính
// accept chỉ là gợi ý UX (lọc bớt lựa chọn trong hộp thoại chọn file);
// validation thật vẫn nằm ở backend.
const ACCEPTED_FILE_EXTENSIONS =
  ".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.mp4,.webm,.mov,.zip,.rar,.7z";

const fileTypeLabel: Record<MaterialFileType, string> = {
  pdf: "PDF",
  doc: "Word",
  excel: "Excel",
  ppt: "PowerPoint",
  video: "Video",
  archive: "Nén",
};

// US2.2 (chương/bài học) + US4.1/US4.2 (tài liệu) + US4.6 (sửa/xóa). Mobile-
// first: mỗi chương là 1 Accordion item để tiết kiệm không gian màn hình
// nhỏ; bài học/tài liệu hiện dạng danh sách dọc, không cần bảng vì ít cột
// dữ liệu.
export function ChaptersSection({ courseId, canManage }: { courseId: number; canManage: boolean }) {
  const queryClient = useQueryClient();
  const [newChapterTitle, setNewChapterTitle] = useState("");
  const [chapterDialogOpen, setChapterDialogOpen] = useState(false);
  const [renameTarget, setRenameTarget] = useState<Chapter | null>(null);
  const [renameTitle, setRenameTitle] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<Chapter | null>(null);

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

  const renameChapterMutation = useMutation({
    mutationFn: () => renameChapter(renameTarget!.id, renameTitle),
    onSuccess: () => {
      toast.success("Đã đổi tên chương");
      setRenameTarget(null);
      queryClient.invalidateQueries({ queryKey: ["chapters", courseId] });
    },
    onError: () => toast.error("Đổi tên chương thất bại"),
  });

  const deleteChapterMutation = useMutation({
    mutationFn: () => deleteChapter(deleteTarget!.id),
    onSuccess: () => {
      toast.success("Đã xóa chương");
      setDeleteTarget(null);
      queryClient.invalidateQueries({ queryKey: ["chapters", courseId] });
    },
    onError: (error: unknown) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Xóa chương thất bại";
      toast.error(message);
      setDeleteTarget(null);
    },
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
            <div className="flex items-center">
              <AccordionTrigger className="flex-1">{chapter.title}</AccordionTrigger>
              {canManage && (
                <div className="flex items-center gap-1 pr-2">
                  <Button
                    type="button"
                    size="icon-sm"
                    variant="ghost"
                    aria-label="Đổi tên chương"
                    onClick={() => {
                      setRenameTarget(chapter);
                      setRenameTitle(chapter.title);
                    }}
                  >
                    <PencilIcon className="size-4" />
                  </Button>
                  <Button
                    type="button"
                    size="icon-sm"
                    variant="ghost"
                    aria-label="Xóa chương"
                    onClick={() => setDeleteTarget(chapter)}
                  >
                    <TrashIcon className="size-4" />
                  </Button>
                </div>
              )}
            </div>
            <AccordionContent>
              <LessonsSection chapterId={chapter.id} canManage={canManage} />
            </AccordionContent>
          </AccordionItem>
        ))}
      </Accordion>

      <Dialog
        open={renameTarget !== null}
        onOpenChange={(open) => !open && setRenameTarget(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Đổi tên chương</DialogTitle>
          </DialogHeader>
          <Input value={renameTitle} onChange={(e) => setRenameTitle(e.target.value)} />
          <DialogFooter>
            <Button
              onClick={() => renameChapterMutation.mutate()}
              disabled={!renameTitle || renameChapterMutation.isPending}
            >
              Lưu
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Xóa chương &quot;{deleteTarget?.title}&quot;?</AlertDialogTitle>
            <AlertDialogDescription>
              Chỉ xóa được nếu chương không còn bài học nào bên trong. Hành động này không thể hoàn tác.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Hủy</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteChapterMutation.mutate()}
              disabled={deleteChapterMutation.isPending}
            >
              Xóa
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function LessonsSection({ chapterId, canManage }: { chapterId: number; canManage: boolean }) {
  const queryClient = useQueryClient();
  const [newLessonTitle, setNewLessonTitle] = useState("");
  const [lessonDialogOpen, setLessonDialogOpen] = useState(false);
  const [renameTarget, setRenameTarget] = useState<Lesson | null>(null);
  const [renameTitle, setRenameTitle] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<Lesson | null>(null);

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

  const renameLessonMutation = useMutation({
    mutationFn: () => renameLesson(renameTarget!.id, renameTitle),
    onSuccess: () => {
      toast.success("Đã đổi tên bài học");
      setRenameTarget(null);
      queryClient.invalidateQueries({ queryKey: ["lessons", chapterId] });
    },
    onError: () => toast.error("Đổi tên bài học thất bại"),
  });

  const deleteLessonMutation = useMutation({
    mutationFn: () => deleteLesson(deleteTarget!.id),
    onSuccess: () => {
      toast.success("Đã xóa bài học");
      setDeleteTarget(null);
      queryClient.invalidateQueries({ queryKey: ["lessons", chapterId] });
    },
    onError: (error: unknown) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Xóa bài học thất bại";
      toast.error(message);
      setDeleteTarget(null);
    },
  });

  return (
    <div className="flex flex-col gap-3 pl-1">
      {isLoading && <p className="text-sm text-muted-foreground">Đang tải bài học...</p>}
      {!isLoading && lessons?.length === 0 && (
        <p className="text-sm text-muted-foreground">Chưa có bài học nào trong chương này.</p>
      )}

      {lessons?.map((lesson) => (
        <div key={lesson.id} className="rounded-md border p-3">
          <div className="flex items-center justify-between gap-2">
            <p className="font-medium">{lesson.title}</p>
            {canManage && (
              <div className="flex items-center gap-1">
                <Button
                  type="button"
                  size="icon-sm"
                  variant="ghost"
                  aria-label="Đổi tên bài học"
                  onClick={() => {
                    setRenameTarget(lesson);
                    setRenameTitle(lesson.title);
                  }}
                >
                  <PencilIcon className="size-4" />
                </Button>
                <Button
                  type="button"
                  size="icon-sm"
                  variant="ghost"
                  aria-label="Xóa bài học"
                  onClick={() => setDeleteTarget(lesson)}
                >
                  <TrashIcon className="size-4" />
                </Button>
              </div>
            )}
          </div>
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

      <Dialog
        open={renameTarget !== null}
        onOpenChange={(open) => !open && setRenameTarget(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Đổi tên bài học</DialogTitle>
          </DialogHeader>
          <Input value={renameTitle} onChange={(e) => setRenameTitle(e.target.value)} />
          <DialogFooter>
            <Button
              onClick={() => renameLessonMutation.mutate()}
              disabled={!renameTitle || renameLessonMutation.isPending}
            >
              Lưu
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Xóa bài học &quot;{deleteTarget?.title}&quot;?</AlertDialogTitle>
            <AlertDialogDescription>
              Chỉ xóa được nếu bài học không còn tài liệu/bài tập bên trong. Hành động này không thể hoàn tác.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Hủy</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteLessonMutation.mutate()}
              disabled={deleteLessonMutation.isPending}
            >
              Xóa
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

export function MaterialsList({ lessonId, canManage }: { lessonId: number; canManage: boolean }) {
  const queryClient = useQueryClient();
  const [deleteTarget, setDeleteTarget] = useState<Material | null>(null);

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
    onError: (error: unknown) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Tải tài liệu lên thất bại";
      toast.error(message);
    },
  });

  const downloadMutation = useMutation({
    mutationFn: ({ id, fileName }: { id: number; fileName: string }) => downloadMaterial(id, fileName),
    onError: () => toast.error("Tải tài liệu thất bại, vui lòng thử lại"),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteMaterial(deleteTarget!.id),
    onSuccess: () => {
      toast.success("Đã xóa tài liệu");
      setDeleteTarget(null);
      queryClient.invalidateQueries({ queryKey: ["materials", lessonId] });
    },
    onError: () => {
      toast.error("Xóa tài liệu thất bại");
      setDeleteTarget(null);
    },
  });

  return (
    <div className="mt-2 flex flex-col gap-2">
      {materials?.map((m) => (
        <div key={m.id} className="flex w-full items-center justify-between gap-2">
          <button
            type="button"
            onClick={() => downloadMutation.mutate({ id: m.id, fileName: m.file_name })}
            disabled={downloadMutation.isPending}
            className="flex min-w-0 items-center gap-2 text-sm text-primary hover:underline disabled:opacity-50"
          >
            <DownloadIcon className="size-4 shrink-0" />
            <span className="truncate">{m.file_name}</span>
            <Badge variant="secondary">{fileTypeLabel[m.file_type]}</Badge>
          </button>
          {canManage && (
            <Button
              type="button"
              size="icon-sm"
              variant="ghost"
              aria-label="Xóa tài liệu"
              onClick={() => setDeleteTarget(m)}
            >
              <TrashIcon className="size-4" />
            </Button>
          )}
        </div>
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
            accept={ACCEPTED_FILE_EXTENSIONS}
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

      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Xóa tài liệu &quot;{deleteTarget?.file_name}&quot;?</AlertDialogTitle>
            <AlertDialogDescription>Hành động này không thể hoàn tác.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Hủy</AlertDialogCancel>
            <AlertDialogAction onClick={() => deleteMutation.mutate()} disabled={deleteMutation.isPending}>
              Xóa
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
