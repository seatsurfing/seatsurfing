import React from "react";
import "flatpickr/dist/themes/airbnb.css";
import Flatpickr from "react-flatpickr";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";

interface State {}

interface Props {
  t: TranslationFunc;
  enableTime?: boolean | undefined;
  required?: boolean | undefined;
  disabled?: boolean | undefined;
  value: Date;
  onChange: (value: Date) => void;
}

class DateTimePicker extends React.Component<Props, State> {
  render() {
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
        }}
      />
    );
  }
}

//export default Loading;
export default withTranslation(DateTimePicker as any);
//export default withTranslation(Loading as any) as unknown as React.Component<Props, State>;
