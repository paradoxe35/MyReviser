package ai

import (
	"errors"
	"strings"
	"sync"
)

// ReasoningStyle is how a provider wants to be told to think less. There is no portable parameter:
// OpenAI 400s on reasoning_effort for a non-reasoning model, and OpenRouter 400s if it sees both
// shapes at once.
type ReasoningStyle int

const (
	ReasoningOpenAIEffort ReasoningStyle = iota
	ReasoningOpenRouter
)

// ReasoningAware is a provider with a reasoning parameter to send. Anthropic has none — its
// thinking is opt-in, and a correction does not want it.
type ReasoningAware interface {
	SetLowReasoning(low bool)
}

// DetectReasoningStyle matches by host, not provider type: a custom provider can point at
// OpenRouter too, and it's the one OpenAI-compatible gateway with its own request shape.
func DetectReasoningStyle(baseURL string) ReasoningStyle {
	if strings.Contains(strings.ToLower(baseURL), "openrouter.ai") {
		return ReasoningOpenRouter
	}
	return ReasoningOpenAIEffort
}

// rejected caches endpoint/model pairs that refused a reasoning parameter, so the wasted round trip
// happens once per launch rather than on every correction. Process-scoped: persisting it would risk
// a stale rejection outliving a model upgrade.
var rejected sync.Map

// withReasoningFallback sends with the reasoning parameter, retrying without it on rejection. It
// matches on status code rather than message text, since every provider words the error
// differently, and only caches the rejection once dropping the parameter is confirmed to fix it —
// a 400 has other causes too, and caching on status alone could silence reasoning for a model that
// never objected.
func withReasoningFallback(endpoint, model string, wanted bool, send func(includeReasoning bool) (string, error)) (string, error) {
	key := endpoint + "::" + model
	if _, refused := rejected.Load(key); !wanted || refused {
		return send(false)
	}

	result, err := send(true)
	if err == nil || !refusedReasoning(err) {
		return result, err
	}

	// If this fails too, the parameter was never the problem and nothing is cached.
	result, retryErr := send(false)
	if retryErr != nil {
		return "", retryErr
	}

	rejected.Store(key, struct{}{})
	return result, nil
}

func refusedReasoning(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == 400 || apiErr.StatusCode == 422
}
