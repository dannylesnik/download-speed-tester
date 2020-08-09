# DOWNLOAD-SPEED-TESTER
Download-Speed-Tester is small utility that measures Download speed of any valid web url. 

## How to set it up: 
All you need to do is to configure one mandatory environment variable and two optional environment variables. 

`HTTP_URL` - Mandatory paramter.  Valid HTTP/HTTPS url that you would like to measure. This url passing validation and the program will be terminated if url is not valid. 

`TOTAL_REQ` - total number of requests. Default is 25

`MAX_PARAL_REQ` - maximum allowed parallel requests. Default is 5


## Running it in Docker

```
docker run -d -it --name speed-test --env-file ./env.list dannylesnik/download-speed-tester:latest | more env.list
HTTP_URL=[MY Download URL]
TOTAL_REQ=30
MAX_PARAL_REQ=4
```

## Running it on Kubernetes.

```
apiVersion: v1
kind: Pod
metadata:
  name: download-speed-tester
spec:
  containers:
  - name: download-speed-tester
    image: 113379206287.dkr.ecr.us-east-1.amazonaws.com/development/tmstech/download-speed-tester:latest
    resources:
      limits:
        memory: "512Mi"
      requests:
        memory: "512Mi"
    env:
      - name: HTTP_URL
        value: "[SOME VALID URL]"   
      - name: TOTAL_REQ
        value: "30"  
      - name: MAX_PARAL_REQ
        value: "4"  
```
