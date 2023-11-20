package utils

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type ApiInstances struct {
	// HttpClient
	HttpClient *datadog.APIClient

	// V1 APIs
	authenticationApiV1                   *datadogV1.AuthenticationApi
	awsIntegrationApiV1                   *datadogV1.AWSIntegrationApi
	awsLogsIntegrationApiV1               *datadogV1.AWSLogsIntegrationApi
	azureIntegrationApiV1                 *datadogV1.AzureIntegrationApi
	dashboardListsApiV1                   *datadogV1.DashboardListsApi
	dashboardsApiV1                       *datadogV1.DashboardsApi
	downtimesApiV1                        *datadogV1.DowntimesApi
	eventsApiV1                           *datadogV1.EventsApi
	gcpIntegrationApiV1                   *datadogV1.GCPIntegrationApi
	hostsApiV1                            *datadogV1.HostsApi
	ipRangesApiV1                         *datadogV1.IPRangesApi
	keyManagementApiV1                    *datadogV1.KeyManagementApi
	logsApiV1                             *datadogV1.LogsApi
	logsIndexesApiV1                      *datadogV1.LogsIndexesApi
	logsPipelinesApiV1                    *datadogV1.LogsPipelinesApi
	metricsApiV1                          *datadogV1.MetricsApi
	monitorsApiV1                         *datadogV1.MonitorsApi
	notebooksApiV1                        *datadogV1.NotebooksApi
	organizationsApiV1                    *datadogV1.OrganizationsApi
	pagerDutyIntegrationApiV1             *datadogV1.PagerDutyIntegrationApi
	securityMonitoringApiV1               *datadogV1.SecurityMonitoringApi
	serviceChecksApiV1                    *datadogV1.ServiceChecksApi
	serviceLevelObjectiveCorrectionsApiV1 *datadogV1.ServiceLevelObjectiveCorrectionsApi
	serviceLevelObjectivesApiV1           *datadogV1.ServiceLevelObjectivesApi
	slackIntegrationApiV1                 *datadogV1.SlackIntegrationApi
	snapshotsApiV1                        *datadogV1.SnapshotsApi
	syntheticsApiV1                       *datadogV1.SyntheticsApi
	tagsApiV1                             *datadogV1.TagsApi
	usageMeteringApiV1                    *datadogV1.UsageMeteringApi
	usersApiV1                            *datadogV1.UsersApi
	webhooksIntegrationApiV1              *datadogV1.WebhooksIntegrationApi

	//V2 APIs
	apmRetentionFiltersApiV2   *datadogV2.APMRetentionFiltersApi
	auditApiV2                 *datadogV2.AuditApi
	authNMappingsApiV2         *datadogV2.AuthNMappingsApi
	cloudflareIntegrationApiV2 *datadogV2.CloudflareIntegrationApi
	cloudWorkloadSecurityApiV2 *datadogV2.CloudWorkloadSecurityApi
	confluentCloudApiV2        *datadogV2.ConfluentCloudApi
	dashboardListsApiV2        *datadogV2.DashboardListsApi
	downtimesApiV2             *datadogV2.DowntimesApi
	eventsApiV2                *datadogV2.EventsApi
	fastlyIntegrationApiV2     *datadogV2.FastlyIntegrationApi
	gcpStsIntegrationApiV2     *datadogV2.GCPIntegrationApi
	incidentServicesApiV2      *datadogV2.IncidentServicesApi
	incidentTeamsApiV2         *datadogV2.IncidentTeamsApi
	incidentsApiV2             *datadogV2.IncidentsApi
	ipAllowlistApiV2           *datadogV2.IPAllowlistApi
	keyManagementApiV2         *datadogV2.KeyManagementApi
	logsApiV2                  *datadogV2.LogsApi
	logsArchivesApiV2          *datadogV2.LogsArchivesApi
	logsMetricsApiV2           *datadogV2.LogsMetricsApi
	metricsApiV2               *datadogV2.MetricsApi
	monitorsApiV2              *datadogV2.MonitorsApi
	opsgenieIntegrationApiV2   *datadogV2.OpsgenieIntegrationApi
	organizationsApiV2         *datadogV2.OrganizationsApi
	processesApiV2             *datadogV2.ProcessesApi
	restrictionPolicyApiV2     *datadogV2.RestrictionPoliciesApi
	rolesApiV2                 *datadogV2.RolesApi
	rumApiV2                   *datadogV2.RUMApi
	securityMonitoringApiV2    *datadogV2.SecurityMonitoringApi
	sensitiveDataScannerApiV2  *datadogV2.SensitiveDataScannerApi
	serviceAccountsApiV2       *datadogV2.ServiceAccountsApi
	spansMetricsApiV2          *datadogV2.SpansMetricsApi
	syntheticsApiV2            *datadogV2.SyntheticsApi
	teamsApiV2                 *datadogV2.TeamsApi
	usageMeteringApiV2         *datadogV2.UsageMeteringApi
	usersApiV2                 *datadogV2.UsersApi
}

