import React from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import Validation from "@/util/Validation";

interface Props {
  children: string;
  inline?: boolean;
  relativeLinksSameTab?: boolean;
}

const MarkdownRenderer: React.FC<Props> = ({
  children,
  inline,
  relativeLinksSameTab = false,
}) => (
  <Markdown
    remarkPlugins={[remarkGfm]}
    skipHtml={true}
    components={{
      a: ({ href, children }) =>
        href ? (
          !relativeLinksSameTab || !Validation.isRelativeUrl(href) ? (
            <a href={href} target="_blank" rel="noopener noreferrer">
              {children}
            </a>
          ) : (
            <a href={href}>{children}</a>
          )
        ) : (
          <>{children}</>
        ),
      img: () => null,
      ...(inline
        ? {
            p: ({ children }) => <>{children}</>,
            ul: ({ children }) => <>{children}</>,
            ol: ({ children }) => <>{children}</>,
            li: ({ children }) => <>{children} </>,
          }
        : {}),
    }}
  >
    {children}
  </Markdown>
);

export default MarkdownRenderer;
