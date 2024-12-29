import React from 'react';
import { Navbar, Nav, Modal, Button, Form, Badge, Container, NavLink } from 'react-bootstrap';
import { Ajax, User, MergeRequest, AjaxCredentials } from 'flexspace-commons';
import RuntimeConfig from './RuntimeConfig';
import { Users as IconMerge, Bell as IconAlert, Settings as IconSettings, Calendar as IconCalendar, PlusSquare as IconPlus } from 'react-feather';
import { WithTranslation, withTranslation } from 'next-i18next';
import { NextRouter } from 'next/router';
import withReadyRouter from './withReadyRouter';
import Link from 'next/link';

interface State {
    redirect: string | null
    showMergeInit: boolean
    showMergeNextStep: boolean
    showMergeRequests: boolean
    targetUserEmail: string
    invalidTargetUserEmail: boolean
    mergeRequests: MergeRequest[]
    allowMergeInit: boolean
    allowAdmin: boolean
}

interface Props extends WithTranslation {
    router: NextRouter
}

class NavBar extends React.Component<Props, State> {
    constructor(props: any) {
        super(props);
        this.state = {
            redirect: null,
            showMergeInit: false,
            showMergeNextStep: false,
            showMergeRequests: false,
            targetUserEmail: "",
            invalidTargetUserEmail: false,
            mergeRequests: [],
            allowMergeInit: false,
            allowAdmin: false
        };
    }

    componentDidMount = () => {
        if (!Ajax.CREDENTIALS.accessToken) {
            return;
        }
        this.loadData();
    }

    loadData = () => {
        User.getMergeRequests().then(list => {
            this.setState({ mergeRequests: list });
        });
        User.getSelf().then(user => {
            if (user.email === user.atlassianId) {
                this.setState({ allowMergeInit: true });
            }
            if (user.role >= User.UserRoleSpaceAdmin) {
                this.setState({ allowAdmin: true });
            }
        });
    }

    logOut = (e: any) => {
        e.preventDefault();
        let logoutUrl = Ajax.CREDENTIALS.logoutUrl;
        Ajax.CREDENTIALS = new AjaxCredentials();
        Ajax.PERSISTER.deleteCredentialsFromSessionStorage().then(() => {
            if (logoutUrl) {
                window.location.href = logoutUrl;
                return;
            }
            this.setState({
                redirect: "/login"
            });
        });
    }

    gotoAdmin = (e: any) => {
        e.preventDefault();
        if (typeof window !== 'undefined') {
            window.location.href = "/admin/dashboard/";
        }
    }

    showMergeModal = (e: any) => {
        e.preventDefault();
        this.setState({ showMergeInit: true });
    }

    showMergeRequestsModal = (e: any) => {
        e.preventDefault();
        this.setState({ showMergeRequests: true });
    }

    initMerge = () => {
        User.initMerge(this.state.targetUserEmail).then(() => {
            this.setState({
                showMergeInit: false,
                showMergeNextStep: true
            })
        }).catch(() => {
            this.setState({
                invalidTargetUserEmail: true
            });
        });
    }

    openWebUI = () => {
        if (typeof window !== 'undefined') {
            window.open(process.env.PUBLIC_URL);
        }
    }

    acceptMergeRequest = (id: string) => {
        User.finishMerge(id).then(() => {
            this.setState({
                showMergeRequests: false
            });
            this.loadData();
        });
    }

    renderMergeRequest = (item: MergeRequest) => {
        return (
            <p key={item.id}>
                {item.email} <Button size="sm" onClick={() => this.acceptMergeRequest(item.id)}>{this.props.t("accept")}</Button>
            </p>
        );
    }

