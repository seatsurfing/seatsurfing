import { useTranslation } from "next-export-i18n";

export type TranslationFunc = (key: string, view?: object) => any;

export const withTranslation = (Component: any) => {
    return function(props: any) {
        const { t } = useTranslation();
        return <Component {...props} t={t} />;
    };
};
