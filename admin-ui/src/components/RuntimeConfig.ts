import { Ajax, AjaxCredentials, User, Settings as OrgSettings } from 'seatsurfing-commons'

interface RuntimeUserInfos {
    superAdmin: boolean;
    spaceAdmin: boolean;
    orgAdmin: boolean;
    pluginMenuItems: any[];
    pluginWelcomeScreens: any[];
    featureGroups: boolean;
}

export default class RuntimeConfig {
    static INFOS: RuntimeUserInfos = {
        superAdmin: false,
        spaceAdmin: false,
        orgAdmin: false,
        pluginMenuItems: [],
        pluginWelcomeScreens: [],
        featureGroups: false,
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
            RuntimeConfig.loadUserAndSettings().then(() => {
                resolve();
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
                    if (s.name === "_sys_admin_menu_items") RuntimeConfig.INFOS.pluginMenuItems = (s.value ? JSON.parse(s.value) : []);
                    if (s.name === "_sys_admin_welcome_screens") RuntimeConfig.INFOS.pluginWelcomeScreens = (s.value ? JSON.parse(s.value) : []);
                    if (s.name === "feature_groups") RuntimeConfig.INFOS.featureGroups = (s.value ? JSON.parse(s.value) : []);
                });
                resolve();
            });
        });
    }

    static loadUserAndSettings = async (): Promise<void> => {
        return User.getSelf().then(user => {
            RuntimeConfig.INFOS.superAdmin = user.superAdmin;
            RuntimeConfig.INFOS.spaceAdmin = user.spaceAdmin;
            RuntimeConfig.INFOS.orgAdmin = user.admin;
            return RuntimeConfig.loadSettings();
        });
    }
}
