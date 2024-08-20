package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"cert-tracker/config"
	"cert-tracker/metrics"
	"cert-tracker/utils"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type CertHandler struct {
	rdb          *redis.Client
	tracer       trace.Tracer
	breaker      *gobreaker.CircuitBreaker
	redisLimiter *rate.Limiter
	logger       *zap.Logger
}

type CertResponse struct {
	Domain               string `json:"domain"`
	ExpiryDate           string `json:"expiry_date"`
	IssuedDate           string `json:"issued_date"`
	DaysRemaining        int    `json:"days_remaining"`
	CertificateAuthority string `json:"certificate_authority"`
	HTTPStatus           int    `json:"http_status"`
	Error                string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const numWorkers = 5

func NewCertHandler(rdb *redis.Client, tracer trace.Tracer, logger *zap.Logger, cfg *config.Config) *CertHandler {
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        cfg.CircuitBreakerName,
		MaxRequests: uint32(cfg.CircuitBreakerMaxRequests),
		Interval:    time.Duration(cfg.CircuitBreakerInterval) * time.Second,
		Timeout:     time.Duration(cfg.CircuitBreakerTimeout) * time.Second,
	})

	limiter := rate.NewLimiter(5, 10)

	return &CertHandler{
		rdb:          rdb,
		tracer:       tracer,
		breaker:      breaker,
		redisLimiter: limiter,
		logger:       logger,
	}
}

func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (h *CertHandler) CheckCertificateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := h.tracer.Start(r.Context(), "CheckCertificateHandler")
		defer span.End()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		// Structured Logging with Context
		logger := h.logger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("method", r.Method),
		)

		if err := h.redisLimiter.Wait(ctx); err != nil {
			logger.Warn("Rate limit exceeded", zap.Error(err))
			handleAPIError(w, logger, err, "rate_limit_exceeded", "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var domains []string
			if err := json.NewDecoder(r.Body).Decode(&domains); err != nil {
				logger.Warn("Invalid request body", zap.Error(err))
				handleAPIError(w, logger, err, "invalid_request", "Invalid request body", http.StatusBadRequest)
				return
			}

			if len(domains) == 0 {
				err := fmt.Errorf("no domains provided")
				logger.Warn("No domains provided", zap.Error(err))
				handleAPIError(w, logger, err, "no_domains_provided", "No domains provided", http.StatusBadRequest)
				return
			}

			responses := h.processDomains(ctx, domains, logger)
			jsonResponse(w, http.StatusOK, responses)

		case http.MethodGet:
			domain := r.URL.Query().Get("domain")
			if domain == "" {
				err := fmt.Errorf("domain parameter is required")
				logger.Warn("Domain parameter is missing", zap.Error(err))
				handleAPIError(w, logger, err, "missing_domain", "Domain parameter is required", http.StatusBadRequest)
				return
			}

			response := h.processDomain(ctx, domain, logger)
			jsonResponse(w, response.HTTPStatus, response)

		default:
			err := fmt.Errorf("method not allowed")
			logger.Warn("Invalid method", zap.Error(err))
			handleAPIError(w, logger, err, "method_not_allowed", "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *CertHandler) processDomains(ctx context.Context, domains []string, logger *zap.Logger) []CertResponse {
	var wg sync.WaitGroup
	responseCh := make(chan CertResponse, len(domains))
	semaphore := make(chan struct{}, numWorkers)

	for _, domain := range domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			responseCh <- h.processDomain(ctx, domain, logger)
		}(domain)
	}

	wg.Wait()
	close(responseCh)

	var responses []CertResponse
	for response := range responseCh {
		responses = append(responses, response)
	}

	return responses
}

func (h *CertHandler) processDomain(ctx context.Context, domain string, logger *zap.Logger) CertResponse {
	ctx, span := h.tracer.Start(ctx, "processDomain", trace.WithAttributes(
		attribute.String("domain", domain),
	))
	defer span.End()

	logger = logger.With(zap.String("domain", domain))

	if !utils.IsValidDomain(domain) {
		err := fmt.Errorf("invalid domain")
		logger.Warn("Invalid domain", zap.Error(err))
		return CertResponse{
			Domain:     domain,
			Error:      "invalid domain",
			HTTPStatus: http.StatusBadRequest,
		}
	}

	certInfo, err := utils.GetCertificateInfo(ctx, domain)
	response := CertResponse{
		Domain:     domain,
		HTTPStatus: http.StatusOK,
	}

	if err != nil {
		logger.Error("Failed to get certificate info", zap.Error(err))
		return CertResponse{
			Domain:     domain,
			Error:      "internal server error",
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	response.ExpiryDate = certInfo.ExpiryDate
	response.IssuedDate = certInfo.IssuedDate
	response.DaysRemaining = certInfo.DaysRemaining
	response.CertificateAuthority = certInfo.CertificateAuthority

	logger.Info("Successfully retrieved certificate info")

	// Update Prometheus metrics with certificate info
	metrics.UpdateCertMetrics(domain, certInfo.ExpiryDate, certInfo.IssuedDate, certInfo.DaysRemaining, certInfo.CertificateAuthority, http.StatusText(response.HTTPStatus), response.Error)

	redisCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = 10 * time.Second

	err = backoff.Retry(func() error {
		_, err := h.breaker.Execute(func() (interface{}, error) {
			pipe := h.rdb.Pipeline()
			h.storeCertInfoInPipeline(redisCtx, pipe, domain, response)
			h.addDomainToSetInPipeline(redisCtx, pipe, domain)
			_, err := pipe.Exec(redisCtx)
			if err != nil {
				return nil, errors.Wrapf(err, "failed Redis pipeline execution for domain %s", domain)
			}
			return nil, nil
		})
		return err
	}, retryPolicy)

	if err != nil {
		logger.Error("Redis operations failed after retries", zap.Error(err))
	} else {
		logger.Info("Successfully executed Redis operations")
	}

	return response
}

func (h *CertHandler) storeCertInfoInPipeline(ctx context.Context, pipe redis.Pipeliner, domain string, info CertResponse) {
	key := "cert:" + domain
	data := map[string]interface{}{
		"IssuedDate":           info.IssuedDate,
		"ExpiryDate":           info.ExpiryDate,
		"DaysRemaining":        info.DaysRemaining,
		"CertificateAuthority": info.CertificateAuthority,
		"Error":                info.Error,
	}
	pipe.HSet(ctx, key, data)
}

func (h *CertHandler) addDomainToSetInPipeline(ctx context.Context, pipe redis.Pipeliner, domain string) {
	pipe.SAdd(ctx, "domains", domain)
}

func handleAPIError(w http.ResponseWriter, logger *zap.Logger, err error, code, message string, status int) {
	logger.Error(message, zap.Error(err))
	response := ErrorResponse{
		Code:    code,
		Message: message,
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
