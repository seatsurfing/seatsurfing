import { Ajax, AjaxCredentials, User, Settings as OrgSettings } from 'seatsurfing-commons'

interface RuntimeUserInfos {
    username: string;
    isLoading: boolean;
    maxBookingsPerUser: number;
    maxConcurrentBookingsPerUser: number;
    maxDaysInAdvance: number;
    maxBookingDurationHours: number;
    maxHoursBeforeDelete: number;
    minBookingDurationHours: number;
    dailyBasisBooking: boolean;
    noAdminRestrictions: boolean;
    showNames: boolean;
    customLogoUrl: string;
    defaultTimezone: string;
    disableBuddies: boolean;
    maxHoursPartiallyBooked: number;
    maxHoursPartiallyBookedEnabled: boolean;
}

export default class RuntimeConfig {
    static EMBEDDED: boolean = false;
    static INFOS: RuntimeUserInfos = {
        username: "",
        isLoading: true,
        maxBookingsPerUser: 0,
        maxConcurrentBookingsPerUser: 0,
        maxDaysInAdvance: 0,
        maxBookingDurationHours: 0,
        maxHoursBeforeDelete: 0,
        minBookingDurationHours: 0,
        dailyBasisBooking: false,
        noAdminRestrictions: false,
        disableBuddies: false,
        customLogoUrl: '',
        maxHoursPartiallyBooked: 0,
        maxHoursPartiallyBookedEnabled: false,
        showNames: false,
        defaultTimezone: "",
    };

    static verifyToken = async (resolve: Function) => {
        Ajax.CREDENTIALS = await Ajax.PERSISTER.readCredentialsFromSessionStorage();
        if (!Ajax.CREDENTIALS.accessToken) {
            Ajax.CREDENTIALS = await Ajax.PERSISTER.readRefreshTokenFromLocalStorage();
            if (Ajax.CREDENTIALS.refreshToken) {
                await Ajax.refreshAccessToken(Ajax.CREDENTIALS.refreshToken);
            }
        }
        if (Ajax.CREDENTIALS.accessToken) {
            User.getSelf().then(user => {
                RuntimeConfig.loadSettings().then(() => {
                    RuntimeConfig.setDetails(user.email);
                    resolve();
                    //this.setState({ isLoading: false });
                });
            }).catch((e) => {
                Ajax.CREDENTIALS = new AjaxCredentials();
                Ajax.PERSISTER.deleteCredentialsFromSessionStorage().then(() => {
                    resolve();
                    //this.setState({ isLoading: false });
                });
            });
        } else {
            resolve();
            //this.setState({ isLoading: false });
        }
    }

    static loadSettings = async (): Promise<void> => {
        return new Promise<void>(function (resolve, reject) {
            OrgSettings.list().then(settings => {
                settings.forEach(s => {
                    if (typeof window !== 'undefined') {
                        if (s.name === "max_bookings_per_user") RuntimeConfig.INFOS.maxBookingsPerUser = window.parseInt(s.value);
                        if (s.name === "max_concurrent_bookings_per_user") RuntimeConfig.INFOS.maxConcurrentBookingsPerUser = window.parseInt(s.value);
                        if (s.name === "max_days_in_advance") RuntimeConfig.INFOS.maxDaysInAdvance = window.parseInt(s.value);
                        if (s.name === "max_booking_duration_hours") RuntimeConfig.INFOS.maxBookingDurationHours = window.parseInt(s.value);
                        if (s.name === "max_hours_before_delete") RuntimeConfig.INFOS.maxHoursBeforeDelete = window.parseInt(s.value);
                        if (s.name === "max_hours_partially_booked") RuntimeConfig.INFOS.maxHoursPartiallyBooked = window.parseInt(s.value);
                        if (s.name === "min_booking_duration_hours") RuntimeConfig.INFOS.minBookingDurationHours = window.parseInt(s.value);
                    }
                    if (s.name === "daily_basis_booking") RuntimeConfig.INFOS.dailyBasisBooking = (s.value === "1");
                    if (s.name === "no_admin_restrictions") RuntimeConfig.INFOS.noAdminRestrictions = (s.value === "1");
                    if (s.name === "max_hours_partially_booked_enabled") RuntimeConfig.INFOS.maxHoursPartiallyBookedEnabled = (s.value === "1");
                    if (s.name === "show_names") RuntimeConfig.INFOS.showNames = (s.value === "1");
                    if (s.name === "disable_buddies") RuntimeConfig.INFOS.disableBuddies = (s.value === "1");
                    if (s.name === "custom_logo_url") RuntimeConfig.INFOS.customLogoUrl = s.value;
                    if (s.name === "default_timezone") RuntimeConfig.INFOS.defaultTimezone = s.value;
                });
                resolve();
            });
        });
    }

    static setDetails = (username: string) => {
        RuntimeConfig.loadSettings().then(() => {
            RuntimeConfig.INFOS.username = username;
        });
    }

    static async setLoginDetails(): Promise<void> {
        return User.getSelf().then(user => {
            RuntimeConfig.setDetails(user.email);
        });
    }
}
