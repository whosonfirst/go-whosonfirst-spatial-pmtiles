# ECS

```
$> aws ecs register-task-definition \
	--family whosonfirst-spatial-pmtiles \
	--cli-input-json file://whosonfirst-spatial-pmtiles.json \
	--ephemeral-storage="sizeInGiB=200" \
	--region {AWS_REGION}
```

If you are processing a large number of (large) Who's On First data repositories it is possible that you will run out of disk space. Assigning an instance with ephemeral storage greater than the default (between 11GB - 200GB) will help with that. It is not possible to assign ephermal storage from the AWS web console so you'll need to register your task programatically, or from the command line using the `aws` tool, using a JSON definition.