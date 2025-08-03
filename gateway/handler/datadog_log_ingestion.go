package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"os"
)

func sendDeployInfoToDatadog(s *ServiceInfo) error {
	body := []datadogV2.HTTPLogItem{
		{
			Ddsource: datadog.PtrString("go"),
			Ddtags:   datadog.PtrString(fmt.Sprintf("env:%s", s.Branch)),
			Hostname: datadog.PtrString("antonio.devops-relay.com"),
			Message:  fmt.Sprintf("[%s] %s service deployed.", s.Branch, s.ApplicationName),
			Service:  datadog.PtrString("devops-gateway"),
			AdditionalProperties: map[string]interface{}{
				"date":                  fmt.Sprintf("%s", s.Date),
				"org":                   fmt.Sprintf("%s", s.Org),
				"repository":            fmt.Sprintf("%s", s.Repo),
				"branch":                fmt.Sprintf("%s", s.Branch),
				"application_name":      fmt.Sprintf("%s", s.ApplicationName),
				"application_namespace": fmt.Sprintf("%s", s.ApplicationNamespace),
				"operator":              fmt.Sprintf("%s", s.Operator),
				"commit":                fmt.Sprintf("%s", s.CommitMessage),
			},
		},
	}
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewLogsApi(apiClient)
	resp, r, err := api.SubmitLog(ctx, body, *datadogV2.NewSubmitLogOptionalParameters())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LogsApi.SubmitLog`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprintf(os.Stdout, "Response from `LogsApi.SubmitLog`:%+v", responseContent)
	return nil
}
