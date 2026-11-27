import React from "react";
import { Image as IconImage } from "react-feather";
import { useTranslation } from "next-export-i18n";

const SpacePhotoIcon: React.FC = () => {
  const { t } = useTranslation();
  const label = t("image");
  return (
    <span title={label} style={{ display: "flex" }}>
      <IconImage size={16} role="img" aria-label={label} />
    </span>
  );
};

export default SpacePhotoIcon;
