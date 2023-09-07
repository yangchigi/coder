import AlertTitle from "@mui/material/AlertTitle"
import Button from "@mui/material/Button"
import Link from "@mui/material/Link"
import Box from "@mui/material/Box"
import { Alert } from "components/Alert/Alert"
import { ReactNode, createContext, useContext, useState } from "react"

const REVIEW_RESULTS_URL =
  "https://cdr.jfrog.io/ui/scans-list/repositories/docker-local/scan-descendants"

const SeverityWarningContext = createContext<
  | {
      severity?: string
      setSeverity: (value: string) => void
    }
  | undefined
>(undefined)

export const SeverityWarningProvider = ({
  children,
}: {
  children: ReactNode
}) => {
  const [severity, setSeverity] = useState<undefined | string>()

  return (
    <SeverityWarningContext.Provider value={{ severity, setSeverity }}>
      {children}
    </SeverityWarningContext.Provider>
  )
}

export const SeverityWarningBanner = () => {
  const { severity } = useSeverityWarning()
  const crit = extractCritValue(severity as string)
  const high = extractHigh(severity as string)

  return crit && high ? (
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
            href={REVIEW_RESULTS_URL}
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
            {crit && (
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
                {crit} critical
              </Box>
            )}
            {high && (
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
                {high} high
              </Box>
            )}
          </Box>
        </Box>
      </Box>
    </Alert>
  ) : null
}

export const useSeverityWarning = () => {
  const context = useContext(SeverityWarningContext)
  if (context === undefined) {
    throw new Error(
      "useSeverityWarning must be used within a SeverityWarningProvider",
    )
  }
  return context
}

function extractCritValue(inputString: string): number | undefined {
  const regex = /crit\((\d+)\)/i // Match "crit" (case-insensitive) followed by a number
  const match = inputString.match(regex)

  if (match) {
    const critValue = parseInt(match[1])
    return critValue
  } else {
    return undefined // Return undefined if "crit" is not found
  }
}

function extractHigh(inputString: string): number | undefined {
  const regex = /high\((\d+)\)/i // Match "high" (case-insensitive) followed by a number
  const match = inputString.match(regex)

  if (match) {
    const highValue = parseInt(match[1])
    return highValue
  } else {
    return undefined // Return undefined if "high" is not found
  }
}
