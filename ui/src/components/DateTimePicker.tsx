import React from "react";
import "flatpickr/dist/themes/airbnb.css";
import Flatpickr from "react-flatpickr";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import { CustomLocale } from "flatpickr/dist/types/locale";
import { Instance } from "flatpickr/dist/types/instance";
import { english as DefaultLocale } from "flatpickr/dist/l10n/default.js";

interface State {
  locale: CustomLocale | undefined;
}

interface Props {
  t: TranslationFunc;
  id?: string;
  enableTime?: boolean | undefined;
  noCalendar?: boolean | undefined;
  required?: boolean | undefined;
  disabled?: boolean | undefined;
  showWeekday?: boolean | undefined;
  value: Date | null;
  minDate?: Date | undefined;
  maxDate?: Date | undefined;
  isDateDisabled?: (date: Date) => boolean;
  // Element the popup is appended to instead of document.body, e.g. a
  // container inside a modal so the popup stays within its focus trap.
  appendTo?: HTMLElement | null;
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
    if (lang === "en" || !/^[a-z]{2,3}$/.test(lang)) {
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

  // Flatpickr positions its popup in document coordinates, which is wrong
  // once it is appended to a positioned container (e.g. inside a modal).
  // Position it relative to the viewport instead.
  positionPopup = (fp: Instance, customElement: HTMLElement | undefined) => {
    const popup = fp.calendarContainer;
    const anchor = customElement ?? fp._positionElement;
    if (!popup || !anchor) {
      return;
    }
    const rect = anchor.getBoundingClientRect();
    const height = popup.offsetHeight;
    const showOnTop =
      window.innerHeight - rect.bottom < height && rect.top > height;
    popup.style.position = "fixed";
    popup.style.top =
      (showOnTop ? rect.top - height - 2 : rect.bottom + 2) + "px";
    popup.style.left =
      Math.max(
        0,
        Math.min(rect.left, window.innerWidth - popup.offsetWidth - 2),
      ) + "px";
    popup.style.right = "auto";
    popup.classList.toggle("arrowTop", !showOnTop);
    popup.classList.toggle("arrowBottom", showOnTop);
    popup.classList.add("arrowLeft");
  };

  render() {
    if (!this.state.locale) {
      return <></>;
    }
    let formatting = this.props.noCalendar
      ? ""
      : RuntimeConfig.INFOS.dateFormat +
        (this.props.showWeekday ? " (D)" : "") +
        " ";
    if (this.props.enableTime) {
      if (RuntimeConfig.INFOS.use24HourTime) {
        formatting += "H:i";
      } else {
        formatting += "h:i K";
      }
    }
    return (
      <Flatpickr
        className="form-control"
        id={this.props.id}
        data-enable-time={this.props.enableTime}
        disabled={this.props.disabled}
        value={this.props.value ?? ""}
        required={this.props.required}
        onClose={([value]: Date[]) => {
          if (value != null && value instanceof Date)
            this.props.onChange(value);
        }}
        options={{
          dateFormat: formatting,
          time_24hr: RuntimeConfig.INFOS.use24HourTime,
          minDate: this.props.minDate,
          maxDate: this.props.maxDate,
          disable: this.props.isDateDisabled ? [this.props.isDateDisabled] : [],
          locale: {
            ...this.state.locale,
            firstDayOfWeek: RuntimeConfig.INFOS.weekStartDay,
          },
          noCalendar: this.props.noCalendar,
          ...(this.props.appendTo
            ? { appendTo: this.props.appendTo, position: this.positionPopup }
            : {}),
        }}
      />
    );
  }
}

export default withTranslation(DateTimePicker as any);
