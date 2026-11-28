import React from "react";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import Loading from "@/components/Loading";
import Navigation, { PreferencesTab } from "@/util/Navigation";

interface Props {
  router: NextRouter;
}

class Preferences extends React.Component<Props> {
  componentDidMount = () => {
    const tab = this.props.router.query.tab;
    this.props.router.replace(
      Navigation.preferences(
        typeof tab === "string" ? (tab as PreferencesTab) : null,
      ),
    );
  };

  render() {
    return <Loading />;
  }
}

export default withReadyRouter(Preferences as any);
