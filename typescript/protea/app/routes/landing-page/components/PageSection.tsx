import React, { ReactNode } from "react";

export interface PageSectionProps {
  padding?: "default" | "wide";
  children: ReactNode;
  className?: string;
}

export function PageSection({ padding = "default", children, className = "" }: PageSectionProps) {
  const padClass = padding === "wide" ? "pad-wide" : "pad-default";
  
  return (
    <section className={`page-section ${padClass} ${className}`.trim()}>
      <div className="page-section-inner">
        {children}
      </div>
    </section>
  );
}
