import React from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { Dropdown, DropdownButton, NavDropdown } from "react-bootstrap";
import RuntimeConfig from "./RuntimeConfig";
import { LanguageSwitcher } from "next-export-i18n";

interface State {}

interface Props {
  inNavbar?: boolean;
  t: TranslationFunc;
}

class LanguageSelector extends React.Component<Props, State> {
  render() {
    if (this.props.inNavbar) {
      return (
        <NavDropdown title={RuntimeConfig.getLanguage()}>
          {RuntimeConfig.getAvailableLanguages()
            .sort()
            .filter((l) => l !== "default")
            .map((l) => (
              <LanguageSwitcher key={"lng-" + l} lang={l}>
                <NavDropdown.Item
                  key={"lng-btn-" + l}
                  active={l === RuntimeConfig.getLanguage()}
                >
                  {l}
                </NavDropdown.Item>
              </LanguageSwitcher>
            ))}
        </NavDropdown>
      );
    }

    return (
      <DropdownButton
        title={RuntimeConfig.getLanguage()}
        className="lng-selector"
        size="sm"
        variant="outline-secondary"
        drop="up"
      >
        {RuntimeConfig.getAvailableLanguages()
          .sort()
          .filter((l) => l !== "default")
          .map((l) => (
            <LanguageSwitcher key={"lng-" + l} lang={l}>
              <Dropdown.Item
                key={"lng-btn-" + l}
                active={l === RuntimeConfig.getLanguage()}
              >
                {l}
              </Dropdown.Item>
            </LanguageSwitcher>
          ))}
      </DropdownButton>
    );
  }
}

export default withTranslation(LanguageSelector as any);
