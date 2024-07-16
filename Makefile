build:
	docker build -t metrics-provider . 
	
tag: build
	docker tag metrics-provider:latest homelab:5000/metrics-provider:latest

push: tag
	docker push homelab:5000/metrics-provider:latest
