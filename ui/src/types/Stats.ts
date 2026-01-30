import Ajax from "../util/Ajax";

export default class Stats {
  numUsers: number;
  numBookings: number;
  numLocations: number;
  numSpaces: number;
  numBookingsCurrent: number;
  numBookingsToday: number;
  numBookingsYesterday: number;
  numBookingsThisWeek: number;
  spaceLoadToday: number;
  spaceLoadYesterday: number;
  spaceLoadThisWeek: number;
  spaceLoadLastWeek: number;

  constructor() {
    this.numUsers = 0;
    this.numBookings = 0;
    this.numLocations = 0;
    this.numSpaces = 0;
    this.numBookingsCurrent = 0;
    this.numBookingsToday = 0;
    this.numBookingsYesterday = 0;
    this.numBookingsThisWeek = 0;
    this.spaceLoadToday = 0;
    this.spaceLoadYesterday = 0;
    this.spaceLoadThisWeek = 0;
    this.spaceLoadLastWeek = 0;
  }

  deserialize(input: any): void {
    this.numUsers = input.numUsers;
    this.numBookings = input.numBookings;
    this.numLocations = input.numLocations;
    this.numSpaces = input.numSpaces;
    this.numBookingsCurrent = input.numBookingsCurrent;
    this.numBookingsToday = input.numBookingsToday;
    this.numBookingsYesterday = input.numBookingsYesterday;
    this.numBookingsThisWeek = input.numBookingsThisWeek;
    this.spaceLoadToday = input.spaceLoadToday;
    this.spaceLoadYesterday = input.spaceLoadYesterday;
    this.spaceLoadThisWeek = input.spaceLoadThisWeek;
    this.spaceLoadLastWeek = input.spaceLoadLastWeek;
  }

  static async get(): Promise<Stats> {
    return Ajax.get("/stats/").then((result) => {
      let e: Stats = new Stats();
      e.deserialize(result.json);
      return e;
    });
  }
}
