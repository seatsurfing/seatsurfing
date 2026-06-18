import React, { useState } from "react";
import { Button } from "react-bootstrap";
import { Clipboard as IconCopy, Check as IconCheck } from "react-feather";

interface Props {
  text: string;
  disabled?: boolean;
  hidden?: boolean;
  small?: boolean;
}

const CopyToClipboardButton: React.FC<Props> = ({
  text,
  disabled,
  hidden,
  small,
}) => {
  const [copied, setCopied] = useState(false);

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const iconStyle = small ? { width: "14px", height: "14px" } : undefined;
  const buttonStyle = small
    ? { marginLeft: "5px", padding: "4px 4px", lineHeight: 1 }
    : undefined;

  return (
    <Button
      onClick={handleClick}
      disabled={disabled ?? text === ""}
      hidden={hidden}
      variant="outline-secondary"
      size={small ? "sm" : undefined}
      style={buttonStyle}
    >
      {copied ? (
        <IconCheck className="feather" style={iconStyle} />
      ) : (
        <IconCopy className="feather" style={iconStyle} />
      )}
    </Button>
  );
};

export default CopyToClipboardButton;
