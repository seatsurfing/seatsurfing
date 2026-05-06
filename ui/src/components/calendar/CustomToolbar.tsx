import React from "react";
import Link from "next/link";
import { ToolbarProps } from "react-big-calendar";
import {
  Trello as IconTrello,
  ArrowLeft as IconArrowLeft,
  ArrowRight as IconArrowRight,
} from "react-feather";
import Formatting from "@/util/Formatting";
import moment from "moment-timezone";
import { TranslationFunc } from "@/components/withTranslation";

interface Props {
  toolbar: ToolbarProps<object, object>;
  t: TranslationFunc;
}

const CustomToolbar: React.FC<Props> = ({ toolbar, t }) => {
  const weekStart = moment(toolbar.date).clone().startOf("week");
  const weekEnd = moment(toolbar.date).clone().endOf("week");
  const formatter = Formatting.getFormatterDate();

  return (
    <div
      className="custom-toolbar"
      style={{ marginBottom: "5px", textAlign: "left" }}
    >
      <Link
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("TODAY")}
      >
        <IconTrello className="feather" /> {t("today")}
      </Link>{" "}
      <Link
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("PREV")}
      >
        <IconArrowLeft className="feather" />
      </Link>{" "}
      <Link
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("NEXT")}
      >
        <IconArrowRight className="feather" />
      </Link>{" "}
      <span
        className="toolbar-label"
        style={{
          display: "flex",
          float: "right",
          height: "100%",
          alignItems: "center",
        }}
      >
        {formatter.format(weekStart.toDate())} –{" "}
        {formatter.format(weekEnd.toDate())}
      </span>
    </div>
  );
};

export default CustomToolbar;
