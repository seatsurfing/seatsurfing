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
    components={{
      a: ({ href, children }) => (
        <a href={href} target="_blank" rel="noopener noreferrer">
          {children}
        </a>
      ),
      ...(inline ? { p: ({ children }) => <>{children}</> } : {}),
    }}
  >
    {children}
  </Markdown>
);

export default MarkdownRenderer;
