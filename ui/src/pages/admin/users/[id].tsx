import React from "react";
import {
  Form,
  Col,
  Row,
  Button,
  Alert,
  InputGroup,
  Modal,
} from "react-bootstrap";
import {
  ChevronLeft as IconBack,
  Save as IconSave,
  Trash2 as IconDelete,
  RefreshCw as IconRefresh,
  XCircle as IconReset,
} from "react-feather";
import { NextRouter } from "next/router";
import FullLayout from "@/components/FullLayout";
import Link from "next/link";
import Loading from "@/components/Loading";
import CopyToClipboardButton from "@/components/CopyToClipboardButton";
import withReadyRouter from "@/components/withReadyRouter";
import withPermission from "@/components/withPermission";
import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";
import User from "@/types/User";
import OrgSettings from "@/types/Settings";
import Organization from "@/types/Organization";

import AuthProvider from "@/types/AuthProvider";
import ErrorText from "@/types/ErrorText";
import AjaxError from "@/util/AjaxError";
import Validation from "@/util/Validation";
import Formatting from "@/util/Formatting";
import RendererUtils from "@/util/RendererUtils";
import Role from "@/types/Role";
import { Permission, PermissionLevel } from "@/types/Permission";
import ConfirmModal from "@/components/ConfirmModal";

type PendingConfirmAction =
  "delete" | "resetPasskeys" | "resetTotp" | "revokeApiToken";

interface State {
  loading: boolean;
  submitting: boolean;
  saved: boolean;
  error: boolean;
  errorText: string;
  goBack: boolean;
  email: string;
  originalEmail: string;
  firstname: string;
  lastname: string;
  requirePassword: boolean;
  password: string;
  changePassword: boolean;
  authMethod: string; // one of User.AuthMethodPassword | User.AuthMethodProvider | User.AuthMethodInvitation
  authProviderId: string;
  sendInvitation: boolean;
  resendInvitation: boolean;
  accountType: number;
  roleIds: string[];
  totpEnabled: boolean;
  hasPasskeys: boolean;
  apiTokenConfigured: boolean;
  showApiTokenModal: boolean;
  generatedToken: string;
  lastActivity: Date | null;
  pendingConfirmAction: PendingConfirmAction | null;
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
  roles: Role[] = [];

  constructor(props: any) {
    super(props);
    this.state = {
      loading: true,
      submitting: false,
      saved: false,
      error: false,
      errorText: "",
      goBack: false,
      email: "",
      originalEmail: "",
      firstname: "",
      lastname: "",
      requirePassword: false,
      password: "",
      changePassword: false,
      authMethod: User.AuthMethodPassword,
      authProviderId: "",
      sendInvitation: false,
      resendInvitation: false,
      accountType: User.AccountTypePerson,
      roleIds: [],
      totpEnabled: false,
      hasPasskeys: false,
      apiTokenConfigured: false,
      showApiTokenModal: false,
      generatedToken: "",
      lastActivity: null,
      pendingConfirmAction: null,
    };
  }

  componentDidMount = () => {
    this.loadData();
  };

  sameRoleIds = (a: string[], b: string[]): boolean => {
    return (
      a.length === b.length && [...a].sort().join() === [...b].sort().join()
    );
  };

  toggleRole = (roleId: string, assigned: boolean) => {
    this.setState((state) => ({
      roleIds: assigned
        ? [...state.roleIds, roleId]
        : state.roleIds.filter((id) => id !== roleId),
    }));
  };

  isServiceAccount = (accountType: number) => {
    return (
      accountType === User.AccountTypeServiceAccountRO ||
      accountType === User.AccountTypeServiceAccountRW
    );
  };

