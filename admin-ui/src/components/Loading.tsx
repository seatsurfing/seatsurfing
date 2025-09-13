import React from "react";
import { Loader as IconLoad } from "react-feather";
import { TranslationFunc, withTranslation } from "./withTranslation";

interface State {}

interface Props {
  showText: boolean;
  paddingTop: boolean;
  visible: boolean;
  t: TranslationFunc;
}

class Loading extends React.Component<Props, State> {
  render() {
    let text = "Loading...";
    let paddingTop = (this.props.paddingTop ?? true) ? "padding-top" : "";
    let display =
      this.props.visible === undefined || this.props.visible === true
        ? "display-block"
        : "display-none";
    return (
      <div className={`${paddingTop} ${display} center loading-overlay`}>
        <IconLoad className="feather loader" />
        {text}
      </div>
    );
  }
}

export default withTranslation(Loading as any);
