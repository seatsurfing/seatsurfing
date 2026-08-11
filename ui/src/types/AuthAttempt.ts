import Ajax from "../util/Ajax";

export interface AuthAttemptListParams {
  start?: Date;
  end?: Date;
  user?: string;
  method?: string;
  success?: boolean;
  errorCode?: string;
  limit?: number;
  offset?: number;
}

export interface AuthAttemptListResult {
  total: number;
  items: AuthAttempt[];
}

export default class AuthAttempt {
  static readonly METHODS = [
    "password",
    "totp",
    "passkey",
    "passkey_2fa",
    "oauth",
    "confluence",
  ];

  static readonly ERROR_CODES = [
    "user_not_found",
    "password_pending",
    "no_password_set",
    "bound_to_auth_provider",
    "user_disabled",
    "service_account",
    "wrong_password",
    "password_update_required",
    "totp_missing",
    "totp_replay",
    "totp_invalid",
    "passkey_state_invalid",
    "passkey_assertion_invalid",
    "passkey_clone_detected",
    "idp_provider_not_found",
    "idp_config_invalid",
    "idp_state_invalid",
    "idp_code_exchange_failed",
    "idp_userinfo_failed",
    "idp_attribute_mapping_failed",
    "idp_provider_mismatch",
    "user_limit_reached",
    "user_create_failed",
    "org_mismatch",
    "internal_error",
    "confluence_jwt_invalid",
  ];

  id: string;
  userId: string;
  email: string;
  timestamp: Date;
  successful: boolean;
  method: string;
  authProviderId: string;
  authProviderName: string;
  errorCode: string;
  errorDetail: string;
  device: string;

  constructor() {
    this.id = "";
    this.userId = "";
    this.email = "";
    this.timestamp = new Date();
    this.successful = false;
    this.method = "";
    this.authProviderId = "";
    this.authProviderName = "";
    this.errorCode = "";
    this.errorDetail = "";
    this.device = "";
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.userId = input.userId;
    this.email = input.email;
    if (input.timestamp) {
      this.timestamp = new Date(input.timestamp);
    }
    this.successful = input.successful;
    this.method = input.method;
    this.authProviderId = input.authProviderId;
    this.authProviderName = input.authProviderName;
    this.errorCode = input.errorCode;
    this.errorDetail = input.errorDetail;
    this.device = input.device;
  }

  static async list(
    params: AuthAttemptListParams,
  ): Promise<AuthAttemptListResult> {
    const query = new URLSearchParams();
    if (params.start) {
      query.append("start", params.start.toISOString());
    }
    if (params.end) {
      query.append("end", params.end.toISOString());
    }
    if (params.user) {
      query.append("user", params.user);
    }
    if (params.method) {
      query.append("method", params.method);
    }
    if (params.success !== undefined) {
      query.append("success", params.success ? "true" : "false");
    }
    if (params.errorCode) {
      query.append("errorCode", params.errorCode);
    }
    if (params.limit) {
      query.append("limit", params.limit.toString());
    }
    if (params.offset) {
      query.append("offset", params.offset.toString());
    }
    const result = await Ajax.get("/auth-attempt/?" + query.toString());
    const items: AuthAttempt[] = [];
    (result.json.items as []).forEach((item) => {
      const e: AuthAttempt = new AuthAttempt();
      e.deserialize(item);
      items.push(e);
    });
    return {
      total: result.json.total,
      items: items,
    };
  }
}
