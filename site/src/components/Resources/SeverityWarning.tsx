import Button from "@mui/material/Button"
import Link from "@mui/material/Link"
import { Alert } from "components/Alert/Alert"
import { ReactNode, createContext, useContext, useState } from "react"

const REVIEW_RESULTS_URL = "https://cdr.jfrog.io/ui/reports"

const SeverityWarningContext = createContext<
  { open: boolean; setOpen: (value: boolean) => void } | undefined
>(undefined)

export const SeverityWarningProvider = ({
  children,
}: {
  children: ReactNode
}) => {
  const [open, setOpen] = useState(false)

  return (
    <SeverityWarningContext.Provider value={{ open, setOpen }}>
      {children}
    </SeverityWarningContext.Provider>
  )
}

export const SeverityWarningBanner = () => (
  <Alert
    severity="warning"
    actions={
      <Button
        variant="text"
        component={Link}
        href={REVIEW_RESULTS_URL}
        target="_blank"
        rel="noreferrer"
      >
        Review results
      </Button>
    }
  >
    Vulnerabilities have been detected for this workspace
  </Alert>
)

export const useSeverityWarning = () => {
  const context = useContext(SeverityWarningContext)
  if (context === undefined) {
    throw new Error(
      "useSeverityWarning must be used within a SeverityWarningProvider",
    )
  }
  return context
}
