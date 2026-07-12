"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { createCourse } from "@/lib/api/courses";
import { useSession } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

// US2.1: Giảng viên tạo khóa học mới (bắt đầu ở trạng thái Draft).
const schema = z.object({
  title: z.string().min(3, "Tiêu đề phải có ít nhất 3 ký tự"),
  description: z.string().optional(),
});

type Values = z.infer<typeof schema>;

export default function NewCoursePage() {
  const router = useRouter();
  const session = useSession();
  const queryClient = useQueryClient();

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { title: "", description: "" },
  });

  const mutation = useMutation({
    mutationFn: (values: Values) => createCourse({ title: values.title, description: values.description ?? "" }),
    onSuccess: (course) => {
      toast.success("Tạo khóa học thành công (đang ở trạng thái Draft)");
      queryClient.invalidateQueries({ queryKey: ["courses"] });
      router.push(`/courses/${course.id}`);
    },
    onError: () => toast.error("Tạo khóa học thất bại, vui lòng thử lại"),
  });

  if (session && session.role === "student") {
    return (
      <div className="mx-auto max-w-md px-4 py-10 text-center text-sm text-muted-foreground">
        Chỉ Giảng viên/Quản trị viên mới có thể tạo khóa học.
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-6 sm:py-10">
      <Card>
        <CardHeader>
          <CardTitle>Tạo khóa học mới</CardTitle>
          <CardDescription>Khóa học sẽ ở trạng thái Draft cho đến khi bạn gửi duyệt.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form
              className="flex flex-col gap-6"
              onSubmit={form.handleSubmit((values) => mutation.mutate(values))}
            >
              <FormField
                control={form.control}
                name="title"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Tiêu đề khóa học</FormLabel>
                    <FormControl>
                      <Input placeholder="Nhập môn Golang" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Mô tả</FormLabel>
                    <FormControl>
                      <Textarea rows={4} placeholder="Mô tả ngắn về khóa học..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? "Đang tạo..." : "Tạo khóa học"}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
