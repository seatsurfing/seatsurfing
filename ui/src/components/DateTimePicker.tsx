import React from "react";
import "flatpickr/dist/themes/airbnb.css";
import Flatpickr from "react-flatpickr";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import { CustomLocale } from "flatpickr/dist/types/locale";
import { english as DefaultLocale } from "flatpickr/dist/l10n/default.js";

interface State {
  locale: CustomLocale | undefined;
}

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
  constructor(props: Props) {
    super(props);
    this.state = {
      locale: undefined,
    };
  }

  componentDidMount() {
    this.loadLocale();
  }

  async loadLocale() {
    let lang = RuntimeConfig.getLanguage();  
    if (lang.indexOf("-") !== -1) {
      lang = lang.split("-")[0];
    }
    if (lang === "en") {
      this.setState({ locale: DefaultLocale });
      return;
    }
      try {
        const localeModule = await import(`flatpickr/dist/l10n/${lang}.js`);
        const name = Object.keys(localeModule)[0];
        this.setState({ locale: localeModule[name] });
      } catch (error) {
        console.error(`Failed to load locale ${lang}:`, error);
        this.setState({ locale: DefaultLocale });
      }
    }

  render() {
    if (!this.state.locale) {
      return <></>;
    }
    return (
      <Flatpickr
        className="form-control"
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
          maxDate: this.props.maxDate,
          locale: this.state.locale,
        }}
      />
    );
  }
}

export default withTranslation(DateTimePicker as any);
