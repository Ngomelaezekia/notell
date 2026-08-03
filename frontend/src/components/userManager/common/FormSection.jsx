import GlassCard from "./GlassCard";
import SectionTitle from "./SectionTitle";

export default function FormSection({
  title,
  description,
  icon,
  badge,
  children,
  className = "",
  contentClassName = "",
}) {
  return (
    <GlassCard className={`space-y-6 ${className}`}>
      <SectionTitle
        title={title}
        description={description}
        icon={icon}
        badge={badge}
      />

      <div className={`space-y-5 ${contentClassName}`}>
        {children}
      </div>
    </GlassCard>
  );
}