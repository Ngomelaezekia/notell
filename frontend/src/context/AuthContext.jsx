import { createContext, useContext, useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import API, { API_BASE_URL, getApiErrorMessage } from "../utils/api";
import { paymentAPI } from "../services/payment/paymentApi";

const AuthContext = createContext(null);
const FEED_CACHE_PREFIX = "notell:feed:v2:";
const clearFeedCache = () => { try { for (let index = sessionStorage.length - 1; index >= 0; index -= 1) { const key = sessionStorage.key(index); if (key?.startsWith(FEED_CACHE_PREFIX)) sessionStorage.removeItem(key); } } catch {} };
const pendingGiftCode = () => { try { const code = new URLSearchParams(window.location.search).get("gift"); return code ? code.trim().toUpperCase().slice(0, 6) : ""; } catch { return ""; } };

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [googleLoading, setGoogleLoading] = useState(false);
  const navigate = useNavigate();

  const fetchCurrentUser = useCallback(async () => {
    try {
      const response = await API.get("/auth/me");
      const currentUser = response.data?.data?.user || response.data?.user || response.data;
      if (!currentUser?.id) throw new Error("Invalid current-user response");
      setUser(currentUser); setError(null); return currentUser;
    } catch (err) {
      if (err?.response?.status === 401) { clearFeedCache(); setUser(null); return null; }
      setError(getApiErrorMessage(err, "Unable to verify your session.")); return null;
    } finally { setLoading(false); }
  }, []);

  const claimReferralGift = useCallback(async (code) => {
    const normalized = String(code || "").trim().toUpperCase();
    if (!normalized) return false;
    try { await paymentAPI.claimGift(normalized); return true; } catch { return false; }
  }, []);

  useEffect(() => {
    const handleSessionExpired = () => { clearFeedCache(); setUser(null); setError("Your session has expired. Please sign in again."); navigate("/auth", { replace: true }); };
    window.addEventListener("notell:session-expired", handleSessionExpired);
    return () => window.removeEventListener("notell:session-expired", handleSessionExpired);
  }, [navigate]);

  useEffect(() => {
    const init = async () => {
      const params = new URLSearchParams(window.location.search);
      const oauthError = params.get("error");
      if (oauthError) {
        const messages = {
          google_account_not_registered: "No Notell account is linked to this Google account. Use Sign Up with Google first.",
          google_account_already_registered: "This Google account is already registered. Use Login with Google instead.",
          invalid_state: "Google authentication expired or could not be verified. Please try again.",
          no_code: "Google did not return an authorization code. Please try again.",
          exchange_failed: "Google sign-in could not be completed. Please try again.",
          user_fetch_failed: "We could not read your Google account details. Please try again.",
          google_login_failed: "We could not sign you in with Google. Please try again.",
          google_signup_failed: "We could not create your Notell account with Google. Please try again.",
          session_failed: "Google sign-in completed, but your Notell session could not be created. Please try again.",
        };
        setError(messages[oauthError] || `OAuth Error: ${oauthError}`);
        window.history.replaceState({}, document.title, window.location.pathname);
      }
      const currentUser = await fetchCurrentUser();
      const gift = pendingGiftCode();
      if (currentUser && gift) {
        await claimReferralGift(gift);
        const cleanPath = `${window.location.pathname}${window.location.hash || ""}`;
        window.history.replaceState({}, document.title, cleanPath || "/");
      }
    };
    void init();
  }, [claimReferralGift, fetchCurrentUser]);

  const loginWithGoogle = (mode = "login", redirectTo = "/") => {
    const selectedMode = mode === "signup" ? "signup" : "login";
    const gift = pendingGiftCode();
    const params = new URLSearchParams({ mode: selectedMode, returnTo: redirectTo || "/" });
    if (gift) params.set("gift", gift);
    setError(null);
    setGoogleLoading(true);
    window.location.assign(API_BASE_URL + "/auth/google?" + params.toString());
  };

  const login = async (credentials, redirectTo = "/") => {
    setError(null);
    try { const response = await API.post("/auth/login", credentials); const currentUser = await fetchCurrentUser(); if (!currentUser) throw new Error("Login succeeded, but the session could not be verified."); navigate(redirectTo || "/", { replace: true }); return response.data; }
    catch (err) { const message = getApiErrorMessage(err, "Login failed"); setError(message); throw new Error(message, { cause: err }); }
  };

  const register = async (formData, redirectTo = "/") => {
    setError(null);
    try { const response = await API.post("/auth/register", formData); const currentUser = await fetchCurrentUser(); if (!currentUser) throw new Error("Registration succeeded, but the session could not be verified."); await claimReferralGift(formData?.giftCode || pendingGiftCode()); navigate(redirectTo || "/", { replace: true }); return response.data; }
    catch (err) { const message = getApiErrorMessage(err, "Registration failed"); setError(message); throw new Error(message, { cause: err }); }
  };

  const logout = async () => { try { await API.post("/auth/logout"); } catch (err) { setError(getApiErrorMessage(err, "Logout failed")); } finally { clearFeedCache(); setUser(null); navigate("/auth", { replace: true }); } };
  const updateUser = useCallback((updatedFields) => setUser((prev) => prev ? { ...prev, ...updatedFields } : null), []);

  return <AuthContext.Provider value={{ user, authenticated: Boolean(user), loading, error, setError, login, register, loginWithGoogle, googleLoading, logout, fetchCurrentUser, updateUser }}>{children}</AuthContext.Provider>;
}

export function useAuth() { const context = useContext(AuthContext); if (!context) throw new Error("useAuth must be used inside AuthProvider"); return context; }
