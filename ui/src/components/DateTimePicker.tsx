import React from "react";
import "flatpickr/dist/themes/airbnb.css";
import Flatpickr from "react-flatpickr";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import Formatting from "@/util/Formatting";

interface State {}

interface Props {
  t: TranslationFunc;
  enableTime?: boolean | undefined;
  required?: boolean | undefined;
  disabled?: boolean | undefined;
  value: Date;
  minDate?: Date | undefined;
  maxDate?: Date | undefined;
  onChange: (value: Date) => void;
}

class DateTimePicker extends React.Component<Props, State> {
  render() {
    const formatter = Formatting.getFormatterShort();
    return (
      <Flatpickr
        data-enable-time={this.props.enableTime}
        disabled={this.props.disabled}
        value={this.props.value}
        required={this.props.required}
        onChange={([value]: Date[]) => {
          if (value != null && value instanceof Date)
            this.props.onChange(value);
        }}
        options={{
          time_24hr: RuntimeConfig.INFOS.use24HourTime,
          minDate: this.props.minDate,
          minTime: "00:00:00",
          maxDate: this.props.maxDate,
          formatDate: (date: Date) => {
            console.log("Formatting date:", date, "->", Formatting.convertToFakeUTCDate(date));
            return formatter.format(Formatting.convertToFakeUTCDate(date));
          }
        }}
      />
    );
  }
}

export default withTranslation(DateTimePicker as any);
