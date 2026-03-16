import Ajax from "../util/Ajax";

export default class Stats {
  spaceLoadToday: number;
  spaceLoadYesterday: number;
  spaceLoadThisWeek: number;
  spaceLoadLastWeek: number;

  constructor() {
    this.spaceLoadToday = 0;
    this.spaceLoadYesterday = 0;
    this.spaceLoadThisWeek = 0;
    this.spaceLoadLastWeek = 0;
  }

  deserialize(input: any): void {
    this.spaceLoadToday = input.spaceLoadToday;
    this.spaceLoadYesterday = input.spaceLoadYesterday;
    this.spaceLoadThisWeek = input.spaceLoadThisWeek;
    this.spaceLoadLastWeek = input.spaceLoadLastWeek;
  }

  static async get(locationId: string | null): Promise<Stats> {
    const queryParams = new URLSearchParams();
    if (locationId) queryParams.set("location", locationId);
    const params = queryParams.toString() ? `?${queryParams.toString()}` : "";
    const result = await Ajax.get(`/stats/load${params}`);
    const e: Stats = new Stats();
    e.deserialize(result.json);
    return e;
  }
}
