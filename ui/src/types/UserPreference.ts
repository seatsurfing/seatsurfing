import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import Navigation from "@/util/Navigation";

export default class UserPreference extends Entity {
  static readonly PREF_ENTER_TIME = "enter_time";
  static readonly PREF_WORKDAY_START = "workday_start";
  static readonly PREF_WORKDAY_END = "workday_end";
  static readonly PREF_WORKDAYS = "workdays";
  static readonly PREF_LOCATION_ID = "location_id";
  static readonly PREF_BOOKED_COLOR = "booked_color";
  static readonly PREF_NOT_BOOKED_COLOR = "not_booked_color";
  static readonly PREF_SELF_BOOKED_COLOR = "self_booked_color";
  static readonly PREF_BUDDY_BOOKED_COLOR = "buddy_booked_color";
  static readonly PREF_PARTIALLY_BOOKED_COLOR = "partially_booked_color";
  static readonly PREF_DISALLOWED_COLOR = "disallowed_color";
  static readonly PREF_CALDAV_URL = "caldav_url";
  static readonly PREF_CALDAV_USER = "caldav_user";
  static readonly PREF_CALDAV_PASS = "caldav_pass";
  static readonly PREF_CALDAV_PATH = "caldav_path";
  static readonly PREF_MAIL_NOTIFICATIONS = "mail_notifications";
  static readonly PREF_USE_24_HOUR_TIME = "use_24_hour_time";
  static readonly PREF_DATE_FORMAT = "date_format";

  name: string;
  value: string;

  constructor(name?: string, value?: string) {
    super();
    this.name = name ?? "";
    this.value = value ?? "";
  }

  serialize(): Object {
    return Object.assign(super.serialize(), {
      name: this.name,
      value: this.value,
    });
  }

  deserialize(input: any): void {
    super.deserialize(input);
    this.name = input.name;
    this.value = input.value;
  }

  getBackendUrl(): string {
    return Navigation.PATH_API_USER_PREFERENCES;
  }

  static async list(): Promise<UserPreference[]> {
    const result = await Ajax.get(Navigation.PATH_API_USER_PREFERENCES);
    return (result.json as []).map((item) => {
      const e = new UserPreference();
      e.deserialize(item);
      return e;
    });
  }

  static async setAll(preferences: UserPreference[]): Promise<void> {
    const payload = preferences.map((e) => e.serialize());
    return Ajax.putData(Navigation.PATH_API_USER_PREFERENCES, payload).then(
      () => undefined,
    );
  }

  static async setOne(name: string, value: string): Promise<void> {
    const payload = { value: value };
    await Ajax.putData(
      `${Navigation.PATH_API_USER_PREFERENCES}${encodeURIComponent(name)}`,
      payload,
    );
  }

  static async getOne(name: string): Promise<string> {
    const res = await Ajax.get(
      `${Navigation.PATH_API_USER_PREFERENCES}${encodeURIComponent(name)}`,
    );
    return res.json;
  }
}