// GetAuthenticationApiV1 get instance of AuthenticationApi
func (i *ApiInstances) GetAuthenticationApiV1() *datadogV1.AuthenticationApi {
	if i.authenticationApiV1 == nil {
		i.authenticationApiV1 = datadogV1.NewAuthenticationApi(i.HttpClient)
	}
	return i.authenticationApiV1
}

// GetAWSIntegrationApiV1 get instance of AWSIntegrationApi
func (i *ApiInstances) GetAWSIntegrationApiV1() *datadogV1.AWSIntegrationApi {
	if i.awsIntegrationApiV1 == nil {
		i.awsIntegrationApiV1 = datadogV1.NewAWSIntegrationApi(i.HttpClient)
	}
	return i.awsIntegrationApiV1
}

// GetAWSLogsIntegrationApiV1 get instance of AwsLogsIntegrationApi
func (i *ApiInstances) GetAWSLogsIntegrationApiV1() *datadogV1.AWSLogsIntegrationApi {
	if i.awsLogsIntegrationApiV1 == nil {
		i.awsLogsIntegrationApiV1 = datadogV1.NewAWSLogsIntegrationApi(i.HttpClient)
	}
	return i.awsLogsIntegrationApiV1
}

// GetAzureIntegrationApiV1 get instance of AzureIntegrationApi
func (i *ApiInstances) GetAzureIntegrationApiV1() *datadogV1.AzureIntegrationApi {
	if i.azureIntegrationApiV1 == nil {
		i.azureIntegrationApiV1 = datadogV1.NewAzureIntegrationApi(i.HttpClient)
	}
	return i.azureIntegrationApiV1
}

// GetDashboardListsApiV1 get instance of DashboardListsApi
func (i *ApiInstances) GetDashboardListsApiV1() *datadogV1.DashboardListsApi {
	if i.dashboardListsApiV1 == nil {
		i.dashboardListsApiV1 = datadogV1.NewDashboardListsApi(i.HttpClient)
	}
	return i.dashboardListsApiV1
}

// GetDashboardsApiV1 get instance of DashboardsApi
func (i *ApiInstances) GetDashboardsApiV1() *datadogV1.DashboardsApi {
	if i.dashboardsApiV1 == nil {
		i.dashboardsApiV1 = datadogV1.NewDashboardsApi(i.HttpClient)
	}
	return i.dashboardsApiV1
}

// GetDowntimesApiV1 get instance of DowntimesApi
func (i *ApiInstances) GetDowntimesApiV1() *datadogV1.DowntimesApi {
	if i.downtimesApiV1 == nil {
		i.downtimesApiV1 = datadogV1.NewDowntimesApi(i.HttpClient)
	}
	return i.downtimesApiV1
}

// GetEventsApiV1 get instance of EventsApi
func (i *ApiInstances) GetEventsApiV1() *datadogV1.EventsApi {
	if i.eventsApiV1 == nil {
		i.eventsApiV1 = datadogV1.NewEventsApi(i.HttpClient)
	}
	return i.eventsApiV1
}

// GetGCPIntegrationApiV1 get instance of GcpIntegrationApi
func (i *ApiInstances) GetGCPIntegrationApiV1() *datadogV1.GCPIntegrationApi {
	if i.gcpIntegrationApiV1 == nil {
		i.gcpIntegrationApiV1 = datadogV1.NewGCPIntegrationApi(i.HttpClient)
	}
	return i.gcpIntegrationApiV1
}

// GetHostsApiV1 get instance of HostsApi
func (i *ApiInstances) GetHostsApiV1() *datadogV1.HostsApi {
	if i.hostsApiV1 == nil {
		i.hostsApiV1 = datadogV1.NewHostsApi(i.HttpClient)
	}
	return i.hostsApiV1
}

