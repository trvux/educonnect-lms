"use client";

import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";
import { BellIcon } from "@phosphor-icons/react";

import { getUnreadCount, listMyNotifications, markNotificationRead } from "@/lib/api/notifications";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

// US6.2 — chuông thông báo + badge số chưa đọc, refetch mỗi 30s để gần như
// real-time mà không cần WebSocket (đủ cho quy mô đồ án).
export function NotificationBell() {
  const queryClient = useQueryClient();

  const { data: unreadCount } = useQuery({
    queryKey: ["notifications-unread-count"],
    queryFn: getUnreadCount,
    refetchInterval: 30_000,
  });

  const { data: notifications } = useQuery({
    queryKey: ["notifications"],
    queryFn: listMyNotifications,
    refetchInterval: 30_000,
  });

  const markReadMutation = useMutation({
    mutationFn: (id: number) => markNotificationRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["notifications-unread-count"] });
    },
  });

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button variant="ghost" size="icon" className="relative" />
        }
      >
        <BellIcon className="size-5" />
        {!!unreadCount && unreadCount > 0 && (
          <Badge
            variant="destructive"
            className="absolute -right-1 -top-1 h-4 min-w-4 justify-center rounded-full px-1 text-xs leading-none"
          >
            {unreadCount > 9 ? "9+" : unreadCount}
          </Badge>
        )}
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-80">
        <DropdownMenuGroup>
          <DropdownMenuLabel>Thông báo</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {!notifications || notifications.length === 0 ? (
            <p className="px-2 py-4 text-center text-sm text-muted-foreground">
              Không có thông báo nào.
            </p>
          ) : (
            notifications.map((n) => (
              <DropdownMenuItem
                key={n.id}
                className="flex flex-col items-start gap-0.5 whitespace-normal"
                onClick={() => {
                  if (!n.read) markReadMutation.mutate(n.id);
                }}
              >
                <div className="flex w-full items-center justify-between gap-2">
                  <span className="font-medium">{n.title}</span>
                  {!n.read && <span className="size-2 shrink-0 rounded-full bg-primary" />}
                </div>
                {n.message && <span className="text-xs text-muted-foreground">{n.message}</span>}
              </DropdownMenuItem>
            ))
          )}
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
