import Ajax from "../util/Ajax";
import { Entity } from "./Entity";

export default class Session {
  id: string;
  userId: string;
  device: string;
  created: Date;

  constructor() {
    this.id = "";
    this.userId = "";
    this.device = "";
    this.created = new Date();
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.userId = input.userId;
    this.device = input.device;
    if (input.created) {
      this.created = new Date(input.created);
    }
  }

  static async list(): Promise<Session[]> {
    return Ajax.get("/user/session").then((result) => {
      let list: Session[] = [];
      (result.json as []).forEach((item) => {
        let e: Session = new Session();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }
}
