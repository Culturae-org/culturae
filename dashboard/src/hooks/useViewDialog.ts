import { useEffect, useState } from "react";
import { useSearchParams } from "react-router";

interface UseViewDialogOptions {
  paramName?: string;
}

export function useViewDialog(
  paramName = "view",
  _options?: UseViewDialogOptions,
) {
  const [searchParams] = useSearchParams();
  const [isOpen, setIsOpen] = useState(false);
  const [scrollPosition, setScrollPosition] = useState(0);

  useEffect(() => {
    const viewId = searchParams.get(paramName);
    setIsOpen(!!viewId);
  }, [searchParams, paramName]);

  const close = () => {
    setScrollPosition(window.scrollY);
    setIsOpen(false);

    const url = new URL(window.location.href);
    url.searchParams.delete(paramName);
    window.history.replaceState({}, "", url.toString());

    setTimeout(() => {
      window.scrollTo(0, scrollPosition);
    }, 0);
  };

  const open = (id: string) => {
    const url = new URL(window.location.href);
    url.searchParams.set(paramName, id);
    window.history.pushState({}, "", url.toString());
    setIsOpen(true);
  };

  const viewId = searchParams.get(paramName);

  return {
    isOpen,
    close,
    open,
    viewId,
    setIsOpen,
  };
}
