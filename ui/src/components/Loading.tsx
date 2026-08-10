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
    const paddingTop = (this.props.paddingTop ?? true) ? "padding-top" : "";
    const display =
      this.props.visible === undefined || this.props.visible === true
        ? "display-block"
        : "display-none";
    return (
      <div className={`${paddingTop} ${display} center loading-overlay`}>
        <IconLoad className="feather loader" />
        &nbsp;Loading &hellip;
      </div>
    );
  }
}

//export default Loading;
export default withTranslation(Loading as any);
//export default withTranslation(Loading as any) as unknown as React.Component<Props, State>;
