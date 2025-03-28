import { JobError } from "api/queries/templates";
import { BuildLogsDrawer } from "./BuildLogsDrawer";
import type { Meta, StoryObj } from "@storybook/react";
import {
  MockProvisionerJob,
  MockTemplateVersion,
  MockWorkspaceBuildLogs,
} from "testHelpers/entities";
import { withWebSocket } from "testHelpers/storybook";

const meta: Meta<typeof BuildLogsDrawer> = {
  title: "pages/CreateTemplatePage/BuildLogsDrawer",
  component: BuildLogsDrawer,
  args: {
    open: true,
  },
};

export default meta;
type Story = StoryObj<typeof BuildLogsDrawer>;

export const Loading: Story = {};

export const MissingVariables: Story = {
  args: {
    templateVersion: MockTemplateVersion,
    error: new JobError(
      {
        ...MockProvisionerJob,
        error_code: "REQUIRED_TEMPLATE_VARIABLES",
      },
      MockTemplateVersion,
    ),
  },
};

export const Logs: Story = {
  args: {
    templateVersion: {
      ...MockTemplateVersion,
      job: {
        ...MockTemplateVersion.job,
        status: "running",
      },
    },
  },
  decorators: [withWebSocket],
  parameters: {
    webSocket: {
      messages: MockWorkspaceBuildLogs.map((log) => JSON.stringify(log)),
    },
  },
};
