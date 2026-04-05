import React from "react";

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
  return (
    <button
      type="button"
      className="ms-2 btn d-flex align-items-center"
      style={{
        padding: "4px 8px",
        borderColor: "#CED4DA",
      }}
      disabled={disabled}
      onClick={onClick}
      title={title}
    >
      {text}
    </button>
  );
};

export default IconTextButton;
