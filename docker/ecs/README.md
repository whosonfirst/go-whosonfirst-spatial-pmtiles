# ECS

```
$> aws ecs register-task-definition \
	--family whosonfirst-spatial-pmtiles \
	--cli-input-json file://whosonfirst-spatial-pmtiles.json \
	--ephemeral-storage="sizeInGiB=200" \
	--region {AWS_REGION}
```
