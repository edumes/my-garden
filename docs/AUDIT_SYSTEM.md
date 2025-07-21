# Audit Notification System

The audit notification system provides comprehensive logging and monitoring of all garden actions and system events. This system helps track user activity, detect anomalies, and maintain security compliance.

## Features

### Automatic Logging
- **HTTP Request Logging**: All API requests are automatically logged with middleware
- **User Action Tracking**: Specific actions like garden creation, plant management, and purchases are logged
- **Error Tracking**: Failed operations and system errors are captured
- **Security Events**: Authentication and authorization events are monitored

### Audit Data Captured
- **User Information**: User ID, username, IP address, user agent
- **Action Details**: Action type, resource affected, success/failure status
- **Timestamps**: Precise timing of all events
- **Context Data**: Request details, response data, and error messages
- **Resource Tracking**: Which gardens, plants, or other resources were affected

## Database Schema

### AuditLog Table
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR NOT NULL,
    resource VARCHAR NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address VARCHAR,
    user_agent VARCHAR,
    status VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## API Endpoints

### Get Audit Logs
```
GET /api/v1/audit/logs
```
Query parameters:
- `user_id`: Filter by specific user
- `action`: Filter by action type
- `resource`: Filter by resource type
- `status`: Filter by status (success/failure/error)
- `start_date`: Start date filter (YYYY-MM-DD)
- `end_date`: End date filter (YYYY-MM-DD)
- `limit`: Number of results (default: 50)
- `offset`: Pagination offset (default: 0)

### Get User Activity
```
GET /api/v1/audit/users/{user_id}/activity
```
Returns recent activity for a specific user.

### Get Garden Activity
```
GET /api/v1/audit/gardens/{garden_id}/activity
```
Returns recent activity for a specific garden.

### Get Audit Statistics
```
GET /api/v1/audit/stats?days=30
```
Returns system-wide audit statistics for the specified number of days.

## Tracked Actions

### Authentication Actions
- `user_register`: New user registration
- `user_login`: User login attempts
- `user_logout`: User logout
- `token_refresh`: JWT token refresh
- `profile_update`: User profile updates

### Garden Actions
- `garden_create`: Creating new gardens
- `garden_update`: Updating garden information
- `garden_delete`: Deleting gardens
- `garden_view`: Viewing garden details

### Plant Actions
- `plant_seed`: Planting seeds
- `plant_harvest`: Harvesting plants
- `plant_remove`: Removing plants
- `plant_water`: Watering plants (future)
- `plant_fertilize`: Fertilizing plants (future)

### Garden Sharing Actions
- `garden_share`: Sharing gardens with users
- `garden_join`: Joining shared gardens
- `share_permission_update`: Updating share permissions
- `share_remove`: Removing garden shares
- `access_link_create`: Creating access links
- `access_link_deactivate`: Deactivating access links

### Store Actions
- `seed_purchase`: Purchasing seeds
- `inventory_view`: Viewing store inventory

### System Actions
- `error`: System errors
- `security`: Security-related events

## Frontend Integration

The audit system includes a React component (`AuditLogs.tsx`) that provides:

- **Real-time Log Viewing**: Display audit logs with filtering and pagination
- **Statistics Dashboard**: View system-wide audit statistics
- **User Activity Tracking**: Monitor specific user actions
- **Garden Activity Monitoring**: Track garden-specific events

## Security Features

### Data Protection
- Sensitive information is not logged (passwords, tokens)
- IP addresses are captured for security analysis
- User agent strings help identify suspicious activity

### Access Control
- Audit endpoints require authentication
- Only authorized users can view audit logs
- Audit data is protected by the same security measures as other API endpoints

## Monitoring and Alerts

### Automated Monitoring
The audit system can be extended to include:

- **Anomaly Detection**: Flag unusual patterns of activity
- **Rate Limiting**: Monitor for excessive requests
- **Security Alerts**: Notify administrators of suspicious activity
- **Performance Monitoring**: Track system performance metrics

### Integration Points
- **Log Aggregation**: Export logs to external monitoring systems
- **Alert Systems**: Integrate with notification services
- **Analytics**: Connect to business intelligence tools

## Usage Examples

### Viewing Recent Activity
```javascript
// Get recent audit logs
const logs = await apiService.getAuditLogs({
  limit: 20,
  status: 'success'
});

// Get user activity
const userActivity = await apiService.getUserActivity(userId, 50);

// Get garden activity
const gardenActivity = await apiService.getGardenActivity(gardenId, 30);
```

### Filtering Audit Data
```javascript
// Filter by date range
const logs = await apiService.getAuditLogs({
  start_date: '2024-01-01',
  end_date: '2024-01-31',
  action: 'garden_create'
});

// Filter by user and resource
const logs = await apiService.getAuditLogs({
  user_id: userId,
  resource: 'plant',
  status: 'success'
});
```

## Best Practices

### Log Management
- Regularly review audit logs for security issues
- Implement log rotation to manage storage
- Archive old logs for compliance requirements
- Monitor log volume and performance impact

### Security Considerations
- Ensure audit logs cannot be tampered with
- Implement proper access controls for audit data
- Regular security reviews of audit data
- Compliance with data protection regulations

### Performance Optimization
- Index audit log tables for efficient querying
- Implement pagination for large result sets
- Use appropriate filters to reduce data volume
- Consider data retention policies

## Future Enhancements

### Planned Features
- **Real-time Notifications**: WebSocket-based real-time audit updates
- **Advanced Analytics**: Machine learning-based anomaly detection
- **Compliance Reporting**: Automated compliance report generation
- **Integration APIs**: Third-party system integration capabilities

### Scalability Improvements
- **Distributed Logging**: Support for multiple database instances
- **Stream Processing**: Real-time log processing capabilities
- **Data Archiving**: Automated archival of old audit data
- **Performance Optimization**: Query optimization and caching 