import { Home, UserCircle, Radio, MessageCircle, Triangle } from "lucide-react";

export const mainNavigation = [
  { name: "Home", path: "/", icon: Home, showIn: ["desktop", "mobile", "drawer"] },
  { name: "Channels", path: "/channels", icon: Radio, showIn: ["desktop", "mobile", "drawer"] },
  { name: "Create", path: "/create-post", icon: Triangle, showIn: ["desktop", "mobile", "drawer"], isCreate: true },
  { name: "Messages", path: "/messages", icon: MessageCircle, showIn: ["desktop", "mobile", "drawer"] },
  { name: "Profile", path: "/user", icon: UserCircle, showIn: ["desktop", "mobile", "drawer"] },
];