// GetIPRangesApiV1 get instance of IPRangesApi
func (i *ApiInstances) GetIPRangesApiV1() *datadogV1.IPRangesApi {
	if i.ipRangesApiV1 == nil {
		i.ipRangesApiV1 = datadogV1.NewIPRangesApi(i.HttpClient)
	}
	return i.ipRangesApiV1
}

// GetKeyManagementApiV1 get instance of KeyManagementApi
func (i *ApiInstances) GetKeyManagementApiV1() *datadogV1.KeyManagementApi {
	if i.keyManagementApiV1 == nil {
		i.keyManagementApiV1 = datadogV1.NewKeyManagementApi(i.HttpClient)
	}
	return i.keyManagementApiV1
}

// GetLogsApiV1 get instance of LogsApi
func (i *ApiInstances) GetLogsApiV1() *datadogV1.LogsApi {
	if i.logsApiV1 == nil {
		i.logsApiV1 = datadogV1.NewLogsApi(i.HttpClient)
	}
	return i.logsApiV1
}

// GetLogsIndexesApiV1 get instance of LogsIndexesApi
func (i *ApiInstances) GetLogsIndexesApiV1() *datadogV1.LogsIndexesApi {
	if i.logsIndexesApiV1 == nil {
		i.logsIndexesApiV1 = datadogV1.NewLogsIndexesApi(i.HttpClient)
	}
	return i.logsIndexesApiV1
}

// GetLogsPipelinesApiV1 get instance of LogsPipelinesApi
func (i *ApiInstances) GetLogsPipelinesApiV1() *datadogV1.LogsPipelinesApi {
	if i.logsPipelinesApiV1 == nil {
		i.logsPipelinesApiV1 = datadogV1.NewLogsPipelinesApi(i.HttpClient)
	}
	return i.logsPipelinesApiV1
}

// GetMetricsApiV1 get instance of MetricsApi
func (i *ApiInstances) GetMetricsApiV1() *datadogV1.MetricsApi {
	if i.metricsApiV1 == nil {
		i.metricsApiV1 = datadogV1.NewMetricsApi(i.HttpClient)
	}
	return i.metricsApiV1
}

// GetMonitorsApiV1 get instance of MonitorsApi
func (i *ApiInstances) GetMonitorsApiV1() *datadogV1.MonitorsApi {
	if i.monitorsApiV1 == nil {
		i.monitorsApiV1 = datadogV1.NewMonitorsApi(i.HttpClient)
	}
	return i.monitorsApiV1
}

// GetNotebooksApiV1 get instance of NotebooksApi
func (i *ApiInstances) GetNotebooksApiV1() *datadogV1.NotebooksApi {
	if i.notebooksApiV1 == nil {
		i.notebooksApiV1 = datadogV1.NewNotebooksApi(i.HttpClient)
	}
	return i.notebooksApiV1
}

// GetOrganizationsApiV1 get instance of OrganizationsApi
func (i *ApiInstances) GetOrganizationsApiV1() *datadogV1.OrganizationsApi {
	if i.organizationsApiV1 == nil {
		i.organizationsApiV1 = datadogV1.NewOrganizationsApi(i.HttpClient)
	}
	return i.organizationsApiV1
}

// GetPagerDutyIntegrationApiV1 get instance of PagerDutyIntegrationApi
func (i *ApiInstances) GetPagerDutyIntegrationApiV1() *datadogV1.PagerDutyIntegrationApi {
	if i.pagerDutyIntegrationApiV1 == nil {
		i.pagerDutyIntegrationApiV1 = datadogV1.NewPagerDutyIntegrationApi(i.HttpClient)
	}
	return i.pagerDutyIntegrationApiV1
}

// GetSecurityMonitoringApiV1 get instance of SecurityMonitoringApi
func (i *ApiInstances) GetSecurityMonitoringApiV1() *datadogV1.SecurityMonitoringApi {
	if i.securityMonitoringApiV1 == nil {
		i.securityMonitoringApiV1 = datadogV1.NewSecurityMonitoringApi(i.HttpClient)
	}
	return i.securityMonitoringApiV1
}

