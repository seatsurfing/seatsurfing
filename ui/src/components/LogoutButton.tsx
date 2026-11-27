import React from "react";
import { Nav } from "react-bootstrap";
import { LogOut as IconLogout } from "react-feather";
import { TranslationFunc, withTranslation } from "./withTranslation";
import RuntimeConfig from "./RuntimeConfig";
import { Breakpoint } from "@/types/Layout";

interface Props {
  compactBreakpoint?: Breakpoint;
  t: TranslationFunc;
}

class LogoutButton extends React.Component<Props, {}> {
  logOut = (e: React.MouseEvent) => {
    e.preventDefault();
    RuntimeConfig.logOut();
  };

  render() {
    const label = this.props.t("logout");
    return (
      <Nav.Link onClick={this.logOut} title={label} aria-label={label}>
        <IconLogout size={16} />
        {this.props.compactBreakpoint && (
          <span className={`d-${this.props.compactBreakpoint}-none`}>
            {" "}
            {label}
          </span>
        )}
      </Nav.Link>
    );
  }
}

export default withTranslation(LogoutButton as any);
