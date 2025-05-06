import React from 'react';
import { Home as IconHome, User as IconUsers, Users as IconGroups, Map as IconMap, Book as IconBook, Settings as IconSettings, Box as IconBox, Activity as IconAnalysis, ExternalLink as IconExternalLink, Icon } from 'react-feather';
import { Ajax, AjaxCredentials, User } from 'seatsurfing-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import { Nav } from 'react-bootstrap';
import { NextRouter } from 'next/router';
import withReadyRouter from './withReadyRouter';
import Link from 'next/link';
import dynamic from 'next/dynamic';
import RuntimeConfig from './RuntimeConfig';

interface State {
}

interface Props extends WithTranslation {
    router: NextRouter
}

class SideBar extends React.Component<Props, State> {
    dynamicIcons: Map<string, any> = new Map();

    constructor(props: any) {
        super(props);
        this.state = {
        };
    }

    componentDidMount = () => {
        if (!RuntimeConfig.INFOS.spaceAdmin) {
            Ajax.CREDENTIALS = new AjaxCredentials();
            Ajax.PERSISTER.deleteCredentialsFromSessionStorage().then(() => {
                this.props.router.push("/login");
            });
            return;
        }
    }

    getActiveKey = () => {
        let path = this.props.router.pathname;
        if (path.startsWith('/plugin/')) {
            return window.location.pathname.replace('/admin', '');
        }
        const startPaths = [
            '/organizations',
            '/users',
            '/groups',
            '/settings',
            '/locations',
            '/bookings'
        ];
        let result = path;
        startPaths.forEach(startPath => {
            if (path.startsWith(startPath)) {
                result = startPath;
            }
        });
        return result;
    }

    render() {
        let orgItem = <></>;
        if (RuntimeConfig.INFOS.superAdmin) {
            orgItem = (
                <li className="nav-item">
                    <Nav.Link as={Link} eventKey="/organizations" href="/organizations"><IconBox className="feather" /> {this.props.t("organizations")}</Nav.Link>
                </li>
            );
        }
        let orgAdminItems = <></>;
        if (RuntimeConfig.INFOS.orgAdmin) {
            orgAdminItems = (
                <>
                    <li className="nav-item">
                        <Nav.Link as={Link} eventKey="/users" href="/users"><IconUsers className="feather" /> {this.props.t("users")}</Nav.Link>
                    </li>
                    <li className="nav-item" hidden={!RuntimeConfig.INFOS.featureGroups}>
                        <Nav.Link as={Link} eventKey="/groups" href="/groups"><IconGroups className="feather" /> {this.props.t("groups")}</Nav.Link>
                    </li>
                    <li className="nav-item">
                        <Nav.Link as={Link} eventKey="/settings" href="/settings"><IconSettings className="feather" /> {this.props.t("settings")}</Nav.Link>
                    </li>
                    {
                        RuntimeConfig.INFOS.pluginMenuItems.map((item) => {
                            if (item.visibility !== 'admin') {
                                return;
                            }
                            let PluginIcon = this.dynamicIcons.get(item.icon);
                            if (!PluginIcon) {
                                PluginIcon = dynamic(() => import('react-feather/dist/icons/' + item.icon.toLowerCase()), { ssr: true }) as Icon;
                                this.dynamicIcons.set(item.icon, PluginIcon);
                            }
                            return (
                                <li className="nav-item" key={'plugin-' + item.id}>
                                    <Nav.Link as={Link} eventKey={'/plugin/' + item.id} href={"/plugin/" + item.id}><PluginIcon className="feather" /> {item.title}</Nav.Link>
                                </li>
                            );
                        })
                    }
                </>
            );
        }
        return (
            <Nav id="sidebarMenu" className="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse" activeKey={this.getActiveKey()}>
                <div className="sidebar-sticky pt-3">
                    <ul className="nav flex-column">
                        <li className="nav-item">
                            <Nav.Link as={Link} eventKey="/dashboard" href="/dashboard"><IconHome className="feather" /> {this.props.t("dashboard")}</Nav.Link>
                        </li>
                        <li className="nav-item">
                            <Nav.Link as={Link} eventKey="/locations" href="/locations"><IconMap className="feather" /> {this.props.t("areas")}</Nav.Link>
                        </li>
                        <li className="nav-item">
                            <Nav.Link as={Link} eventKey="/bookings" href="/bookings"><IconBook className="feather" /> {this.props.t("bookings")}</Nav.Link>
                        </li>
                        <li className="nav-item">
                            <Nav.Link as={Link} eventKey="/report/analysis" href="/report/analysis"><IconAnalysis className="feather" /> {this.props.t("analysis")}</Nav.Link>
                        </li>
                        {
                            RuntimeConfig.INFOS.pluginMenuItems.map((item) => {
                                if (item.visibility !== 'spaceadmin') {
                                    return;
                                }
                                let PluginIcon = this.dynamicIcons.get(item.icon);
                                if (!PluginIcon) {
                                    PluginIcon = dynamic(() => import('react-feather/dist/icons/' + item.icon.toLowerCase()), { ssr: true }) as Icon;
                                    this.dynamicIcons.set(item.icon, PluginIcon);
                                }
                                return (
                                    <li className="nav-item" key={'plugin-' + item.id}>
                                        <Nav.Link as={Link} eventKey={'/plugin/' + item.id} href={"/plugin/" + item.id}><PluginIcon className="feather" /> {item.title}</Nav.Link>
                                    </li>
                                );
                            })
                        }
                        {orgAdminItems}
                        {orgItem}
                        <li className="nav-item">
                            <Nav.Link onClick={(e) => { e.preventDefault(); window.location.href = "/ui/"; }}><IconExternalLink className="feather" /> {this.props.t("bookingui")}</Nav.Link>
                        </li>
                    </ul>
                </div>
            </Nav>
        );
    }
}

export default withTranslation(['admin'])(withReadyRouter(SideBar as any));
