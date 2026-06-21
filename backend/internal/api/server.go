package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kubeops/ops-kubernetes/internal/config"
	"github.com/kubeops/ops-kubernetes/internal/database"
	"github.com/kubeops/ops-kubernetes/internal/handler"
	"github.com/kubeops/ops-kubernetes/internal/middleware"
	"github.com/kubeops/ops-kubernetes/internal/service"
	"github.com/kubeops/ops-kubernetes/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type ClientGetterInterface interface {
	GetK8sClient() kubernetes.Interface
	GetDynamicClient() dynamic.Interface
	GetRestConfig() *rest.Config
}

func Run(cfg *config.Config) error {
	// Database
	db := database.GetDB()

	// Initialize multi-cluster manager
	multiClusterMgr := InitMultiClusterManager(db.DB)

	// Initialize with fake clients for demo mode
	scheme := runtime.NewScheme()
	demoCM := &ClusterManager{}
	demoCM.k8sClient = fake.NewSimpleClientset()
	demoCM.dynamicClient = dynamicfake.NewSimpleDynamicClient(scheme)
	demoCM.restConfig = &rest.Config{Host: "http://localhost:8080"}
	demoCM.connected = false
	multiClusterMgr.clusters["demo"] = demoCM

	log.Println("KubeOps server starting (demo mode - no cluster connected)")

	// Auth
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "kubeops-secret-key-change-in-production"
	}
	authSvc := service.NewAuthService(db.DB, jwtSecret)
	authHandler := handler.NewAuthHandler(authSvc)
	auditSvc := service.NewAuditService(db.DB)

	// Get client getter from multi-cluster manager
	clientGetter := multiClusterMgr.GetClientGetter()
	backupSvc := service.NewBackupService(clientGetter.GetK8sClient(), clientGetter.GetDynamicClient(), "")

	// Handlers use ClientGetter to get client dynamically per request
	h := handler.NewHandler(clientGetter)
	rh := handler.NewResourceHandler(clientGetter)
	wh := handler.NewWorkloadHandler(clientGetter)
	nh := handler.NewNetworkHandler(clientGetter)
	rbh := handler.NewRBACHandler(clientGetter)
	sh := handler.NewStorageHandler(clientGetter)
	ph := handler.NewPolicyHandler(clientGetter)
	lh := handler.NewLogHandler(clientGetter)
	eh := handler.NewExecHandler(clientGetter)
	mh := handler.NewMonitorHandler(clientGetter)
	alertSvc := service.NewAlertService(clientGetter.GetK8sClient(), db.DB)
	ah := handler.NewAlertHandler(alertSvc)
	auh := handler.NewAuditHandler(auditSvc)
	bh := handler.NewBackupHandler(backupSvc)
	gh := handler.NewGraphHandler(clientGetter)

	// Start alert checker
	monitorSvc := service.NewMonitorService(clientGetter.GetK8sClient(), "")
	alertChecker := service.NewAlertChecker(alertSvc, monitorSvc)
	alertChecker.Start(context.Background())

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/api/health", h.HandleHealth)

	// Auth
	mux.HandleFunc("/api/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("/api/auth/register", authHandler.HandleCreateUser)
	mux.HandleFunc("/api/auth/me", authHandler.HandleGetCurrentUser)
	mux.HandleFunc("/api/auth/password", authHandler.HandleUpdatePassword)
	mux.HandleFunc("/api/auth/permissions", authHandler.HandleGetUserPermissions)
	
	// Admin-only endpoints
	mux.Handle("/api/auth/users", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleListUsers)))
	mux.Handle("/api/auth/users/create", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleCreateUser)))
	mux.Handle("/api/auth/users/update", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleUpdateUser)))
	mux.Handle("/api/auth/users/delete", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleDeleteUser)))
	mux.Handle("/api/auth/roles", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleListRoles)))
	mux.Handle("/api/auth/roles/get", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleGetRole)))
	mux.Handle("/api/auth/roles/create", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleCreateRole)))
	mux.Handle("/api/auth/roles/update", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleUpdateRole)))
	mux.Handle("/api/auth/roles/delete", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleDeleteRole)))
	mux.Handle("/api/auth/audit", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleListAuditLogs)))
	mux.Handle("/api/auth/users/clusters", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleGetUserClusters)))
	mux.Handle("/api/auth/users/clusters/set", middleware.AdminMiddleware(http.HandlerFunc(authHandler.HandleSetUserClusters)))
	mux.HandleFunc("/api/auth/check-cluster", authHandler.HandleCheckClusterAccess)

	// Multi-cluster management
	mux.HandleFunc("/api/clusters", handleListClusters(multiClusterMgr, authSvc))
	mux.HandleFunc("/api/clusters/add", handleAddCluster(multiClusterMgr))
	mux.HandleFunc("/api/clusters/remove", handleRemoveCluster(multiClusterMgr))
	mux.HandleFunc("/api/clusters/switch", handleSwitchCluster(multiClusterMgr))
	mux.HandleFunc("/api/clusters/connect", handleClusterConnectMulti(multiClusterMgr))
	mux.HandleFunc("/api/clusters/disconnect", handleDisconnectClusterMulti(multiClusterMgr))

	// Legacy cluster endpoints (backwards compatibility)
	mux.HandleFunc("/api/cluster/connect", handleClusterConnectLegacy(multiClusterMgr))
	mux.HandleFunc("/api/cluster/disconnect", handleClusterDisconnectLegacy(multiClusterMgr))
	mux.HandleFunc("/api/cluster/status", handleClusterStatusMulti(multiClusterMgr))

	// Cluster
	mux.HandleFunc("/api/cluster/overview", h.HandleClusterOverview)

	// Core resources
	mux.HandleFunc("/api/nodes", h.HandleListNodes)
	mux.HandleFunc("/api/nodes/get", h.HandleGetNode)
	mux.HandleFunc("/api/nodes/drain", h.HandleDrainNode)
	mux.HandleFunc("/api/nodes/uncordon", h.HandleUncordonNode)
	mux.HandleFunc("/api/namespaces", h.HandleListNamespaces)
	mux.HandleFunc("/api/namespaces/create", h.HandleCreateNamespace)
	mux.HandleFunc("/api/namespaces/delete", h.HandleDeleteNamespace)
	mux.HandleFunc("/api/pods", h.HandleListPods)
	mux.HandleFunc("/api/pods/get", rh.HandleGetPod)
	mux.HandleFunc("/api/pods/logs", rh.HandleGetPodLogs)
	mux.HandleFunc("/api/pods/restart", rh.HandleRestartPod)
	mux.HandleFunc("/api/pods/delete", rh.HandleDeletePod)
	mux.HandleFunc("/api/deployments", h.HandleListDeployments)
	mux.HandleFunc("/api/deployments/get", rh.HandleGetDeployment)
	mux.HandleFunc("/api/deployments/scale", rh.HandleScaleDeployment)
	mux.HandleFunc("/api/deployments/rollback", rh.HandleRollbackDeployment)
	mux.HandleFunc("/api/deployments/restart", rh.HandleRestartDeployment)
	mux.HandleFunc("/api/deployments/update", rh.HandleUpdateDeployment)
	mux.HandleFunc("/api/deployments/delete", rh.HandleDeleteDeployment)
	mux.HandleFunc("/api/services", h.HandleListServices)
	mux.HandleFunc("/api/services/get", h.HandleGetService)
	mux.HandleFunc("/api/services/create", h.HandleCreateService)
	mux.HandleFunc("/api/services/update", h.HandleUpdateService)
	mux.HandleFunc("/api/services/delete", h.HandleDeleteService)
	mux.HandleFunc("/api/events", h.HandleListEvents)

	// Config resources
	mux.HandleFunc("/api/configmaps", rh.HandleListConfigMaps)
	mux.HandleFunc("/api/configmaps/get", rh.HandleGetConfigMap)
	mux.HandleFunc("/api/configmaps/create", rh.HandleCreateConfigMap)
	mux.HandleFunc("/api/configmaps/update", rh.HandleUpdateConfigMap)
	mux.HandleFunc("/api/configmaps/delete", rh.HandleDeleteConfigMap)
	mux.HandleFunc("/api/secrets", rh.HandleListSecrets)
	mux.HandleFunc("/api/secrets/create", rh.HandleCreateSecret)
	mux.HandleFunc("/api/secrets/update", rh.HandleUpdateSecret)
	mux.HandleFunc("/api/secrets/delete", rh.HandleDeleteSecret)

	// Resource quotas
	mux.HandleFunc("/api/resourcequotas", rh.HandleListResourceQuotas)
	mux.HandleFunc("/api/resourcequotas/get", rh.HandleGetResourceQuota)
	mux.HandleFunc("/api/resourcequotas/create", rh.HandleCreateResourceQuota)
	mux.HandleFunc("/api/resourcequotas/update", rh.HandleUpdateResourceQuota)
	mux.HandleFunc("/api/resourcequotas/delete", rh.HandleDeleteResourceQuota)

	// Workloads
	mux.HandleFunc("/api/deployments/create", wh.HandleCreateDeployment)
	mux.HandleFunc("/api/statefulsets", wh.HandleListStatefulSets)
	mux.HandleFunc("/api/statefulsets/get", wh.HandleGetStatefulSet)
	mux.HandleFunc("/api/statefulsets/create", wh.HandleCreateStatefulSet)
	mux.HandleFunc("/api/statefulsets/scale", wh.HandleScaleStatefulSet)
	mux.HandleFunc("/api/statefulsets/restart", wh.HandleRestartStatefulSet)
	mux.HandleFunc("/api/statefulsets/update", wh.HandleUpdateStatefulSet)
	mux.HandleFunc("/api/statefulsets/delete", wh.HandleDeleteStatefulSet)
	mux.HandleFunc("/api/daemonsets", wh.HandleListDaemonSets)
	mux.HandleFunc("/api/daemonsets/get", wh.HandleGetDaemonSet)
	mux.HandleFunc("/api/daemonsets/create", wh.HandleCreateDaemonSet)
	mux.HandleFunc("/api/daemonsets/restart", wh.HandleRestartDaemonSet)
	mux.HandleFunc("/api/daemonsets/update", wh.HandleUpdateDaemonSet)
	mux.HandleFunc("/api/daemonsets/delete", wh.HandleDeleteDaemonSet)
	mux.HandleFunc("/api/cronjobs", wh.HandleListCronJobs)
	mux.HandleFunc("/api/cronjobs/create", wh.HandleCreateCronJob)
	mux.HandleFunc("/api/cronjobs/get", wh.HandleGetCronJob)
	mux.HandleFunc("/api/cronjobs/update", wh.HandleUpdateCronJob)
	mux.HandleFunc("/api/cronjobs/suspend", wh.HandleSuspendCronJob)
	mux.HandleFunc("/api/cronjobs/delete", wh.HandleDeleteCronJob)
	mux.HandleFunc("/api/jobs", wh.HandleListJobs)
	mux.HandleFunc("/api/jobs/get", wh.HandleGetJob)
	mux.HandleFunc("/api/jobs/create", wh.HandleCreateJob)
	mux.HandleFunc("/api/jobs/update", wh.HandleUpdateJob)
	mux.HandleFunc("/api/jobs/delete", wh.HandleDeleteJob)

	// Networking
	mux.HandleFunc("/api/ingresses", nh.HandleListIngresses)
	mux.HandleFunc("/api/ingresses/get", nh.HandleGetIngress)
	mux.HandleFunc("/api/ingresses/create", nh.HandleCreateIngress)
	mux.HandleFunc("/api/ingresses/delete", nh.HandleDeleteIngress)
	mux.HandleFunc("/api/networkpolicies", nh.HandleListNetworkPolicies)
	mux.HandleFunc("/api/networkpolicies/get", nh.HandleGetNetworkPolicy)
	mux.HandleFunc("/api/networkpolicies/create", nh.HandleCreateNetworkPolicy)
	mux.HandleFunc("/api/networkpolicies/delete", nh.HandleDeleteNetworkPolicy)
	mux.HandleFunc("/api/endpoints", nh.HandleListEndpoints)

	// RBAC
	mux.HandleFunc("/api/roles", rbh.HandleListRoles)
	mux.HandleFunc("/api/roles/get", rbh.HandleGetRole)
	mux.HandleFunc("/api/roles/create", rbh.HandleCreateRole)
	mux.HandleFunc("/api/roles/update", rbh.HandleUpdateRole)
	mux.HandleFunc("/api/roles/delete", rbh.HandleDeleteRole)
	mux.HandleFunc("/api/clusterroles", rbh.HandleListClusterRoles)
	mux.HandleFunc("/api/clusterroles/get", rbh.HandleGetClusterRole)
	mux.HandleFunc("/api/clusterroles/create", rbh.HandleCreateClusterRole)
	mux.HandleFunc("/api/clusterroles/update", rbh.HandleUpdateClusterRole)
	mux.HandleFunc("/api/clusterroles/delete", rbh.HandleDeleteClusterRole)
	mux.HandleFunc("/api/rolebindings", rbh.HandleListRoleBindings)
	mux.HandleFunc("/api/rolebindings/create", rbh.HandleCreateRoleBinding)
	mux.HandleFunc("/api/rolebindings/delete", rbh.HandleDeleteRoleBinding)
	mux.HandleFunc("/api/clusterrolebindings", rbh.HandleListClusterRoleBindings)
	mux.HandleFunc("/api/clusterrolebindings/create", rbh.HandleCreateClusterRoleBinding)
	mux.HandleFunc("/api/clusterrolebindings/delete", rbh.HandleDeleteClusterRoleBinding)
	mux.HandleFunc("/api/serviceaccounts", rbh.HandleListServiceAccounts)
	mux.HandleFunc("/api/serviceaccounts/create", rbh.HandleCreateServiceAccount)
	mux.HandleFunc("/api/serviceaccounts/delete", rbh.HandleDeleteServiceAccount)

	// Storage
	mux.HandleFunc("/api/persistentvolumes", sh.HandleListPersistentVolumes)
	mux.HandleFunc("/api/persistentvolumes/get", sh.HandleGetPersistentVolume)
	mux.HandleFunc("/api/persistentvolumes/create", sh.HandleCreatePersistentVolume)
	mux.HandleFunc("/api/persistentvolumes/delete", sh.HandleDeletePersistentVolume)
	mux.HandleFunc("/api/persistentvolumeclaims", sh.HandleListPersistentVolumeClaims)
	mux.HandleFunc("/api/persistentvolumeclaims/create", sh.HandleCreatePersistentVolumeClaim)
	mux.HandleFunc("/api/persistentvolumeclaims/delete", sh.HandleDeletePersistentVolumeClaim)
	mux.HandleFunc("/api/storageclasses", sh.HandleListStorageClasses)
	mux.HandleFunc("/api/storageclasses/create", sh.HandleCreateStorageClass)
	mux.HandleFunc("/api/storageclasses/delete", sh.HandleDeleteStorageClass)

	// Policy
	mux.HandleFunc("/api/limitranges", ph.HandleListLimitRanges)
	mux.HandleFunc("/api/limitranges/create", ph.HandleCreateLimitRange)
	mux.HandleFunc("/api/limitranges/delete", ph.HandleDeleteLimitRange)
	mux.HandleFunc("/api/hpas", ph.HandleListHPAs)
	mux.HandleFunc("/api/hpas/create", ph.HandleCreateHPA)
	mux.HandleFunc("/api/hpas/delete", ph.HandleDeleteHPA)

	// Exec
	mux.HandleFunc("/api/pods/exec", eh.HandleExecCommand)
	mux.HandleFunc("/api/pods/exec/ws", eh.HandleWebSocketExec)
	mux.HandleFunc("/api/pods/logstream", eh.HandleStreamPodLogs)

	// Monitor
	mux.HandleFunc("/api/monitor/cluster", mh.HandleGetClusterMetrics)
	mux.HandleFunc("/api/monitor/node", mh.HandleGetNodeMetrics)
	mux.HandleFunc("/api/monitor/pod", mh.HandleGetPodMetrics)
	mux.HandleFunc("/api/monitor/cpu", mh.HandleGetCPUUsage)
	mux.HandleFunc("/api/monitor/memory", mh.HandleGetMemoryUsage)
	mux.HandleFunc("/api/monitor/disk", mh.HandleGetDiskUsage)
	mux.HandleFunc("/api/monitor/network", mh.HandleGetNetworkUsage)
	mux.HandleFunc("/api/monitor/prometheus", mh.HandleQueryPrometheus)
	mux.HandleFunc("/api/monitor/grafana", mh.HandleGetGrafanaDashboards)
	mux.HandleFunc("/api/monitor/capabilities", mh.HandleDetectCapabilities)
	mux.HandleFunc("/api/monitor/prometheus/status", mh.HandleGetPrometheusStatus)
	mux.HandleFunc("/api/monitor/prometheus/targets", mh.HandleGetPrometheusTargets)
	mux.HandleFunc("/api/monitor/prometheus/metrics", mh.HandleGetMetricNames)
	mux.HandleFunc("/api/monitor/prometheus/labels", mh.HandleGetLabelValues)

	// Alert Rules
	mux.HandleFunc("/api/alerts/rules", ah.HandleListRules)
	mux.HandleFunc("/api/alerts/rules/get", ah.HandleGetRule)
	mux.HandleFunc("/api/alerts/rules/create", ah.HandleCreateRule)
	mux.HandleFunc("/api/alerts/rules/update", ah.HandleUpdateRule)
	mux.HandleFunc("/api/alerts/rules/delete", ah.HandleDeleteRule)
	mux.HandleFunc("/api/alerts/rules/enable", ah.HandleEnableRule)

	// Alert History
	mux.HandleFunc("/api/alerts/history", ah.HandleListAlerts)
	mux.HandleFunc("/api/alerts/history/get", ah.HandleGetAlert)
	mux.HandleFunc("/api/alerts/history/resolve", ah.HandleResolveAlert)
	mux.HandleFunc("/api/alerts/history/delete", ah.HandleDeleteAlert)
	mux.HandleFunc("/api/alerts/stats", ah.HandleGetAlertStats)

	// Notifications
	mux.HandleFunc("/api/alerts/notifications", ah.HandleListNotifications)
	mux.HandleFunc("/api/alerts/notifications/create", ah.HandleCreateNotification)
	mux.HandleFunc("/api/alerts/notifications/delete", ah.HandleDeleteNotification)

	// Log Management
	mux.HandleFunc("/api/logs", lh.HandleGetPodLogs)
	mux.HandleFunc("/api/logs/stream", lh.HandleStreamPodLogs)
	mux.HandleFunc("/api/logs/download", lh.HandleDownloadPodLogs)
	mux.HandleFunc("/api/logs/containers", lh.HandleListContainers)
	mux.HandleFunc("/api/logs/search", lh.HandleSearchLogs)
	mux.HandleFunc("/api/logs/level", lh.HandleGetLogsByLevel)
	mux.HandleFunc("/api/logs/previous", lh.HandleGetPreviousLogs)
	mux.HandleFunc("/api/logs/since", lh.HandleGetLogsFromTime)

	// Audit Logs
	mux.HandleFunc("/api/audit/query", auh.HandleQueryLogs)
	mux.HandleFunc("/api/audit/get", auh.HandleGetLog)
	mux.HandleFunc("/api/audit/stats", auh.HandleGetStats)
	mux.HandleFunc("/api/audit/export", auh.HandleExportLogs)
	mux.HandleFunc("/api/audit/cleanup", auh.HandleCleanupLogs)
	mux.HandleFunc("/api/audit/users", auh.HandleListUsers)
	mux.HandleFunc("/api/audit/actions", auh.HandleListActions)

	// Backup & Restore
	mux.HandleFunc("/api/backup/create", bh.HandleCreateBackup)
	mux.HandleFunc("/api/backup/list", bh.HandleListBackups)
	mux.HandleFunc("/api/backup/get", bh.HandleGetBackup)
	mux.HandleFunc("/api/backup/delete", bh.HandleDeleteBackup)
	mux.HandleFunc("/api/backup/content", bh.HandleGetBackupContent)
	mux.HandleFunc("/api/backup/restore", bh.HandleRestore)
	mux.HandleFunc("/api/backup/export", bh.HandleExportBackup)
	mux.HandleFunc("/api/backup/import", bh.HandleImportBackup)
	mux.HandleFunc("/api/backup/resources", bh.HandleListResources)

	// Resource Graph
	mux.HandleFunc("/api/graph", gh.HandleGetGraph)

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.InfoLogger.Printf("server starting on %s (mode: %s)", addr, cfg.Mode)
	log.Println("===========================================")
	log.Printf("  KubeOps Server started on %s", addr)
	log.Printf("  Frontend: http://localhost:3000")
	log.Printf("  API: http://localhost:%d/api/health", cfg.Port)
	log.Printf("  Default login: admin / admin123")
	log.Println("===========================================")

	// Permission checker
	permChecker := middleware.NewPermissionChecker(authSvc)

	// Audit middleware
	auditMiddleware := middleware.NewAuditMiddleware(auditSvc)

	handler := middleware.CORS(
		middleware.Logger(
			middleware.AuthMiddleware(authSvc)(
				permChecker.AutoPermission()(
					auditMiddleware.RecordAudit(mux),
				),
			),
		),
	)
	return http.ListenAndServe(addr, handler)
}