// GetServiceChecksApiV1 get instance of ServiceChecksApi
func (i *ApiInstances) GetServiceChecksApiV1() *datadogV1.ServiceChecksApi {
	if i.serviceChecksApiV1 == nil {
		i.serviceChecksApiV1 = datadogV1.NewServiceChecksApi(i.HttpClient)
	}
	return i.serviceChecksApiV1
}

// GetServiceLevelObjectiveCorrectionsApiV1 get instance of ServiceLevelObjectiveCorrectionsApi
func (i *ApiInstances) GetServiceLevelObjectiveCorrectionsApiV1() *datadogV1.ServiceLevelObjectiveCorrectionsApi {
	if i.serviceLevelObjectiveCorrectionsApiV1 == nil {
		i.serviceLevelObjectiveCorrectionsApiV1 = datadogV1.NewServiceLevelObjectiveCorrectionsApi(i.HttpClient)
	}
	return i.serviceLevelObjectiveCorrectionsApiV1
}

// GetServiceLevelObjectivesApiV1 get instance of ServiceLevelObjectivesApi
func (i *ApiInstances) GetServiceLevelObjectivesApiV1() *datadogV1.ServiceLevelObjectivesApi {
	if i.serviceLevelObjectivesApiV1 == nil {
		i.serviceLevelObjectivesApiV1 = datadogV1.NewServiceLevelObjectivesApi(i.HttpClient)
	}
	return i.serviceLevelObjectivesApiV1
}

// GetSlackIntegrationApiV1 get instance of SlackIntegrationApi
func (i *ApiInstances) GetSlackIntegrationApiV1() *datadogV1.SlackIntegrationApi {
	if i.slackIntegrationApiV1 == nil {
		i.slackIntegrationApiV1 = datadogV1.NewSlackIntegrationApi(i.HttpClient)
	}
	return i.slackIntegrationApiV1
}

// GetSnapshotsApiV1 get instance of SnapshotsApi
func (i *ApiInstances) GetSnapshotsApiV1() *datadogV1.SnapshotsApi {
	if i.snapshotsApiV1 == nil {
		i.snapshotsApiV1 = datadogV1.NewSnapshotsApi(i.HttpClient)
	}
	return i.snapshotsApiV1
}

// GetSyntheticsApiV1 get instance of SyntheticsApi
func (i *ApiInstances) GetSyntheticsApiV1() *datadogV1.SyntheticsApi {
	if i.syntheticsApiV1 == nil {
		i.syntheticsApiV1 = datadogV1.NewSyntheticsApi(i.HttpClient)
	}
	return i.syntheticsApiV1
}

// GetTagsApiV1 get instance of TagsApi
func (i *ApiInstances) GetTagsApiV1() *datadogV1.TagsApi {
	if i.tagsApiV1 == nil {
		i.tagsApiV1 = datadogV1.NewTagsApi(i.HttpClient)
	}
	return i.tagsApiV1
}

// GetUsageMeteringApiV1 get instance of UsageMeteringApi
func (i *ApiInstances) GetUsageMeteringApiV1() *datadogV1.UsageMeteringApi {
	if i.usageMeteringApiV1 == nil {
		i.usageMeteringApiV1 = datadogV1.NewUsageMeteringApi(i.HttpClient)
	}
	return i.usageMeteringApiV1
}

// GetUsersApiV1 get instance of UsersApi
func (i *ApiInstances) GetUsersApiV1() *datadogV1.UsersApi {
	if i.usersApiV1 == nil {
		i.usersApiV1 = datadogV1.NewUsersApi(i.HttpClient)
	}
	return i.usersApiV1
}

// GetWebhooksIntegrationApiV1 get instance of WebhooksIntegrationApi
func (i *ApiInstances) GetWebhooksIntegrationApiV1() *datadogV1.WebhooksIntegrationApi {
	if i.webhooksIntegrationApiV1 == nil {
		i.webhooksIntegrationApiV1 = datadogV1.NewWebhooksIntegrationApi(i.HttpClient)
	}
	return i.webhooksIntegrationApiV1
}

// GetAuditApiV2 get instance of AuditApi
func (i *ApiInstances) GetAuditApiV2() *datadogV2.AuditApi {
	if i.auditApiV2 == nil {
		i.auditApiV2 = datadogV2.NewAuditApi(i.HttpClient)
	}
	return i.auditApiV2
}

