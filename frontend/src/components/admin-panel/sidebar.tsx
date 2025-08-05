"use client";
import { Menu } from "@/components/admin-panel/menu";
import { SidebarToggle } from "@/components/admin-panel/sidebar-toggle";
import { Button } from "@/components/ui/button";
import { useSidebar } from "@/hooks/use-sidebar";
import { useStore } from "@/hooks/use-store";
import { cn } from "@/lib/utils";
import { PanelsTopLeft, Menu as MenuIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { useState } from "react";

export function Sidebar() {
  const sidebar = useStore(useSidebar, (x) => x);
  const [isMobileOpen, setIsMobileOpen] = useState(false);
  
  if (!sidebar) return null;
  const { isOpen, toggleOpen, getOpenState, setIsHover, settings } = sidebar;
  
  return (
    <>
      {/* Mobile Menu Button */}
      <div className="lg:hidden fixed top-4 left-4 z-50">
        <Button
          onClick={() => setIsMobileOpen(!isMobileOpen)}
          variant="outline"
          size="icon"
          className="w-10 h-10"
        >
          <MenuIcon className="w-5 h-5" />
        </Button>
      </div>

      {/* Mobile Overlay */}
      {isMobileOpen && (
        <div 
          className="lg:hidden fixed inset-0 bg-black/50 z-40"
          onClick={() => setIsMobileOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          "fixed top-0 left-0 z-40 h-screen transition-all duration-300",
          // Mobile: slide in/out from left
          "lg:translate-x-0",
          isMobileOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
          // Desktop: always visible, just collapsed/expanded
          !getOpenState() ? "w-[90px]" : "w-52",
          settings.disabled && "hidden"
        )}
      >
        <SidebarToggle isOpen={isOpen} setIsOpen={toggleOpen} />
        <div
          onMouseEnter={() => setIsHover(true)}
          onMouseLeave={() => setIsHover(false)}
          className="relative h-full flex flex-col px-3 py-4 overflow-y-auto shadow-md dark:shadow-zinc-800 bg-background"
        >
          <Button
            className={cn(
              "transition-transform ease-in-out duration-300 mb-1",
              !getOpenState() ? "translate-x-1" : "translate-x-0"
            )}
            variant="link"
            asChild
          >
            <Link to="/" className="flex items-center gap-2">
              <PanelsTopLeft className="w-6 h-6 mr-1" />
              <h1
                className={cn(
                  "font-bold text-lg whitespace-nowrap transition-[transform,opacity,display] ease-in-out duration-300",
                  !getOpenState()
                    ? "-translate-x-96 opacity-0 hidden"
                    : "translate-x-0 opacity-100"
                )}
              >
                Virtual Garden
              </h1>
            </Link>
          </Button>
          <Menu isOpen={getOpenState()} onMobileClose={() => setIsMobileOpen(false)} />
        </div>
      </aside>
    </>
  );
}
