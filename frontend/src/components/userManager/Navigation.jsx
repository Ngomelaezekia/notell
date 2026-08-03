import { useCallback } from "react";
import { motion } from "framer-motion";
import {
  ArrowLeft,
  ArrowRight,
  Save,
} from "lucide-react";

import GradientButton from "./common/GradientButton";
import { fadeUp } from "./animation";

export default function Navigation({
  currentStep,
  totalSteps,

  setCurrentStep,
  onSave,

  loading = false,
  saving = false,

  previousLabel = "Previous",
  nextLabel = "Continue",
  submitLabel = "Save Changes",

  showProgress = true,
  showPrevious = true,
  showNext = true,

  className = "",
}) {
  const isFirst = currentStep === 1;
  const isLast = currentStep === totalSteps;
  const isBusy = loading || saving;

  const handlePrevious = useCallback(() => {
    if (!isFirst) {
      setCurrentStep((prev) => prev - 1);
    }
  }, [isFirst, setCurrentStep]);

  const handleNext = useCallback(() => {
    if (isLast) {
      onSave?.();
      return;
    }

    setCurrentStep((prev) => prev + 1);
  }, [isLast, onSave, setCurrentStep]);

  return (
    <motion.footer
      variants={fadeUp}
      initial="hidden"
      animate="visible"
      className={`mt-8 flex flex-col gap-4 md:flex-row md:items-center md:justify-between ${className}`}
    >
      {/* Left */}
      <div className="flex items-center gap-3">
        {showPrevious && (
          <GradientButton
            variant="secondary"
            icon={ArrowLeft}
            disabled={isFirst || isBusy}
            onClick={handlePrevious}
          >
            {previousLabel}
          </GradientButton>
        )}
      </div>

      {/* Center */}
      {showProgress && (
        <div className="order-first text-center text-sm font-medium text-slate-500 md:order-none">
          Step{" "}
          <span className="font-semibold text-slate-800">
            {currentStep}
          </span>{" "}
          of{" "}
          <span className="font-semibold text-slate-800">
            {totalSteps}
          </span>
        </div>
      )}

      {/* Right */}
      <div className="flex justify-end">
        {showNext &&
          (isLast ? (
            <GradientButton
              variant="success"
              icon={Save}
              loading={isBusy}
              onClick={handleNext}
            >
              {submitLabel}
            </GradientButton>
          ) : (
            <GradientButton
              icon={ArrowRight}
              disabled={isBusy}
              onClick={handleNext}
            >
              {nextLabel}
            </GradientButton>
          ))}
      </div>
    </motion.footer>
  );
}