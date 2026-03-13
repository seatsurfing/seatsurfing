import React from "react";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { withTranslation } from "@/components/withTranslation";

interface State {}

interface Props {
  router: NextRouter;
}

class AdminIndex extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {};
  }

  componentDidMount() {
    this.props.router.push("/admin/dashboard/");
  }

  render() {
    return <></>;
  }
}

export default withTranslation(withReadyRouter(AdminIndex as any));
