import React from "react";

interface Props {
  className?: string;
}

const SeatsurfingLogo: React.FC<Props> = ({ className }) => {
  return (
    <img
      src="/ui/seatsurfing.svg"
      alt="Seatsurfing"
      className={["seatsurfing-logo", className].filter(Boolean).join(" ")}
    />
  );
};

export default SeatsurfingLogo;
