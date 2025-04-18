# Assement

This is a simple assessment to test your knowledge of infrastructure programming. This is 15 days mini project to evaluate your kubernetes, coding and git flow skills. You can free to ask questions at original repositories issues.

# Tasks

There is no strich order of tasks.

### Task 0:

Analyze this repository and understand how it works. The repository contains a simple web application with a Go backend and a React frontend. The backend is using its own framework, and the frontend is built using React and Material-UI (MUI). Frontend is SPA. Hence frontend is served by backend by embedding it in the backend binary. The backend is a simple REST API that serves data to the frontend. There is hidden SSO authentication inside frontend. Understand how to generate SPA with SSO.

### Task 1:

Install minikube on your local machine and create a Kubernetes cluster with at least 2 nodes. Or you can create normal kubernetes cluster. Ensure you have a CSI storage plugin.

For SSO you can deploy Keycloak or any other SSO provider into this cluster. Also you can deploy Minio for object storage. Also you can deploy Nexus for docker images.

Please provide each deployment yamls into deploy/k8s directory. Task 2 also needs two deployments one for webserver and one for watching configmaps. Ypu should also provide them.

### Task 2:

There are some templates under the `templates` directory. You will use them create several Kubernetes resources. Yamls have easy to configured by string replace. At backend you should create resources with string replacement and apply with go kubernes client. Understand variables. Your backend application should listen configmaps with annotation:

```
example.org/postgres-cluster
```

Configmap should have all required variables. You can extend or define default variables. Then configmap created on cluster, your backend should create all resources. You can use go templates. However you should check for injections. When configmap is deleted, your backend should delete all resources. Don't forget to scale replicas to 0 before deleting. Also Satefulset creates pvc's. Don't forget to delete them.

There is no update.

There is server command for webserver. You should implement similar command for watching configmaps.

### Task 3:

Frontend should able to monitor configmaps. List, create and delete them. Don't forget frontend is SPA. Don't use more pages. There is /api endpoint at backend which supports GET/POST, also checks authentication. Hence you should generally write logical operations.

### Task 4:

Write unittests. Add them to the build scripts. Also write github workflow to check uinttests. Use ginkgo and gomeka.

### Task 5:

There may be intentional logical bugs at provided code. Find them and fix them. 

### Task 6:

Fork this repository. Work on your fork. Create branches. Create Merge Requests. I will review and may merge them. Hence you should push simple changes. Follow git flow rules. All commits should be signed, and with signed-off header. There are pre-commit and pre-push commits. Ensure they are working. Also there is github workflow which performs checks. If MR fails, you should fix it. Don't use force push.

# Evaluation Criteria

- Git flow: commits, branches, merge requests etc. There are some rules for commit titles. Respect them. like add, update, fix, remove etc labels.
- Aprroved merge requests. Worflows should be passed.
- Respection of application format. SPA and restful API.
- Code quality. There are some rules for golang and react. Respect them. Use golangci-lint and eslint. They are already configured. You can add more rules, but don't remove them. You can add more linters, but don't remove them. More strict rules are better.
- Unittests are extra points. However using ginkgo and gomeka is required.
- Frontend and Configmap watching at backend are seperate tasks. Firstly you should implement backend. It has more points. After that, frontend is only required api calls.
- There are some intentional bugs. Find them and fix them. They are not required, but they are extra points.
