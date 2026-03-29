import React from "react";
import { Button, Form } from "react-bootstrap";
import Link from "next/link";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Organization from "@/types/Organization";
import Ajax from "@/util/Ajax";

interface State {
  loading: boolean;
  complete: boolean;
  success: boolean;
  email: string;
}

interface Props {
  t: TranslationFunc;
}

class InitPasswordReset extends React.Component<Props, State> {
  org: Organization | null;

  constructor(props: any) {
    super(props);
    this.org = null;
    this.state = {
      loading: false,
      complete: false,
      success: false,
      email: "",
    };
  }

  componentDidMount = () => {
    this.loadOrgDetails();
  };

  loadOrgDetails = async () => {
    const domain = window.location.host.split(":").shift();
    try {
      const res = await Ajax.get("/auth/org/" + domain);
      this.org = new Organization();
      this.org.deserialize(res.json.organization);
    } catch {
      const res = await Ajax.get("/auth/singleorg");
      this.org = new Organization();
      this.org.deserialize(res.json.organization);
    }
  };

  onPasswordSubmit = async (e: any) => {
    e.preventDefault();
    this.setState({ loading: true, complete: false, success: false });
    const payload = {
      email: this.state.email,
      organizationId: this.org ? this.org.id : "",
    };
    try {
      const res = await Ajax.postData("/auth/initpwreset", payload);
      const success = res.status >= 200 && res.status <= 299;
      this.setState({ loading: false, complete: true, success });
    } catch {
      this.setState({ loading: false, complete: true, success: false });
    }
  };

  renderContent() {
    if (this.state.complete) {
      const message = this.state.success
        ? this.props.t("initPasswordResetEmail")
        : this.props.t("initPasswordResetFailed");
      return <p>{message}</p>;
    }

    return (
      <>
        <Form.Group>
          <Form.Control
            type="email"
            placeholder={this.props.t("emailPlaceholder")}
            value={this.state.email}
            onChange={(e: any) => this.setState({ email: e.target.value })}
            required={true}
            autoFocus={true}
          />
        </Form.Group>
        <Button
          className="margin-top-10"
          variant="primary"
          type="submit"
          disabled={this.state.loading}
        >
          {this.props.t("changePassword")}
        </Button>
      </>
    );
  }

  render() {
    return (
      <div className="container-center">
        <Form
          className="container-center-inner"
          onSubmit={this.onPasswordSubmit}
        >
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          {this.renderContent()}
          <p className="margin-top-50">
            <Link href="/login">{this.props.t("back")}</Link>
          </p>
        </Form>
      </div>
    );
  }
}

export default withTranslation(InitPasswordReset as any);
