import React from "react";
import { Button, Form } from "react-bootstrap";
import { NextRouter } from "next/router";
import Link from "next/link";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Ajax from "@/util/Ajax";

interface State {
  loading: boolean;
  complete: boolean;
  success: boolean;
  newPassword: string;
  linkValid: boolean | null; // null = checking, true = valid, false = expired
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class CompleteUserInvitation extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: false,
      complete: false,
      success: false,
      newPassword: "",
      linkValid: null, // Start with null to indicate checking
    };
  }

  componentDidMount() {
    const { id } = this.props.router.query;
    if (!id) {
      this.setState({ linkValid: false });
      return;
    }
    // Validate the invite link
    Ajax.get("/auth/setpw/" + id)
      .then((res) => {
        if (res.status === 204 || res.status === 200) {
          this.setState({ linkValid: true });
        } else if (res.status === 410) {
          // Password already set - link is no longer valid
          this.setState({ linkValid: false });
        } else {
          // Link not found or other error
          this.setState({ linkValid: false });
        }
      })
      .catch((e) => {
        this.setState({ linkValid: false });
      });
  }

  onPasswordSubmit = (e: any) => {
    e.preventDefault();
    const { id } = this.props.router.query;
    if (!id || this.state.newPassword.length < 8) {
      return;
    }
    this.setState({ loading: true, complete: false, success: false });
    let payload = {
      password: this.state.newPassword,
    };
    Ajax.postData("/auth/setpw/" + id, payload)
      .then((res) => {
        if (res.status >= 200 && res.status <= 299) {
          this.setState({ loading: false, complete: true, success: true });
        } else {
          this.setState({ loading: false, complete: true, success: false });
        }
      })
      .catch((e) => {
        this.setState({ loading: false, complete: true, success: false });
      });
  };

  render() {
    // Show loading state while checking link validity
    if (this.state.linkValid === null) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <p>{this.props.t("loading")}</p>
          </div>
        </div>
      );
    }

    // Show error if link is expired or invalid
    if (this.state.linkValid === false) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <p>{this.props.t("inviteLinkExpired")}</p>
            <p>
              <Link href="/login" className="btn btn-primary">
                {this.props.t("proceedToLogin")}
              </Link>
            </p>
          </div>
        </div>
      );
    }

    if (this.state.complete && this.state.success) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <p>{this.props.t("passwordChanged")}</p>
            <p>
              <Link href="/login" className="btn btn-primary">
                {this.props.t("proceedToLogin")}
              </Link>
            </p>
          </div>
        </div>
      );
    }

    return (
      <div className="container-center">
        <Form
          className="container-center-inner"
          onSubmit={this.onPasswordSubmit}
        >
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <p>{this.props.t("welcomeSetPassword")}</p>
          <Form.Group>
            <Form.Control
              type="password"
              placeholder={this.props.t("newPassword")}
              value={this.state.newPassword}
              onChange={(e: any) =>
                this.setState({ newPassword: e.target.value, complete: false })
              }
              required={true}
              autoFocus={true}
              minLength={8}
              disabled={this.state.loading}
              isInvalid={this.state.complete && !this.state.success}
            />
            <Form.Control.Feedback type="invalid">
              {this.props.t("errorInvalidPassword")}
            </Form.Control.Feedback>
          </Form.Group>
          <Button
            className="margin-top-10"
            variant="primary"
            type="submit"
            disabled={this.state.loading}
          >
            {this.props.t("setPassword")}
          </Button>
        </Form>
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(CompleteUserInvitation as any));
