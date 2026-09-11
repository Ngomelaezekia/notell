import { lazy, Suspense } from "react";
import { createBrowserRouter, Outlet } from "react-router-dom";
import { AuthProvider } from "../context/AuthContext";
import { ProtectedRoute, PublicOnlyRoute } from "../components/RoutGuards";
import AppLayout from "../layout/AppLayout";

const AuthPage = lazy(() => import("../pages/AuthPage"));
const Posts = lazy(() => import("../pages/PostPage"));
const PostDetailPage = lazy(() => import("../pages/PostDetailPage"));
const VideoFeedView = lazy(() => import("../pages/VideoFeedView"));
const CreatePost = lazy(() => import("../components/CreatePost").then((module) => ({ default: module.CreatePost })));
const UserManage = lazy(() => import("../components/userManager/UserManager"));
const UserPage = lazy(() => import("../pages/UserPage"));
const CurrentUserProfile = lazy(() => import("../pages/CurrentUserProfile"));
const FollowersPage = lazy(() => import("../pages/RelationshipListPage").then((module) => ({ default: module.FollowersPage })));
const FollowingPage = lazy(() => import("../pages/RelationshipListPage").then((module) => ({ default: module.FollowingPage })));
const SearchPage = lazy(() => import("../pages/SearchPage"));
const NotificationsPage = lazy(() => import("../pages/NotificationsPage"));
const SettingsPage = lazy(() => import("../pages/SettingsPage"));
const SettingsDetailPage = lazy(() => import("../pages/SettingsDetailPage"));

const PageFallback = () => (
  <div className="flex min-h-[40vh] items-center justify-center px-4 text-sm text-neutral-500">Loading…</div>
);

const withSuspense = (element) => <Suspense fallback={<PageFallback />}>{element}</Suspense>;

const RootLayout = () => (
  <AuthProvider>
    <Outlet />
  </AuthProvider>
);

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      {
        element: <ProtectedRoute />,
        children: [
          {
            element: <AppLayout />,
            children: [
              { path: "/", element: withSuspense(<Posts />) },
              { path: "create-post", element: withSuspense(<CreatePost />) },
              { path: "video-feed/:id", element: withSuspense(<VideoFeedView />) },
              { path: "user", element: withSuspense(<CurrentUserProfile />) },
              { path: "profile", element: withSuspense(<UserManage />) },
              { path: "notifications", element: withSuspense(<NotificationsPage />) },
              { path: "settings", element: withSuspense(<SettingsPage />) },
              { path: "settings/privacy", element: withSuspense(<SettingsDetailPage section="privacy" />) },
              { path: "settings/content", element: withSuspense(<SettingsDetailPage section="content" />) },
              { path: "settings/notifications", element: withSuspense(<SettingsDetailPage section="notifications" />) },
              { path: "settings/ads", element: withSuspense(<SettingsDetailPage section="ads" />) },
              { path: "settings/history", element: withSuspense(<SettingsDetailPage section="history" />) },
              { path: "settings/downloads", element: withSuspense(<SettingsDetailPage section="downloads" />) },
              { path: "settings/storage", element: withSuspense(<SettingsDetailPage section="storage" />) },
              { path: "settings/about", element: withSuspense(<SettingsDetailPage section="about" />) },
              { path: "settings/terms", element: withSuspense(<SettingsDetailPage section="terms" />) },
              { path: "settings/more", element: withSuspense(<SettingsDetailPage section="more" />) },
              { path: "settings/help", element: withSuspense(<SettingsDetailPage section="help" />) },
              { path: "settings/email", element: withSuspense(<SettingsDetailPage section="email" />) },
              { path: "search", element: withSuspense(<SearchPage />) },
              { path: "posts/:id", element: withSuspense(<PostDetailPage />) },
              { path: "users/:id", element: withSuspense(<UserPage />) },
              { path: "users/:id/followers", element: withSuspense(<FollowersPage />) },
              { path: "users/:id/following", element: withSuspense(<FollowingPage />) },
            ],
          },
        ],
      },
      {
        element: <PublicOnlyRoute />,
        children: [{ path: "/auth", element: withSuspense(<AuthPage />) }],
      },
    ],
  },
]);

export default router;
