package utils

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/common"
	datadogV1 "github.com/DataDog/datadog-api-client-go/v2/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/v2/datadog"
)

var (
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
	auditApiV2                 *datadogV2.AuditApi
	authNMappingsApiV2         *datadogV2.AuthNMappingsApi
	cloudWorkloadSecurityApiV2 *datadogV2.CloudWorkloadSecurityApi
	dashboardListsApiV2        *datadogV2.DashboardListsApi
	eventsApiV2                *datadogV2.EventsApi
	incidentServicesApiV2      *datadogV2.IncidentServicesApi
	incidentTeamsApiV2         *datadogV2.IncidentTeamsApi
	incidentsApiV2             *datadogV2.IncidentsApi
	keyManagementApiV2         *datadogV2.KeyManagementApi
	logsApiV2                  *datadogV2.LogsApi
	logsArchivesApiV2          *datadogV2.LogsArchivesApi
	logsMetricsApiV2           *datadogV2.LogsMetricsApi
	metricsApiV2               *datadogV2.MetricsApi
	opsgenieIntegrationApiV2   *datadogV2.OpsgenieIntegrationApi
	organizationsApiV2         *datadogV2.OrganizationsApi
	processesApiV2             *datadogV2.ProcessesApi
	rolesApiV2                 *datadogV2.RolesApi
	rumApiV2                   *datadogV2.RUMApi
	securityMonitoringApiV2    *datadogV2.SecurityMonitoringApi
	serviceAccountsApiV2       *datadogV2.ServiceAccountsApi
	usageMeteringApiV2         *datadogV2.UsageMeteringApi
	usersApiV2                 *datadogV2.UsersApi
)

// GetAuthenticationApiV1 get instance of AuthenticationApi
func GetAuthenticationApiV1(client *common.APIClient) *datadogV1.AuthenticationApi {
	if authenticationApiV1 == nil {
		authenticationApiV1 = datadogV1.NewAuthenticationApi(client)
	}
	return authenticationApiV1
}

// GetAwsIntegrationApiV1 get instance of AwsIntegrationApi
func GetAwsIntegrationApiV1(client *common.APIClient) *datadogV1.AWSIntegrationApi {
	if awsIntegrationApiV1 == nil {
		awsIntegrationApiV1 = datadogV1.NewAWSIntegrationApi(client)
	}
	return awsIntegrationApiV1
}

// GetAwsLogsIntegrationApiV1 get instance of AwsLogsIntegrationApi
func GetAwsLogsIntegrationApiV1(client *common.APIClient) *datadogV1.AWSLogsIntegrationApi {
	if awsLogsIntegrationApiV1 == nil {
		awsLogsIntegrationApiV1 = datadogV1.NewAWSLogsIntegrationApi(client)
	}
	return awsLogsIntegrationApiV1
}

// GetAzureIntegrationApiV1 get instance of AzureIntegrationApi
func GetAzureIntegrationApiV1(client *common.APIClient) *datadogV1.AzureIntegrationApi {
	if azureIntegrationApiV1 == nil {
		azureIntegrationApiV1 = datadogV1.NewAzureIntegrationApi(client)
	}
	return azureIntegrationApiV1
}

// GetDashboardListsApiV1 get instance of DashboardListsApi
func GetDashboardListsApiV1(client *common.APIClient) *datadogV1.DashboardListsApi {
	if dashboardListsApiV1 == nil {
		dashboardListsApiV1 = datadogV1.NewDashboardListsApi(client)
	}
	return dashboardListsApiV1
}

// GetDashboardsApiV1 get instance of DashboardsApi
func GetDashboardsApiV1(client *common.APIClient) *datadogV1.DashboardsApi {
	if dashboardsApiV1 == nil {
		dashboardsApiV1 = datadogV1.NewDashboardsApi(client)
	}
	return dashboardsApiV1
}

