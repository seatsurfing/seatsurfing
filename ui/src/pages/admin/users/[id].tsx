import React from "react";
import { Form, Col, Row, Button, Alert, InputGroup } from "react-bootstrap";
import {
  ChevronLeft as IconBack,
  Save as IconSave,
  Trash2 as IconDelete,
  RefreshCw as IconRefresh,
  Clipboard as IconCopy,
  Check as IconCheck,
} from "react-feather";
import { NextRouter } from "next/router";
import FullLayout from "@/components/FullLayout";
import Link from "next/link";
import Loading from "@/components/Loading";
import withReadyRouter from "@/components/withReadyRouter";
import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import User from "@/types/User";
import Ajax from "@/util/Ajax";
import OrgSettings from "@/types/Settings";
import RedirectUtil from "@/util/RedirectUtil";
import AuthProvider from "@/types/AuthProvider";

interface State {
  loading: boolean;
  submitting: boolean;
  saved: boolean;
  error: boolean;
  goBack: boolean;
  email: string;
  firstname: string;
  lastname: string;
  requirePassword: boolean;
  password: string;
  changePassword: boolean;
  authMethod: string; // "password" | "provider" | "invitation"
  authProviderId: string;
  sendInvitation: boolean;
  role: number;
  showUsernameCopied: boolean;
  showPasswordCopied: boolean;
}

interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

