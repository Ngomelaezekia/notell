import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";

import { useNavigate } from "react-router-dom";

import API, { API_BASE_URL, getApiErrorMessage } from "../utils/api";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const navigate = useNavigate();

  const fetchCurrentUser = useCallback(async () => {
    try {
      const response = await API.get("/auth/me");
      const currentUser =
        response.data?.data?.user ||
        response.data?.user ||
        response.data;

      if (!currentUser?.id) {
        throw new Error("Invalid current-user response");
      }

      setUser(currentUser);
      setError(null);
      return currentUser;
    } catch (err) {
      if (err?.response?.status === 401) {
        setUser(null);
        return null;
      }

      setError(getApiErrorMessage(err, "Unable to verify your session."));
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const handleSessionExpired = () => {
      setUser(null);
      setError("Your session has expired. Please sign in again.");
      navigate("/auth", { replace: true });
    };

    window.addEventListener("notell:session-expired", handleSessionExpired);
    return () => {
      window.removeEventListener("notell:session-expired", handleSessionExpired);
    };
  }, [navigate]);

  useEffect(() => {
    const init = async () => {
      const params = new URLSearchParams(window.location.search);
      const oauthError = params.get("error");

      if (oauthError) {
        setError(`OAuth Error: ${oauthError}`);
        window.history.replaceState({}, document.title, window.location.pathname);
      }

      await fetchCurrentUser();
    };

    void init();
  }, [fetchCurrentUser]);

  const loginWithGoogle = () => {
    window.location.href = `${API_BASE_URL}/auth/google`;
  };

  const login = async (credentials, redirectTo = "/") => {
    setError(null);

    try {
      const response = await API.post("/auth/login", credentials);
      const currentUser = await fetchCurrentUser();
      if (!currentUser) {
        throw new Error("Login succeeded, but the session could not be verified.");
      }
      navigate(redirectTo || "/", { replace: true });
      return response.data;
    } catch (err) {
      const message = getApiErrorMessage(err, "Login failed");
      setError(message);
      throw new Error(message, { cause: err });
    }
  };

  const register = async (formData, redirectTo = "/") => {
    setError(null);

    try {
      const response = await API.post("/auth/register", formData);
      const currentUser = await fetchCurrentUser();
      if (!currentUser) {
        throw new Error("Registration succeeded, but the session could not be verified.");
      }
      navigate(redirectTo || "/", { replace: true });
      return response.data;
    } catch (err) {
      const message = getApiErrorMessage(err, "Registration failed");
      setError(message);
      throw new Error(message, { cause: err });
    }
  };

  const logout = async () => {
    try {
      await API.post("/auth/logout");
    } catch (err) {
      setError(getApiErrorMessage(err, "Logout failed"));
    } finally {
      setUser(null);
      navigate("/auth", { replace: true });
    }
  };

  const updateUser = useCallback((updatedFields) => {
    setUser((prev) =>
      prev
        ? {
            ...prev,
            ...updatedFields,
          }
        : null
    );
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        authenticated: Boolean(user),
        loading,
        error,
        setError,
        login,
        register,
        loginWithGoogle,
        logout,
        fetchCurrentUser,
        updateUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);

  if (!context) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return context;
}
