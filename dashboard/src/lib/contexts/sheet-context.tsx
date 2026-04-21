"use client";

import * as React from "react";

interface SheetContextType {
  isInSheet: boolean;
}

const SheetContext = React.createContext<SheetContextType | undefined>(
  undefined,
);

export function SheetProvider({ children }: { children: React.ReactNode }) {
  return (
    <SheetContext.Provider value={{ isInSheet: false }}>
      {children}
    </SheetContext.Provider>
  );
}

export function SheetContentProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SheetContext.Provider value={{ isInSheet: true }}>
      {children}
    </SheetContext.Provider>
  );
}

export function useSheetContext() {
  return React.useContext(SheetContext) ?? { isInSheet: false };
}
