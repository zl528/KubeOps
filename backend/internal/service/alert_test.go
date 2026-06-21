package service

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestAlertService_ListRules(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	rules, err := svc.ListRules(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) < 5 {
		t.Errorf("expected at least 5 default rules, got %d", len(rules))
	}

	foundCPU := false
	for _, rule := range rules {
		if rule.ID == "high-cpu" {
			foundCPU = true
			if rule.Name != "CPU 使用率过高" {
				t.Errorf("expected name 'CPU 使用率过高', got %s", rule.Name)
			}
			if rule.Threshold != 80 {
				t.Errorf("expected threshold 80, got %f", rule.Threshold)
			}
		}
	}

	if !foundCPU {
		t.Error("expected to find high-cpu rule")
	}
}

func TestAlertService_CreateRule(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	newRule := AlertRule{
		Name:        "测试规则",
		Description: "测试描述",
		Metric:      "test_metric",
		Condition:   "greater_than",
		Threshold:   50,
		Severity:    "info",
		Enabled:     true,
	}

	err := svc.CreateRule(context.Background(), newRule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules, _ := svc.ListRules(context.Background())
	found := false
	for _, rule := range rules {
		if rule.Name == "测试规则" {
			found = true
			if rule.ID == "" {
				t.Error("expected rule ID to be set")
			}
		}
	}

	if !found {
		t.Error("expected to find created rule")
	}
}

func TestAlertService_DeleteRule(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	err := svc.DeleteRule(context.Background(), "high-cpu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules, _ := svc.ListRules(context.Background())
	for _, rule := range rules {
		if rule.ID == "high-cpu" {
			t.Error("expected high-cpu rule to be deleted")
		}
	}
}

func TestAlertService_AddAlert(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	alert := AlertHistory{
		RuleID:    "high-cpu",
		RuleName:  "CPU 使用率过高",
		Severity:  "warning",
		Message:   "CPU 使用率达到 85%",
		Value:     85,
		Threshold: 80,
	}

	svc.AddAlert(alert)

	alerts, _ := svc.ListAlerts(context.Background(), "")
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Status != "firing" {
		t.Errorf("expected status 'firing', got %s", alerts[0].Status)
	}
}

func TestAlertService_ResolveAlert(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	alert := AlertHistory{
		RuleID:    "high-cpu",
		RuleName:  "CPU 使用率过高",
		Severity:  "warning",
		Message:   "CPU 使用率达到 85%",
		Value:     85,
		Threshold: 80,
	}

	svc.AddAlert(alert)

	alerts, _ := svc.ListAlerts(context.Background(), "")
	if len(alerts) == 0 {
		t.Fatal("expected at least 1 alert")
	}

	err := svc.ResolveAlert(context.Background(), alerts[0].ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolved, _ := svc.GetAlert(context.Background(), alerts[0].ID)
	if resolved.Status != "resolved" {
		t.Errorf("expected status 'resolved', got %s", resolved.Status)
	}

	if resolved.ResolvedAt == nil {
		t.Error("expected ResolvedAt to be set")
	}
}

func TestAlertService_GetAlertStats(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	svc.AddAlert(AlertHistory{RuleID: "1", Severity: "warning", Message: "test1"})
	svc.AddAlert(AlertHistory{RuleID: "2", Severity: "critical", Message: "test2"})

	stats := svc.GetAlertStats(context.Background())

	if stats["total"] != 2 {
		t.Errorf("expected total 2, got %d", stats["total"])
	}

	if stats["firing"] != 2 {
		t.Errorf("expected firing 2, got %d", stats["firing"])
	}
}

func TestAlertService_ListAlertsBySeverity(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	svc.AddAlert(AlertHistory{RuleID: "1", Severity: "warning", Message: "test1"})
	svc.AddAlert(AlertHistory{RuleID: "2", Severity: "critical", Message: "test2"})
	svc.AddAlert(AlertHistory{RuleID: "3", Severity: "warning", Message: "test3"})

	warnings, _ := svc.ListAlerts(context.Background(), "warning")
	if len(warnings) != 2 {
		t.Errorf("expected 2 warning alerts, got %d", len(warnings))
	}

	criticals, _ := svc.ListAlerts(context.Background(), "critical")
	if len(criticals) != 1 {
		t.Errorf("expected 1 critical alert, got %d", len(criticals))
	}
}

func TestAlertService_EnableDisableRule(t *testing.T) {
	client := fake.NewSimpleClientset()
	svc := NewAlertService(client)

	err := svc.EnableRule(context.Background(), "high-cpu", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule, _ := svc.GetRule(context.Background(), "high-cpu")
	if rule.Enabled {
		t.Error("expected rule to be disabled")
	}

	err = svc.EnableRule(context.Background(), "high-cpu", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule, _ = svc.GetRule(context.Background(), "high-cpu")
	if !rule.Enabled {
		t.Error("expected rule to be enabled")
	}
}
