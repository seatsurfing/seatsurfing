import React from 'react';
import { Form, Col, Row, Button, Alert } from 'react-bootstrap';
import { ChevronLeft as IconBack, Save as IconSave, Trash2 as IconDelete } from 'react-feather';
import { Ajax, SpaceAttribute } from 'seatsurfing-commons';
import { NextRouter } from 'next/router';
import { WithTranslation, withTranslation } from 'next-i18next';
import FullLayout from '@/components/FullLayout';
import Loading from '@/components/Loading';
import Link from 'next/link';
import withReadyRouter from '@/components/withReadyRouter';

interface State {
  loading: boolean
  submitting: boolean
  saved: boolean
  error: boolean
  goBack: boolean
  label: string
  type: number;
  spaceApplicable: boolean
  locationApplicable: boolean
}

interface Props extends WithTranslation {
  router: NextRouter
}

class EditAttribute extends React.Component<Props, State> {
  entity: SpaceAttribute = new SpaceAttribute();

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      goBack: false,
      label: "",
      type: 1, 
      spaceApplicable: false, 
      locationApplicable: false,
    };
  }

  componentDidMount = () => {
    if (!Ajax.CREDENTIALS.accessToken) {
      this.props.router.push("/login");
      return;
    }
    this.loadData();
  }

  loadData = () => {
    const { id } = this.props.router.query;
    if (id && (typeof id === "string") && (id !== 'add')) {
      SpaceAttribute.get(id).then(e => {
        this.entity = e;
        this.setState({
          label: e.label,
          type: Number(e.type),
          spaceApplicable: e.spaceApplicable,
          locationApplicable: e.locationApplicable,
          loading: false,
        });
      });
    } else {
      this.setState({ loading: false });
    }
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false
    });
    this.entity.label = this.state.label;
    this.entity.type = Number(this.state.type);
    this.entity.spaceApplicable = this.state.spaceApplicable;
    this.entity.locationApplicable = this.state.locationApplicable;
    this.entity.save().then(() => {
      this.props.router.push("/attributes/" + this.entity.id);
      this.setState({ saved: true });
    }).catch(() => {
      this.setState({ error: true });
    });
  }

  deleteItem = () => {
    if (window.confirm(this.props.t("confirmDeleteAttribute"))) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  }

  render() {
    if (this.state.goBack) {
      this.props.router.push('/attributes');
      return <></>
    }

    let backButton = <Link href="/attributes" className="btn btn-sm btn-outline-secondary"><IconBack className="feather" /> {this.props.t("back")}</Link>;
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editAttribute")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.props.t("errorSave")}</Alert>
    }

    let buttonDelete = <Button className="btn-sm" variant="outline-secondary" onClick={this.deleteItem} disabled={false}><IconDelete className="feather" /> {this.props.t("delete")}</Button>;
    let buttonSave = <Button className="btn-sm" variant="outline-secondary" type="submit" form="form"><IconSave className="feather" /> {this.props.t("save")}</Button>;
    if (this.entity.id) {
      buttons = <>{backButton} {buttonDelete} {buttonSave}</>;
    } else {
      buttons = <>{backButton} {buttonSave}</>;
    }

    return (
      <FullLayout headline={this.props.t("editAttribute")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("name")}</Form.Label>
            <Col sm="4">
              <Form.Control type="text" value={this.state.label} onChange={(e: any) => this.setState({ label: e.target.value })} required={true} autoFocus={true} />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("type")}</Form.Label>
            <Col sm="4">
              <Form.Select value={this.state.type} onChange={(e: any) => this.setState({ type: e.target.value })}>
                <option value="1">{this.props.t("number")}</option>
                <option value="2">{this.props.t("boolean")}</option>
                <option value="3">{this.props.t("text")}</option>
              </Form.Select>
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("applicableTo")}</Form.Label>
            <Col sm="4">
              <Form.Check type="checkbox" id="check-locationApplicable" label={this.props.t("areas")} checked={this.state.locationApplicable} onChange={(e: any) => this.setState({ locationApplicable: e.target.checked })} />
              <Form.Check type="checkbox" id="check-spaceApplicable" label={this.props.t("spaces")} checked={this.state.spaceApplicable} onChange={(e: any) => this.setState({ spaceApplicable: e.target.checked })} />
            </Col>
          </Form.Group>
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(EditAttribute as any));
