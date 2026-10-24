import {
  Home,
  Leaf,
  ShoppingCart,
  TrendingUp,
  Trophy,
  Activity,
  Settings,
  LucideIcon
} from "lucide-react";

type Submenu = {
  href: string;
  label: string;
  active?: boolean;
};

type Menu = {
  href: string;
  label: string;
  active?: boolean;
  icon: LucideIcon;
  submenus?: Submenu[];
};

type Group = {
  groupLabel: string;
  menus: Menu[];
};

export function getMenuList(): Group[] {
  return [
    {
      groupLabel: "",
      menus: [
        {
          href: "/",
          label: "Dashboard",
          icon: Home,
          submenus: []
        }
      ]
    },
    {
      groupLabel: "Garden",
      menus: [
        {
          href: "/garden",
          label: "My Gardens",
          icon: Leaf,
          submenus: []
        },
        {
          href: "/profile",
          label: "Progress",
          icon: Trophy,
          submenus: []
        }
      ]
    },
    {
      groupLabel: "Marketplace",
      menus: [
        {
          href: "/marketplace",
          label: "Marketplace",
          icon: TrendingUp,
          submenus: []
        },
        {
          href: "/store",
          label: "Store",
          icon: ShoppingCart,
          submenus: []
        }
      ]
    },
    {
      groupLabel: "System",
      menus: [
        {
          href: "/audit",
          label: "Audit Logs",
          icon: Activity
        },
        {
          href: "/settings",
          label: "Settings",
          icon: Settings
        }
      ]
    }
  ];
}
