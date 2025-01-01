import React from 'react';
import { Ajax, Formatting } from 'flexspace-commons';
import Loading from '../components/Loading';

interface State {
  loading: boolean
  now: Date
  fakeUtcTime: Date
  res: any
}

interface Props {
}

class DebugTimeIssues extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    let now = new Date();
    this.state = {
      loading: true,
      now: now,
      fakeUtcTime: Formatting.convertToFakeUTCDate(now),
      res: {}
    };
  }

  componentDidMount = () => {
    let payload = {
      "time": this.state.fakeUtcTime.toISOString()
    };
    Ajax.postData("/booking/debugtimeissues/", payload).then(result => {
      this.setState({
        loading: false,
        res: result.json
      });
    });
  }

  render() {
    let serverSideContent = <Loading />;
    if (!this.state.loading) {
      let resDate = new Date(this.state.res.result);
      let strippedString = Formatting.stripTimezoneDetails(this.state.res.result);
      let strippedDate = new Date(strippedString);
      serverSideContent = (
        <>
          <p>Server Error:<br />{this.state.res.error}</p>
          <p>Location Timezone:<br />{this.state.res.tz}</p>
          <p>Received Time:<br />{this.state.res.receivedTime}</p>
          <p>Received Time Transformed:<br />{this.state.res.receivedTimeTransformed}</p>
          <p>Database Time:<br />{this.state.res.dbTime}</p>
          <p>Result:<br />{this.state.res.result}</p>
          <hr />
          <p>Browser Received Date:<br />{resDate.toString()}</p>
          <p>Browser Transformed String:<br />{strippedString}</p>
          <p>Browser Transformed Time:<br />{strippedDate.toString()}</p>
          <p>Browser Formatted Time:<br />{Formatting.getFormatter().format(strippedDate)}</p>
        </>
      );
    }

    return (
      <div className="container-center">
        <div className="container-center-inner-wide">
          <p>User Agent:<br />{typeof window !== 'undefined' ? window.navigator.userAgent : ''}</p>
          <p>Browser Language:<br />{typeof window !== 'undefined' ? window.navigator.language : ''}</p>
          <hr />
          <p>Current Browser Time:<br />{this.state.now.toString()}</p>
          <p>Current Browser Time ISO String:<br />{this.state.now.toISOString()}</p>
          <p>Current Browser Time Offset:<br />{this.state.now.getTimezoneOffset()}</p>
          <p>Fake UTC Time:<br />{this.state.fakeUtcTime.toISOString()}</p>
          <hr />
          {serverSideContent}
        </div>
      </div>
    );
  }
}

export default DebugTimeIssues;
