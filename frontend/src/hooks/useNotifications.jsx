import { useCallback, useEffect, useRef, useState } from "react";
import { notificationsAPI } from "../services/notification/notificationsApi";
import { getApiErrorMessage } from "../utils/api";

const POLL_INTERVAL = 15000;

export const useNotifications = (enabled = true) => {
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const fetchingRef = useRef(false);

  const fetchNotifications = useCallback(async (showLoading = true) => {
    if (!enabled || fetchingRef.current) return;
    fetchingRef.current = true;
    if (showLoading) setLoading(true);
    setError(null);
    try {
      const response = await notificationsAPI.getAll(1, 20);
      const data = response?.data ?? {};
      setNotifications(data.notifications ?? []);
      setUnreadCount(Number(data.unreadCount ?? 0));
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to load notifications."));
    } finally {
      fetchingRef.current = false;
      if (showLoading) setLoading(false);
    }
  }, [enabled]);

  useEffect(() => {
    if (!enabled) return undefined;

    const poll = () => {
      void fetchNotifications(false);
    };
    void fetchNotifications(true);

    const intervalId = window.setInterval(poll, POLL_INTERVAL);

    const handleVisibility = () => {
      if (document.visibilityState === "visible") poll();
    };

    document.addEventListener("visibilitychange", handleVisibility);
    return () => {
      window.clearInterval(intervalId);
      document.removeEventListener("visibilitychange", handleVisibility);
    };
  }, [enabled, fetchNotifications]);

  const markRead = useCallback(async (notificationId) => {
    const target = notifications.find((notification) => notification.notificationId === notificationId);
    if (!target || target.read) return;

    setError(null);
    try {
      await notificationsAPI.markRead(notificationId);
      setNotifications((current) =>
        current.map((notification) =>
          notification.notificationId === notificationId
            ? { ...notification, read: true }
            : notification
        )
      );
      setUnreadCount((current) => Math.max(0, current - 1));
    } catch (err) {
      const message = getApiErrorMessage(err, "Failed to mark notification as read.");
      setError(message);
      throw new Error(message, { cause: err });
    }
  }, [notifications]);

  const markAllRead = useCallback(async () => {
    if (unreadCount === 0) return;

    setError(null);
    try {
      await notificationsAPI.markAllRead();
      setNotifications((current) => current.map((notification) => ({ ...notification, read: true })));
      setUnreadCount(0);
    } catch (err) {
      const message = getApiErrorMessage(err, "Failed to mark notifications as read.");
      setError(message);
      throw new Error(message, { cause: err });
    }
  }, [unreadCount]);

  return {
    notifications,
    unreadCount,
    loading,
    error,
    refetch: () => fetchNotifications(true),
    markRead,
    markAllRead,
  };
};
