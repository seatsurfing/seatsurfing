import React from "react";
import { Button } from "react-bootstrap";
import { Calendar as IconCalendar } from "react-feather";
import { useTranslation } from "next-export-i18n";

interface Props {
  onClick: () => void;
  disabled?: boolean;
  iconOnly?: boolean;
}

const CalendarButton: React.FC<Props> = ({ onClick, disabled, iconOnly }) => {
  const { t } = useTranslation();
  if (iconOnly) {
    return (
      <Button
        variant="outline-secondary"
        onClick={onClick}
        disabled={disabled}
        title={t("calendar")}
        aria-label={t("calendar")}
      >
        <IconCalendar className="feather" />
      </Button>
    );
  }
  return (
    <Button variant="secondary" onClick={onClick} disabled={disabled}>
      <IconCalendar className="feather" style={{ marginRight: "5px" }} />
      {t("calendar")}
    </Button>
  );
};

export default CalendarButton;