  loadData = () => {
    const promises: Promise<any>[] = [
      OrgSettings.getOne(Organization.PREF_FEATURE_NO_USER_LIMIT),
      User.getCount(),
      User.getSelf().then((me) => {
        return [me];
      }),
      AuthProvider.list(),
      RuntimeConfig.hasPermission(Permission.Roles, PermissionLevel.Read)
        ? Role.list()
        : Promise.resolve([]),
    ];
    const { id } = this.props.router.query;
    if (id && typeof id === "string" && id !== "add") {
      promises.push(User.get(id));
    }
    Promise.all(promises).then(async (values) => {
      this.usersMax = values[0] === "1" ? 1000000 : 10;
      this.usersCur = values[1];
      this.authProviders = values[3];
      this.roles = values[4];
      if (values.length >= 6) {
        const user = values[5];
        this.entity = user;
        // Determine auth method from user data
        let authMethod = User.AuthMethodPassword;
        if (user.passwordPending) {
          authMethod = User.AuthMethodInvitation;
        } else if (user.authProviderId) {
          authMethod = User.AuthMethodProvider;
        } else if (user.requirePassword) {
          authMethod = User.AuthMethodPassword;
        }
        const isServiceAccount =
          user.accountType === User.AccountTypeServiceAccountRO ||
          user.accountType === User.AccountTypeServiceAccountRW;
        const apiTokenConfigured = isServiceAccount
          ? await User.getApiTokenStatus(user.id).catch(() => false)
          : false;
        this.setState({
          email: user.email,
          originalEmail: user.email,
          firstname: user.firstname,
          lastname: user.lastname,
          requirePassword: user.requirePassword,
          authMethod,
          authProviderId: user.authProviderId || "",
          accountType: user.accountType,
          roleIds: user.roleIds,
          totpEnabled: user.totpEnabled,
          hasPasskeys: user.hasPasskeys,
          apiTokenConfigured,
          lastActivity: user.lastActivity,
        });
      }
      this.setState({
        loading: false,
      });
    });
  };

