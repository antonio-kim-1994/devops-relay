package handler

type HealthCheckRequest struct {
	ApplicationName string `json:"application_name"`
	Org             string `json:"org"`
	Branch          string `json:"branch"`
}

type ServiceInfo struct {
	Date                 string `json:"date" binding:"required"`
	Org                  string `json:"org" binding:"required"`
	Operator             string `json:"operator" binding:"required"`
	Repo                 string `json:"repo" binding:"required"`
	DockerTag            string `json:"docker_tag" binding:"required"`
	CommitMessage        string `json:"commit_message" binding:"required"`
	SlackWebhookUrl      string `json:"slack_webhook_url" binding:"required"`
	Branch               string `json:"branch" binding:"required"`
	ApplicationName      string `json:"application_name" binding:"required"`
	ApplicationNamespace string `json:"application_namespace" binding:"required"`
}

type SlackResponse struct {
	Button      ButtonValue `json:"button"`
	User        User        `json:"user"`
	ResponseURL string      `json:"response_url"`
}

// Button Value
// Org/Branch/ApplicationName/ApplicationNamespace/deploy/approve, reject
type ButtonValue struct {
	Org                  string `json:"org"`
	Branch               string `json:"branch"`
	ApplicationName      string `json:"application_name"`
	ApplicationNamespace string `json:"application_namespace"`
	RequestType          string `json:"request_type"`
	Result               string `json:"result"`
}

type User struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
