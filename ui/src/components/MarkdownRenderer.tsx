import React from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface Props {
  children: string;
  inline?: boolean;
}

const MarkdownRenderer: React.FC<Props> = ({ children, inline }) => (
  <Markdown
    remarkPlugins={[remarkGfm]}
    skipHtml={true}
    components={{
      a: ({ href, children }) =>
        href ? (
          <a href={href} target="_blank" rel="noopener noreferrer">
            {children}
          </a>
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
