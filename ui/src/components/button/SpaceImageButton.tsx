import React from "react";
import { Button } from "react-bootstrap";
import { Image as IconImage } from "react-feather";
import { useTranslation } from "next-export-i18n";

interface Props {
  onClick: () => void;
  disabled?: boolean;
}

const SpaceImageButton: React.FC<Props> = ({ onClick, disabled }) => {
  const { t } = useTranslation();
  return (
    <Button variant="secondary" onClick={onClick} disabled={disabled}>
      <IconImage className="feather" style={{ marginRight: "5px" }} />
      {t("spaceImage")}
    </Button>
  );
};

export default SpaceImageButton;
