import React from "react";
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
  return (
    <button
      type="button"
      className={`ms-2 btn d-flex align-items-center ${
        active ? "btn-primary" : "btn-light"
      }`}
      style={{
        padding: "4px 8px",
        borderColor: "#CED4DA",
        backgroundColor: disabled ? "var(--bs-secondary-bg)" : undefined,
        opacity: disabled ? 1 : undefined,
      }}
      disabled={disabled}
      onClick={onClick}
      title={title}
    >
      <Icon
        title={title}
        color={active ? "#fff" : "#555"}
        height="20px"
        width="20px"
      />
    </button>
  );
};

export default IconButton;