type connectRequest struct {
	Kubeconfig string `json:"kubeconfig"`
	Token      string `json:"token"`
	Server     string `json:"server"`
	Name       string `json:"name"`
}

type connectResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Name    string `json:"name,omitempty"`
}

type clusterListResponse struct {
	Success bool         `json:"success"`
	Clusters []ClusterInfo `json:"clusters"`
}

func handleListClusters(mgr *MultiClusterManager, authSvc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		// Get user from context
		userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
		role, _ := r.Context().Value(middleware.UserRoleKey).(string)

		log.Printf("[DEBUG] handleListClusters: userID=%d, role=%s", userID, role)

		clusters := mgr.ListClusters()

		// Non-admin users can only see their authorized clusters
		if role != "admin" && userID > 0 {
			ctx := r.Context()
			authorizedClusters, err := authSvc.GetUserClusters(ctx, userID)
			if err != nil {
				log.Printf("[DEBUG] handleListClusters: GetUserClusters error: %v", err)
				writeJSONResp(w, http.StatusInternalServerError, connectResponse{Success: false, Message: "failed to get user clusters"})
				return
			}
			
			log.Printf("[DEBUG] handleListClusters: authorizedClusters=%v", authorizedClusters)
			
			// Filter clusters
			var filteredClusters []ClusterInfo
			for _, c := range clusters {
				for _, ac := range authorizedClusters {
					if c.Name == ac {
						filteredClusters = append(filteredClusters, c)
						break
					}
				}
			}
			clusters = filteredClusters
		}

		writeJSONResp(w, http.StatusOK, clusterListResponse{
			Success:  true,
			Clusters: clusters,
		})
	}
}

