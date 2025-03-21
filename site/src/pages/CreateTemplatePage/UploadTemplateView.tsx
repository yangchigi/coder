import { useQuery, useMutation } from "react-query";
import { useNavigate } from "react-router-dom";
import {
  templateVersionLogs,
  JobError,
  templateVersionVariables,
} from "api/queries/templates";
import { uploadFile } from "api/queries/files";
import { useOrganizationId } from "contexts/auth/useOrganizationId";
import { useDashboard } from "modules/dashboard/useDashboard";
import { CreateTemplateForm } from "./CreateTemplateForm";
import { firstVersionFromFile, getFormPermissions, newTemplate } from "./utils";
import { FC } from "react";
import { CreateTemplatePageViewProps } from "./types";

export const UploadTemplateView: FC<CreateTemplatePageViewProps> = ({
  onCreateTemplate,
  onOpenBuildLogsDrawer,
  variablesSectionRef,
  isCreating,
  error,
}) => {
  const navigate = useNavigate();
  const organizationId = useOrganizationId();

  const dashboard = useDashboard();
  const formPermissions = getFormPermissions(dashboard.entitlements);

  const uploadFileMutation = useMutation(uploadFile());
  const uploadedFile = uploadFileMutation.data;

  const isJobError = error instanceof JobError;
  const templateVersionLogsQuery = useQuery({
    ...templateVersionLogs(isJobError ? error.version.id : ""),
    enabled: isJobError,
  });

  const missedVariables = useQuery({
    ...templateVersionVariables(isJobError ? error.version.id : ""),
    enabled:
      isJobError && error.job.error_code === "REQUIRED_TEMPLATE_VARIABLES",
  });

  return (
    <CreateTemplateForm
      {...formPermissions}
      onOpenBuildLogsDrawer={onOpenBuildLogsDrawer}
      variablesSectionRef={variablesSectionRef}
      variables={missedVariables.data}
      error={error}
      isSubmitting={isCreating}
      onCancel={() => navigate(-1)}
      jobError={isJobError ? error.job.error : undefined}
      logs={templateVersionLogsQuery.data}
      upload={{
        onUpload: uploadFileMutation.mutateAsync,
        isUploading: uploadFileMutation.isLoading,
        onRemove: uploadFileMutation.reset,
        file: uploadFileMutation.variables,
      }}
      onSubmit={async (formData) => {
        await onCreateTemplate({
          organizationId,
          version: firstVersionFromFile(
            uploadedFile!.hash,
            formData.user_variable_values,
          ),
          template: newTemplate(formData),
        });
      }}
    />
  );
};
