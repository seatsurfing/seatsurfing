import React, { useState } from "react";
import { IconType } from "react-icons";

interface Props {
  icon: IconType;
  active: boolean;
  title: string;
  onClick: () => void;
  disabled?: boolean;
}

const IconButton: React.FC<Props> = ({
  icon: Icon,
  active,
  title,
  onClick,
  disabled,
}) => {
  const [hover, setHover] = useState(false);
  const hovered = hover && !disabled && !active;
  return (
    <button
      type="button"
      className={`ms-2 btn d-flex align-items-center ${
        active ? "btn-primary" : "btn-light"
      }`}
      style={{
        padding: "4px 8px",
        borderColor: hovered ? "#6c757d" : "#CED4DA",
        backgroundColor: disabled
          ? "var(--bs-secondary-bg)"
          : active
            ? undefined
            : hovered
              ? "#6c757d"
              : "#fff",
        opacity: disabled ? 1 : undefined,
      }}
      disabled={disabled}
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      title={title}
    >
      <Icon
        title={title}
        color={disabled ? "#555" : active ? "#fff" : hovered ? "#fff" : "#555"}
        height="20px"
        width="20px"
      />
    </button>
  );
};

export default IconButton;
