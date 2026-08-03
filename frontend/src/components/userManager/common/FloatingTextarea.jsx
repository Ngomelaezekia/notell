import {
  forwardRef,
  useId,
  useMemo,
  useState,
} from "react";

import { motion } from "framer-motion";
import { AlertCircle } from "lucide-react";

const textareaBase =
  "peer w-full rounded-2xl border bg-white/80 backdrop-blur-xl resize-none outline-none transition-all px-4 pt-7 pb-4 text-[15px] font-medium";

const textareaNormal =
  "border-slate-200 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100";

const textareaError =
  "border-red-400 focus:ring-4 focus:ring-red-100";

const textareaDisabled =
  "opacity-50 cursor-not-allowed";

const helperId = (id) => `${id}-helper`;

const FloatingTextarea = forwardRef(function FloatingTextarea(
  {
    id,
    name,
    label,
    value = "",
    onChange,
    placeholder = "",
    rows = 5,
    helper = "",
    error = "",
    maxLength,
    disabled = false,
    required = false,
    autoComplete = "off",
    className = "",
  },
  ref
) {
  const generatedId = useId();
  const textareaId = id ?? generatedId;

  const [focused, setFocused] = useState(false);

  const floating = useMemo(
    () => focused || value.length > 0,
    [focused, value]
  );

  return (
    <div className={`w-full ${className}`}>
      <div className="relative">
        <motion.textarea
          ref={ref}
          id={textareaId}
          name={name}
          rows={rows}
          value={value}
          onChange={onChange}
          disabled={disabled}
          required={required}
          autoComplete={autoComplete}
          maxLength={maxLength}
          placeholder={floating ? placeholder : ""}
          aria-invalid={!!error}
          aria-describedby={
            helper || error
              ? helperId(textareaId)
              : undefined
          }
          initial={false}
          animate={{
            scale: focused ? 1.01 : 1,
          }}
          transition={{
            duration: 0.15,
          }}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          className={`
            ${textareaBase}
            ${error ? textareaError : textareaNormal}
            ${disabled ? textareaDisabled : ""}
          `}
        />

        <label
          htmlFor={textareaId}
          className={`absolute left-4 bg-white/90 px-1 transition-all duration-200 pointer-events-none ${
            floating
              ? "-top-2 text-[11px] font-semibold text-indigo-600"
              : "top-5 text-slate-400"
          }`}
        >
          {label}
        </label>
      </div>

      <div className="mt-2 flex items-center justify-between">
        <div
          id={helperId(textareaId)}
          className={`flex items-center gap-2 text-sm ${
            error
              ? "text-red-500"
              : "text-slate-500"
          }`}
        >
          {error && <AlertCircle size={16} />}
          <span>{error || helper}</span>
        </div>

        {maxLength && (
          <span
            className={`text-xs ${
              value.length >= maxLength
                ? "text-red-500"
                : "text-slate-400"
            }`}
          >
            {value.length}/{maxLength}
          </span>
        )}
      </div>
    </div>
  );
});

export default FloatingTextarea;