// GetDowntimesApiV1 get instance of DowntimesApi
func GetDowntimesApiV1(client *common.APIClient) *datadogV1.DowntimesApi {
	if downtimesApiV1 == nil {
		downtimesApiV1 = datadogV1.NewDowntimesApi(client)
	}
	return downtimesApiV1
}

// GetEventsApiV1 get instance of EventsApi
func GetEventsApiV1(client *common.APIClient) *datadogV1.EventsApi {
	if eventsApiV1 == nil {
		eventsApiV1 = datadogV1.NewEventsApi(client)
	}
	return eventsApiV1
}

// GetGcpIntegrationApiV1 get instance of GcpIntegrationApi
func GetGcpIntegrationApiV1(client *common.APIClient) *datadogV1.GCPIntegrationApi {
	if gcpIntegrationApiV1 == nil {
		gcpIntegrationApiV1 = datadogV1.NewGCPIntegrationApi(client)
	}
	return gcpIntegrationApiV1
}

// GetHostsApiV1 get instance of HostsApi
func GetHostsApiV1(client *common.APIClient) *datadogV1.HostsApi {
	if hostsApiV1 == nil {
		hostsApiV1 = datadogV1.NewHostsApi(client)
	}
	return hostsApiV1
}

// GetIpRangesApiV1 get instance of IpRangesApi
func GetIpRangesApiV1(client *common.APIClient) *datadogV1.IPRangesApi {
	if ipRangesApiV1 == nil {
		ipRangesApiV1 = datadogV1.NewIPRangesApi(client)
	}
	return ipRangesApiV1
}

// GetKeyManagementApiV1 get instance of KeyManagementApi
func GetKeyManagementApiV1(client *common.APIClient) *datadogV1.KeyManagementApi {
	if keyManagementApiV1 == nil {
		keyManagementApiV1 = datadogV1.NewKeyManagementApi(client)
	}
	return keyManagementApiV1
}

// GetLogsApiV1 get instance of LogsApi
func GetLogsApiV1(client *common.APIClient) *datadogV1.LogsApi {
	if logsApiV1 == nil {
		logsApiV1 = datadogV1.NewLogsApi(client)
	}
	return logsApiV1
}

// GetLogsIndexesApiV1 get instance of LogsIndexesApi
func GetLogsIndexesApiV1(client *common.APIClient) *datadogV1.LogsIndexesApi {
	if logsIndexesApiV1 == nil {
		logsIndexesApiV1 = datadogV1.NewLogsIndexesApi(client)
	}
	return logsIndexesApiV1
}

// GetLogsPipelinesApiV1 get instance of LogsPipelinesApi
func GetLogsPipelinesApiV1(client *common.APIClient) *datadogV1.LogsPipelinesApi {
	if logsPipelinesApiV1 == nil {
		logsPipelinesApiV1 = datadogV1.NewLogsPipelinesApi(client)
	}
	return logsPipelinesApiV1
}

// GetMetricsApiV1 get instance of MetricsApi
func GetMetricsApiV1(client *common.APIClient) *datadogV1.MetricsApi {
	if metricsApiV1 == nil {
		metricsApiV1 = datadogV1.NewMetricsApi(client)
	}
	return metricsApiV1
}

// GetMonitorsApiV1 get instance of MonitorsApi
func GetMonitorsApiV1(client *common.APIClient) *datadogV1.MonitorsApi {
	if monitorsApiV1 == nil {
		monitorsApiV1 = datadogV1.NewMonitorsApi(client)
	}
	return monitorsApiV1
}

// GetNotebooksApiV1 get instance of NotebooksApi
func GetNotebooksApiV1(client *common.APIClient) *datadogV1.NotebooksApi {
	if notebooksApiV1 == nil {
		notebooksApiV1 = datadogV1.NewNotebooksApi(client)
	}
	return notebooksApiV1
}

// GetOrganizationsApiV1 get instance of OrganizationsApi
func GetOrganizationsApiV1(client *common.APIClient) *datadogV1.OrganizationsApi {
	if organizationsApiV1 == nil {
		organizationsApiV1 = datadogV1.NewOrganizationsApi(client)
	}
	return organizationsApiV1
}

