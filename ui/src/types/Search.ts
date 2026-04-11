import Ajax from "../util/Ajax";

export class UserSearchResult {
  id: string;
  email: string;
  firstname: string;
  lastname: string;

  constructor() {
    this.id = "";
    this.email = "";
    this.firstname = "";
    this.lastname = "";
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.email = input.email;
    this.firstname = input.firstname;
    this.lastname = input.lastname;
  }

  static from(input: any): UserSearchResult {
    const e = new UserSearchResult();
    e.deserialize(input);
    return e;
  }
}

export class LocationSearchResult {
  id: string;
  name: string;
  description: string;

  constructor() {
    this.id = "";
    this.name = "";
    this.description = "";
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.name = input.name;
    this.description = input.description;
  }

  static from(input: any): LocationSearchResult {
    const e = new LocationSearchResult();
    e.deserialize(input);
    return e;
  }
}

export class SpaceSearchResult {
  id: string;
  name: string;
  location: LocationSearchResult | null;

  constructor() {
    this.id = "";
    this.name = "";
    this.location = null;
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.name = input.name;
    if (input.location) {
      this.location = LocationSearchResult.from(input.location);
    }
  }

  static from(input: any): SpaceSearchResult {
    const e = new SpaceSearchResult();
    e.deserialize(input);
    return e;
  }
}

export class GroupSearchResult {
  id: string;
  name: string;
  organizationId: string;

  constructor() {
    this.id = "";
    this.name = "";
    this.organizationId = "";
  }

  deserialize(input: any): void {
    this.id = input.id;
    this.name = input.name;
    this.organizationId = input.organizationId;
  }

  static from(input: any): GroupSearchResult {
    const e = new GroupSearchResult();
    e.deserialize(input);
    return e;
  }
}

export class SearchOptions {
  keyword: string = "";
  includeUsers: boolean = false;
  includeLocations: boolean = false;
  includeSpaces: boolean = false;
  includeGroups: boolean = false;
  expandLocations: boolean = false;

  getSearchParams(): URLSearchParams {
    const params = new URLSearchParams();
    for (const key of Object.keys(this) as (keyof SearchOptions)[]) {
      if (key === "keyword" && this[key]) {
        params.append("query", this[key]);
      } else if (this[key]) {
        params.append(key, "1");
      }
    }
    return params;
  }
}

export default class Search {
  users: UserSearchResult[];
  locations: LocationSearchResult[];
  spaces: SpaceSearchResult[];
  groups: GroupSearchResult[];

  constructor() {
    this.users = [];
    this.locations = [];
    this.spaces = [];
    this.groups = [];
  }

  deserialize(input: any): void {
    if (input.users) this.users = input.users.map(UserSearchResult.from);
    if (input.groups) this.groups = input.groups.map(GroupSearchResult.from);
    if (input.locations)
      this.locations = input.locations.map(LocationSearchResult.from);
    if (input.spaces) this.spaces = input.spaces.map(SpaceSearchResult.from);
  }

  static from(input: any): Search {
    const e = new Search();
    e.deserialize(input);
    return e;
  }

  static async search(options: SearchOptions): Promise<Search> {
    const { json } = await Ajax.get(
      `/search/?${options.getSearchParams().toString()}`,
    );
    return Search.from(json);
  }
}
