import withReadyRouter from '@/components/withReadyRouter';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import React from 'react';
import { Button } from 'react-bootstrap';
import { Copy as IconCopy } from 'react-feather';

interface State {
}

interface Props extends WithTranslation {
  router: NextRouter
}

class ConfluenceHint extends React.Component<Props, State> {
  onCreateAccountClick = (e: any) => {
    e.preventDefault();
    if (typeof window !== 'undefined') {
      window.open("https://seatsurfing.io/sign-up/");
    }
  }

  onAdminClick = (e: any) => {
    e.preventDefault();
    if (typeof window !== 'undefined') {
      window.open("https://app.seatsurfing.io/admin/");
    }
  }

  onHelpClick = (e: any) => {
    e.preventDefault();
    if (typeof window !== 'undefined') {
      window.open("https://seatsurfing.io/contact/");
    }
  }

  onInputClick = (e: any) => {
    e.preventDefault();
    let input = document.querySelector("input.copy-input");
    if (input) {
      (input as HTMLInputElement).select();
    }
  }

  onCopyClick = () => {
    let input = document.querySelector("input.copy-input");
    if (input) {
      (input as HTMLInputElement).select();
      document.execCommand("copy");
    }
  }

  render() {
    const { id } = this.props.router.query;
    return (
      <div className="container-confluence">
          <h1>{this.props.t("errorConfluenceClientIdUnknown")}</h1>
          <p>{this.props.t("confluenceClientIdHint")}</p>
          <ol>
            <li><Button variant="link" className="button-link" onClick={this.onCreateAccountClick}>{this.props.t("confluenceClientIdStep1")}</Button></li>
            <li><Button variant="link" className="button-link" onClick={this.onAdminClick}>{this.props.t("confluenceClientIdStep2")}</Button></li>
            <li>{this.props.t("confluenceClientIdStep3")}</li>
            <li>
              {this.props.t("confluenceClientIdStep4")}
              <br />
              <input type="text" className="copy-input" size={36} onClick={this.onInputClick} value={id} readOnly={true} />
              <Button variant="link" size="sm" className="copy-button" onClick={this.onCopyClick}><IconCopy className="feather" /></Button>
            </li>
            <li>{this.props.t("confluenceClientIdStep5")}</li>
            <li>{this.props.t("confluenceClientIdStep6")}</li>
          </ol>
          <p>{this.props.t("confluenceClientIdHint2")} <Button variant="link" className="button-link" onClick={this.onHelpClick}>{this.props.t("confluenceClientIdHint3")}</Button></p>
      </div>
    )
  }
}

export default withTranslation()(withReadyRouter(ConfluenceHint as any));
