// OAuth now uses the HttpOnly auth_token cookie set by the backend.
// There is no client-side token to extract or persist after the OAuth callback.
export function handleOAuthRedirect() {
  const params = new URLSearchParams(window.location.search);
  if (params.has("token")) {
    params.delete("token");
    const query = params.toString();
    const nextURL = `${window.location.pathname}${query ? `?${query}` : ""}${window.location.hash}`;
    window.history.replaceState({}, document.title, nextURL);
  }
}
