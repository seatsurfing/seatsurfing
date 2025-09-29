import React from "react";
import withReadyRouter from "@/components/withReadyRouter";
import { NextRouter } from "next/router";

interface State {}

interface Props {
  router: NextRouter;
}

class Index extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
  }

  render() {
    this.props.router.push("/login");
    return <></>;
  }
}

export default withReadyRouter(Index as any);
