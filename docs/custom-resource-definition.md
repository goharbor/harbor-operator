# Custom Resource Harbor

## Conditional Statuses

`Phase` field is deprecated in favor of `Conditions` list: <https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties>

## Default value

Default value is setted thanks to `Default()`. It must be auto-applied thanks to the conversion webhook.
_This does not work at the moment_
