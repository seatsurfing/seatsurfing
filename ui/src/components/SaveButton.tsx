import React from "react";
import { Button } from "react-bootstrap";
import { useTranslation } from "next-export-i18n";

interface Props {
  submitting: boolean;
  className?: string;
  disabled?: boolean;
}

const SaveButton: React.FC<Props> = ({ submitting, className, disabled }) => {
  const { t } = useTranslation();
  return (
    <Button type="submit" disabled={submitting || disabled} className={className}>
      {submitting && (
        <span
          className="spinner-border spinner-border-sm me-2"
          role="status"
          aria-hidden="true"
        />
      )}
      {t("save")}
    </Button>
  );
};

export default SaveButton;
