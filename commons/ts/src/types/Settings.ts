import { Entity } from "./Entity";
import Ajax from "../util/Ajax";

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
    return "/setting/";
  }

  static async list(): Promise<Settings[]> {
    return Ajax.get("/setting/").then((result) => {
      let list: Settings[] = [];
      (result.json as []).forEach((item) => {
        let e: Settings = new Settings();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }

  static async setAll(settings: Settings[]): Promise<void> {
    let payload = settings.map((e) => e.serialize());
    return Ajax.putData("/setting/", payload).then(() => undefined);
  }

  static async setOne(name: string, value: string): Promise<void> {
    let payload = { value: value };
    return Ajax.putData("/setting/" + name, payload).then(() => undefined);
  }

  static async getOne(name: string): Promise<string> {
    return Ajax.get("/setting/" + name).then((res) => res.json);
  }
}
