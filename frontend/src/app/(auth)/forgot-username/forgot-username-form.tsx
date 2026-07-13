"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";

import { forgotUsername } from "@/lib/api/auth";
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
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

// US1.8 — quên tên đăng nhập (thực chất là quên email đã đăng ký), tra qua
// số điện thoại và trả về email đã che (không lộ toàn bộ, chống dò tài khoản).
export function ForgotUsernameForm({ className, ...props }: React.ComponentProps<"div">) {
  const [phone, setPhone] = useState("");

  const mutation = useMutation({
    mutationFn: () => forgotUsername(phone),
  });

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Quên tên đăng nhập</CardTitle>
          <CardDescription>Nhập số điện thoại đã đăng ký để tìm lại email đăng nhập</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            className="flex flex-col gap-6"
            onSubmit={(e) => {
              e.preventDefault();
              mutation.mutate();
            }}
          >
            <div className="flex flex-col gap-2">
              <Label>Số điện thoại</Label>
              <Input
                placeholder="0987654321"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                required
              />
            </div>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? "Đang tìm..." : "Tìm email"}
            </Button>

            {mutation.isSuccess && (
              <Alert>
                <AlertTitle>Đã tìm thấy tài khoản</AlertTitle>
                <AlertDescription>Email đăng nhập của bạn là: {mutation.data.masked_email}</AlertDescription>
              </Alert>
            )}
            {mutation.isError && (
              <Alert variant="destructive">
                <AlertTitle>Không tìm thấy</AlertTitle>
                <AlertDescription>
                  Không có tài khoản nào ứng với số điện thoại này.
                </AlertDescription>
              </Alert>
            )}
          </form>
          <p className="mt-6 text-center text-sm text-muted-foreground">
            Nhớ email rồi?{" "}
            <Link href="/login" className="underline underline-offset-4">
              Đăng nhập
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