// GetPagerDutyIntegrationApiV1 get instance of PagerDutyIntegrationApi
func GetPagerDutyIntegrationApiV1(client *common.APIClient) *datadogV1.PagerDutyIntegrationApi {
	if pagerDutyIntegrationApiV1 == nil {
		pagerDutyIntegrationApiV1 = datadogV1.NewPagerDutyIntegrationApi(client)
	}
	return pagerDutyIntegrationApiV1
}

// GetSecurityMonitoringApiV1 get instance of SecurityMonitoringApi
func GetSecurityMonitoringApiV1(client *common.APIClient) *datadogV1.SecurityMonitoringApi {
	if securityMonitoringApiV1 == nil {
		securityMonitoringApiV1 = datadogV1.NewSecurityMonitoringApi(client)
	}
	return securityMonitoringApiV1
}

// GetServiceChecksApiV1 get instance of ServiceChecksApi
func GetServiceChecksApiV1(client *common.APIClient) *datadogV1.ServiceChecksApi {
	if serviceChecksApiV1 == nil {
		serviceChecksApiV1 = datadogV1.NewServiceChecksApi(client)
	}
	return serviceChecksApiV1
}

// GetServiceLevelObjectiveCorrectionsApiV1 get instance of ServiceLevelObjectiveCorrectionsApi
func GetServiceLevelObjectiveCorrectionsApiV1(client *common.APIClient) *datadogV1.ServiceLevelObjectiveCorrectionsApi {
	if serviceLevelObjectiveCorrectionsApiV1 == nil {
		serviceLevelObjectiveCorrectionsApiV1 = datadogV1.NewServiceLevelObjectiveCorrectionsApi(client)
	}
	return serviceLevelObjectiveCorrectionsApiV1
}

// GetServiceLevelObjectivesApiV1 get instance of ServiceLevelObjectivesApi
func GetServiceLevelObjectivesApiV1(client *common.APIClient) *datadogV1.ServiceLevelObjectivesApi {
	if serviceLevelObjectivesApiV1 == nil {
		serviceLevelObjectivesApiV1 = datadogV1.NewServiceLevelObjectivesApi(client)
	}
	return serviceLevelObjectivesApiV1
}

// GetSlackIntegrationApiV1 get instance of SlackIntegrationApi
func GetSlackIntegrationApiV1(client *common.APIClient) *datadogV1.SlackIntegrationApi {
	if slackIntegrationApiV1 == nil {
		slackIntegrationApiV1 = datadogV1.NewSlackIntegrationApi(client)
	}
	return slackIntegrationApiV1
}

// GetSnapshotsApiV1 get instance of SnapshotsApi
func GetSnapshotsApiV1(client *common.APIClient) *datadogV1.SnapshotsApi {
	if snapshotsApiV1 == nil {
		snapshotsApiV1 = datadogV1.NewSnapshotsApi(client)
	}
	return snapshotsApiV1
}

// GetSyntheticsApiV1 get instance of SyntheticsApi
func GetSyntheticsApiV1(client *common.APIClient) *datadogV1.SyntheticsApi {
	if syntheticsApiV1 == nil {
		syntheticsApiV1 = datadogV1.NewSyntheticsApi(client)
	}
	return syntheticsApiV1
}

// GetTagsApiV1 get instance of TagsApi
func GetTagsApiV1(client *common.APIClient) *datadogV1.TagsApi {
	if tagsApiV1 == nil {
		tagsApiV1 = datadogV1.NewTagsApi(client)
	}
	return tagsApiV1
}

// GetUsageMeteringApiV1 get instance of UsageMeteringApi
func GetUsageMeteringApiV1(client *common.APIClient) *datadogV1.UsageMeteringApi {
	if usageMeteringApiV1 == nil {
		usageMeteringApiV1 = datadogV1.NewUsageMeteringApi(client)
	}
	return usageMeteringApiV1
}

