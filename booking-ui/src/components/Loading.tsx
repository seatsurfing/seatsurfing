import { WithTranslation, withTranslation } from 'next-i18next';
import React from 'react';
import { Loader as IconLoad } from 'react-feather';

interface State {
}

interface Props extends WithTranslation {
    showText: boolean
    paddingTop: boolean
    visible: boolean
}

class Loading extends React.Component<Props, State> {
    render() {
        let text = "Loading...";
        let paddingTop = (this.props.paddingTop ?? true) ? 'padding-top' : '';
        let display = (this.props.visible === undefined || this.props.visible === true) ? 'display-block' : 'display-none';
        return (
            <div className={`${paddingTop} ${display} center loading-overlay`}><IconLoad className="feather loader" />{text}</div>
        );
    }
}

export default withTranslation()(Loading as any);
