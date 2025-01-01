import React, { JSX } from 'react';
import { NextRouter, useRouter } from 'next/router';
import { NextComponentType, NextPageContext } from 'next';
import { BaseContext } from 'next/dist/shared/lib/utils';

export type WithRouterProps = {
    router: NextRouter;
};

export type ExcludeRouterProps<P> = Pick<
    P,
    Exclude<keyof P, keyof WithRouterProps>
>;

export default function withReadyRouter<
    P extends WithRouterProps,
    C extends BaseContext = NextPageContext
>(
    ComposedComponent: NextComponentType<C, any, P>
): React.ComponentType<ExcludeRouterProps<P>> {
    function WithReadyRouterWrapper(props: any): JSX.Element {
        const router = useRouter();
        if (!router.isReady) {
            return <></>;
        }
        return <ComposedComponent router={router} {...props} />;
    }
    WithReadyRouterWrapper.getInitialProps = ComposedComponent.getInitialProps;
    // This is needed to allow checking for custom getInitialProps in _app
    (WithReadyRouterWrapper as any).origGetInitialProps = (ComposedComponent as any).origGetInitialProps;
    if (process.env.NODE_ENV !== 'production') {
        const name = ComposedComponent.displayName || ComposedComponent.name || 'Unknown';
        WithReadyRouterWrapper.displayName = `withReadyRouter(${name})`;
    }

    return WithReadyRouterWrapper;
}
