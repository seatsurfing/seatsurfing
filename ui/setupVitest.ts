// fix test warnings "The current testing environment is not configured to support act(...)"
(globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
