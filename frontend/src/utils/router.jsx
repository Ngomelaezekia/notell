import { createBrowserRouter, Outlet } from "react-router-dom";
import { AuthProvider } from "../context/AuthContext";
import { ProtectedRoute, PublicOnlyRoute } from "../components/RoutGuards";
import AuthPage from "../pages/AuthPage";
import Posts from "../pages/PostPage"
import { CreatePost } from "../components/CreatePost";
import UserManage from "../components/userManager/UserManager";
import AppLayout from "../layout/AppLayout";

const RootLayout = () => {
  return (
    <AuthProvider>
      <Outlet />
    </AuthProvider>
  );
};

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      // Protected Routes
      {
        element: <ProtectedRoute />,
        children: [
          {
            element: <AppLayout />,
            children: [
          
          {
            path: "/",
            element: <Posts />,
          },
          {
            path:'create-post',
            element: <CreatePost />
          },
          {
            path: "profile",
            element: <UserManage />
          },
        ],
      },
        ],
      },
      // Public Only Routes (Redirects to '/' if authenticated)
      {
        element: <PublicOnlyRoute />,
        children: [
          {
            path: "/auth",
            element: <AuthPage />,
          },
        ],
      },
    ],
  },
]);

export default router;