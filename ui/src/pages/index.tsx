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
    const { query } = this.props.router;
    this.props.router.push({
      pathname: "/login",
      query: query,
    });
    return <></>;
  }
}

export default withReadyRouter(Index as any);