// GetUsersApiV1 get instance of UsersApi
func GetUsersApiV1(client *common.APIClient) *datadogV1.UsersApi {
	if usersApiV1 == nil {
		usersApiV1 = datadogV1.NewUsersApi(client)
	}
	return usersApiV1
}

// GetWebhooksIntegrationApiV1 get instance of WebhooksIntegrationApi
func GetWebhooksIntegrationApiV1(client *common.APIClient) *datadogV1.WebhooksIntegrationApi {
	if webhooksIntegrationApiV1 == nil {
		webhooksIntegrationApiV1 = datadogV1.NewWebhooksIntegrationApi(client)
	}
	return webhooksIntegrationApiV1
}

// GetAuditApiV2 get instance of AuditApi
func GetAuditApiV2(client *common.APIClient) *datadogV2.AuditApi {
	if auditApiV2 == nil {
		auditApiV2 = datadogV2.NewAuditApi(client)
	}
	return auditApiV2
}

// GetAuthNMappingsApiV2 get instance of AuthNMappingsApi
func GetAuthNMappingsApiV2(client *common.APIClient) *datadogV2.AuthNMappingsApi {
	if authNMappingsApiV2 == nil {
		authNMappingsApiV2 = datadogV2.NewAuthNMappingsApi(client)
	}
	return authNMappingsApiV2
}

// GetCloudWorkloadSecurityApiV2 get instance of CloudWorkloadSecurityApi
func GetCloudWorkloadSecurityApiV2(client *common.APIClient) *datadogV2.CloudWorkloadSecurityApi {
	if cloudWorkloadSecurityApiV2 == nil {
		cloudWorkloadSecurityApiV2 = datadogV2.NewCloudWorkloadSecurityApi(client)
	}
	return cloudWorkloadSecurityApiV2
}

// GetDashboardListsApiV2 get instance of DashboardListsApi
func GetDashboardListsApiV2(client *common.APIClient) *datadogV2.DashboardListsApi {
	if dashboardListsApiV2 == nil {
		dashboardListsApiV2 = datadogV2.NewDashboardListsApi(client)
	}
	return dashboardListsApiV2
}

// GetEventsApiV2 get instance of EventsApi
func GetEventsApiV2(client *common.APIClient) *datadogV2.EventsApi {
	if eventsApiV2 == nil {
		eventsApiV2 = datadogV2.NewEventsApi(client)
	}
	return eventsApiV2
}

// GetIncidentServicesApiV2 get instance of IncidentServicesApi
func GetIncidentServicesApiV2(client *common.APIClient) *datadogV2.IncidentServicesApi {
	if incidentServicesApiV2 == nil {
		incidentServicesApiV2 = datadogV2.NewIncidentServicesApi(client)
	}
	return incidentServicesApiV2
}

// GetIncidentTeamsApiV2 get instance of IncidentTeamsApi
func GetIncidentTeamsApiV2(client *common.APIClient) *datadogV2.IncidentTeamsApi {
	if incidentTeamsApiV2 == nil {
		incidentTeamsApiV2 = datadogV2.NewIncidentTeamsApi(client)
	}
	return incidentTeamsApiV2
}

// GetIncidentsApiV2 get instance of IncidentsApi
func GetIncidentsApiV2(client *common.APIClient) *datadogV2.IncidentsApi {
	if incidentsApiV2 == nil {
		incidentsApiV2 = datadogV2.NewIncidentsApi(client)
	}
	return incidentsApiV2
}

// GetKeyManagementApiV2 get instance of KeyManagementApi
func GetKeyManagementApiV2(client *common.APIClient) *datadogV2.KeyManagementApi {
	if keyManagementApiV2 == nil {
		keyManagementApiV2 = datadogV2.NewKeyManagementApi(client)
	}
	return keyManagementApiV2
}

// GetLogsApiV2 get instance of LogsApi
func GetLogsApiV2(client *common.APIClient) *datadogV2.LogsApi {
	if logsApiV2 == nil {
		logsApiV2 = datadogV2.NewLogsApi(client)
	}
	return logsApiV2
}

