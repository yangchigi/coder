import AlertTitle from "@mui/material/AlertTitle";
import Button from "@mui/material/Button";
import Link from "@mui/material/Link";
import Box from "@mui/material/Box";
import { Alert } from "components/Alert/Alert";
import { ReactNode, createContext, useContext, useState } from "react";
import { JFrogXrayScan } from "api/typesGenerated";

const SeverityWarningContext = createContext<
  | {
      results?: JFrogXrayScan;
      setResults: (value?: JFrogXrayScan) => void;
    }
  | undefined
>(undefined);

export const SeverityWarningProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const [results, setResults] = useState<undefined | JFrogXrayScan>();

  return (
    <SeverityWarningContext.Provider value={{ results, setResults }}>
      {children}
    </SeverityWarningContext.Provider>
  );
};

export const SeverityWarningBanner = () => {
  const { results } = useSeverityWarning();

  return results && (results.critical > 0 || results.high > 0) ? (
    <Alert
      icon={false}
      severity="warning"
      actions={
        <Box
          sx={{
            padding: "11px 8px",
          }}
        >
          <Button
            variant="text"
            component={Link}
            href={results.results_url}
            target="_blank"
            rel="noreferrer"
          >
            Review results
          </Button>
        </Box>
      }
    >
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          gap: 4,
          padding: "11px 8px",
        }}
      >
        <Box
          component="img"
          src="/icon/jfrog.png"
          sx={{ height: 40, width: 42, display: "block" }}
        />
        <Box>
          <AlertTitle>
            JFrog Xray detected new vulnerabilities for this workspace
          </AlertTitle>

          <Box sx={{ display: "flex", gap: 1.5, mt: 0.25 }}>
            {results.critical > 0 && (
              <Box
                sx={{
                  display: "flex",
                  fontSize: 13,
                  gap: 0.75,
                  color: (theme) => theme.palette.error.light,
                  alignItems: "center",
                }}
              >
                <Box
                  sx={{
                    width: 8,
                    height: 8,
                    backgroundColor: (theme) => theme.palette.error.main,
                    borderRadius: "100%",
                  }}
                />
                {results.critical} critical
              </Box>
            )}
            {results.high && (
              <Box
                sx={{
                  display: "flex",
                  fontSize: 13,
                  gap: 0.75,
                  color: (theme) => theme.palette.warning.light,
                  alignItems: "center",
                }}
              >
                <Box
                  sx={{
                    width: 8,
                    height: 8,
                    backgroundColor: (theme) => theme.palette.warning.main,
                    borderRadius: "100%",
                  }}
                />
                {results.high} high
              </Box>
            )}
          </Box>
        </Box>
      </Box>
    </Alert>
  ) : null;
};

export const useSeverityWarning = () => {
  const context = useContext(SeverityWarningContext);
  if (context === undefined) {
    throw new Error(
      "useSeverityWarning must be used within a SeverityWarningProvider",
    );
  }
  return context;
};
