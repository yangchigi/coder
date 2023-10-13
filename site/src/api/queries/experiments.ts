import * as API from "api/api";
import { ExperimentsResponse } from "api/typesGenerated";
import { getMetadataAsJSON } from "utils/metadata";

export const experiments = () => {
  return {
    queryKey: ["experiments"],
    queryFn: async () =>
      getMetadataAsJSON<ExperimentsResponse>("experiments") ??
      API.getExperiments(),
  };
};
