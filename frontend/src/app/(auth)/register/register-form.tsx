"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import Link from "next/link";

import { apiClient } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

// US1.1/US1.7: đăng ký công khai luôn tạo tài khoản Học viên. Muốn trở
// thành Giảng viên phải gửi yêu cầu trong trang Hồ sơ để quản trị viên duyệt
// (backend cũng ép Student bất kể client gửi gì — xem AllowRoleOnRegister).
const registerSchema = z.object({
  fullName: z.string().min(2, "Họ tên phải có ít nhất 2 ký tự"),
  email: z.string().min(1, "Vui lòng nhập email").email("Email không hợp lệ"),
  password: z.string().min(8, "Mật khẩu phải có ít nhất 8 ký tự"),
});

type RegisterValues = z.infer<typeof registerSchema>;

export function RegisterForm({ className, ...props }: React.ComponentProps<"div">) {
  const router = useRouter();
  const form = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { fullName: "", email: "", password: "" },
  });

  const mutation = useMutation({
    mutationFn: async (values: RegisterValues) => {
      const res = await apiClient.post("/auth/register", {
        full_name: values.fullName,
        email: values.email,
        password: values.password,
      });
      return res.data;
    },
    onSuccess: (_data, values) => {
      toast.success("Đăng ký thành công, kiểm tra email để lấy mã OTP xác thực");
      router.push(`/verify-email?email=${encodeURIComponent(values.email)}`);
    },
    onError: (error: unknown) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Đăng ký thất bại, vui lòng thử lại";
      toast.error(message);
    },
  });

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Tạo tài khoản EduConnect LMS</CardTitle>
          <CardDescription>
            Tài khoản mới là Học viên; muốn trở thành Giảng viên, hãy gửi yêu cầu trong Hồ sơ sau khi
            đăng nhập
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form
              className="flex flex-col gap-6"
              onSubmit={form.handleSubmit((values) => mutation.mutate(values))}
            >
              <FormField
                control={form.control}
                name="fullName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Họ và tên</FormLabel>
                    <FormControl>
                      <Input placeholder="Nguyễn Văn A" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input type="email" placeholder="ban@vlu.edu.vn" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Mật khẩu</FormLabel>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className="flex flex-col gap-3">
                <Button type="submit" disabled={mutation.isPending}>
                  {mutation.isPending ? "Đang đăng ký..." : "Đăng ký"}
                </Button>
                <p className="text-center text-sm text-muted-foreground">
                  Đã có tài khoản?{" "}
                  <Link href="/login" className="underline underline-offset-4">
                    Đăng nhập
                  </Link>
                </p>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
