load('ext://restart_process', 'docker_build_with_restart')

docker_build_with_restart('sthanguy/fc-gateway',
							context='./services/gateway',
							entrypoint='go run main.go',
							dockerfile='./services/gateway/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/gateway', '/usr/gateway'),
							]
)


docker_build_with_restart('sthanguy/fc-auth',
							context='./services/auth',
							entrypoint='go run main.go',
							dockerfile='./services/auth/Dockerfile',
							extra_tag='latest',
							live_update=[
								sync('./services/auth', '/usr/auth'),
							]
)

# gateway
k8s_yaml(['./services/gateway/service.yaml', './services/gateway/ingress.yaml'])
k8s_yaml(kustomize('./services/gateway'))

# auth
k8s_yaml(['./services/auth/service.yaml'])
k8s_yaml(kustomize('./services/auth'))

# session-cache
k8s_yaml(['./services/session-cache/service.yaml', './services/session-cache/configmap.yaml', './services/session-cache/deployment.yaml'])

# postgres
k8s_yaml(['./services/postgres/configmap.yaml', './services/postgres/pvc.yaml', './services/postgres/deployment.yaml', './services/postgres/service.yaml'])
