import { Navigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function CurrentUserProfile() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex min-h-[60vh] w-full items-center justify-center bg-neutral-950 text-sm text-white/50">
        Loading your profile...
      </div>
    );
  }

  if (!user?.id) {
    return <Navigate to="/auth" replace />;
  }

  return <Navigate to={`/users/${user.id}`} replace />;
}
