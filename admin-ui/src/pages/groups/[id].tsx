import React from 'react';
import { Form, Col, Row, Button, Alert, Table, InputGroup } from 'react-bootstrap';
import { ChevronLeft as IconBack, Save as IconSave, Trash2 as IconDelete } from 'react-feather';
import { Ajax, Group, Search, SearchOptions, User } from 'seatsurfing-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import { AsyncTypeahead } from 'react-bootstrap-typeahead';
import FullLayout from '@/components/FullLayout';
import Link from 'next/link';
import Loading from '@/components/Loading';
import withReadyRouter from '@/components/withReadyRouter';
import 'react-bootstrap-typeahead/css/Typeahead.css';
import ProfilePicture from '@/components/ProfilePicture';

interface State {
  loading: boolean
  typeaheadOptions: any[]
  typeaheadLoading: boolean
  submitting: boolean
  saved: boolean
  error: boolean
  goBack: boolean
  name: string
  addUserIds: string[]
  members: User[]
  removeUserIds: string[]
}

interface Props extends WithTranslation {
  router: NextRouter
}

class EditUser extends React.Component<Props, State> {
  entity: Group = new Group();
  typeahead: any = null;

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      typeaheadOptions: [],
      typeaheadLoading: false,
      submitting: false,
      saved: false,
      error: false,
      goBack: false,
      name: "",
      addUserIds: [],
      members: [],
      removeUserIds: []
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
    let promises: Promise<any>[] = [
    ];
    const { id } = this.props.router.query;
    if (id && (typeof id === "string") && (id !== 'add')) {
      promises.push(Group.get(id));
    }
    Promise.all(promises).then(values => {
      if (values.length >= 1) {
        let group = values[0];
        this.entity = group;
        this.loadMembers().then(() => {
          this.setState({
            name: group.name,
          });
        });
      }
      this.setState({
        loading: false
      });
    });
  }

  loadMembers = () => {
    return this.entity.getMembers().then((members) => {
      this.setState({
        members: members,
      });
    });
  }

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false
    });
    this.entity.name = this.state.name;
    this.entity.save().then((e) => {
      this.entity.id = e.id;
      this.props.router.push("/groups/" + this.entity.id);
      this.setState({ saved: true });
    }).catch(() => {
      this.setState({ error: true });
    });
  }

  deleteItem = () => {
    if (window.confirm(this.props.t("confirmDeleteGroup"))) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  }

  filterSearch = () => {
    return true;
  }

  onSearchSelected = (selected: any) => {
    this.setState({
      addUserIds: selected.map((user: any) => user.id)
    });
  }

  handleSearch = (query: string) => {
    this.setState({ typeaheadLoading: true });
    let options = new SearchOptions();
    options.includeUsers = true;
    Search.search(query ? query : "", options).then(res => {
      this.setState({
        typeaheadOptions: res.users,
        typeaheadLoading: false
      });
    });
  }

  addMembers = () => {
    if (this.typeahead !== null) {
      this.entity.addMembers(this.state.addUserIds).then(() => {
        this.typeahead.clear();
        this.setState({
          typeaheadOptions: [],
          typeaheadLoading: false,
          addUserIds: []
        });
        this.loadMembers();
      }
      ).catch(() => {
        this.setState({
          typeaheadLoading: false
        });
      });
    }
  }

  getMemberRow = (user: User) => {
    return (
      <tr key={user.id}>
        <td style={{ tableLayout: "fixed", width: "20px" }}>
          <Form.Check type="checkbox" onChange={(e: any) => this.selectMember(user.id, e.target.checked)} checked={this.state.removeUserIds.includes(user.id)} />
        </td>
        <td style={{ tableLayout: "fixed", width: "64px" }}>
          <ProfilePicture width={48} height={48} />
        </td>
        <td style={{ tableLayout: "auto"  }}>
          <span style={{ marginLeft: "10px" }}>{user.email}</span>
        </td>
      </tr>
    );
  }

  selectMember = (userId: string, checked: boolean) => {
    let removeUserIds = this.state.removeUserIds;
    if (checked && !removeUserIds.includes(userId)) {
      removeUserIds.push(userId);
    } else if (!checked && removeUserIds.includes(userId)) {
      removeUserIds.splice(removeUserIds.indexOf(userId), 1);
    }
    this.setState({
      removeUserIds: removeUserIds
    });
  }

  persistRemoveMembers = () => {
    this.entity.removeMembers(this.state.removeUserIds).then(() => {
      this.setState({
        removeUserIds: []
      });
      this.loadMembers();
    });
  }

  render() {
    if (this.state.goBack) {
      this.props.router.push('/groups');
      return <></>
    }

    let backButton = <Link href="/groups" className="btn btn-sm btn-outline-secondary"><IconBack className="feather" /> {this.props.t("back")}</Link>;
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editGroup")} buttons={buttons}>
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

    let memberTable = <></>
    if (this.entity.id) {
      memberTable = (
        <>
          <div className="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom" style={{ "marginTop": "50px" }}>
            <h4>{this.props.t("members")}</h4>
          </div>
          <Form>
            <Form.Group as={Row}>
              <Col sm="6">
                <InputGroup>
                  <AsyncTypeahead
                    filterBy={this.filterSearch}
                    id="search-users"
                    isLoading={this.state.typeaheadLoading}
                    labelKey="email"
                    multiple={true}
                    minLength={3}
                    onChange={this.onSearchSelected}
                    onSearch={this.handleSearch}
                    options={this.state.typeaheadOptions}
                    placeholder={this.props.t("searchForUser")}
                    ref={(ref: any) => { this.typeahead = ref; }}
                    renderMenuItemChildren={(option: any) => (
                      <div className="d-flex">
                        <ProfilePicture width={24} height={24} />
                        <span style={{ marginLeft: "10px" }}>{option.email}</span>
                      </div>
                    )}
                  />
                  <Button
                    onClick={() => { this.addMembers() }}
                    variant="outline-secondary">
                    {this.props.t("add")}
                  </Button>
                </InputGroup>
              </Col>
            </Form.Group>
          </Form>
          <Table>
            <tbody>
              {this.state.members.map((user: User) => this.getMemberRow(user))}
            </tbody>
          </Table>
          <Button className="btn-sm" variant="outline-secondary" hidden={this.state.removeUserIds.length === 0} onClick={() => { this.persistRemoveMembers() }}>
            {this.props.t("remove")}
          </Button>
        </>
      );
    }

    return (
      <FullLayout headline={this.props.t("editGroup")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">{this.props.t("name")}</Form.Label>
            <Col sm="4">
              <Form.Control type="name" value={this.state.name} onChange={(e: any) => this.setState({ name: e.target.value })} required={true} />
            </Col>
          </Form.Group>
        </Form>
        {memberTable}
      </FullLayout>
    );
  }
}

export default withTranslation(['admin'])(withReadyRouter(EditUser as any));