// GetAuthNMappingsApiV2 get instance of AuthNMappingsApi
func (i *ApiInstances) GetAuthNMappingsApiV2() *datadogV2.AuthNMappingsApi {
	if i.authNMappingsApiV2 == nil {
		i.authNMappingsApiV2 = datadogV2.NewAuthNMappingsApi(i.HttpClient)
	}
	return i.authNMappingsApiV2
}

// GetCloudWorkloadSecurityApiV2 get instance of CloudWorkloadSecurityApi
func (i *ApiInstances) GetCloudWorkloadSecurityApiV2() *datadogV2.CloudWorkloadSecurityApi {
	if i.cloudWorkloadSecurityApiV2 == nil {
		i.cloudWorkloadSecurityApiV2 = datadogV2.NewCloudWorkloadSecurityApi(i.HttpClient)
	}
	return i.cloudWorkloadSecurityApiV2
}

// GetDowntimesApiV2 get instance of DowntimesApi
func (i *ApiInstances) GetDowntimesApiV2() *datadogV2.DowntimesApi {
	if i.downtimesApiV2 == nil {
		i.downtimesApiV2 = datadogV2.NewDowntimesApi(i.HttpClient)
	}
	return i.downtimesApiV2
}

// GetDashboardListsApiV2 get instance of DashboardListsApi
func (i *ApiInstances) GetDashboardListsApiV2() *datadogV2.DashboardListsApi {
	if i.dashboardListsApiV2 == nil {
		i.dashboardListsApiV2 = datadogV2.NewDashboardListsApi(i.HttpClient)
	}
	return i.dashboardListsApiV2
}

// GetEventsApiV2 get instance of EventsApi
func (i *ApiInstances) GetEventsApiV2() *datadogV2.EventsApi {
	if i.eventsApiV2 == nil {
		i.eventsApiV2 = datadogV2.NewEventsApi(i.HttpClient)
	}
	return i.eventsApiV2
}

// GetGCPStsIntegrationApiV2 get instance of GetGCPStsIntegration
func (i *ApiInstances) GetGCPIntegrationApiV2() *datadogV2.GCPIntegrationApi {
	if i.gcpStsIntegrationApiV2 == nil {
		i.gcpStsIntegrationApiV2 = datadogV2.NewGCPIntegrationApi(i.HttpClient)
	}
	return i.gcpStsIntegrationApiV2
}

// GetIncidentServicesApiV2 get instance of IncidentServicesApi
func (i *ApiInstances) GetIncidentServicesApiV2() *datadogV2.IncidentServicesApi {
	if i.incidentServicesApiV2 == nil {
		i.incidentServicesApiV2 = datadogV2.NewIncidentServicesApi(i.HttpClient)
	}
	return i.incidentServicesApiV2
}

// GetIncidentTeamsApiV2 get instance of IncidentTeamsApi
func (i *ApiInstances) GetIncidentTeamsApiV2() *datadogV2.IncidentTeamsApi {
	if i.incidentTeamsApiV2 == nil {
		i.incidentTeamsApiV2 = datadogV2.NewIncidentTeamsApi(i.HttpClient)
	}
	return i.incidentTeamsApiV2
}

// GetIncidentsApiV2 get instance of IncidentsApi
func (i *ApiInstances) GetIncidentsApiV2() *datadogV2.IncidentsApi {
	if i.incidentsApiV2 == nil {
		i.incidentsApiV2 = datadogV2.NewIncidentsApi(i.HttpClient)
	}
	return i.incidentsApiV2
}

func (i *ApiInstances) GetIPAllowlistApiV2() *datadogV2.IPAllowlistApi {
	if i.ipAllowlistApiV2 == nil {
		i.ipAllowlistApiV2 = datadogV2.NewIPAllowlistApi(i.HttpClient)
	}
	return i.ipAllowlistApiV2
}

// GetKeyManagementApiV2 get instance of KeyManagementApi
func (i *ApiInstances) GetKeyManagementApiV2() *datadogV2.KeyManagementApi {
	if i.keyManagementApiV2 == nil {
		i.keyManagementApiV2 = datadogV2.NewKeyManagementApi(i.HttpClient)
	}
	return i.keyManagementApiV2
}

