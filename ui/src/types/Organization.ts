import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export default class Organization extends Entity {
  static readonly ENFORCE_TOTP_DISABLED = 0;
  static readonly ENFORCE_TOTP_ALL_USERS = 1;
  static readonly ENFORCE_TOTP_ADMINS_ONLY = 2;

  static readonly PREF_ALLOW_ANY_USER = "allow_any_user";
  static readonly PREF_DEFAULT_TIMEZONE = "default_timezone";
  static readonly PREF_CONFLUENCE_SERVER_SHARED_SECRET =
    "confluence_server_shared_secret";
  static readonly PREF_CUSTOM_LOGO_URL = "custom_logo_url";
  static readonly PREF_MAX_BOOKINGS_PER_USER = "max_bookings_per_user";
  static readonly PREF_MAX_CONCURRENT_BOOKINGS_PER_USER =
    "max_concurrent_bookings_per_user";
  static readonly PREF_MAX_DAYS_IN_ADVANCE = "max_days_in_advance";
  static readonly PREF_BOOKING_RETENTION_ENABLED = "booking_retention_enabled";
  static readonly PREF_BOOKING_RETENTION_DAYS = "booking_retention_days";
  static readonly PREF_ENABLE_MAX_HOURS_BEFORE_DELETE =
    "enable_max_hours_before_delete";
  static readonly PREF_MAX_HOURS_BEFORE_DELETE = "max_hours_before_delete";
  static readonly PREF_DAILY_BASIS_BOOKING = "daily_basis_booking";
  static readonly PREF_MAX_BOOKING_DURATION_HOURS =
    "max_booking_duration_hours";
  static readonly PREF_MIN_BOOKING_DURATION_HOURS =
    "min_booking_duration_hours";
  static readonly PREF_TARGET_UTILIZATION_HOURS_PER_WEEK =
    "target_utilization_hours_per_week";
  static readonly PREF_SUBJECT_DEFAULT = "subject_default";
  static readonly PREF_NO_ADMIN_RESTRICTIONS = "no_admin_restrictions";
  static readonly PREF_SHOW_NAMES = "show_names";
  static readonly PREF_ALLOW_BOOKING_NONEXIST_USERS =
    "allow_booking_nonexist_users";
  static readonly PREF_DISABLE_BUDDIES = "disable_buddies";
  static readonly PREF_MAX_HOURS_PARTIALLY_BOOKED_ENABLED =
    "max_hours_partially_booked_enabled";
  static readonly PREF_MAX_HOURS_PARTIALLY_BOOKED =
    "max_hours_partially_booked";
  static readonly PREF_FEATURE_NO_USER_LIMIT = "feature_no_user_limit";
  static readonly PREF_FEATURE_CUSTOM_DOMAINS = "feature_custom_domains";
  static readonly PREF_ALLOW_RECURRING_BOOKINGS = "allow_recurring_bookings";
  static readonly PREF_NEW_USER_DEFAULT_MAIL_NOTIFICATION =
    "new_user_default_mail_notification";
  static readonly PREF_ENFORCE_TOTP = "enforce_totp";
  static readonly PREF_KIOSK_MODE_ENABLED = "kiosk_mode_enabled";
  static readonly PREF_KIOSK_ACCESS_SECRET = "kiosk_access_secret";
  static readonly PREF_SYS_ORG_SIGNUP_DELETE = "_sys_org_signup_delete";
  static readonly PREF_HIDE_REPORTS = "hide_reports";
  static readonly PREF_HIDE_STATS = "hide_stats";

  name: string;
  contactFirstname: string;
  contactLastname: string;
  contactEmail: string;
  language: string;

  constructor() {
    super();
    this.name = "";
    this.contactFirstname = "";
    this.contactLastname = "";
    this.contactEmail = "";
    this.language = "";
  }

  serialize(): Object {
    const obj: any = {
      name: this.name,
      firstname: this.contactFirstname,
      lastname: this.contactLastname,
      email: this.contactEmail,
      language: this.language,
    };
    return Object.assign(super.serialize(), obj);
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.contactFirstname = input.firstname;
    this.contactLastname = input.lastname;
    this.contactEmail = input.email;
    this.language = input.language;
  }

  getBackendUrl(): string {
    return "/organization/";
  }

  async save(): Promise<Organization> {
    await Ajax.saveEntity(this, this.getBackendUrl());
    return this;
  }

  async delete(): Promise<number> {
    const result = await Ajax.delete(
      `${this.getBackendUrl()}${encodeURIComponent(this.id)}`,
    );
    return result.json.code as number;
  }

  static async get(id: string): Promise<Organization> {
    const result = await Ajax.get(`/organization/${encodeURIComponent(id)}`);
    const e: Organization = new Organization();
    e.deserialize(result.json);
    return e;
  }

  static async list(): Promise<Organization[]> {
    const result = await Ajax.get("/organization/");
    const list: Organization[] = [];
    (result.json as []).forEach((item) => {
      const e: Organization = new Organization();
      e.deserialize(item);
      list.push(e);
    });
    return list;
  }

  static async getOrgForDomain(domain: string): Promise<Organization> {
    const result = await Ajax.get(
      `/organization/domain/${encodeURIComponent(domain)}`,
    );
    const e: Organization = new Organization();
    e.deserialize(result.json);
    return e;
  }
}
