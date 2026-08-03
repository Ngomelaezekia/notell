import { forwardRef, useId, useMemo, useState,} from "react";
import { motion } from "framer-motion";
import {Eye, EyeOff,  AlertCircle, } from "lucide-react";
import { buttonAnimation } from "../animation";

const baseInput =
  "peer w-full h-14 rounded-2xl bg-white/80 backdrop-blur-xl border outline-none transition-all text-[15px] font-medium pt-5 pb-2 pl-12";

const normalState =
  "border-slate-200 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100";

const errorState =
  "border-red-400 focus:ring-4 focus:ring-red-100";

const disabledState =
  "opacity-50 cursor-not-allowed";

const helperId = (id) => `${id}-helper`;

const FloatingInput = forwardRef(function FloatingInput(
  {
    id,
    name,
    label,
    icon: Icon,
    type = "text",
    value = "",
    onChange,
    placeholder = "",
    helper = "",
    error = "",
    disabled = false,
    required = false,
    autoComplete = "off",
    className = "",
  },
  ref
) {
  const generatedId = useId();
  const inputId = id ?? generatedId;

  const [focused, setFocused] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const isPassword = useMemo(
    () => type === "password",
    [type]
  );

  const inputType = useMemo(() => {
    if (!isPassword) return type;
    return showPassword ? "text" : "password";
  }, [isPassword, showPassword, type]);

  const floating = useMemo(
    () => focused || value.length > 0,
    [focused, value]
  );

  return (
    <div className={`w-full ${className}`}>
      <div className="relative">
        <motion.input
          ref={ref}
          id={inputId}
          name={name}
          type={inputType}
          value={value}
          required={required}
          disabled={disabled}
          autoComplete={autoComplete}
          placeholder={floating ? placeholder : ""}
          aria-invalid={!!error}
          aria-describedby={
            helper || error
              ? helperId(inputId)
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
          onChange={onChange}
          className={`
            ${baseInput}
            ${isPassword ? "pr-14" : "pr-4"}
            ${error ? errorState : normalState}
            ${disabled ? disabledState : ""}
          `}
        />

        {Icon && (
          <Icon
            size={19}
            className={`absolute left-4 top-1/2 -translate-y-1/2 transition-colors ${
              focused
                ? "text-indigo-600"
                : "text-slate-400"
            }`}
          />
        )}

        <label
          htmlFor={inputId}
          className={`absolute left-12 bg-white/90 px-1 transition-all duration-200 pointer-events-none ${
            floating
              ? "-top-2 text-[11px] font-semibold text-indigo-600"
              : "top-4 text-slate-400"
          }`}
        >
          {label}
        </label>

        {isPassword && (
          <motion.button
            type="button"
            {...buttonAnimation}
            tabIndex={-1}
            aria-label={
              showPassword
                ? "Hide password"
                : "Show password"
            }
            onClick={() =>
              setShowPassword((v) => !v)
            }
            className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 transition-colors hover:text-indigo-600"
          >
            {showPassword ? (
              <EyeOff size={19} />
            ) : (
              <Eye size={19} />
            )}
          </motion.button>
        )}
      </div>

      {(helper || error) && (
        <div
          id={helperId(inputId)}
          className={`mt-2 flex items-center gap-2 text-sm ${
            error
              ? "text-red-500"
              : "text-slate-500"
          }`}
        >
          {error && <AlertCircle size={16} />}
          <span>{error || helper}</span>
        </div>
      )}
    </div>
  );
});

export default FloatingInput;