func handleAddCluster(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		var req connectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "invalid request body"})
			return
		}

		if req.Name == "" {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "cluster name is required"})
			return
		}

		err := mgr.AddCluster(req.Name, req.Server, req.Token, req.Kubeconfig)
		if err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: err.Error()})
			return
		}

		writeJSONResp(w, http.StatusOK, connectResponse{
			Success: true,
			Message: "cluster added successfully",
			Name:    req.Name,
		})
	}
}

func handleRemoveCluster(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "invalid request body"})
			return
		}

		if req.Name == "" {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "cluster name is required"})
			return
		}

		err := mgr.RemoveCluster(req.Name)
		if err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: err.Error()})
			return
		}

		writeJSONResp(w, http.StatusOK, connectResponse{
			Success: true,
			Message: "cluster removed successfully",
		})
	}
}

func handleSwitchCluster(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "invalid request body"})
			return
		}

		if req.Name == "" {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "cluster name is required"})
			return
		}

		err := mgr.SwitchCluster(req.Name)
		if err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: err.Error()})
			return
		}

		writeJSONResp(w, http.StatusOK, connectResponse{
			Success: true,
			Message: "cluster switched successfully",
			Name:    req.Name,
		})
	}
}

func handleClusterConnectMulti(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		var req connectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "invalid request body"})
			return
		}

		if req.Name == "" {
			req.Name = "default"
		}

		err := mgr.Connect(req.Name, req.Server, req.Token, req.Kubeconfig)
		if err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: err.Error()})
			return
		}

		writeJSONResp(w, http.StatusOK, connectResponse{
			Success: true,
			Message: "connected successfully",
			Name:    req.Name,
		})
	}
}

