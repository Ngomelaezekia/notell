import { Home, UserCircle, Radio, MessageCircle, Triangle } from "lucide-react";

export const mainNavigation = [
  {
    name: "Posts",
    path: "/",
    icon: Home,
    showIn: ["desktop", "mobile", "drawer"],
  },
  {
    name: "Create",
    path: "/create-post",
    icon: Triangle,
    showIn: ["mobile"],
    isCreate: true,
  },
  {
    name: "Live",
    path: "/live",
    icon: Radio,
    showIn: ["desktop", "drawer"],
    comingSoon: true,
  },
  {
    name: "Messages",
    path: "/messages",
    icon: MessageCircle,
    showIn: ["desktop", "drawer"],
    comingSoon: true,
  },
  {
    name: "Profile",
    path: "/user",
    icon: UserCircle,
    showIn: ["desktop", "mobile", "drawer"],
  },
];
