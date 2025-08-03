package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

const argoRolloutsUrl = "http://argocd-argo-rollouts-dashboard.argocd.svc.cluster.local"

func promoteApplication(rolloutsName, namespace string) error {
	uriPath := fmt.Sprintf("api/v1/rollouts/%s/%s/promote", namespace, rolloutsName)

	payload := map[string]string{
		"name":      rolloutsName,
		"namespace": namespace,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Err(err).Msg("failed to marshal payload")
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", argoRolloutsUrl, uriPath), bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return err
	}
	defer resp.Body.Close()

	return nil
}

func abortApplication(rolloutsName, namespace string) error {
	uriPath := fmt.Sprintf("api/v1/rollouts/%s/%s/abort", namespace, rolloutsName)

	payload := map[string]string{
		"name":      rolloutsName,
		"namespace": namespace,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Err(err).Msg("failed to marshal payload")
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", argoRolloutsUrl, uriPath), bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("")
		return err
	}
	defer resp.Body.Close()

	return nil
}
