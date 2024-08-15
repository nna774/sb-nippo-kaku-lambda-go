all: app

app:
	GOARCH=amd64 GOOS=linux go build

REGION := ap-northeast-1

deploy: app
	sam package --template-file template.yml --region $(REGION) --s3-bucket nana-lambda --output-template-file packaged-template.yml
	sam deploy --template-file packaged-template.yml --region $(REGION) --capabilities CAPABILITY_IAM --stack-name sb-nippo-kaku-lambda-go
