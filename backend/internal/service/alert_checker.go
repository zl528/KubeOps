package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type AlertChecker struct {
	alertSvc   *AlertService
	monitorSvc *MonitorService
	interval   time.Duration
}

func NewAlertChecker(alertSvc *AlertService, monitorSvc *MonitorService) *AlertChecker {
	return &AlertChecker{
		alertSvc:   alertSvc,
		monitorSvc: monitorSvc,
		interval:   1 * time.Minute,
	}
}

func (ac *AlertChecker) Start(ctx context.Context) {
	go func() {
		log.Println("[ALERT] Starting alert checker...")
		ticker := time.NewTicker(ac.interval)
		defer ticker.Stop()

		// Run immediately on start
		ac.check(ctx)

		for {
			select {
			case <-ctx.Done():
				log.Println("[ALERT] Alert checker stopped")
				return
			case <-ticker.C:
				ac.check(ctx)
			}
		}
	}()
}

func (ac *AlertChecker) check(ctx context.Context) {
	rules, err := ac.alertSvc.ListRules(ctx)
	if err != nil {
		log.Printf("[ALERT] Error listing rules: %v", err)
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		value, err := ac.queryMetric(ctx, rule.Metric)
		if err != nil {
			log.Printf("[ALERT] Error querying metric %s: %v", rule.Metric, err)
			continue
		}

		triggered := ac.evaluateCondition(value, rule.Condition, rule.Threshold)
		if triggered {
			ac.alertSvc.FireAlert(ctx, rule.ID, rule.Name, rule.Severity,
				rule.Description, value, rule.Threshold)
			log.Printf("[ALERT] Alert fired: %s (value=%.2f, threshold=%.2f)", rule.Name, value, rule.Threshold)
		}
	}
}

func (ac *AlertChecker) queryMetric(ctx context.Context, metric string) (float64, error) {
	promURL := ac.monitorSvc.getPrometheusURL(ctx)
	if promURL == "" {
		return 0, nil
	}
	ac.monitorSvc.prometheusURL = promURL

	query := ""
	switch metric {
	case "cpu_usage":
		query = `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	case "memory_usage":
		query = `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	case "disk_usage":
		query = `(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100`
	default:
		return 0, nil
	}

	result, err := ac.monitorSvc.QueryPrometheus(query)
	if err != nil {
		return 0, err
	}

	if data, ok := result.(PrometheusData); ok && len(data.Result) > 0 {
		var metric PrometheusMetric
		if err := json.Unmarshal(data.Result[0], &metric); err == nil {
			if len(metric.Value) >= 2 {
				if val, ok := metric.Value[1].(string); ok {
					var usage float64
					fmt.Sscanf(val, "%f", &usage)
					return usage, nil
				}
			}
		}
	}

	return 0, nil
}

func (ac *AlertChecker) evaluateCondition(value float64, condition string, threshold float64) bool {
	switch condition {
	case "greater_than", "gt":
		return value > threshold
	case "greater_than_or_equal", "gte":
		return value >= threshold
	case "less_than", "lt":
		return value < threshold
	case "less_than_or_equal", "lte":
		return value <= threshold
	case "equals", "eq":
		return value == threshold
	default:
		return false
	}
}
