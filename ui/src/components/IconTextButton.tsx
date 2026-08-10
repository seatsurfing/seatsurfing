import React, { useState } from "react";

interface Props {
  text: string;
  title: string;
  onClick: () => void;
  disabled?: boolean;
}

const IconTextButton: React.FC<Props> = ({
  text,
  title,
  onClick,
  disabled,
}) => {
  const [hover, setHover] = useState(false);
  const hovered = hover && !disabled;
  return (
    <button
      type="button"
      className="ms-2 btn d-flex align-items-center"
      style={{
        padding: "4px 8px",
        borderColor: hovered ? "#6c757d" : "#CED4DA",
        backgroundColor: disabled
          ? "var(--bs-secondary-bg)"
          : hovered
            ? "#6c757d"
            : "#fff",
        color: disabled ? undefined : hovered ? "#fff" : "#555",
        opacity: disabled ? 1 : undefined,
      }}
      disabled={disabled}
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      title={title}
    >
      {text}
    </button>
  );
};

export default IconTextButton;
