export default function OAuthDivider({ text = "or continue with" }) {
  return (
    <div className="flex items-center gap-4 my-6">
      <div className="h-px flex-1 bg-white/20" />
      <span className="text-sm text-white/70">{text}</span>
      <div className="h-px flex-1 bg-white/20" />
    </div>
  );
}