// GetLogsApiV2 get instance of LogsApi
func (i *ApiInstances) GetLogsApiV2() *datadogV2.LogsApi {
	if i.logsApiV2 == nil {
		i.logsApiV2 = datadogV2.NewLogsApi(i.HttpClient)
	}
	return i.logsApiV2
}

// GetLogsArchivesApiV2 get instance of LogsArchivesApi
func (i *ApiInstances) GetLogsArchivesApiV2() *datadogV2.LogsArchivesApi {
	if i.logsArchivesApiV2 == nil {
		i.logsArchivesApiV2 = datadogV2.NewLogsArchivesApi(i.HttpClient)
	}
	return i.logsArchivesApiV2
}

// GetLogsMetricsApiV2 get instance of LogsMetricsApi
func (i *ApiInstances) GetLogsMetricsApiV2() *datadogV2.LogsMetricsApi {
	if i.logsMetricsApiV2 == nil {
		i.logsMetricsApiV2 = datadogV2.NewLogsMetricsApi(i.HttpClient)
	}
	return i.logsMetricsApiV2
}

// GetMetricsApiV2 get instance of MetricsApi
func (i *ApiInstances) GetMetricsApiV2() *datadogV2.MetricsApi {
	if i.metricsApiV2 == nil {
		i.metricsApiV2 = datadogV2.NewMetricsApi(i.HttpClient)
	}
	return i.metricsApiV2
}

// GetMonitorsApiV2 get instance of MonitorsApi
func (i *ApiInstances) GetMonitorsApiV2() *datadogV2.MonitorsApi {
	if i.monitorsApiV2 == nil {
		i.monitorsApiV2 = datadogV2.NewMonitorsApi(i.HttpClient)
	}
	return i.monitorsApiV2
}

// GetOpsgenieIntegrationApiV2 get instance of OpsgenieIntegrationApi
func (i *ApiInstances) GetOpsgenieIntegrationApiV2() *datadogV2.OpsgenieIntegrationApi {
	if i.opsgenieIntegrationApiV2 == nil {
		i.opsgenieIntegrationApiV2 = datadogV2.NewOpsgenieIntegrationApi(i.HttpClient)
	}
	return i.opsgenieIntegrationApiV2
}

// GetOrganizationsApiV2 get instance of OrganizationsApi
func (i *ApiInstances) GetOrganizationsApiV2() *datadogV2.OrganizationsApi {
	if i.organizationsApiV2 == nil {
		i.organizationsApiV2 = datadogV2.NewOrganizationsApi(i.HttpClient)
	}
	return i.organizationsApiV2
}

// GetProcessesApiV2 get instance of ProcessesApi
func (i *ApiInstances) GetProcessesApiV2() *datadogV2.ProcessesApi {
	if i.processesApiV2 == nil {
		i.processesApiV2 = datadogV2.NewProcessesApi(i.HttpClient)
	}
	return i.processesApiV2
}

// GetRolesApiV2 get instance of RolesApi
func (i *ApiInstances) GetRolesApiV2() *datadogV2.RolesApi {
	if i.rolesApiV2 == nil {
		i.rolesApiV2 = datadogV2.NewRolesApi(i.HttpClient)
	}
	return i.rolesApiV2
}

// GetRumApiV2 get instance of RumApi
func (i *ApiInstances) GetRumApiV2() *datadogV2.RUMApi {
	if i.rumApiV2 == nil {
		i.rumApiV2 = datadogV2.NewRUMApi(i.HttpClient)
	}
	return i.rumApiV2
}

// GetSecurityMonitoringApiV2 get instance of SecurityMonitoringApi
func (i *ApiInstances) GetSecurityMonitoringApiV2() *datadogV2.SecurityMonitoringApi {
	if i.securityMonitoringApiV2 == nil {
		i.securityMonitoringApiV2 = datadogV2.NewSecurityMonitoringApi(i.HttpClient)
	}
	return i.securityMonitoringApiV2

}

// GetSensitiveDataScannerApiV2 get instance of SensitiveDataScannerApi
func (i *ApiInstances) GetSensitiveDataScannerApiV2() *datadogV2.SensitiveDataScannerApi {
	if i.sensitiveDataScannerApiV2 == nil {
		i.sensitiveDataScannerApiV2 = datadogV2.NewSensitiveDataScannerApi(i.HttpClient)
	}
	return i.sensitiveDataScannerApiV2
}

