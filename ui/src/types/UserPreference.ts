import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

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
    this.name = name ? name : "";
    this.value = value ? value : "";
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
    return "/preference/";
  }

  static async list(): Promise<UserPreference[]> {
    return Ajax.get("/preference/").then((result) => {
      let list: UserPreference[] = [];
      (result.json as []).forEach((item) => {
        let e: UserPreference = new UserPreference();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async setAll(preferences: UserPreference[]): Promise<void> {
    let payload = preferences.map((e) => e.serialize());
    return Ajax.putData("/preference/", payload).then(() => undefined);
  }

  static async setOne(name: string, value: string): Promise<void> {
    let payload = { value: value };
    return Ajax.putData("/preference/" + name, payload).then(() => undefined);
  }

  static async getOne(name: string): Promise<string> {
    return Ajax.get("/preference/" + name).then((res) => res.json);
  }
}
