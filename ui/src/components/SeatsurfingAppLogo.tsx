import React from "react";

interface Props {
  className?: string;
  customLogoUrl?: string;
}

const SeatsurfingAppLogo: React.FC<Props> = ({ className, customLogoUrl }) => {
  if (customLogoUrl) {
    return (
      <img
        src={customLogoUrl}
        alt="Seatsurfing"
        className={["custom-logo", className].filter(Boolean).join(" ")}
      />
    );
  }
  return (
    <img
      src="/ui/seatsurfing.svg"
      alt="Seatsurfing"
      className={["seatsurfing-logo", className].filter(Boolean).join(" ")}
    />
  );
};

export default SeatsurfingAppLogo;
