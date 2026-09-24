import React from "react";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import SeatsurfingAppLogo from "@/components/SeatsurfingAppLogo";

interface Props {
  t: TranslationFunc;
}

class Error404 extends React.Component<Props> {
  constructor(props: any) {
    super(props);
  }

  render() {
    return (
      <div className="container-center">
        <div className="container-center-inner">
          <SeatsurfingAppLogo />
          <p>
            <a href="/ui/">{this.props.t("error404")}</a>
          </p>
        </div>
      </div>
    );
  }
}

export default withTranslation(Error404 as any);
