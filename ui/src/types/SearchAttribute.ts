import Ajax from "../util/Ajax";
import DateUtil from "../util/DateUtil";
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
      attributeId: this.attributeId,
      comparator: this.comparator,
      value: this.value,
    };
  }

  static async search(
    enter: Date,
    leave: Date,
    attributes: SearchAttribute[],
  ): Promise<Location[]> {
    let payload = {
      enter: DateUtil.convertToFakeUTCDate(enter).toISOString(),
      leave: DateUtil.convertToFakeUTCDate(leave).toISOString(),
      attributes: attributes.map((a) => a.serialize()),
    };
    return Ajax.postData("/location/search", payload).then((result) => {
      let list: Location[] = [];
      (result.json as []).forEach((item) => {
        let e: Location = new Location();
        e.deserialize(item);
        list.push(e);
      });
      return list;
    });
  }
}
