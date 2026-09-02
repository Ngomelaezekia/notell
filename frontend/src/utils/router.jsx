import { createBrowserRouter, Outlet } from "react-router-dom";
import { AuthProvider } from "../context/AuthContext";
import { ProtectedRoute, PublicOnlyRoute } from "../components/RoutGuards";
import AuthPage from "../pages/AuthPage";
import Posts from "../pages/PostPage";
import { CreatePost } from "../components/CreatePost";
import UserManage from "../components/userManager/UserManager";
import UserPage from "../pages/UserPage";
import { FollowersPage, FollowingPage } from "../pages/RelationshipListPage";
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
