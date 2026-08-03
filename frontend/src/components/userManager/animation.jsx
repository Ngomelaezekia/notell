import { cubicBezier } from "framer-motion";

/* -------------------------------------------------------------------------- */
/*                                  EASING                                    */
/* -------------------------------------------------------------------------- */

export const ease = cubicBezier(0.22, 1, 0.36, 1);

/* -------------------------------------------------------------------------- */
/*                               TRANSITIONS                                  */
/* -------------------------------------------------------------------------- */

export const transition = {
  fast: {
    duration: 0.2,
    ease,
  },

  normal: {
    duration: 0.35,
    ease,
  },

  slow: {
    duration: 0.45,
    ease,
  },

  spring: {
    type: "spring",
    stiffness: 280,
    damping: 20,
  },

  springSoft: {
    type: "spring",
    stiffness: 220,
    damping: 24,
  },
};

/* -------------------------------------------------------------------------- */
/*                            PAGE TRANSITIONS                                */
/* -------------------------------------------------------------------------- */

export const pageTransition = transition.slow;

/* -------------------------------------------------------------------------- */
/*                              PAGE VARIANTS                                 */
/* -------------------------------------------------------------------------- */

export const fadeUp = {
  hidden: {
    opacity: 0,
    y: 24,
  },

  visible: {
    opacity: 1,
    y: 0,
    transition: pageTransition,
  },

  exit: {
    opacity: 0,
    y: -24,
    transition: transition.fast,
  },
};

export const slideLeft = {
  hidden: {
    opacity: 0,
    x: 60,
  },

  visible: {
    opacity: 1,
    x: 0,
    transition: pageTransition,
  },

  exit: {
    opacity: 0,
    x: -60,
    transition: transition.normal,
  },
};

export const scaleIn = {
  hidden: {
    opacity: 0,
    scale: 0.96,
  },

  visible: {
    opacity: 1,
    scale: 1,
    transition: pageTransition,
  },

  exit: {
    opacity: 0,
    scale: 0.96,
    transition: transition.fast,
  },
};

/* -------------------------------------------------------------------------- */
/*                            STAGGER CONTAINERS                              */
/* -------------------------------------------------------------------------- */

export const staggerContainer = {
  hidden: {},

  visible: {
    transition: {
      staggerChildren: 0.08,
      delayChildren: 0.05,
    },
  },
};

/* -------------------------------------------------------------------------- */
/*                             CARD ANIMATIONS                                */
/* -------------------------------------------------------------------------- */

export const cardAnimation = {
  hidden: {
    opacity: 0,
    y: 20,
  },

  visible: {
    opacity: 1,
    y: 0,
    transition: transition.normal,
  },
};

export const floatingCard = {
  whileHover: {
    y: -6,
    transition: transition.fast,
  },
};

/* -------------------------------------------------------------------------- */
/*                              AVATAR ANIMATION                              */
/* -------------------------------------------------------------------------- */

export const avatarAnimation = {
  whileHover: {
    scale: 1.05,
    rotate: 2,
    transition: transition.spring,
  },

  whileTap: {
    scale: 0.98,
  },
};

/* -------------------------------------------------------------------------- */
/*                             BUTTON ANIMATION                               */
/* -------------------------------------------------------------------------- */

export const buttonAnimation = {
  whileHover: {
    y: -2,
    scale: 1.02,
    transition: transition.fast,
  },

  whileTap: {
    scale: 0.97,
  },
};

/* -------------------------------------------------------------------------- */
/*                           STEP INDICATOR                                   */
/* -------------------------------------------------------------------------- */

export const stepAnimation = {
  inactive: {
    scale: 1,
    opacity: 0.7,
  },

  active: {
    scale: 1.08,
    opacity: 1,
    transition: transition.spring,
  },
};

/* -------------------------------------------------------------------------- */
/*                             PROGRESS BAR                                   */
/* -------------------------------------------------------------------------- */

export const progressAnimation = {
  initial: {
    width: 0,
  },

  animate: (width) => ({
    width,
    transition: transition.normal,
  }),
};

/* -------------------------------------------------------------------------- */
/*                                  TOAST                                     */
/* -------------------------------------------------------------------------- */

export const toastAnimation = {
  hidden: {
    opacity: 0,
    y: 50,
    scale: 0.9,
  },

  visible: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: transition.normal,
  },

  exit: {
    opacity: 0,
    y: 30,
    scale: 0.95,
    transition: transition.fast,
  },
};

/* -------------------------------------------------------------------------- */
/*                           REUSABLE HELPERS                                 */
/* -------------------------------------------------------------------------- */

export const hoverLift = {
  whileHover: {
    y: -4,
    transition: transition.fast,
  },
};

export const hoverScale = {
  whileHover: {
    scale: 1.03,
    transition: transition.fast,
  },
};

export const tapScale = {
  whileTap: {
    scale: 0.97,
  },
};

export const fade = {
  hidden: {
    opacity: 0,
  },

  visible: {
    opacity: 1,
    transition: transition.normal,
  },

  exit: {
    opacity: 0,
    transition: transition.fast,
  },
};