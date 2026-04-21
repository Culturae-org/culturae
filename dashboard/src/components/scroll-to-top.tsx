"use client";

import { IconChevronUp } from "@tabler/icons-react";
import { useEffect, useState } from "react";

export default function ScrollToTop() {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const toggleVisibility = () => {
      if (window.pageYOffset > 100) {
        setIsVisible(true);
      } else {
        setIsVisible(false);
      }
    };

    window.addEventListener("scroll", toggleVisibility);
    return () => window.removeEventListener("scroll", toggleVisibility);
  }, []);

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  return (
    <>
      {isVisible && (
        <button
          type="button"
          onClick={scrollToTop}
          className="fixed bottom-4 right-4 bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground p-2 rounded-full shadow-sm transition-colors z-50 border border-border-scroll-button"
          aria-label="Scroll to top"
        >
          <IconChevronUp className="h-4 w-4" />
        </button>
      )}
    </>
  );
}