    render() {
        if (this.state.redirect != null) {
            let target = this.state.redirect;
            this.setState({ redirect: null });
            this.props.router.push(target);
            return <></>;
        }

        let signOffButton = <></>;
        let adminButton = <></>;
        let initMergeButton = <></>;
        let mergeRequestsButton = <></>;
        let collapsable = <></>;
        let buddies = <></>;

        if (!RuntimeConfig.EMBEDDED) {
            if (this.state.allowAdmin) {
                adminButton = <Nav.Link onClick={this.gotoAdmin}>{this.props.t("administration")}</Nav.Link>;
            }
            signOffButton = <Nav.Link onClick={this.logOut}>{this.props.t("signout")}</Nav.Link>;
            if (this.state.mergeRequests.length > 0) {
                mergeRequestsButton = <Nav.Link onClick={this.showMergeRequestsModal} className="icon-link"><IconAlert className="feather feather-lg" /><Badge pill={true} bg="light" className="badge-top">{this.state.mergeRequests.length}</Badge></Nav.Link>;
            }
        } else {
            if (this.state.allowMergeInit) {
                initMergeButton = <Nav.Link onClick={this.showMergeModal} className="icon-link"><IconMerge className="feather feather-lg" /></Nav.Link>;
            }
        }

        if (RuntimeConfig.INFOS.showNames && !RuntimeConfig.INFOS.disableBuddies) {
            buddies = <Nav.Link as={Link} eventKey="/buddies" href="/buddies">{RuntimeConfig.EMBEDDED ? <IconCalendar className="feather feather-lg" /> : this.props.t("myBuddies")}</Nav.Link>
        }

        collapsable = (
            <>
                <Nav activeKey={this.props.router.pathname}>
                    <Nav.Link as={Link} eventKey="/search" href="/search">{RuntimeConfig.EMBEDDED ? <IconPlus className="feather feather-lg" /> : this.props.t("bookSeat")}</Nav.Link>
                    <Nav.Link as={Link} eventKey="/bookings" href="/bookings">{RuntimeConfig.EMBEDDED ? <IconCalendar className="feather feather-lg" /> : this.props.t("myBookings")}</Nav.Link>
                    {buddies}
                    <Nav.Link as={Link} eventKey="/preferences" href="/preferences">{RuntimeConfig.EMBEDDED ? <IconSettings className="feather feather-lg" /> : this.props.t("preferences")}</Nav.Link>
                    {adminButton}
                    {signOffButton}
                </Nav>
                <Nav className="ms-auto">
                    {initMergeButton}
                    {mergeRequestsButton}
                </Nav>
            </>
        );

        if (!RuntimeConfig.EMBEDDED) {
            collapsable = (
                <>
                    <Navbar.Toggle aria-controls="basic-navbar-nav" />
                    <Navbar.Collapse id="basic-navbar-nav">
                        {collapsable}
                    </Navbar.Collapse>
                </>
            );
        }


        const logoUrl = RuntimeConfig.INFOS.customLogoUrl || "/ui/seatsurfing.svg";

        return (
            <>
                <Navbar bg="light" variant="light" fixed="top" expand={RuntimeConfig.EMBEDDED ? true : "xl"}>
                    <Container fluid={true}>
                        <Navbar.Brand as={NavLink} to="/search"><img src={logoUrl} alt="Seatsurfing" /></Navbar.Brand>
                        {collapsable}
                    </Container>
                </Navbar>
                <Modal show={this.state.showMergeInit} onHide={() => this.setState({ showMergeInit: false })}>
                    <Modal.Header closeButton>
                        <Modal.Title>{this.props.t("mergeUserAccounts")}</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <p>{this.props.t("mergeUserAccountsHint")}</p>
                        <Form onSubmit={(e) => { e.preventDefault(); this.initMerge() }}>
                            <Form.Group>
                                <Form.Control type="email" placeholder={this.props.t("emailPlaceholder")} required={true} value={this.state.targetUserEmail} onChange={(e: any) => this.setState({ targetUserEmail: e.target.value, invalidTargetUserEmail: false })} isInvalid={this.state.invalidTargetUserEmail} autoFocus={true} />
                                <Form.Control.Feedback type="invalid">{this.props.t("errorInvalidEmail")}</Form.Control.Feedback>
                            </Form.Group>
                        </Form>
                    </Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={() => this.setState({ showMergeInit: false })}>
                            {this.props.t("cancel")}
                        </Button>
                        <Button variant="primary" onClick={this.initMerge}>
                            {this.props.t("requestMerge")}
                        </Button>
                    </Modal.Footer>
                </Modal>
                <Modal show={this.state.showMergeNextStep} onHide={() => this.setState({ showMergeNextStep: false })}>
                    <Modal.Header closeButton>
                        <Modal.Title>{this.props.t("mergeUserAccounts")}</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <p>{this.props.t("mergeUserAccountsNextStepHint")}</p>
                        <Button onClick={this.openWebUI}>{this.props.t("openWebUI")}</Button>
                    </Modal.Body>
                </Modal>
                <Modal show={this.state.showMergeRequests} onHide={() => this.setState({ showMergeRequests: false })}>
                    <Modal.Header closeButton>
                        <Modal.Title>{this.props.t("mergeUserAccounts")}</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <p>{this.props.t("introIncomingMergeRequests")}</p>
                        {this.state.mergeRequests.map(item => this.renderMergeRequest(item))}
                    </Modal.Body>
                </Modal>
            </>
        );
    }
}

export default withTranslation()(withReadyRouter(NavBar as any));
