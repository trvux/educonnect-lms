"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import Link from "next/link";

import { verifyEmail, resendVerification } from "@/lib/api/auth";
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
import { Label } from "@/components/ui/label";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";

// US1.9 — xác thực OTP gửi lúc đăng ký trước khi tài khoản được active để
// đăng nhập. Prefill email từ query string (?email=...) khi tới từ trang
// đăng ký hoặc từ cảnh báo "chưa xác thực email" lúc đăng nhập.
export function VerifyEmailForm({ className, ...props }: React.ComponentProps<"div">) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [email, setEmail] = useState(searchParams.get("email") ?? "");
  const [otp, setOtp] = useState("");

  const verifyMutation = useMutation({
    mutationFn: () => verifyEmail({ email, otp }),
    onSuccess: () => {
      toast.success("Xác thực thành công, mời đăng nhập");
      router.push("/login");
    },
    onError: (error: unknown) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Mã OTP không đúng hoặc đã hết hạn";
      toast.error(message);
    },
  });

  const resendMutation = useMutation({
    mutationFn: () => resendVerification(email),
    onSuccess: () => toast.success("Nếu email chưa xác thực, mã OTP mới đã được gửi"),
    onError: () => toast.error("Gửi lại mã thất bại, vui lòng thử lại"),
  });

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Xác thực email</CardTitle>
          <CardDescription>Nhập mã OTP đã gửi tới email đăng ký để kích hoạt tài khoản</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            className="flex flex-col gap-6"
            onSubmit={(e) => {
              e.preventDefault();
              verifyMutation.mutate();
            }}
          >
            <div className="flex flex-col gap-2">
              <Label>Email</Label>
              <Input
                type="email"
                placeholder="ban@vlu.edu.vn"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>Mã OTP</Label>
              <InputOTP maxLength={6} value={otp} onChange={setOtp}>
                <InputOTPGroup>
                  {[0, 1, 2, 3, 4, 5].map((i) => (
                    <InputOTPSlot key={i} index={i} />
                  ))}
                </InputOTPGroup>
              </InputOTP>
            </div>
            <div className="flex flex-col gap-3">
              <Button type="submit" disabled={otp.length !== 6 || verifyMutation.isPending}>
                {verifyMutation.isPending ? "Đang xác thực..." : "Xác thực"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => resendMutation.mutate()}
                disabled={!email || resendMutation.isPending}
              >
                Gửi lại mã OTP
              </Button>
            </div>
          </form>
          <p className="mt-6 text-center text-sm text-muted-foreground">
            Đã xác thực rồi?{" "}
            <Link href="/login" className="underline underline-offset-4">
              Đăng nhập
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
