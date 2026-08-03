import {
  Home,
  Compass,
  UserCircle,
  Settings,
  Radio,
  MessageCircle,
} from "lucide-react";


export const mainNavigation = [

  {
    name:"Posts",

    path:"/",

    icon:Home,

    showIn:[
      "desktop",
      "mobile",
      "drawer"
    ],

  },


  {
    name:"Explore",

    path:"/explore",

    icon:Compass,

    showIn:[
      "desktop",
      "mobile",
      "drawer"
    ],

  },


  {
    name:"Live",

    path:"/live",

    icon:Radio,

    showIn:[
      "desktop",
      "drawer"
    ],

    comingSoon:true,

  },


  {
    name:"Messages",

    path:"/messages",

    icon:MessageCircle,

    showIn:[
      "desktop",
      "drawer"
    ],

    comingSoon:true,

  },


  {
    name:"Account",

    path:"/profile",

    icon:UserCircle,

    showIn:[
      "desktop",
      "mobile",
      "drawer"
    ],

  },


  {
    name:"Settings",

    path:"/settings",

    icon:Settings,

    showIn:[
      "drawer"
    ],

  },

];