// GetServiceAccountsApiV2 get instance of ServiceAccountsApi
func (i *ApiInstances) GetServiceAccountsApiV2() *datadogV2.ServiceAccountsApi {
	if i.serviceAccountsApiV2 == nil {
		i.serviceAccountsApiV2 = datadogV2.NewServiceAccountsApi(i.HttpClient)
	}
	return i.serviceAccountsApiV2
}

// GetUsageMeteringApiV2 get instance of UsageMeteringApi
func (i *ApiInstances) GetUsageMeteringApiV2() *datadogV2.UsageMeteringApi {
	if i.usageMeteringApiV2 == nil {
		i.usageMeteringApiV2 = datadogV2.NewUsageMeteringApi(i.HttpClient)
	}
	return i.usageMeteringApiV2
}

// GetUsersApiV2 get instance of UsersApi
func (i *ApiInstances) GetUsersApiV2() *datadogV2.UsersApi {
	if i.usersApiV2 == nil {
		i.usersApiV2 = datadogV2.NewUsersApi(i.HttpClient)
	}
	return i.usersApiV2
}

// GetCloudflareIntegrationApiV2 get instance of CloudflareIntegrationApi
func (i *ApiInstances) GetCloudflareIntegrationApiV2() *datadogV2.CloudflareIntegrationApi {
	if i.cloudflareIntegrationApiV2 == nil {
		i.cloudflareIntegrationApiV2 = datadogV2.NewCloudflareIntegrationApi(i.HttpClient)
	}
	return i.cloudflareIntegrationApiV2
}

// GetConfluentCloudApiV2 get instance of GetConfluentCloudApi
func (i *ApiInstances) GetConfluentCloudApiV2() *datadogV2.ConfluentCloudApi {
	if i.confluentCloudApiV2 == nil {
		i.confluentCloudApiV2 = datadogV2.NewConfluentCloudApi(i.HttpClient)
	}
	return i.confluentCloudApiV2
}

// GetFastlyIntegrationApiV2 get instance of FastlyIntegrationApi
func (i *ApiInstances) GetFastlyIntegrationApiV2() *datadogV2.FastlyIntegrationApi {
	if i.fastlyIntegrationApiV2 == nil {
		i.fastlyIntegrationApiV2 = datadogV2.NewFastlyIntegrationApi(i.HttpClient)
	}
	return i.fastlyIntegrationApiV2
}

// GetRestrictionPoliciesApiV2 get instance of RestrictionPoliciesApi
func (i *ApiInstances) GetRestrictionPoliciesApiV2() *datadogV2.RestrictionPoliciesApi {
	if i.restrictionPolicyApiV2 == nil {
		i.restrictionPolicyApiV2 = datadogV2.NewRestrictionPoliciesApi(i.HttpClient)
	}
	return i.restrictionPolicyApiV2
}

// GetTeamsApiV2 get instance of TeamsApi
func (i *ApiInstances) GetTeamsApiV2() *datadogV2.TeamsApi {
	if i.teamsApiV2 == nil {
		i.teamsApiV2 = datadogV2.NewTeamsApi(i.HttpClient)
	}
	return i.teamsApiV2
}

// GetSpansMetricsApiV2 get instance of SpansMetricsApi
func (i *ApiInstances) GetSpansMetricsApiV2() *datadogV2.SpansMetricsApi {
	if i.spansMetricsApiV2 == nil {
		i.spansMetricsApiV2 = datadogV2.NewSpansMetricsApi(i.HttpClient)
	}
	return i.spansMetricsApiV2
}

// GetSyntheticsApiV2 get instance of SyntheticsApi
func (i *ApiInstances) GetSyntheticsApiV2() *datadogV2.SyntheticsApi {
	if i.syntheticsApiV2 == nil {
		i.syntheticsApiV2 = datadogV2.NewSyntheticsApi(i.HttpClient)
	}
	return i.syntheticsApiV2
}

// GetSpansMetricsApiV2 get instance of SpansMetricsApi
func (i *ApiInstances) GetApmRetentionFiltersApiV2() *datadogV2.APMRetentionFiltersApi {
	if i.apmRetentionFiltersApiV2 == nil {
		i.apmRetentionFiltersApiV2 = datadogV2.NewAPMRetentionFiltersApi(i.HttpClient)
	}
	return i.apmRetentionFiltersApiV2
}
