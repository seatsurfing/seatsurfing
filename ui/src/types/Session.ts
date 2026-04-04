import Ajax from "../util/Ajax";

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

  async delete(): Promise<void> {
    await Ajax.get(`/auth/logout/${decodeURIComponent(this.id)}`);
  }

  static async list(): Promise<Session[]> {
    const result = await Ajax.get("/user/session");
    const list: Session[] = [];
    (result.json as []).forEach((item) => {
      const e: Session = new Session();
      e.deserialize(item);
      list.push(e);
    });
    return list;
  }
}
