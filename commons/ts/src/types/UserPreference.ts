import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

export default class UserPreference extends Entity {
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
