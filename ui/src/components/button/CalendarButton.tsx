import React from "react";
import { Button } from "react-bootstrap";
import { Calendar as IconCalendar } from "react-feather";
import { useTranslation } from "next-export-i18n";

interface Props {
  onClick: () => void;
  disabled?: boolean;
}

const CalendarButton: React.FC<Props> = ({ onClick, disabled }) => {
  const { t } = useTranslation();
  return (
    <Button variant="secondary" onClick={onClick} disabled={disabled}>
      <IconCalendar className="feather" style={{ marginRight: "5px" }} />
      {t("calendar")}
    </Button>
  );
};

export default CalendarButton;
