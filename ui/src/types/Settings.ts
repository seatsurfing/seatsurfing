import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import Navigation from "../util/Navigation";

export default class Settings extends Entity {
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
    return Navigation.PATH_API_SETTINGS;
  }

  static async list(): Promise<Settings[]> {
    const result = await Ajax.get(Navigation.PATH_API_SETTINGS);
    const list: Settings[] = [];
    (result.json as []).forEach((item) => {
      const e: Settings = new Settings();
      e.deserialize(item);
      list.push(e);
    });
    return list;
  }

  static async setAll(settings: Settings[]): Promise<void> {
    const payload = settings.map((e) => e.serialize());
    return Ajax.putData(Navigation.PATH_API_SETTINGS, payload).then(
      () => undefined,
    );
  }

  static async setOne(name: string, value: string): Promise<void> {
    const payload = { value: value };
    return Ajax.putData(
      `${Navigation.PATH_API_SETTINGS}${encodeURIComponent(name)}`,
      payload,
    ).then(() => undefined);
  }

  static async getOne(name: string): Promise<string> {
    return Ajax.get(
      `${Navigation.PATH_API_SETTINGS}${encodeURIComponent(name)}`,
    ).then((res) => res.json);
  }
}