func handleDisconnectClusterMulti(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		var req struct {
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "" {
			mgr.Disconnect(req.Name)
		} else {
			mgr.DisconnectAll()
		}

		writeJSONResp(w, http.StatusOK, connectResponse{Success: true, Message: "disconnected"})
	}
}

func handleClusterConnectLegacy(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		var req connectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: "invalid request body"})
			return
		}

		if req.Name == "" {
			req.Name = "default"
		}

		err := mgr.Connect(req.Name, req.Server, req.Token, req.Kubeconfig)
		if err != nil {
			writeJSONResp(w, http.StatusBadRequest, connectResponse{Success: false, Message: err.Error()})
			return
		}

		writeJSONResp(w, http.StatusOK, connectResponse{
			Success: true,
			Message: "connected successfully",
			Name:    req.Name,
		})
	}
}

func handleClusterDisconnectLegacy(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResp(w, http.StatusMethodNotAllowed, connectResponse{Success: false, Message: "method not allowed"})
			return
		}

		mgr.DisconnectAll()
		writeJSONResp(w, http.StatusOK, connectResponse{Success: true, Message: "disconnected"})
	}
}

func handleClusterStatusMulti(mgr *MultiClusterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connected := mgr.IsConnected()
		name := mgr.GetActiveClusterName()

		writeJSONResp(w, http.StatusOK, map[string]interface{}{
			"connected": connected,
			"name":      name,
		})
	}
}

func writeJSONResp(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
