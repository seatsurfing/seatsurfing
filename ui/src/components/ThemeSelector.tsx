import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Dropdown, DropdownButton, NavDropdown } from "react-bootstrap";
import {
  Sun as IconLight,
  Moon as IconDark,
  Monitor as IconAuto,
} from "react-feather";
import Theme, { ThemeMode } from "@/util/Theme";

interface State {
  mode: ThemeMode;
}

interface Props {
  inNavbar?: boolean;
  drop?: "up" | "down" | "start" | "end";
  align?: "start" | "end";
  variant?: string;
  t: TranslationFunc;
}

class ThemeSelector extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      mode: Theme.AUTO,
    };
  }

  componentDidMount(): void {
    this.setState({ mode: Theme.get() });
  }

  setMode = (mode: ThemeMode) => {
    Theme.set(mode);
    this.setState({ mode });
  };

  getIcon = (mode: ThemeMode) => {
    switch (mode) {
      case Theme.LIGHT:
        return <IconLight size={16} />;
      case Theme.DARK:
        return <IconDark size={16} />;
      default:
        return <IconAuto size={16} />;
    }
  };

  render() {
    const options: { mode: ThemeMode; label: string }[] = [
      { mode: Theme.LIGHT, label: this.props.t("themeLight") },
      { mode: Theme.DARK, label: this.props.t("themeDark") },
      { mode: Theme.AUTO, label: this.props.t("themeSystem") },
    ];

    const title = <>{this.getIcon(this.state.mode)}</>;

    const ItemComponent = this.props.inNavbar
      ? NavDropdown.Item
      : Dropdown.Item;

    const items = options.map((option) => (
      <ItemComponent
        key={"theme-" + option.mode}
        active={option.mode === this.state.mode}
        onClick={() => this.setMode(option.mode)}
      >
        {this.getIcon(option.mode)} {option.label}
      </ItemComponent>
    ));

    if (this.props.inNavbar) {
      return (
        <NavDropdown
          align={this.props.align ?? "start"}
          title={title}
          drop={this.props.drop ?? "down"}
        >
          {items}
        </NavDropdown>
      );
    }

    return (
      <DropdownButton
        title={title}
        className="theme-selector"
        size="sm"
        variant={this.props.variant ?? "outline-secondary"}
        drop={this.props.drop ?? "up"}
      >
        {items}
      </DropdownButton>
    );
  }
}

export default withTranslation(ThemeSelector as any);
