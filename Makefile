build:
	docker build -t fable --target production . 
push:
	docker push docker.io/shiva0612/fable:latest 
up:
	docker-compose up -d
down: 
	docker-compose down