import React from "react";
import { ButtonGroup, ToggleButton } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface Props {
  t: TranslationFunc;
  id?: string;
  weekStartDay?: number;
  value: number[];
  onChange: (value: number[]) => void;
  preventEmpty?: boolean;
}

class WeekdaySelection extends React.Component<Props> {
  onCheck = (day: number, checked: boolean) => {
    const value = checked
      ? [...this.props.value, day]
      : this.props.value.filter((d) => d !== day);
    if (!checked && this.props.preventEmpty && value.length < 1) {
      return;
    }
    this.props.onChange(value);
  };

  render() {
    const weekStartDay = this.props.weekStartDay ?? 1;
    return (
      <ButtonGroup id={this.props.id}>
        {[0, 1, 2, 3, 4, 5, 6]
          .map((offset) => (weekStartDay + offset) % 7)
          .map((day) => (
            <ToggleButton
              type="checkbox"
              variant={
                this.props.value.includes(day) ? "primary" : "outline-secondary"
              }
              key={"weekday-" + day}
              id={(this.props.id ? this.props.id + "-" : "") + "weekday-" + day}
              value={day}
              checked={this.props.value.includes(day)}
              onChange={(e: any) => this.onCheck(day, e.target.checked)}
            >
              {this.props.t("workday-short-" + day)}
            </ToggleButton>
          ))}
      </ButtonGroup>
    );
  }
}

export default withTranslation(WeekdaySelection as any);
