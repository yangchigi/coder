import { type FC } from "react";
import { useProxy } from "contexts/ProxyContext";
import { Section } from "../Section";
import { WorkspaceProxyView } from "./WorkspaceProxyView";

export const WorkspaceProxyPage: FC = () => {
  const description =
    "Workspace proxies improve terminal and web app connections to workspaces.";

  const {
    proxyLatencies,
    proxies,
    error: proxiesError,
    isFetched: proxiesFetched,
    isLoading: proxiesLoading,
    proxy,
  } = useProxy();

  return (
    <Section
      title="Workspace Proxies"
      css={(theme) => ({
        "& code": {
          background: theme.palette.divider,
          fontSize: 12,
          padding: "2px 4px",
          color: theme.palette.text.primary,
          borderRadius: 2,
        },
      })}
      description={description}
      layout="fluid"
    >
      <WorkspaceProxyView
        proxyLatencies={proxyLatencies}
        proxies={proxies}
        isLoading={proxiesLoading}
        hasLoaded={proxiesFetched}
        getWorkspaceProxiesError={proxiesError}
        preferredProxy={proxy.proxy}
      />
    </Section>
  );
};

export default WorkspaceProxyPage;
