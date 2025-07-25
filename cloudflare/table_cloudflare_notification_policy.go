package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/alerting"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableCloudflareNotificationPolicy(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "cloudflare_notification_policy",
		Description: "Cloudflare Notifications help you stay up to date with your Cloudflare account.",
		List: &plugin.ListConfig{
			KeyColumns: []*plugin.KeyColumn{
				{Name: "account_id", Require: plugin.Required},
			},
			Hydrate: listNotificationPolicies,
		},
		Get: &plugin.GetConfig{
			KeyColumns: []*plugin.KeyColumn{
				{Name: "id", Require: plugin.Required},
				{Name: "account_id", Require: plugin.Required},
			},
			ShouldIgnoreError: isNotFoundError([]string{"Invalid notification policy identifier"}),
			Hydrate:           getNotificationPolicy,
		},
		Columns: commonColumns([]*plugin.Column{
			// Top columns
			{Name: "id", Type: proto.ColumnType_STRING, Transform: transform.FromField("ID"), Description: "Notification policy identifier."},
			{Name: "alert_interval", Type: proto.ColumnType_STRING, Description: "Specification of how often to re-alert from the same incident, not support on all alert types."},
			{Name: "alert_type", Type: proto.ColumnType_STRING, Description: "Refers to which event will trigger a Notification dispatch."},
			{Name: "created", Type: proto.ColumnType_TIMESTAMP, Description: "When the notification policy was created."},
			{Name: "description", Type: proto.ColumnType_STRING, Description: "Description for the Notification policy."},
			{Name: "enabled", Type: proto.ColumnType_BOOL, Description: "Whether or not the Notification policy is enabled."},
			{Name: "filters", Type: proto.ColumnType_JSON, Description: "Filters that allow you to be alerted only on a subset of events for that alert type based on some criteria."},
			{Name: "mechanisms", Type: proto.ColumnType_JSON, Description: "List of IDs that will be used when dispatching a notification."},
			{Name: "modified", Type: proto.ColumnType_TIMESTAMP, Description: "When the notification policy was last modified."},
			{Name: "name", Type: proto.ColumnType_STRING, Description: "Name of the policy."},
			
			// Query columns for filtering
			{Name: "account_id", Type: proto.ColumnType_STRING, Transform: transform.FromQual("account_id"), Description: "The account ID to filter rulesets."},
		}),
	}
}

//// LIST FUNCTION

// listNotificationPolicies retrieves all notification policies for the specified account_id.
//
// - Account-level notification policies (account_id)
func listNotificationPolicies(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	conn, err := connectV4(ctx, d)
	if err != nil {
		logger.Error("cloudflare_notification_policy.listNotificationPolicies", "connection_error", err)
		return nil, err
	}

	// Get the qualifiers
	quals := d.EqualsQuals
	accountID := quals["account_id"].GetStringValue()

	// Empty check
	if accountID == "" {
		return nil, nil
	}

	// Build API parameters
	input := alerting.PolicyListParams{
		AccountID: cloudflare.F(accountID),
	}

	// Execute paginated API call
	iter := conn.Alerting.Policies.ListAutoPaging(ctx, input)
	for iter.Next() {
		ruleset := iter.Current()
		d.StreamListItem(ctx, ruleset)

		// Context can be cancelled due to manual cancellation or the limit has been hit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	if err := iter.Err(); err != nil {
		logger.Error("cloudflare_notification_policy.listNotificationPolicies", "ListAutoPaging error", err)
		return nil, err
	}

	return nil, nil
}

//// GET FUNCTION

// getNotificationPolicy retrieves a specific notification policy by ID.
//
// Parameters:
// - id: The ruleset identifier (required)
// - account_id OR zone_id: The account or zone context (at least one required)
func getNotificationPolicy(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	conn, err := connectV4(ctx, d)
	if err != nil {
		logger.Error("cloudflare_notification_policy.getNotificationPolicy", "connection_error", err)
		return nil, err
	}

	quals := d.EqualsQuals
	notificationPolicyID := quals["id"].GetStringValue()
	accountID := quals["account_id"].GetStringValue()

	input := alerting.PolicyGetParams{
		AccountID: cloudflare.F(accountID),
	}

	// Execute API call to get the specific ruleset
	notificationPolicy, err := conn.Alerting.Policies.Get(ctx, notificationPolicyID, input)
	if err != nil {
		logger.Error("cloudflare_notification_policy.getNotificationPolicy", "error", err)
		return nil, err
	}

	return notificationPolicy, nil
}
