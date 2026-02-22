import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import Organization from "./Organization";
import MergeRequest from "./MergeRequest";
import { BuddyBooking } from "./Buddy";

export default class User extends Entity {
  static UserRoleUser: number = 0;
  static UserRoleSpaceAdmin: number = 10;
  static UserRoleOrgAdmin: number = 20;
  static UserRoleServiceAccountRO: number = 21;
  static UserRoleServiceAccountRW: number = 22;
  static UserRoleSuperAdmin: number = 90;

  id: string;
  email: string;
  firstname: string;
  lastname: string;
  atlassianId: string;
  organizationId: string;
  organization: Organization;
  authProviderId: string;
  requirePassword: boolean;
  passwordPending: boolean;
  role: number;
  spaceAdmin: boolean;
  admin: boolean;
  superAdmin: boolean;
  password: string;
  sendInvitation: boolean;
  firstBooking: BuddyBooking | null;
  totpEnabled: boolean;
  hasPasskeys: boolean;

  constructor() {
    super();
    this.id = "";
    this.email = "";
    this.firstname = "";
    this.lastname = "";
    this.atlassianId = "";
    this.organizationId = "";
    this.organization = new Organization();
    this.authProviderId = "";
    this.requirePassword = false;
    this.passwordPending = false;
    this.role = User.UserRoleUser;
    this.spaceAdmin = false;
    this.admin = false;
    this.superAdmin = false;
    this.password = "";
    this.sendInvitation = false;
    this.firstBooking = null;
    this.totpEnabled = false;
    this.hasPasskeys = false;
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      email: this.email,
      firstname: this.firstname,
      lastname: this.lastname,
      role: this.role,
      password: this.password,
      sendInvitation: this.sendInvitation,
      authProviderId: this.authProviderId,
      organizationId: this.organizationId,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.email = input.email;
    this.firstname = input.firstname;
    this.lastname = input.lastname;
    this.organizationId = input.organizationId;
    if (input.organization) {
      this.organization.deserialize(input.organization);
    }
    if (input.atlassianId) {
      this.atlassianId = input.atlassianId;
    }
    if (input.authProviderId) {
      this.authProviderId = input.authProviderId;
    }
    if (input.requirePassword) {
      this.requirePassword = input.requirePassword;
    }
    if (input.passwordPending !== undefined) {
      this.passwordPending = input.passwordPending;
    }
    this.role = input.role;
    this.spaceAdmin = input.spaceAdmin;
    this.admin = input.admin;
    this.superAdmin = input.superAdmin;
    this.totpEnabled = input.totpEnabled;
    this.hasPasskeys = input.hasPasskeys ?? false;
  }

  getBackendUrl(): string {
    return "/user/";
  }

  async save(): Promise<User> {
    return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
  }

  async delete(): Promise<void> {
    return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
  }

  async setPassword(password: string): Promise<void> {
    let payload = { password: password };
    return Ajax.putData(
      this.getBackendUrl() + this.id + "/password",
      payload,
    ).then(() => undefined);
  }

  static async initMerge(targetUserEmail: string): Promise<void> {
    let payload = { email: targetUserEmail };
    return Ajax.postData("/user/merge/init", payload).then(() => undefined);
  }

  static async finishMerge(actionId: string): Promise<void> {
    return Ajax.postData("/user/merge/finish/" + actionId, null).then(
      () => undefined,
    );
  }

  static async getMergeRequests(): Promise<MergeRequest[]> {
    return Ajax.get("/user/merge").then((result) => {
      let list: MergeRequest[] = [];
      (result.json as []).forEach((item: any) => {
        let e: MergeRequest = new MergeRequest(
          item.id,
          item.email,
          item.userId,
        );
        list.push(e);
      });
      return list;
    });
  }

  static async getCount(): Promise<number> {
    return Ajax.get("/user/count").then((result) => {
      return result.json.count;
    });
  }

  static async getSelf(): Promise<User> {
    return Ajax.get("/user/me").then((result) => {
      let e: User = new User();
      e.deserialize(result.json);
      return e;
    });
  }

  static async get(id: string): Promise<User> {
    return Ajax.get("/user/" + id).then((result) => {
      let e: User = new User();
      e.deserialize(result.json);
      return e;
    });
  }

  static async list(params?: { search: string | null }): Promise<User[]> {
    return Ajax.get(
      "/user/" +
        (params && params.search
          ? "?q=" + encodeURIComponent(params.search)
          : ""),
    ).then((result) => {
      let list: User[] = [];
      (result.json as []).forEach((item) => {
        let e: User = new User();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async getByEmail(email: string): Promise<User> {
    return Ajax.get("/user/byEmail/" + email).then((result) => {
      let e: User = new User();
      e.deserialize(result.json);
      return e;
    });
  }

  static async generateTotp(): Promise<TotpGenerateResponse> {
    return Ajax.get("/user/totp/generate").then((result) => {
      let e: TotpGenerateResponse = new TotpGenerateResponse();
      e.qrCode = result.json.image;
      e.stateId = result.json.stateId;
      return e;
    });
  }

  static async validateTotp(stateId: string, code: string): Promise<void> {
    let payload = {
      code: code,
      stateId: stateId,
    };
    return Ajax.postData("/user/totp/validate", payload).then(() => undefined);
  }

  static async getTotpSecret(stateId: string): Promise<string> {
    return Ajax.get("/user/totp/" + stateId + "/secret").then(
      (result) => result.json.secret,
    );
  }

  static async disableTotp(): Promise<void> {
    return Ajax.postData("/user/totp/disable", null).then(() => undefined);
  }

  static async adminResetPasskeys(userId: string): Promise<void> {
    return Ajax.delete("/user/" + userId + "/passkeys").then(() => undefined);
  }

  static async adminResetTotp(userId: string): Promise<void> {
    return Ajax.delete("/user/" + userId + "/totp").then(() => undefined);
  }
}

export class TotpGenerateResponse {
  qrCode: string;
  stateId: string;

  constructor() {
    this.qrCode = "";
    this.stateId = "";
  }
}
