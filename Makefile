build:
	docker build -t antelope .
run:
	-docker stop antelope
	-docker rm antelope
	docker run -d -p 9081:9081 --name antelope antelope