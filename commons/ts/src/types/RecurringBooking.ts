import { Entity } from "./Entity";
import Ajax from "../util/Ajax";
import Formatting from "../util/Formatting";

export default class RecurringBooking extends Entity {
    static CadenceDaily: number = 1;
    static CadenceWeekly: number = 2;

    spaceId: string;
    subject: string;
    enter: Date;
    leave: Date;
    end: Date;
    cadence: number;
    cycle: number;
    weekdays: number[];

    constructor() {
        super();
        this.spaceId = "";
        this.subject = "";
        this.enter = new Date();
        this.leave = new Date();
        this.end = new Date();
        this.cadence = 0;
        this.cycle = 0;
        this.weekdays = [];
    }

    serialize(): Object {
        // Convert the local dates to UTC dates without changing the date/time ("fake" UTC)
        let enter = Formatting.convertToFakeUTCDate(this.enter);
        let leave = Formatting.convertToFakeUTCDate(this.leave);
        let end = Formatting.convertToFakeUTCDate(this.end);

        return Object.assign(super.serialize(), {
            "enter": enter.toISOString(),
            "leave": leave.toISOString(),
            "end": end.toISOString(),
            "spaceId": this.spaceId,
            "subject": this.subject,
            "cadence": this.cadence,
            "cycle": this.cycle,
            "weekdays": this.weekdays,
        });
    }

    deserialize(input: any): void {
        super.deserialize(input);
        // Discard time zone information from date
        input.enter = Formatting.stripTimezoneDetails(input.enter);
        input.leave = Formatting.stripTimezoneDetails(input.leave);
        input.end = Formatting.stripTimezoneDetails(input.end);
        this.enter = new Date(input.enter);
        this.leave = new Date(input.leave);
        this.end = new Date(input.end);
        this.spaceId = input.spaceId || "";
        this.cadence = input.cadence || 0;
        this.cycle = input.cycle || 0;
        if (input.subject) {
            this.subject = input.subject;
        }
    }

    getBackendUrl(): string {
        return "/recurring-booking/";
    }

    async save(): Promise<RecurringBooking> {
        return Ajax.saveEntity(this, this.getBackendUrl()).then(() => this);
    }

    async delete(): Promise<void> {
        return Ajax.delete(this.getBackendUrl() + this.id).then(() => undefined);
    }

    static async get(id: string): Promise<RecurringBooking> {
        return Ajax.get("/recurring-booking/" + id).then(result => {
            let e: RecurringBooking = new RecurringBooking();
            e.deserialize(result.json);
            return e;
        });
    }
}