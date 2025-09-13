import React, { CSSProperties } from "react";
import { withTranslation } from "./withTranslation";
import { IoDiamond as DiamondIcon } from "react-icons/io5";
import RuntimeConfig from "./RuntimeConfig";

interface State {}

interface Props {
  hidden?: boolean;
  style?: CSSProperties | undefined;
}

class PremiumFeatureIcon extends React.Component<Props, State> {
  render() {
    if (this.props.hidden) {
      return <></>;
    }
    if (
      RuntimeConfig.INFOS.cloudHosted &&
      !RuntimeConfig.INFOS.subscriptionActive
    ) {
      return (
        <DiamondIcon
          className="feather"
          style={{ marginLeft: "5px", ...this.props.style }}
          color="#007bff"
        />
      );
    }
  }
}

export default withTranslation(PremiumFeatureIcon as any);
