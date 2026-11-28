import React from "react";
import { Button } from "react-bootstrap";
import { Image as IconImage } from "react-feather";
import { useTranslation } from "next-export-i18n";

interface Props {
  onClick: () => void;
  disabled?: boolean;
  iconOnly?: boolean;
}

const SpaceImageButton: React.FC<Props> = ({ onClick, disabled, iconOnly }) => {
  const { t } = useTranslation();
  if (iconOnly) {
    return (
      <Button
        variant="outline-secondary"
        onClick={onClick}
        disabled={disabled}
        title={t("image")}
        aria-label={t("image")}
      >
        <IconImage className="feather" />
      </Button>
    );
  }
  return (
    <Button variant="secondary" onClick={onClick} disabled={disabled}>
      <IconImage className="feather" style={{ marginRight: "5px" }} />
      {t("image")}
    </Button>
  );
};

export default SpaceImageButton;
