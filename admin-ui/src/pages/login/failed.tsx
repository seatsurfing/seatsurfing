import { TranslationFunc, withTranslation } from '@/components/withTranslation';
import Link from 'next/link';
import React from 'react';
import { Form, Alert } from 'react-bootstrap';

interface State {
}

interface Props {
  t: TranslationFunc;
}

class LoginFailed extends React.Component<Props, State> {
  render() {
    return (
      <div className="container-signin">
        <Form className="form-signin">
          <Alert variant="danger">{this.props.t("errorLoginFailed")}</Alert>
          <p>{this.props.t("loginFailedDescription")}</p>
          <Link className="btn btn-primary" href="/login">{this.props.t("back")}</Link>
        </Form>
      </div>
    )
  }
}

export default withTranslation(LoginFailed as any);
