build:
	docker build -t antelope .
run:
	-docker stop antelope
	-docker rm antelope
	docker run -d -p 9081:9081
	-e mongo_host = "10.248.9.173"
	-e mongo_auth = "true"
    -e mongo_db = "antelope"
	-e mongo_user = "antelope"
	-e mongo_pass = "deltapsi@#22237"
    -e mongo_aws_template_collection = "aws_template"
    -e mongo_aws_cluster_collection = "aws_cluster"
	-e mongo_azure_template_collection = "azure_template"
    -e mongo_azure_cluster_collection = "azure_cluster"
	-e redis_url = "10.248.9.173"
	-e logger_url = "10.248.9.173"
    -e network_url = "10.248.9.173"
     --name antelope antelope
