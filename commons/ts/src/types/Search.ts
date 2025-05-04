import User from "./User";
import Location from "./Location";
import Space from "./Space";
import Ajax from "../util/Ajax";

export class SearchOptions {
    includeUsers: boolean = false;
    includeLocations: boolean = false;
    includeSpaces: boolean = false;
    includeGroups: boolean = false;
}

export default class Search {
    users: User[]
    locations: Location[]
    spaces: Space[]

    constructor() {
        this.users = [];
        this.locations = [];
        this.spaces = [];
    }

    deserialize(input: any): void {
        if (input.users) {
            this.users = input.users.map((user: any) => {
                let e = new User();
                e.deserialize(user);
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

    static async search(keyword: string, options: SearchOptions): Promise<Search> {
        let params = new URLSearchParams();
        params.append("query", keyword);
        params.append("includeUsers", options.includeUsers ? "1" : "0");
        params.append("includeLocations", options.includeLocations ? "1" : "0");
        params.append("includeSpaces", options.includeSpaces ? "1" : "0");
        params.append("includeGroups", options.includeGroups ? "1" : "0");
        return Ajax.get("/search/?" + params.toString()).then(result => {
            let e: Search = new Search();
            e.deserialize(result.json);
            return e;
        });
    }
}