  onSubmit = async (e: any) => {
    e.preventDefault();
    this.setState({
      error: false,
      saved: false,
    });
    this.entity.email = this.state.email;
    this.entity.firstname = this.state.firstname;
    this.entity.lastname = this.state.lastname;
    this.entity.accountType = this.state.accountType;

    // Set authentication fields based on selected auth method
    if (this.state.authMethod === User.AuthMethodInvitation) {
      // Only send invitation if email changed or explicitly requested
      const emailChanged = this.state.email !== this.state.originalEmail;
      const isNewUser = !this.entity.id;
      this.entity.sendInvitation =
        isNewUser || emailChanged || this.state.resendInvitation;
      this.entity.password = "";
      this.entity.authProviderId = "";
    } else if (this.state.authMethod === User.AuthMethodProvider) {
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

    try {
      await this.entity.save();
      if (
        RuntimeConfig.hasPermission(Permission.Roles, PermissionLevel.Admin) &&
        !this.sameRoleIds(this.entity.roleIds, this.state.roleIds)
      ) {
        await Role.setRoleIdsForUser(this.entity.id, this.state.roleIds);
        this.entity.roleIds = this.state.roleIds;
      }
      this.props.router.push(
        `/admin/users/${encodeURIComponent(this.entity.id)}`,
      );
      this.setState({
        saved: true,
        resendInvitation: false,
        originalEmail: this.entity.email, // don't (re)send invitation mails on second save
      });
    } catch (e) {
      let code: number = 0;
      if (e instanceof AjaxError) {
        code = e.appErrorCode;
      }
      this.setState({
        error: true,
        errorText: code
          ? ErrorText.getTextForAppCode(code, this.props.t)
          : this.props.t("errorSave"),
      });
    }
  };

  deleteItem = () => {
    this.setState({ pendingConfirmAction: "delete" });
  };

  resetPasskeys = () => {
    this.setState({ pendingConfirmAction: "resetPasskeys" });
  };

  resetTotp = () => {
    this.setState({ pendingConfirmAction: "resetTotp" });
  };

  generatePassword = () => {
    const password = Validation.generatePassword();
    this.setState({ password, changePassword: true });
  };

  changeAccountType = (accountType: number) => {
    let changePassword = this.isServiceAccount(accountType)
      ? true
      : this.state.changePassword;
    this.setState({ accountType, changePassword });
    if (changePassword) {
      this.generatePassword();
    }
  };

  generateApiToken = async () => {
    const token = await User.generateApiToken(this.entity.id);
    this.setState({
      apiTokenConfigured: true,
      showApiTokenModal: true,
      generatedToken: token,
    });
  };

  revokeApiToken = () => {
    this.setState({ pendingConfirmAction: "revokeApiToken" });
  };

  confirmPendingAction = async () => {
    const action = this.state.pendingConfirmAction;
    this.setState({ pendingConfirmAction: null });
    switch (action) {
      case "delete":
        await this.entity.delete();
        this.setState({ goBack: true });
        break;
      case "resetPasskeys":
        await User.adminResetPasskeys(this.entity.id);
        this.setState({ hasPasskeys: false });
        break;
      case "resetTotp":
        await User.adminResetTotp(this.entity.id);
        this.setState({ totpEnabled: false });
        break;
      case "revokeApiToken":
        await User.revokeApiToken(this.entity.id);
        this.setState({ apiTokenConfigured: false });
        break;
    }
  };

  getPendingConfirmMessage = () => {
    switch (this.state.pendingConfirmAction) {
      case "delete":
        return this.props.t("confirmDeleteUser");
      case "resetPasskeys":
        return this.props.t("confirmResetPasskeys");
      case "resetTotp":
        return this.props.t("confirmResetTotp");
      case "revokeApiToken":
        return this.props.t("confirmRevokeApiToken");
      default:
        return "";
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
      hint = <Alert variant="danger">{this.state.errorText}</Alert>;
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
    const isOwnUser = RuntimeConfig.INFOS.userId === this.entity.id;
    const canManageServiceAccounts = RuntimeConfig.hasPermission(
      Permission.ServiceAccounts,
      PermissionLevel.Admin,
    );
    // What remains on the user itself is how the account authenticates.
    // Access is granted by the roles assigned below.
    let accountTypeSelect = <></>;
    if (!isOwnUser && canManageServiceAccounts) {
      accountTypeSelect = (
        <Form.Select
          id="accountType"
          value={this.state.accountType}
          onChange={(e: any) =>
            this.changeAccountType(parseInt(e.target.value))
          }
          required={true}
        >
          <option value={User.AccountTypePerson}>
            {this.props.t("roleUser")}
          </option>
          <option value={User.AccountTypeServiceAccountRO}>
            {this.props.t("roleServiceAccountRO")}
          </option>
          <option value={User.AccountTypeServiceAccountRW}>
            {this.props.t("roleServiceAccountRW")}
          </option>
        </Form.Select>
      );
    } else {
      const accountTypeName = RendererUtils.accountTypeName(
        this.state.accountType,
        this.props.t,
      );
      accountTypeSelect = (
        <>
          <Form.Control
            id="accountType"
            plaintext={true}
            readOnly={true}
            defaultValue={accountTypeName}
          />
          {isOwnUser && (
            <Form.Text className="text-muted">
              {this.props.t("cannotChangeOwnAccountType")}
            </Form.Text>
          )}
        </>
      );
    }
    return (
      <FullLayout headline={this.props.t("editUser")} buttons={buttons}>
        <Form onSubmit={this.onSubmit} id="form">
          {hint}
          <Form.Group as={Row}>
            <Form.Label htmlFor="accountType" column sm="2">
              {this.props.t("accountType")}
            </Form.Label>
            <Col sm="4">{accountTypeSelect}</Col>
          </Form.Group>
          {RuntimeConfig.hasPermission(
            Permission.Roles,
            PermissionLevel.Read,
          ) && (
            <Form.Group as={Row}>
              <Form.Label column sm="2">
                {this.props.t("roles")}
              </Form.Label>
              <Col sm="4">
                {this.roles.length === 0 && (
                  <Form.Text className="text-muted">
                    {this.props.t("noRolesDefined")}
                  </Form.Text>
                )}
                {this.roles.map((role) => (
                  <Form.Check
                    key={role.id}
                    type="checkbox"
                    id={"role-" + role.id}
                    label={role.name}
                    checked={this.state.roleIds.includes(role.id)}
                    disabled={
                      !RuntimeConfig.hasPermission(
                        Permission.Roles,
                        PermissionLevel.Admin,
                      )
                    }
                    onChange={(e: any) =>
                      this.toggleRole(role.id, e.target.checked)
                    }
                  />
                ))}
                {isOwnUser && (
                  <Form.Text className="text-muted">
                    {this.props.t("cannotRemoveOwnRoleManagement")}
                  </Form.Text>
                )}
              </Col>
            </Form.Group>
          )}
          {!this.isServiceAccount(this.state.accountType) && (
            <Form.Group as={Row}>
              <Form.Label htmlFor="email" column sm="2">
                {this.props.t("emailAddress")}
              </Form.Label>
              <Col sm="4">
                <Form.Control
                  id="email"
                  type="email"
                  placeholder="some@domain.com"
                  value={this.state.email}
                  onChange={(e: any) =>
                    this.setState({ email: e.target.value })
                  }
                  required={true}
                />
              </Col>
            </Form.Group>
          )}
          <Form.Group as={Row}>
            <Form.Label htmlFor="firstname" column sm="2">
              {this.props.t("firstname")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="firstname"
                type="text"
                placeholder=""
                value={this.state.firstname}
                onChange={(e: any) =>
                  this.setState({ firstname: e.target.value })
                }
                required={true}
                minLength={2}
                maxLength={64}
                pattern={Validation.HUMAN_NAME_PATTERN}
                title={this.props.t("nameRequirements")}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label htmlFor="lastname" column sm="2">
              {this.props.t("lastname")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="lastname"
                type="text"
                placeholder=""
                value={this.state.lastname}
                onChange={(e: any) =>
                  this.setState({ lastname: e.target.value })
                }
                required={true}
                minLength={2}
                maxLength={64}
                pattern={Validation.HUMAN_NAME_PATTERN}
                title={this.props.t("nameRequirements")}
              />
            </Col>
          </Form.Group>
          <Form.Group as={Row}>
            <Form.Label htmlFor="username" column sm="2">
              {this.props.t("username")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="username"
                  type="text"
                  readOnly={!this.isServiceAccount(this.state.accountType)}
                  value={this.state.email}
                  onChange={
                    this.isServiceAccount(this.state.accountType)
                      ? (e: any) => this.setState({ email: e.target.value })
                      : undefined
                  }
                  required={this.isServiceAccount(this.state.accountType)}
                />
                <CopyToClipboardButton text={this.state.email} />
              </InputGroup>
            </Col>
          </Form.Group>

          {/* Auth method selection for non-service accounts */}
          <Form.Group
            as={Row}
            hidden={
              this.isServiceAccount(this.state.accountType) ||
              RuntimeConfig.INFOS.disablePasswordLogin
            }
          >
            <Form.Label htmlFor="auth-method-password" column sm="2">
              {this.props.t("authMethod")}
            </Form.Label>
            <Col sm="4">
              <Form.Check
                type="radio"
                id="auth-method-password"
                name="authMethod"
                label={this.props.t("authMethodPassword")}
                checked={this.state.authMethod === User.AuthMethodPassword}
                onChange={() =>
                  this.setState({ authMethod: User.AuthMethodPassword })
                }
              />
              {this.authProviders.length > 0 && (
                <Form.Check
                  type="radio"
                  id="auth-method-provider"
                  name="authMethod"
                  label={this.props.t("authMethodProvider")}
                  checked={this.state.authMethod === User.AuthMethodProvider}
                  onChange={() =>
                    this.setState({ authMethod: User.AuthMethodProvider })
                  }
                />
              )}
              <Form.Check
                type="radio"
                id="auth-method-invitation"
                name="authMethod"
                label={this.props.t("authMethodInvitation")}
                checked={this.state.authMethod === User.AuthMethodInvitation}
                onChange={() =>
                  this.setState({ authMethod: User.AuthMethodInvitation })
                }
              />
            </Col>
          </Form.Group>

          {/* Auth provider selection */}
          <Form.Group
            as={Row}
            hidden={
              this.isServiceAccount(this.state.accountType) ||
              this.state.authMethod !== User.AuthMethodProvider ||
              RuntimeConfig.INFOS.disablePasswordLogin
            }
          >
            <Form.Label htmlFor="authProvider" column sm="2">
              {this.props.t("chooseAuthProvider")}
            </Form.Label>
            <Col sm="4">
              <Form.Select
                id="authProvider"
                value={this.state.authProviderId}
                onChange={(e: any) =>
                  this.setState({ authProviderId: e.target.value })
                }
                required={this.state.authMethod === User.AuthMethodProvider}
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
              this.isServiceAccount(this.state.accountType) ||
              !this.entity.id ||
              this.state.authMethod !== User.AuthMethodPassword ||
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

          {/* Resend invitation checkbox for existing users with invitation auth method */}
          <Form.Group
            as={Row}
            hidden={
              this.isServiceAccount(this.state.accountType) ||
              !this.entity.id ||
              this.state.authMethod !== User.AuthMethodInvitation
            }
          >
            <Col sm="6">
              <Form.Check
                type="checkbox"
                id="check-resendInvitation"
                label={this.props.t("resendInvitation")}
                checked={this.state.resendInvitation}
                onChange={(e: any) =>
                  this.setState({ resendInvitation: e.target.checked })
                }
              />
            </Col>
          </Form.Group>

          {/* Password field */}
          <Form.Group
            as={Row}
            hidden={
              (RuntimeConfig.INFOS.disablePasswordLogin &&
                !this.isServiceAccount(this.state.accountType)) ||
              (!this.isServiceAccount(this.state.accountType) &&
                this.state.authMethod !== User.AuthMethodPassword)
            }
          >
            <Form.Label htmlFor="password" column sm="2">
              {this.props.t("password")}
            </Form.Label>
            <Col sm="4">
              <InputGroup>
                <Form.Control
                  id="password"
                  type={
                    this.isServiceAccount(this.state.accountType)
                      ? "text"
                      : "password"
                  }
                  value={this.state.password}
                  onChange={(e: any) =>
                    this.setState({ password: e.target.value })
                  }
                  required={
                    !!(
                      this.isServiceAccount(this.state.accountType) ||
                      (!this.entity.id &&
                        this.state.authMethod === User.AuthMethodPassword) ||
                      (this.entity.id &&
                        this.state.changePassword &&
                        this.state.authMethod === User.AuthMethodPassword)
                    )
                  }
                  disabled={
                    (!this.isServiceAccount(this.state.accountType) &&
                      this.entity.id &&
                      !this.state.changePassword) ||
                    this.isServiceAccount(this.state.accountType)
                  }
                  minLength={
                    this.isServiceAccount(this.state.accountType)
                      ? Validation.PASSWORD_MIN_LENGTH_SA
                      : Validation.PASSWORD_MIN_LENGTH
                  }
                  pattern={Validation.PASSWORD_PATTERN}
                  title={this.props.t("passwordRequirements")}
                />
                <Button
                  onClick={() => this.generatePassword()}
                  hidden={!this.isServiceAccount(this.state.accountType)}
                  variant="outline-secondary"
                >
                  <IconRefresh className="feather" />
                </Button>
                <CopyToClipboardButton
                  text={this.state.password}
                  hidden={!this.isServiceAccount(this.state.accountType)}
                />
              </InputGroup>
            </Col>
          </Form.Group>

          {/* Last activity — only shown for existing users */}
          <Form.Group as={Row} hidden={!this.entity.id}>
            <Form.Label htmlFor="lastActivity" column sm="2">
              {this.props.t("lastActivity")}
            </Form.Label>
            <Col sm="4">
              <Form.Control
                id="lastActivity"
                plaintext={true}
                readOnly={true}
                value={
                  this.state.lastActivity
                    ? Formatting.getFormatterShort(true).format(
                        this.state.lastActivity,
                      )
                    : "-"
                }
              />
            </Col>
          </Form.Group>

          {/* Second factor reset — only shown for existing non-service-account users */}
          <Form.Group
            as={Row}
            hidden={
              !this.entity.id ||
              this.isServiceAccount(this.state.accountType) ||
              (!this.state.hasPasskeys && !this.state.totpEnabled)
            }
          >
            <Form.Label column sm="2">
              {this.props.t("secondFactor")}
            </Form.Label>
            <Col sm="4" className="d-flex gap-2 align-items-center">
              {this.state.hasPasskeys && (
                <Button
                  className="btn-sm"
                  variant="outline-danger"
                  onClick={this.resetPasskeys}
                >
                  <IconReset className="feather" />{" "}
                  {this.props.t("resetPasskeys")}
                </Button>
              )}
              {this.state.totpEnabled && (
                <Button
                  className="btn-sm"
                  variant="outline-danger"
                  onClick={this.resetTotp}
                >
                  <IconReset className="feather" /> {this.props.t("resetTotp")}
                </Button>
              )}
            </Col>
          </Form.Group>

          {/* API Token section — only shown for existing service accounts */}
          {this.entity.id && this.isServiceAccount(this.state.accountType) && (
            <Form.Group as={Row}>
              <Form.Label column sm="2">
                {this.props.t("apiToken")}
              </Form.Label>
              <Col sm="4" className="d-flex gap-2 align-items-center">
                <span
                  className={
                    this.state.apiTokenConfigured
                      ? "text-success"
                      : "text-secondary"
                  }
                >
                  {this.state.apiTokenConfigured
                    ? this.props.t("apiTokenConfigured")
                    : this.props.t("apiTokenNotConfigured")}
                </span>
                <Button
                  className="btn-sm"
                  variant="outline-secondary"
                  onClick={this.generateApiToken}
                >
                  {this.props.t("generateApiToken")}
                </Button>
                {this.state.apiTokenConfigured && (
                  <Button
                    className="btn-sm"
                    variant="outline-danger"
                    onClick={this.revokeApiToken}
                  >
                    {this.props.t("revokeApiToken")}
                  </Button>
                )}
              </Col>
            </Form.Group>
          )}
        </Form>

        <Modal
          show={this.state.showApiTokenModal}
          onHide={() =>
            this.setState({ showApiTokenModal: false, generatedToken: "" })
          }
        >
          <Modal.Header closeButton>
            <Modal.Title>{this.props.t("apiToken")}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Alert variant="warning">
              {this.props.t("apiTokenOnceWarning")}
            </Alert>
            <InputGroup>
              <Form.Control
                type="text"
                readOnly
                value={this.state.generatedToken}
              />
              <CopyToClipboardButton text={this.state.generatedToken} />
            </InputGroup>
          </Modal.Body>
          <Modal.Footer>
            <Button
              variant="secondary"
              onClick={() =>
                this.setState({ showApiTokenModal: false, generatedToken: "" })
              }
            >
              {this.props.t("close")}
            </Button>
          </Modal.Footer>
        </Modal>

        <ConfirmModal
          show={this.state.pendingConfirmAction !== null}
          message={this.getPendingConfirmMessage()}
          onCancel={() => this.setState({ pendingConfirmAction: null })}
          onConfirm={this.confirmPendingAction}
        />
      </FullLayout>
    );
  }
}

export default withTranslation(
  withReadyRouter(
    withPermission(EditUser as any, Permission.Users, PermissionLevel.Read) as any,
  ),
);
