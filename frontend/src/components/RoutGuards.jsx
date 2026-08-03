import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";


function LoadingSpinner() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950">
      <div
        className="
          h-9
          w-9
          animate-spin
          rounded-full
          border-4
          border-zinc-700
          border-t-lime-400
        "
      />
    </div>
  );
}




export function ProtectedRoute() {

  const {
    authenticated,
    loading,
  } = useAuth();


  const location = useLocation();



  if (loading) {
    return <LoadingSpinner />;
  }



  if (!authenticated) {

    return (
      <Navigate
        to="/auth"
        replace
        state={{
          from: location,
        }}
      />
    );

  }



  return <Outlet />;

}







export function PublicOnlyRoute() {

  const {
    authenticated,
    loading,
  } = useAuth();



  if (loading) {
    return <LoadingSpinner />;
  }




  if (authenticated) {

    return (
      <Navigate
        to="/"
        replace
      />
    );

  }



  return <Outlet />;

}