gen:
	operator-sdk generate k8s
	operator-sdk generate crds

apply:
	kubectl apply -f deploy/crds/*_crd.yaml

run:
	export OPERATOR_NAME=application-operator
	operator-sdk run --local --namespace=default