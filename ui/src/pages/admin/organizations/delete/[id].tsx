import React from "react";
import { Button, Form } from "react-bootstrap";
import { NextRouter } from "next/router";
import withReadyRouter from "@/components/withReadyRouter";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import Ajax from "@/util/Ajax";

interface State {
  loading: boolean;
  complete: boolean;
  success: boolean;
  orgContactEmail: string;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class CompleteOrgDeletion extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      loading: false,
      complete: false,
      success: false,
      orgContactEmail: "",
    };
  }

  onDeleteOrgSubmit = (e: any) => {
    e.preventDefault();
    const { id } = this.props.router.query;
    if (!id || this.state.orgContactEmail.length < 3) {
      return;
    }
    this.setState({ loading: true, complete: false, success: false });
    let payload = {
      orgContactEmail: this.state.orgContactEmail,
    };

    Ajax.postData("/organization/deleteorg/" + id, payload)
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
    if (this.state.complete && this.state.success) {
      return (
        <div className="container-center">
          <div className="container-center-inner">
            <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
            <p>{this.props.t("confirmDeleteOrgSuccess")}</p>
          </div>
        </div>
      );
    }

    return (
      <div className="container-center">
        <Form
          className="container-center-inner"
          onSubmit={this.onDeleteOrgSubmit}
        >
          <img src="/ui/seatsurfing.svg" alt="Seatsurfing" className="logo" />
          <Form.Group>
            <Form.Control
              type="email"
              placeholder={this.props.t("emailPlaceholder")}
              value={this.state.orgContactEmail}
              onChange={(e: any) =>
                this.setState({
                  orgContactEmail: e.target.value,
                  complete: false,
                })
              }
              required={true}
              autoFocus={true}
              minLength={3}
              disabled={this.state.loading}
              isInvalid={this.state.complete && !this.state.success}
            />
            <Form.Control.Feedback type="invalid">
              {this.props.t("confirmDeleteOrgError")}
            </Form.Control.Feedback>
          </Form.Group>
          <Button
            className="margin-top-10"
            variant="primary"
            type="submit"
            disabled={this.state.loading}
          >
            {this.props.t("confirmDeleteOrg")}
          </Button>
        </Form>
      </div>
    );
  }
}

export default withTranslation(withReadyRouter(CompleteOrgDeletion as any));
