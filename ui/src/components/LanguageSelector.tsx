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
    const icon = (
      <svg
        viewBox="0 0 24 24"
        width="20"
        height="20"
        aria-hidden="true"
        className="iconLanguage_nlXk"
      >
        <path
          fill="currentColor"
          d="M12.87 15.07l-2.54-2.51.03-.03c1.74-1.94 2.98-4.17 3.71-6.53H17V4h-7V2H8v2H1v1.99h11.17C11.5 7.92 10.44 9.75 9 11.35 8.07 10.32 7.3 9.19 6.69 8h-2c.73 1.63 1.73 3.17 2.98 4.56l-5.09 5.02L4 19l5-5 3.11 3.11.76-2.04zM18.5 10h-2L12 22h2l1.12-3h4.75L21 22h2l-4.5-12zm-2.62 7l1.62-4.33L19.12 17h-3.24z"
        ></path>
      </svg>
    );

    if (this.props.inNavbar) {
      return (
        <NavDropdown
          title={
            <>
              {icon}
              {RuntimeConfig.getLanguage()}
            </>
          }
        >
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
      <>
        <DropdownButton
          title={
            <>
              {icon}
              {RuntimeConfig.getLanguage()}
            </>
          }
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
      </>
    );
  }
}

export default withTranslation(LanguageSelector as any);
