
export default class SpaceAttributeValue {
    attributeId: string;
    value: string;

    constructor() {
        this.attributeId = "";
        this.value = "";
    }

    serialize(): Object {
        return Object.assign({}, {
            "attributeId": this.attributeId,
            "value": this.value,
        });
    }

    deserialize(input: any): void {
        this.attributeId = input.attributeId;
        this.value = input.value;
    }
}