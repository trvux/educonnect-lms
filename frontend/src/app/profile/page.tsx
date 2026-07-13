"use client";

import { useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CameraIcon } from "@phosphor-icons/react";

import { getMe, updateMe, uploadAvatar, changePassword, avatarUrl } from "@/lib/api/auth";
import { createRoleUpgradeRequest } from "@/lib/api/role-upgrade";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

const roleLabel: Record<string, string> = {
  student: "Học viên",
  teacher: "Giảng viên",
  admin: "Quản trị viên",
};

// US1.4 — thông tin liên hệ, khớp regex backend (0xxxxxxxxx hoặc +84xxxxxxxxx).
const phonePattern = /^(0\d{9}|\+84\d{9})$/;

const profileSchema = z.object({
  fullName: z.string().min(2, "Họ tên phải có ít nhất 2 ký tự"),
  phone: z
    .string()
    .refine((v) => v === "" || phonePattern.test(v), "Số điện thoại không hợp lệ (vd 0987654321)"),
  studentCode: z.string(),
});
type ProfileValues = z.infer<typeof profileSchema>;

const passwordSchema = z
  .object({
    currentPassword: z.string().min(1, "Vui lòng nhập mật khẩu hiện tại"),
    newPassword: z.string().min(8, "Mật khẩu mới phải có ít nhất 8 ký tự"),
    confirmPassword: z.string().min(1, "Vui lòng nhập lại mật khẩu mới"),
  })
  .refine((v) => v.newPassword === v.confirmPassword, {
    message: "Mật khẩu nhập lại không khớp",
    path: ["confirmPassword"],
  });
type PasswordValues = z.infer<typeof passwordSchema>;

export default function ProfilePage() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [reason, setReason] = useState("");

  const { data: me, isLoading } = useQuery({ queryKey: ["me"], queryFn: getMe });

  const profileForm = useForm<ProfileValues>({
    resolver: zodResolver(profileSchema),
    // Luôn truyền object đã định nghĩa đủ field (không phải undefined) — nếu
    // không, lần render đầu tiên (trước khi `me` tải xong) sẽ khởi tạo field
    // với value undefined (uncontrolled), rồi effect sync của react-hook-form
    // set lại thành string thật ngay sau đó (controlled), gây cảnh báo
    // Base UI "uncontrolled to controlled".
    values: {
      fullName: me?.full_name ?? "",
      phone: me?.phone ?? "",
      studentCode: me?.student_code ?? "",
    },
  });

  const updateProfileMutation = useMutation({
    mutationFn: (values: ProfileValues) =>
      updateMe({ full_name: values.fullName, phone: values.phone, student_code: values.studentCode }),
    onSuccess: (updated) => {
      queryClient.setQueryData(["me"], updated);
      toast.success("Đã cập nhật hồ sơ");
    },
    onError: () => toast.error("Cập nhật hồ sơ thất bại"),
  });

  const avatarMutation = useMutation({
    mutationFn: uploadAvatar,
    onSuccess: (updated) => {
      queryClient.setQueryData(["me"], updated);
      toast.success("Đã cập nhật ảnh đại diện");
    },
    onError: () => toast.error("Tải ảnh đại diện thất bại"),
  });

  const passwordForm = useForm<PasswordValues>({
    resolver: zodResolver(passwordSchema),
    defaultValues: { currentPassword: "", newPassword: "", confirmPassword: "" },
  });

  const changePasswordMutation = useMutation({
    mutationFn: (values: PasswordValues) =>
      changePassword({ current_password: values.currentPassword, new_password: values.newPassword }),
    onSuccess: () => {
      toast.success("Đã đổi mật khẩu");
      passwordForm.reset();
    },
    onError: (error: unknown) => {
      const status = (error as { response?: { status?: number } })?.response?.status;
      toast.error(status === 401 ? "Mật khẩu hiện tại không đúng" : "Đổi mật khẩu thất bại");
    },
  });

  const roleUpgradeMutation = useMutation({
    mutationFn: () => createRoleUpgradeRequest(reason),
    onSuccess: () => {
      toast.success("Đã gửi yêu cầu, chờ quản trị viên duyệt");
      setReason("");
    },
    onError: (error: unknown) => {
      const status = (error as { response?: { status?: number } })?.response?.status;
      toast.error(
        status === 409 ? "Bạn đã có yêu cầu đang chờ duyệt" : "Gửi yêu cầu thất bại, vui lòng thử lại"
      );
    },
  });

  if (isLoading || !me) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-6 sm:py-10">
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6 px-4 py-6 sm:py-10">
      <div className="flex items-center gap-4">
        <div className="relative">
          <Avatar size="lg" className="size-16">
            {me.avatar_path && <AvatarImage src={avatarUrl(me.avatar_path)} alt={me.full_name} />}
            <AvatarFallback className="text-lg">{me.full_name.charAt(0).toUpperCase()}</AvatarFallback>
          </Avatar>
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className="absolute -right-1 -bottom-1 flex size-6 items-center justify-center rounded-full bg-primary text-primary-foreground"
            aria-label="Đổi ảnh đại diện"
          >
            <CameraIcon className="size-3.5" />
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) avatarMutation.mutate(file);
              e.target.value = "";
            }}
          />
        </div>
        <div>
          <h1 className="text-xl font-semibold tracking-tight">{me.full_name}</h1>
          <p className="text-sm text-muted-foreground">{me.email}</p>
          <Badge variant="secondary" className="mt-1">
            {roleLabel[me.role]}
          </Badge>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Thông tin cá nhân</CardTitle>
          <CardDescription>Cập nhật họ tên, số điện thoại và mã số sinh viên/giảng viên</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...profileForm}>
            <form
              className="flex flex-col gap-4"
              onSubmit={profileForm.handleSubmit((values) => updateProfileMutation.mutate(values))}
            >
              <FormField
                control={profileForm.control}
                name="fullName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Họ và tên</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={profileForm.control}
                name="phone"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Số điện thoại</FormLabel>
                    <FormControl>
                      <Input placeholder="0987654321" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={profileForm.control}
                name="studentCode"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Mã số sinh viên/giảng viên</FormLabel>
                    <FormControl>
                      <Input placeholder="2180xxxxxx" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" disabled={updateProfileMutation.isPending} className="self-start">
                {updateProfileMutation.isPending ? "Đang lưu..." : "Lưu thay đổi"}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Đổi mật khẩu</CardTitle>
        </CardHeader>
        <CardContent>
          <Form {...passwordForm}>
            <form
              className="flex flex-col gap-4"
              onSubmit={passwordForm.handleSubmit((values) => changePasswordMutation.mutate(values))}
            >
              <FormField
                control={passwordForm.control}
                name="currentPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Mật khẩu hiện tại</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={passwordForm.control}
                name="newPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Mật khẩu mới</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={passwordForm.control}
                name="confirmPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Nhập lại mật khẩu mới</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" disabled={changePasswordMutation.isPending} className="self-start">
                {changePasswordMutation.isPending ? "Đang đổi..." : "Đổi mật khẩu"}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>

      {me.role === "student" && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Trở thành Giảng viên</CardTitle>
            <CardDescription>
              Gửi yêu cầu kèm lý do, quản trị viên sẽ xem xét và duyệt.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <Textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Vd: Tôi muốn tạo khóa học phụ đạo cho lớp..."
            />
            <Button
              onClick={() => roleUpgradeMutation.mutate()}
              disabled={!reason.trim() || roleUpgradeMutation.isPending}
              className="self-start"
            >
              {roleUpgradeMutation.isPending ? "Đang gửi..." : "Gửi yêu cầu"}
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
