import Ajax from "../util/Ajax";
import Location from "./Location";

export default class SearchAttribute {
    attributeId: string;
    comparator: string;
    value: string;

    constructor() {
        this.attributeId = "";
        this.comparator = "";
        this.value = "";
    }

    serialize(): Object {
        return {
            "attributeId": this.attributeId,
            "comparator": this.comparator,
            "value": this.value,
        };
    }

    static async search(attributes: SearchAttribute[]): Promise<Location[]> {
        return Ajax.postData("/location/search", attributes.map(a => a.serialize())).then(result => {
            let list: Location[] = [];
            (result.json as []).forEach(item => {
                let e: Location = new Location();
                e.deserialize(item);
                list.push(e);
            });
            return list;
        });
    }
}