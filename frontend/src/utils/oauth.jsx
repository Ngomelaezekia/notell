import { setToken } from "./api";

export function handleOAuthRedirect() {
  const params = new URLSearchParams(window.location.search);
  const token = params.get("token");
  if (token) {
    setToken(token);
    window.history.replaceState({}, document.title, window.location.pathname);
  }
}