class EditUser extends React.Component<Props, State> {
  entity: User = new User();
  authProviders: AuthProvider[] = [];
  usersMax: number = 0;
  usersCur: number = -1;
  adminUserRole: number = 0;

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      goBack: false,
      email: "",
      firstname: "",
      lastname: "",
      requirePassword: false,
      password: "",
      changePassword: false,
      authMethod: "password",
      authProviderId: "",
      sendInvitation: false,
      role: User.UserRoleUser,
      showUsernameCopied: false,
      showPasswordCopied: false,
    };
  }

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    this.loadData();
  };

  isServiceAccount = (role: number) => {
    return (
      role === User.UserRoleServiceAccountRO ||
      role === User.UserRoleServiceAccountRW
    );
  };

  loadData = () => {
    let promises: Promise<any>[] = [
      OrgSettings.getOne("feature_no_user_limit"),
      User.getCount(),
      User.getSelf().then((me) => {
        return [me];
      }),
      AuthProvider.list(),
    ];
    const { id } = this.props.router.query;
    if (id && typeof id === "string" && id !== "add") {
      promises.push(User.get(id));
    }
    Promise.all(promises).then((values) => {
      this.usersMax = values[0] === "1" ? 1000000 : 10;
      this.usersCur = values[1];
      this.adminUserRole = values[2][0].role;
      this.authProviders = values[3];
      if (values.length >= 5) {
        let user = values[4];
        this.entity = user;
        // Determine auth method from user data
        let authMethod = "password";
        if (user.passwordPending) {
          authMethod = "invitation";
        } else if (user.authProviderId) {
          authMethod = "provider";
        } else if (user.requirePassword) {
          authMethod = "password";
        }
        this.setState({
          email: user.email,
          firstname: user.firstname,
          lastname: user.lastname,
          requirePassword: user.requirePassword,
          authMethod: authMethod,
          authProviderId: user.authProviderId || "",
          role: user.role,
        });
      }
      this.setState({
        loading: false,
      });
    });
  };

  onSubmit = (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false,
    });
    this.entity.email = this.state.email;
    this.entity.firstname = this.state.firstname;
    this.entity.lastname = this.state.lastname;
    this.entity.role = this.state.role;

    // Set authentication fields based on selected auth method
    if (this.state.authMethod === "invitation") {
      this.entity.sendInvitation = true;
      this.entity.password = "";
      this.entity.authProviderId = "";
    } else if (this.state.authMethod === "provider") {
      this.entity.sendInvitation = false;
      this.entity.password = "";
      this.entity.authProviderId = this.state.authProviderId;
    } else {
      // password method
      this.entity.sendInvitation = false;
      this.entity.authProviderId = "";
      if (this.state.changePassword || !this.entity.id) {
        this.entity.password = this.state.password;
      } else {
        this.entity.password = "";
      }
    }

    this.entity
      .save()
      .then(() => {
        this.props.router.push("/admin/users/" + this.entity.id);
        this.setState({ saved: true });
      })
      .catch(() => {
        this.setState({ error: true });
      });
  };

  deleteItem = () => {
    if (window.confirm(this.props.t("confirmDeleteUser"))) {
      this.entity.delete().then(() => {
        this.setState({ goBack: true });
      });
    }
  };

  generatePassword = () => {
    const length = 32;
    const charset =
      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    let password = "";
    for (let i = 0, n = charset.length; i < length; ++i) {
      password += charset.charAt(Math.floor(Math.random() * n));
    }
    this.setState({ password, changePassword: true });
  };

  copyUsernameToClipboard = () => {
    navigator.clipboard
      .writeText(RuntimeConfig.INFOS.organizationId + "_" + this.state.email)
      .then(() => {
        this.setState({ showUsernameCopied: true });
        setTimeout(() => {
          this.setState({ showUsernameCopied: false });
        }, 2000);
      })
      .catch(() => {
        // do nothing
      });
  };

  copyPasswordToClipboard = () => {
    navigator.clipboard
      .writeText(this.state.password)
      .then(() => {
        this.setState({ showPasswordCopied: true });
        setTimeout(() => {
          this.setState({ showPasswordCopied: false });
        }, 2000);
      })
      .catch(() => {
        // do nothing
      });
  };

  changeRole = (role: number) => {
    let changePassword = this.isServiceAccount(role)
      ? true
      : this.state.changePassword;
    this.setState({ role: role, changePassword });
    if (changePassword) {
      this.generatePassword();
    }
  };

  render() {
    if (this.state.goBack) {
      this.props.router.push("/admin/users");
      return <></>;
    }

    let backButton = (
      <Link href="/admin/users" className="btn btn-sm btn-outline-secondary">
        <IconBack className="feather" /> {this.props.t("back")}
      </Link>
    );
    let buttons = backButton;

    if (this.state.loading) {
      return (
        <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
          <Loading />
        </FullLayout>
      );
    }

    if (this.usersCur >= this.usersMax && !this.entity.id) {
      return (
        <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
          <p>{this.props.t("errorSubscriptionLimit")}</p>
        </FullLayout>
      );
    }

    let hint = <></>;
    if (this.state.saved) {
      hint = <Alert variant="success">{this.props.t("entryUpdated")}</Alert>;
    } else if (this.state.error) {
      hint = <Alert variant="danger">{this.props.t("errorSave")}</Alert>;
    }

    const buttonDelete = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        onClick={this.deleteItem}
        disabled={RuntimeConfig.INFOS.userId === this.entity.id}
      >
        <IconDelete className="feather" /> {this.props.t("delete")}
      </Button>
    );
    const buttonSave = (
      <Button
        className="btn-sm"
        variant="outline-secondary"
        type="submit"
        form="form"
      >
        <IconSave className="feather" /> {this.props.t("save")}
      </Button>
    );
    if (this.entity.id) {
      buttons = (
        <>
          {backButton} {buttonDelete} {buttonSave}
        </>
      );
    } else {
      buttons = (
        <>
          {backButton} {buttonSave}
        </>
      );
    }
    let roleSelect = <></>;
    if (this.adminUserRole >= this.state.role) {
      roleSelect = (
        <Form.Select
          value={this.state.role}
          onChange={(e: any) => this.changeRole(parseInt(e.target.value))}
          required={true}
        >
          <option value={User.UserRoleUser}>{this.props.t("roleUser")}</option>
          {this.adminUserRole >= User.UserRoleSpaceAdmin ? (
            <option value={User.UserRoleSpaceAdmin}>
              {this.props.t("roleSpaceAdmin")}
            </option>
          ) : (
            <></>
          )}
          {this.adminUserRole >= User.UserRoleOrgAdmin ? (
            <option value={User.UserRoleOrgAdmin}>
              {this.props.t("roleOrgAdmin")}
            </option>
          ) : (
            <></>
          )}
          {this.adminUserRole >= User.UserRoleOrgAdmin ? (
            <option value={User.UserRoleServiceAccountRO}>
              {this.props.t("roleServiceAccountRO")}
            </option>
          ) : (
            <></>
          )}
          {this.adminUserRole >= User.UserRoleOrgAdmin ? (
            <option value={User.UserRoleServiceAccountRW}>
              {this.props.t("roleServiceAccountRW")}
            </option>
          ) : (
            <></>
          )}
          {this.adminUserRole >= User.UserRoleSuperAdmin ? (
            <option value={User.UserRoleSuperAdmin}>
              {this.props.t("roleSuperAdmin")}
            </option>
          ) : (
            <></>
          )}
        </Form.Select>
      );
    } else {
      let role = this.props.t("roleUser");
      if (this.state.role === User.UserRoleSpaceAdmin) {
        role = this.props.t("roleSpaceAdmin");
      }
      if (this.state.role === User.UserRoleOrgAdmin) {
        role = this.props.t("roleOrgAdmin");
      }
      if (this.state.role === User.UserRoleServiceAccountRO) {
        role = this.props.t("roleServiceAccountRO");
      }
      if (this.state.role === User.UserRoleServiceAccountRW) {
        role = this.props.t("roleServiceAccountRW");
      }
      if (this.state.role === User.UserRoleSuperAdmin) {
        role = this.props.t("roleSuperAdmin");
      }
      roleSelect = (
        <Form.Control plaintext={true} readOnly={true} defaultValue={role} />
      );
    }
    let copyPasswordIcon = <IconCopy className="feather" />;
    if (this.state.showPasswordCopied) {
      copyPasswordIcon = <IconCheck className="feather" />;
    }
    let copyUsernameIcon = <IconCopy className="feather" />;
    if (this.state.showUsernameCopied) {
      copyUsernameIcon = <IconCheck className="feather" />;
    }
    return (
      <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label column sm="2">
              {this.props.t("role")}
            </Form.Label>
            <Col sm="4">{roleSelect}</Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">
              {this.props.t("emailAddress")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                type="email"
                placeholder="some@domain.com"
                value={this.state.email}
                onChange={(e: any) => this.setState({ email: e.target.value })}
                required={true}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">
              {this.props.t("firstname")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                type="firstname"
                placeholder=""
                value={this.state.firstname}
                onChange={(e: any) =>
                  this.setState({ firstname: e.target.value })
                }
                required={true}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label column sm="2">
              {this.props.t("lastname")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                type="lastname"
                placeholder=""
                value={this.state.lastname}
                onChange={(e: any) =>
                  this.setState({ lastname: e.target.value })
                }
                required={true}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row} hidden={this.isServiceAccount(this.state.role)}>
            <Form.Label column sm="2">
              {this.props.t("username")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  type="text"
                  readOnly={true}
                  defaultValue={
                    this.isServiceAccount(this.state.role)
                      ? RuntimeConfig.INFOS.organizationId +
                        "_" +
                        this.state.email
                      : this.state.email
                  }
                />
                <Button
                  onClick={() => this.copyUsernameToClipboard()}
                  disabled={this.state.email === ""}
                  variant="outline-secondary"
                >
                  {copyUsernameIcon}
                </Button>
              </InputGroup>
            </Col>
          </Form.Group>

          {/* Auth method selection for non-service accounts */}
          <Form.Group
            as={Row}
            hidden={
              this.isServiceAccount(this.state.role) ||
              RuntimeConfig.INFOS.disablePasswordLogin
            }
          >
            <Form.Label column sm="2">
              {this.props.t("authMethod")}
            </Form.Label>
            <Col sm="4">
              <Form.Check
                type="radio"
                id="auth-method-password"
                name="authMethod"
                label={this.props.t("authMethodPassword")}
                checked={this.state.authMethod === "password"}
                onChange={() => this.setState({ authMethod: "password" })}
              />
              {this.authProviders.length > 0 && (
                <Form.Check
                  type="radio"
                  id="auth-method-provider"
                  name="authMethod"
                  label={this.props.t("authMethodProvider")}
                  checked={this.state.authMethod === "provider"}
                  onChange={() => this.setState({ authMethod: "provider" })}
                />
              )}
              <Form.Check
                type="radio"
                id="auth-method-invitation"
                name="authMethod"
                label={this.props.t("authMethodInvitation")}
                checked={this.state.authMethod === "invitation"}
                onChange={() => this.setState({ authMethod: "invitation" })}
              />
            </Col>
          </Form.Group>

          {/* Auth provider selection */}
          <Form.Group
            as={Row}
            hidden={
              this.isServiceAccount(this.state.role) ||
              this.state.authMethod !== "provider" ||
              RuntimeConfig.INFOS.disablePasswordLogin
            }
          >
            <Form.Label column sm="2">
              {this.props.t("chooseAuthProvider")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                value={this.state.authProviderId}
                onChange={(e: any) =>
                  this.setState({ authProviderId: e.target.value })
                }
                required={this.state.authMethod === "provider"}
              >
                <option value="">{this.props.t("pleaseSelect")}</option>
                {this.authProviders.map((provider) => (
                  <option key={provider.id} value={provider.id}>
                    {provider.name}
                  </option>
                ))}
              </Form.Select>
            </Col>
          </Form.Group>

          {/* Password change checkbox for existing users */}
          <Form.Group
            as={Row}
            hidden={
              this.isServiceAccount(this.state.role) ||
              !this.entity.id ||
              this.state.authMethod !== "password" ||
              RuntimeConfig.INFOS.disablePasswordLogin
            }
          >
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-changePassword"
                label={this.props.t("passwordChange")}
                checked={this.state.changePassword}
                onChange={(e: any) =>
                  this.setState({ changePassword: e.target.checked })
                }
              />
            </Col>
          </Form.Group>

          {/* Password field */}
          <Form.Group
            as={Row}
            hidden={
              (RuntimeConfig.INFOS.disablePasswordLogin &&
                !this.isServiceAccount(this.state.role)) ||
              (!this.isServiceAccount(this.state.role) &&
                this.state.authMethod !== "password")
            }
          >
            <Form.Label column sm="2">
              {this.props.t("password")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  type={
                    this.isServiceAccount(this.state.role) ? "text" : "password"
                  }
                  value={this.state.password}
                  onChange={(e: any) =>
                    this.setState({ password: e.target.value })
                  }
                  required={
                    this.isServiceAccount(this.state.role) ||
                    (!this.entity.id && this.state.authMethod === "password") ||
                    (this.entity.id &&
                      this.state.changePassword &&
                      this.state.authMethod === "password")
                  }
                  disabled={
                    (!this.isServiceAccount(this.state.role) &&
                      this.entity.id &&
                      !this.state.changePassword) ||
                    this.isServiceAccount(this.state.role)
                  }
                  minLength={this.isServiceAccount(this.state.role) ? 32 : 8}
                />
                <Button
                  onClick={() => this.generatePassword()}
                  hidden={!this.isServiceAccount(this.state.role)}
                  variant="outline-secondary"
                >
                  <IconRefresh className="feather" />
                </Button>
                <Button
                  onClick={() => this.copyPasswordToClipboard()}
                  disabled={this.state.password === ""}
                  hidden={!this.isServiceAccount(this.state.role)}
                  variant="outline-secondary"
                >
                  {copyPasswordIcon}
                </Button>
              </InputGroup>
            </Col>
          </Form.Group>
        </Form>
      </FullLayout>
    );
  }
}

export default withTranslation(withReadyRouter(EditUser as any));
