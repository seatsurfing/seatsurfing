import React, { CSSProperties, use } from "react";
import { TranslationFunc, withTranslation } from "./withTranslation";
import { IoChatboxOutline, IoSend } from "react-icons/io5";
import styles from "./FeedbackButton.module.css";
import RuntimeConfig from "./RuntimeConfig";
import { Button, Form } from "react-bootstrap";
import Ajax from "@/util/Ajax";

interface State {
  feedbackText: string;
  submitting: boolean;
  success: boolean;
  error: boolean;
  showForm: boolean;
}

interface Props {
  t: TranslationFunc;
  hidden?: boolean;
  style?: CSSProperties | undefined;
}

class FeedbackButton extends React.Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      feedbackText: "",
      submitting: false,
      success: false,
      error: false,
      showForm: false,
    };
  }

  showFeedbackForm = () => {
    this.setState({
      showForm: !this.state.showForm,
      feedbackText: "",
      success: false,
      error: false,
      submitting: false,
    });
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({ submitting: true, success: false, error: false });
    let payload = {
      ua: navigator.userAgent,
      text: this.state.feedbackText,
    };
    Ajax.postData("/feedback/", payload)
      .then(() => {
        this.setState({ feedbackText: "", success: true, submitting: false });
      })
      .catch(() => {
        this.setState({ error: true, submitting: false });
      });
  };

  render() {
    if (!RuntimeConfig.INFOS.cloudHosted) {
      return <></>;
    }
    return (
      <>
        <div
          className={styles.feedbackButtonContainer}
          hidden={!this.state.showForm}
        >
          <p style={{ fontWeight: "bold" }}>
            {this.props.t("cloudFeedbackHeadline")}
          </p>
          <p hidden={this.state.success}>
            {this.props.t("cloudFeedbackPrompt")}
          </p>
          <Form onSubmit={this.onSubmit}>
            <p hidden={!this.state.success}>
              {this.props.t("cloudFeedbackThanks")} 🙏
            </p>
            <Form.Control
              hidden={this.state.success}
              as="textarea"
              minLength={10}
              required={true}
              rows={3}
              placeholder={this.props.t("cloudFeedbackPlaceholder")}
              value={this.state.feedbackText}
              onChange={(e) => this.setState({ feedbackText: e.target.value })}
              disabled={this.state.submitting}
            />
            <p hidden={!this.state.error} style={{ marginTop: "10px" }}>
              {this.props.t("cloudFeedbackError")}
            </p>
            <Button
              type="submit"
              variant="link"
              className={styles.feedbackSendButton}
              hidden={this.state.success}
              disabled={this.state.submitting}
            >
              {this.props.t("send")} <IoSend className="feather" />
            </Button>
          </Form>
        </div>
        <div
          id="feedback-button"
          className={
            this.state.showForm
              ? styles.feedbackButtonActive
              : styles.feedbackButton
          }
          onClick={() => this.showFeedbackForm()}
        >
          <IoChatboxOutline className={styles.feedbackButtonIcon} />
        </div>
      </>
    );
  }
}

export default withTranslation(FeedbackButton as any);
