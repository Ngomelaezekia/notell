import { createBrowserRouter, Outlet } from "react-router-dom";
import { AuthProvider } from "../context/AuthContext";
import { ProtectedRoute, PublicOnlyRoute } from "../components/RoutGuards";
import AuthPage from "../pages/AuthPage";
import Posts from "../pages/PostPage";
import PostDetailPage from "../pages/PostDetailPage";
import { CreatePost } from "../components/CreatePost";
import UserManage from "../components/userManager/UserManager";
import UserPage from "../pages/UserPage";
import { FollowersPage, FollowingPage } from "../pages/RelationshipListPage";
import SearchPage from "../pages/SearchPage";
import SettingsPage from "../pages/SettingsPage";
import SettingsDetailPage from "../pages/SettingsDetailPage";
import AppLayout from "../layout/AppLayout";

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
              { path: "/", element: <Posts /> },
              { path: "create-post", element: <CreatePost /> },
              { path: "profile", element: <UserManage /> },
              { path: "settings", element: <SettingsPage /> },
              { path: "settings/privacy", element: <SettingsDetailPage section="privacy" /> },
              { path: "settings/content", element: <SettingsDetailPage section="content" /> },
              { path: "settings/notifications", element: <SettingsDetailPage section="notifications" /> },
              { path: "settings/ads", element: <SettingsDetailPage section="ads" /> },
              { path: "settings/history", element: <SettingsDetailPage section="history" /> },
              { path: "settings/downloads", element: <SettingsDetailPage section="downloads" /> },
              { path: "settings/storage", element: <SettingsDetailPage section="storage" /> },
              { path: "settings/about", element: <SettingsDetailPage section="about" /> },
              { path: "settings/terms", element: <SettingsDetailPage section="terms" /> },
              { path: "settings/more", element: <SettingsDetailPage section="more" /> },
              { path: "settings/help", element: <SettingsDetailPage section="help" /> },
              { path: "settings/email", element: <SettingsDetailPage section="email" /> },
              { path: "search", element: <SearchPage /> },
              { path: "posts/:id", element: <PostDetailPage /> },
              { path: "users/:id", element: <UserPage /> },
              { path: "users/:id/followers", element: <FollowersPage /> },
              { path: "users/:id/following", element: <FollowingPage /> },
            ],
          },
        ],
      },
      {
        element: <PublicOnlyRoute />,
        children: [{ path: "/auth", element: <AuthPage /> }],
      },
    ],
  },
]);

export default router;
