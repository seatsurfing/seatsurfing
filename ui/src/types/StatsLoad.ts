import Ajax from "../util/Ajax";

export default class Stats {
  spaceLoadNextWeek: number;
  spaceLoadThisWeek: number;
  spaceLoadLastWeek: number;
  spaceLoadLastMonth: number;

  constructor() {
    this.spaceLoadNextWeek = 0;
    this.spaceLoadThisWeek = 0;
    this.spaceLoadLastWeek = 0;
    this.spaceLoadLastMonth = 0;
  }

  deserialize(input: any): void {
    this.spaceLoadNextWeek = input.spaceLoadNextWeek;
    this.spaceLoadThisWeek = input.spaceLoadThisWeek;
    this.spaceLoadLastWeek = input.spaceLoadLastWeek;
    this.spaceLoadLastMonth = input.spaceLoadLastMonth;
  }

  static async getLoad(locationId: string | null): Promise<Stats> {
    const queryParams = new URLSearchParams();
    if (locationId) queryParams.set("location", locationId);
    const params = queryParams.toString() ? `?${queryParams.toString()}` : "";
    const result = await Ajax.get(`/stats/load${params}`);
    const e: Stats = new Stats();
    e.deserialize(result.json);
    return e;
  }

  static async getWeekday(locationId: string | null, period: string | null): Promise<number[]> {
    const queryParams = new URLSearchParams();
    if (locationId) queryParams.set("location", locationId);
    if (period) queryParams.set("period", period);
    const params = queryParams.toString() ? `?${queryParams.toString()}` : "";
    const result = await Ajax.get(`/stats/weekday${params}`);
    return result.json.bookingsByWeekday ?? [0, 0, 0, 0, 0, 0, 0];
  }
}
