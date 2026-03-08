import User from "./User";
import Location from "./Location";
import Space from "./Space";
import Ajax from "../util/Ajax";
import Group from "./Group";

export class SearchOptions {
  includeUsers: boolean = false;
  includeLocations: boolean = false;
  includeSpaces: boolean = false;
  includeGroups: boolean = false;
}

export default class Search {
  users: User[];
  locations: Location[];
  spaces: Space[];
  groups: Group[];

  constructor() {
    this.users = [];
    this.locations = [];
    this.spaces = [];
    this.groups = [];
  }

  deserialize(input: any): void {
    if (input.users) {
      this.users = input.users.map((user: any) => {
        let e = new User();
        e.deserialize(user);
        return e;
      });
    }
    if (input.groups) {
      this.groups = input.groups.map((group: any) => {
        let e = new Group();
        e.deserialize(group);
        return e;
      });
    }
    if (input.locations) {
      this.locations = input.locations.map((location: any) => {
        let e = new Location();
        e.deserialize(location);
        return e;
      });
    }
    if (input.spaces) {
      this.spaces = input.spaces.map((space: any) => {
        let e = new Space();
        e.deserialize(space);
        return e;
      });
    }
  }

  static async search(
    keyword: string,
    options: SearchOptions,
  ): Promise<Search> {
    const params = new URLSearchParams();
    params.append("query", keyword);
    if (options.includeUsers) {
      params.append("includeUsers", "1");
    }
    if (options.includeLocations) {
      params.append("includeLocations", "1");
    }
    if (options.includeSpaces) {
      params.append("includeSpaces", "1");
    }
    if (options.includeGroups) {
      params.append("includeGroups", "1");
    }
    const result = await Ajax.get("/search/?" + params.toString());
    const e: Search = new Search();
    e.deserialize(result.json);
    return e;
  }
}
