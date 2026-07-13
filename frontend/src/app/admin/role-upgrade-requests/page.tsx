"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { listPendingRoleUpgrades, approveRoleUpgrade, rejectRoleUpgrade } from "@/lib/api/role-upgrade";
import { useSession } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

// US1.7: Quản trị viên xem hàng chờ và duyệt/từ chối yêu cầu nâng cấp lên
// Giảng viên.
export default function RoleUpgradeRequestsPage() {
  const session = useSession();
  const queryClient = useQueryClient();

  const { data: requests, isLoading } = useQuery({
    queryKey: ["role-upgrade-requests-pending"],
    queryFn: listPendingRoleUpgrades,
    enabled: session?.role === "admin",
  });

  const approveMutation = useMutation({
    mutationFn: (id: number) => approveRoleUpgrade(id),
    onSuccess: () => {
      toast.success("Đã duyệt, tài khoản đã trở thành Giảng viên");
      queryClient.invalidateQueries({ queryKey: ["role-upgrade-requests-pending"] });
    },
    onError: () => toast.error("Duyệt yêu cầu thất bại"),
  });

  const rejectMutation = useMutation({
    mutationFn: (id: number) => rejectRoleUpgrade(id),
    onSuccess: () => {
      toast.success("Đã từ chối yêu cầu");
      queryClient.invalidateQueries({ queryKey: ["role-upgrade-requests-pending"] });
    },
    onError: () => toast.error("Từ chối yêu cầu thất bại"),
  });

  if (session && session.role !== "admin") {
    return (
      <div className="mx-auto max-w-md px-4 py-10 text-center text-sm text-muted-foreground">
        Chỉ Quản trị viên mới truy cập được trang này.
      </div>
    );
  }

  const pending = approveMutation.isPending || rejectMutation.isPending;

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight">Hàng chờ nâng cấp Giảng viên</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Danh sách học viên đang yêu cầu nâng cấp lên vai trò Giảng viên.
      </p>

      <div className="mt-6 flex flex-col gap-3">
        {isLoading && <Skeleton className="h-24 w-full" />}
        {!isLoading && requests?.length === 0 && (
          <p className="text-sm text-muted-foreground">Không có yêu cầu nào đang chờ duyệt.</p>
        )}
        {requests?.map((req) => (
          <Card key={req.id}>
            <CardHeader>
              <CardTitle className="text-base">Học viên #{req.user_id}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-muted-foreground">{req.reason}</p>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => rejectMutation.mutate(req.id)}
                  disabled={pending}
                  className="w-fit"
                >
                  Từ chối
                </Button>
                <Button
                  size="sm"
                  onClick={() => approveMutation.mutate(req.id)}
                  disabled={pending}
                  className="w-fit"
                >
                  Duyệt
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
