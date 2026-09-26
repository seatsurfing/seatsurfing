import React from "react";
import ThemeSelector from "@/components/ThemeSelector";
import LanguageSelector from "@/components/LanguageSelector";

const CopyrightFooter: React.FC = () => {
  return (
    <div className="copyright-footer">
      &copy;&nbsp;
      <a
        href="https://seatsurfing.io"
        target="_blank"
        rel="noopener noreferrer"
      >
        Seatsurfing
      </a>
      <div className="footer-selectors">
        <ThemeSelector compactBreakpoint="md" />
        <LanguageSelector compactBreakpoint="md" />
      </div>
    </div>
  );
};

export default CopyrightFooter;
