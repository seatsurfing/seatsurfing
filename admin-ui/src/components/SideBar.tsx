import React from 'react';
import { Home as IconHome, Users as IconUsers, Map as IconMap, Book as IconBook, Settings as IconSettings, Box as IconBox, Activity as IconAnalysis, ExternalLink as IconExternalLink, Gift as UpgradeIcon } from 'react-feather';
import { Ajax, AjaxCredentials, User } from 'flexspace-commons';
import { WithTranslation, withTranslation } from 'next-i18next';
import { Nav } from 'react-bootstrap';
import { NextRouter } from 'next/router';
import withReadyRouter from './withReadyRouter';
import Link from 'next/link';

interface State {
    superAdmin: boolean
    spaceAdmin: boolean
    orgAdmin: boolean
}

interface Props extends WithTranslation {
    router: NextRouter
}

class SideBar extends React.Component<Props, State> {
    constructor(props: any) {
        super(props);
        this.state = {
            superAdmin: false,
            spaceAdmin: false,
            orgAdmin: false,
        };
    }

    componentDidMount = () => {
        User.getSelf().then(user => {
            if (user.role === 0) {
                Ajax.CREDENTIALS = new AjaxCredentials();
                Ajax.PERSISTER.deleteCredentialsFromSessionStorage().then(() => {
                    this.props.router.push("/login");
                });
                return;
            }
            this.setState({
                superAdmin: user.superAdmin,
                spaceAdmin: user.spaceAdmin,
                orgAdmin: user.admin,
            });
        });
    }

    getActiveKey = () => {
        const startPaths = [
            '/organizations',
            '/users',
            '/settings',
            '/locations',
            '/bookings'
        ];
        let path = this.props.router.pathname;
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
        if (this.state.superAdmin) {
            orgItem = (
                <li className="nav-item">
                    <Nav.Link as={Link} eventKey="/organizations" href="/organizations"><IconBox className="feather" /> {this.props.t("organizations")}</Nav.Link>
                </li>
            );
        }
        let orgAdminItems = <></>;
        if (this.state.orgAdmin) {
            orgAdminItems = (
                <>
                    <li className="nav-item">
                        <Nav.Link as={Link} eventKey="/users" href="/users"><IconUsers className="feather" /> {this.props.t("users")}</Nav.Link>
                    </li>
                    <li className="nav-item">
                        <Nav.Link as={Link} eventKey="/settings" href="/settings"><IconSettings className="feather" /> {this.props.t("settings")}</Nav.Link>
                    </li>
                    <li className="nav-item" hidden={window.location.host.split(':').shift() !== 'app.seatsurfing.io'}>
                        <Nav.Link as={Link} eventKey="/upgrade-cloud" href="/upgrade-cloud"><UpgradeIcon className="feather" /> {this.props.t("upgradeCloud")}</Nav.Link>
                    </li>
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
                        {orgAdminItems}
                        {orgItem}
                        <li className="nav-item">
                            <Nav.Link onClick={(e) => {e.preventDefault(); window.location.href="/ui/";}}><IconExternalLink className="feather" /> {this.props.t("bookingui")}</Nav.Link>
                        </li>
                    </ul>
                </div>
            </Nav>
        );
    }
}

export default withTranslation(['admin'])(withReadyRouter(SideBar as any));
