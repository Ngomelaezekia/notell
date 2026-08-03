import { useMemo } from "react";
import { motion } from "framer-motion";
import { Check } from "lucide-react";

import {
  buttonAnimation,
  progressAnimation,
} from "./animation";

export default function Stepper({
  steps = [],

  currentStep,

  setCurrentStep,

  clickable = true,

  showLabels = true,

  className = "",
}) {
  const totalSteps = steps.length;

  const progress = useMemo(() => {
    if (totalSteps <= 1) return 0;

    return (
      ((currentStep - 1) / (totalSteps - 1)) *
      100
    );
  }, [currentStep, totalSteps]);

  return (
    <div
      className={`
        rounded-3xl
        border
        border-white/60
        bg-white/70
        p-6
        shadow-lg
        backdrop-blur-xl
        ${className}
      `}
    >
      <div className="relative overflow-x-auto">
        <div className="relative flex min-w-max items-start justify-between gap-8">
          {/* Track */}

          <div
            className="
              absolute
              left-0
              right-0
              top-6
              h-1
              rounded-full
              bg-slate-200
            "
          />

          {/* Progress */}

          <motion.div
            className="
              absolute
              left-0
              top-6
              h-1
              rounded-full
              bg-gradient-to-r
              from-indigo-500
              via-blue-500
              to-cyan-400
            "
            initial="initial"
            animate="animate"
            custom={`${progress}%`}
            variants={progressAnimation}
          />

          {steps.map((step) => {
            const active =
              currentStep === step.id;

            const completed =
              currentStep > step.id;

            const Icon = step.icon;

            return (
              <motion.button
                key={step.id}
                {...buttonAnimation}
                type="button"
                disabled={!clickable}
                aria-current={
                  active ? "step" : undefined
                }
                aria-label={step.title}
                onClick={() =>
                  clickable &&
                  setCurrentStep(step.id)
                }
                className="
                  relative
                  z-10
                  flex
                  flex-col
                  items-center
                  gap-3
                  disabled:cursor-default
                "
              >
                <motion.div
                  animate={{
                    scale: active ? 1.08 : 1,
                  }}
                  transition={{
                    duration: 0.2,
                  }}
                  className={`
                    flex
                    h-12
                    w-12
                    items-center
                    justify-center
                    rounded-full
                    border-2
                    transition-all

                    ${
                      completed
                        ? `
                          border-emerald-500
                          bg-emerald-500
                          text-white
                        `
                        : active
                        ? `
                          border-indigo-500
                          bg-gradient-to-br
                          from-indigo-500
                          via-blue-500
                          to-cyan-400
                          text-white
                          shadow-lg
                          shadow-indigo-500/30
                        `
                        : `
                          border-slate-300
                          bg-white
                          text-slate-400
                        `
                    }
                  `}
                >
                  {completed ? (
                    <Check size={20} />
                  ) : (
                    Icon && <Icon size={20} />
                  )}
                </motion.div>

                {showLabels && (
                  <div className="text-center">
                    <p
                      className={`
                        text-sm
                        font-semibold

                        ${
                          active
                            ? "text-indigo-600"
                            : completed
                            ? "text-emerald-600"
                            : "text-slate-500"
                        }
                      `}
                    >
                      {step.title}
                    </p>

                    {step.description && (
                      <p className="mt-1 text-xs text-slate-400">
                        {step.description}
                      </p>
                    )}
                  </div>
                )}
              </motion.button>
            );
          })}
        </div>
      </div>
    </div>
  );
}