// GetLogsArchivesApiV2 get instance of LogsArchivesApi
func GetLogsArchivesApiV2(client *common.APIClient) *datadogV2.LogsArchivesApi {
	if logsArchivesApiV2 == nil {
		logsArchivesApiV2 = datadogV2.NewLogsArchivesApi(client)
	}
	return logsArchivesApiV2
}

// GetLogsMetricsApiV2 get instance of LogsMetricsApi
func GetLogsMetricsApiV2(client *common.APIClient) *datadogV2.LogsMetricsApi {
	if logsMetricsApiV2 == nil {
		logsMetricsApiV2 = datadogV2.NewLogsMetricsApi(client)
	}
	return logsMetricsApiV2
}

// GetMetricsApiV2 get instance of MetricsApi
func GetMetricsApiV2(client *common.APIClient) *datadogV2.MetricsApi {
	if metricsApiV2 == nil {
		metricsApiV2 = datadogV2.NewMetricsApi(client)
	}
	return metricsApiV2
}

// GetOpsgenieIntegrationApiV2 get instance of OpsgenieIntegrationApi
func GetOpsgenieIntegrationApiV2(client *common.APIClient) *datadogV2.OpsgenieIntegrationApi {
	if opsgenieIntegrationApiV2 == nil {
		opsgenieIntegrationApiV2 = datadogV2.NewOpsgenieIntegrationApi(client)
	}
	return opsgenieIntegrationApiV2
}

// GetOrganizationsApiV2 get instance of OrganizationsApi
func GetOrganizationsApiV2(client *common.APIClient) *datadogV2.OrganizationsApi {
	if organizationsApiV2 == nil {
		organizationsApiV2 = datadogV2.NewOrganizationsApi(client)
	}
	return organizationsApiV2
}

// GetProcessesApiV2 get instance of ProcessesApi
func GetProcessesApiV2(client *common.APIClient) *datadogV2.ProcessesApi {
	if processesApiV2 == nil {
		processesApiV2 = datadogV2.NewProcessesApi(client)
	}
	return processesApiV2
}

// GetRolesApiV2 get instance of RolesApi
func GetRolesApiV2(client *common.APIClient) *datadogV2.RolesApi {
	if rolesApiV2 == nil {
		rolesApiV2 = datadogV2.NewRolesApi(client)
	}
	return rolesApiV2
}

// GetRumApiV2 get instance of RumApi
func GetRumApiV2(client *common.APIClient) *datadogV2.RUMApi {
	if rumApiV2 == nil {
		rumApiV2 = datadogV2.NewRUMApi(client)
	}
	return rumApiV2
}

// GetSecurityMonitoringApiV2 get instance of SecurityMonitoringApi
func GetSecurityMonitoringApiV2(client *common.APIClient) *datadogV2.SecurityMonitoringApi {
	if securityMonitoringApiV2 == nil {
		securityMonitoringApiV2 = datadogV2.NewSecurityMonitoringApi(client)
	}
	return securityMonitoringApiV2
}

// GetServiceAccountsApiV2 get instance of ServiceAccountsApi
func GetServiceAccountsApiV2(client *common.APIClient) *datadogV2.ServiceAccountsApi {
	if serviceAccountsApiV2 == nil {
		serviceAccountsApiV2 = datadogV2.NewServiceAccountsApi(client)
	}
	return serviceAccountsApiV2
}

// GetUsageMeteringApiV2 get instance of UsageMeteringApi
func GetUsageMeteringApiV2(client *common.APIClient) *datadogV2.UsageMeteringApi {
	if usageMeteringApiV2 == nil {
		usageMeteringApiV2 = datadogV2.NewUsageMeteringApi(client)
	}
	return usageMeteringApiV2
}

// GetUsersApiV2 get instance of UsersApi
func GetUsersApiV2(client *common.APIClient) *datadogV2.UsersApi {
	if usersApiV2 == nil {
		usersApiV2 = datadogV2.NewUsersApi(client)
	}
	return usersApiV2
}
