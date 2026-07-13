"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import Link from "next/link";

import { forgotPassword, resetPassword } from "@/lib/api/auth";
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

// US1.6 — quên mật khẩu qua OTP 6 số gửi email. Bước 1 nhập email để nhận
// OTP, bước 2 nhập OTP + mật khẩu mới.
export function ForgotPasswordForm({ className, ...props }: React.ComponentProps<"div">) {
  const router = useRouter();
  const [step, setStep] = useState<"email" | "reset">("email");
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [newPassword, setNewPassword] = useState("");

  const sendOtpMutation = useMutation({
    mutationFn: () => forgotPassword(email),
    onSuccess: () => {
      toast.success("Nếu email tồn tại, mã OTP đã được gửi. Vui lòng kiểm tra hộp thư.");
      setStep("reset");
    },
    onError: () => toast.error("Gửi OTP thất bại, vui lòng thử lại"),
  });

  const resetMutation = useMutation({
    mutationFn: () => resetPassword({ email, otp, new_password: newPassword }),
    onSuccess: () => {
      toast.success("Đã đặt lại mật khẩu, mời đăng nhập");
      router.push("/login");
    },
    onError: (error: unknown) => {
      const message =
        (error as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Mã OTP không đúng hoặc đã hết hạn";
      toast.error(message);
    },
  });

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Quên mật khẩu</CardTitle>
          <CardDescription>
            {step === "email"
              ? "Nhập email đã đăng ký để nhận mã OTP"
              : `Nhập mã OTP đã gửi tới ${email} và mật khẩu mới`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {step === "email" ? (
            <form
              className="flex flex-col gap-6"
              onSubmit={(e) => {
                e.preventDefault();
                sendOtpMutation.mutate();
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
              <Button type="submit" disabled={sendOtpMutation.isPending}>
                {sendOtpMutation.isPending ? "Đang gửi..." : "Gửi mã OTP"}
              </Button>
            </form>
          ) : (
            <form
              className="flex flex-col gap-6"
              onSubmit={(e) => {
                e.preventDefault();
                resetMutation.mutate();
              }}
            >
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
              <div className="flex flex-col gap-2">
                <Label>Mật khẩu mới</Label>
                <Input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  minLength={8}
                  required
                />
              </div>
              <div className="flex flex-col gap-3">
                <Button type="submit" disabled={otp.length !== 6 || resetMutation.isPending}>
                  {resetMutation.isPending ? "Đang đặt lại..." : "Đặt lại mật khẩu"}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => sendOtpMutation.mutate()}
                  disabled={sendOtpMutation.isPending}
                >
                  Gửi lại mã OTP
                </Button>
              </div>
            </form>
          )}
          <p className="mt-6 text-center text-sm text-muted-foreground">
            Nhớ mật khẩu?{" "}
            <Link href="/login" className="underline underline-offset-4">
              Đăng nhập
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
